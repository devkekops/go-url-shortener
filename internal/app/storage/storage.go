package storage

import (
	"fmt"
)

//model
type LinkRepository interface {
	FindByID(id int64) (string, error)
	Save(link string) int64
}

//implementation
type LinkRepo struct {
	db map[int64]string
}

func NewLinkRepo(db map[int64]string) *LinkRepo {
	return &LinkRepo{
		db: db,
	}
}

func (r *LinkRepo) FindByID(id int64) (string, error) {
	url, exist := r.db[id]
	if !exist {
		return "", fmt.Errorf("not found row %d", id)
	}
	return url, nil

}

func (r *LinkRepo) Save(link string) int64 {
	index := len(r.db) + 1
	r.db[int64(index)] = link

	return int64(index)
}
