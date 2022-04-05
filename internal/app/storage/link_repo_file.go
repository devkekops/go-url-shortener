package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type URLEntry struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type LinkRepoFile struct {
	mutex       sync.RWMutex
	file        *os.File
	encoder     *json.Encoder
	idToLinkMap map[int64]string
}

func NewLinkRepoFile(filename string) (*LinkRepoFile, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	idToLinkMap := make(map[int64]string)

	decoder := json.NewDecoder(file)
	for {
		URLEntry := URLEntry{}
		err := decoder.Decode(&URLEntry)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		idToLinkMap[URLEntry.ID] = URLEntry.URL
	}

	return &LinkRepoFile{
		mutex:       sync.RWMutex{},
		file:        file,
		encoder:     json.NewEncoder(file),
		idToLinkMap: idToLinkMap,
	}, nil
}

func (r *LinkRepoFile) Close() error {
	return r.file.Close()
}

func (r *LinkRepoFile) FindByID(id int64) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	url, exist := r.idToLinkMap[id]
	if !exist {
		return "", fmt.Errorf("not found row %d", id)
	}
	return url, nil
}

func (r *LinkRepoFile) Save(link string) (int64, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	index := len(r.idToLinkMap) + 1
	id := int64(index)
	r.idToLinkMap[id] = link

	URLEntry := URLEntry{id, link}
	err := r.encoder.Encode(&URLEntry)
	if err != nil {
		return 0, err
	}

	return id, nil
}
