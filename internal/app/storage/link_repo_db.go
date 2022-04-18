package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
)

type LinkRepoDB struct {
	db             *pgx.Conn
	linkRepoMemory *LinkRepoMemory
}

func NewLinkRepoDB(dsn string) (*LinkRepoDB, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &LinkRepoDB{
		db:             conn,
		linkRepoMemory: NewLinkRepoMemory(),
	}, nil
}

func (r *LinkRepoDB) GetLongByShortLink(shortURL string) (string, error) {
	return r.linkRepoMemory.GetLongByShortLink(shortURL)
}

func (r *LinkRepoDB) SaveLongLink(link string, userID string) (string, error) {
	return r.linkRepoMemory.SaveLongLink(link, userID)
}

func (r *LinkRepoDB) GetUserLinks(userID string) ([]URLPair, error) {
	return r.linkRepoMemory.GetUserLinks(userID)
}

func (r *LinkRepoDB) Close() error {
	return r.db.Close(context.Background())
}

func (r *LinkRepoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := r.db.Ping(ctx); err != nil {
		return err
	}

	return nil
}
