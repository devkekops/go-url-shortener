package storage

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

type LinkRepoFile struct {
	mutex          sync.RWMutex
	file           *os.File
	encoder        *json.Encoder
	linkRepoMemory *LinkRepoMemory
}

func NewLinkRepoFile(filename string) (*LinkRepoFile, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	linkRepoMemory := NewLinkRepoMemory()

	decoder := json.NewDecoder(file)
	for {
		URLEntry := URLEntry{}
		err = decoder.Decode(&URLEntry)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		linkRepoMemory.idToLinkMap[URLEntry.ID] = URLEntry.URL
		linkRepoMemory.userIDToLinksIDMap[URLEntry.UserID] = append(linkRepoMemory.userIDToLinksIDMap[URLEntry.UserID], URLEntry.ID)
	}

	return &LinkRepoFile{
		mutex:          sync.RWMutex{},
		file:           file,
		encoder:        json.NewEncoder(file),
		linkRepoMemory: linkRepoMemory,
	}, nil
}

func (r *LinkRepoFile) GetLongByShortLink(shortURL string) (string, error) {
	return r.linkRepoMemory.GetLongByShortLink(shortURL)
}

func (r *LinkRepoFile) SaveLongLink(link string, userID string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	shortURL, _ := r.linkRepoMemory.SaveLongLink(link, userID)
	linkID := base62ToBase10(shortURL)

	URLEntry := URLEntry{linkID, userID, link}
	err := r.encoder.Encode(&URLEntry)
	if err != nil {
		return "", err
	}

	return shortURL, nil
}

func (r *LinkRepoFile) GetUserLinks(userID string) ([]URLPair, error) {
	return r.linkRepoMemory.GetUserLinks(userID)
}

func (r *LinkRepoFile) Close() error {
	return r.file.Close()
}

func (r *LinkRepoFile) Ping() error {
	return nil
}
