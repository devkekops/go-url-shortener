package storage

import (
	"fmt"
)

type LinkRepoMemory struct {
	idToLinkMap map[int64]string
}

func NewLinkRepoMemory() *LinkRepoMemory {
	return &LinkRepoMemory{
		idToLinkMap: make(map[int64]string),
	}
}

func (r *LinkRepoMemory) FindByID(id int64) (string, error) {
	url, exist := r.idToLinkMap[id]
	if !exist {
		return "", fmt.Errorf("not found row %d", id)
	}
	return url, nil

}

func (r *LinkRepoMemory) Save(link string) (int64, error) {
	index := len(r.idToLinkMap) + 1
	r.idToLinkMap[int64(index)] = link

	return int64(index), nil
}
