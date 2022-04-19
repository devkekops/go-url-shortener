package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type LinkRepoDB struct {
	dbpool *pgxpool.Pool
}

func NewLinkRepoDB(dsn string) (*LinkRepoDB, error) {
	ctx := context.Background()
	dbpool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	queryCreateTable := `
	CREATE TABLE IF NOT EXISTS urls(
		id          serial PRIMARY KEY,
		user_id   	uuid NOT NULL,
		url 		text NOT NULL
	);`

	_, err = dbpool.Exec(ctx, queryCreateTable)
	if err != nil {
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
	var linkID int64
	querySaveLink := `INSERT INTO urls (user_id, url) VALUES ($1, $2) RETURNING id`
	err := r.dbpool.QueryRow(context.Background(), querySaveLink, userID, link).Scan(&linkID)
	if err != nil {
		return "", err
	}
	shortURL := base10ToBase62(linkID)

	return shortURL, nil
}

func (r *LinkRepoDB) GetUserLinks(userID string) ([]URLPair, error) {
	var userLinks []URLPair

	queryGetUserLinks := `SELECT id, url FROM urls WHERE user_id = $1`
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
