package storage

import (
	"fmt"
	"sync"
)

type LinkRepoMemory struct {
	mutex       sync.RWMutex
	idToLinkMap map[int64]string
}

func NewLinkRepoMemory() *LinkRepoMemory {
	return &LinkRepoMemory{
		mutex:       sync.RWMutex{},
		idToLinkMap: make(map[int64]string),
	}
}

func (r *LinkRepoMemory) FindByID(id int64) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	url, exist := r.idToLinkMap[id]
	if !exist {
		return "", fmt.Errorf("not found row %d", id)
	}
	return url, nil

}

func (r *LinkRepoMemory) Save(link string) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	index := len(r.idToLinkMap) + 1
	r.idToLinkMap[int64(index)] = link

	return int64(index), nil
}
