package storage

import (
	"database/sql"
)

func (r *LinkRepo) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS links(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT NOT NULL
    );
    `
	_, err := r.db.Exec(query)

	return err
}

//model
type LinkRepository interface {
	FindByID(id int64) (string, error)
	Save(link string) (int64, error)
}

//implementation
type LinkRepo struct {
	db *sql.DB
}

func NewLinkRepo(db *sql.DB) *LinkRepo {
	return &LinkRepo{
		db: db,
	}
}

func (r *LinkRepo) FindByID(id int64) (string, error) {
	var url string
	err := r.db.QueryRow("select url from links WHERE id = ?", id).Scan(&url)
	if err != nil {
		return "", err
	}

	return url, err
}

func (r *LinkRepo) Save(link string) (int64, error) {
	result, err := r.db.Exec("insert into links (url) values (?)", link)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, err
}
