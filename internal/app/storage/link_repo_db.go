package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/devkekops/go-url-shortener/internal/app/myerrors"
)

type LinkRepoDB struct {
	dbpool     *pgxpool.Pool //concurrency safe (see https://github.com/jackc/pgx/wiki/Getting-started-with-pgx#using-a-connection-pool)
	toDeleteCh chan int64
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
		url 		TEXT NOT NULL UNIQUE,
		deleted		BOOLEAN NOT NULL DEFAULT FALSE
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

	r := &LinkRepoDB{
		dbpool:     dbpool,
		toDeleteCh: make(chan int64),
	}

	go r.clearBuffer()

	return r, nil
}

func (r *LinkRepoDB) GetLongByShortLink(shortURL string) (string, error) {
	linkID := base62ToBase10(shortURL)
	var (
		url     string
		deleted bool
	)
	queryGetLink := `SELECT url, deleted FROM urls WHERE id = $1`

	err := r.dbpool.QueryRow(context.Background(), queryGetLink, linkID).Scan(&url, &deleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", myerrors.NewNotFoundURLError(shortURL)
		} else {
			return "", err
		}
	}

	if deleted {
		return "", myerrors.NewDeletedURLError(url)
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
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				return nil, err
			}
			if !pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return nil, err
			}
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
		return nil, myerrors.NewUserHasNoURLsError(userID)
	}

	return userLinks, nil
}

func (r *LinkRepoDB) deleteLinks(ids []int64) error {
	queryDeleteLinks := `UPDATE urls SET deleted = TRUE WHERE id = ANY ($1)`
	_, err := r.dbpool.Exec(context.Background(), queryDeleteLinks, ids)
	return err
}

func (r *LinkRepoDB) clearBuffer() {
	ticker := time.NewTicker(10 * time.Second)
	buffer := make([]int64, 0, 10)
	for {
		select {
		case id := <-r.toDeleteCh:
			buffer = append(buffer, id)
			if len(buffer) == cap(buffer) {
				//fmt.Printf("clear buffer %v because it's full", buffer)
				err := r.deleteLinks(buffer)
				if err != nil {
					log.Println(err)
				}
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				//fmt.Printf("clear buffer %v because time is end", buffer)
				err := r.deleteLinks(buffer)
				if err != nil {
					log.Println(err)
				}
				buffer = buffer[:0]
			}
		}
	}
}

func (r *LinkRepoDB) DeleteUserLinks(userID string, shortURLs []string) {
	queryGetExistedUserLinks := `SELECT id FROM urls WHERE deleted = false AND id IN (SELECT url_id FROM user_urls WHERE id = ($1) AND url_id = ANY ($2))`

	go func() {
		ids := make([]int64, 0, len(shortURLs))
		for _, shortURL := range shortURLs {
			ids = append(ids, base62ToBase10(shortURL))
		}

		rows, err := r.dbpool.Query(context.Background(), queryGetExistedUserLinks, userID, ids)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		ids = ids[:0]
		for rows.Next() {
			var id int64
			err := rows.Scan(&id)
			if err != nil {
				log.Println(err)
			}
			ids = append(ids, id)
		}
		//fmt.Println(ids)

		for _, id := range ids {
			r.toDeleteCh <- id
		}
	}()
}

func (r *LinkRepoDB) Close() error {
	r.dbpool.Close()
	close(r.toDeleteCh)
	return nil
}

func (r *LinkRepoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return r.dbpool.Ping(ctx)
}
