package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

type LinkRepoDB struct {
	dbpool *pgxpool.Pool //concurrency safe (see https://github.com/jackc/pgx/wiki/Getting-started-with-pgx#using-a-connection-pool)
}

func NewLinkRepoDB(dsn string) (*LinkRepoDB, error) {
	ctx := context.Background()
	dbpool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	queryCreateTableURLs := `
	CREATE TABLE IF NOT EXISTS urls(
		id          SERIAL PRIMARY KEY,
		url 		TEXT NOT NULL UNIQUE
	);`

	queryCreateTableUserURLs := `
	CREATE TABLE IF NOT EXISTS user_urls(
		id UUID NOT NULL,
		url_id INTEGER NOT NULL REFERENCES urls(id),
		UNIQUE (id, url_id)
	);`

	tx, err := dbpool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, queryCreateTableURLs); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, queryCreateTableUserURLs); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &LinkRepoDB{
		dbpool: dbpool,
	}, nil
}

func (r *LinkRepoDB) GetLongByShortLink(shortURL string) (string, error) {
	linkID := base62ToBase10(shortURL)
	var url string
	queryGetLink := `SELECT url FROM urls WHERE id = $1`

	err := r.dbpool.QueryRow(context.Background(), queryGetLink, linkID).Scan(&url)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("not found shortURL %s (linkID %d)", shortURL, linkID)
		} else {
			return "", err
		}
	}

	return url, nil
}

func (r *LinkRepoDB) SaveLongLink(link string, userID string) (string, error) {
	ctx := context.Background()
	var linkID int64
	var err error
	querySaveURL := `INSERT INTO urls (url) VALUES ($1) RETURNING id`
	queryGetIDByURL := `SELECT id FROM urls WHERE url = $1`
	querySaveUserURL := `INSERT INTO user_urls (id, url_id) VALUES ($1, $2) ON CONFLICT (id, url_id) DO NOTHING`

	if err = r.dbpool.QueryRow(ctx, querySaveURL, link).Scan(&linkID); err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return "", err
		}
		if !pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return "", err
		}
		if err1 := r.dbpool.QueryRow(ctx, queryGetIDByURL, link).Scan(&linkID); err1 != nil {
			return "", err1
		}
	}

	if _, err2 := r.dbpool.Exec(ctx, querySaveUserURL, userID, linkID); err2 != nil {
		var pgErr *pgconn.PgError
		if errors.As(err2, &pgErr) {
			if !pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return "", err2
			}
		}
	}

	shortURL := base10ToBase62(linkID)

	return shortURL, err
}

func (r *LinkRepoDB) SaveLongLinks(longURLUnits []LongURLUnit, userID string) ([]ShortURLUnit, error) {
	var shortURLUnits []ShortURLUnit

	for _, longURLUnit := range longURLUnits {
		shortURL, err := r.SaveLongLink(longURLUnit.OriginalURL, userID)
		if err != nil {
			return nil, err
		}
		shortURLUnits = append(shortURLUnits, ShortURLUnit{longURLUnit.CorrelationID, shortURL})
	}

	return shortURLUnits, nil
}

func (r *LinkRepoDB) GetUserLinks(userID string) ([]URLPair, error) {
	var userLinks []URLPair
	queryGetUserLinks := `SELECT id, url FROM urls WHERE id IN (SELECT url_id FROM user_urls WHERE id = $1)`

	rows, err := r.dbpool.Query(context.Background(), queryGetUserLinks, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			urlPair URLPair
			linkID  int64
		)

		err = rows.Scan(&linkID, &urlPair.OriginalURL)
		if err != nil {
			return nil, err
		}
		urlPair.ShortURL = base10ToBase62(linkID)
		userLinks = append(userLinks, urlPair)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if userLinks == nil {
		return nil, fmt.Errorf("not found URLs for userID %s", userID)
	}

	return userLinks, nil
}

func (r *LinkRepoDB) Close() error {
	r.dbpool.Close()
	return nil
}

func (r *LinkRepoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return r.dbpool.Ping(ctx)
}
