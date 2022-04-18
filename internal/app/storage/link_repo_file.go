package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type LinkRepoFile struct {
	mutex              sync.RWMutex
	file               *os.File
	encoder            *json.Encoder
	idToLinkMap        map[int64]string
	userIDToLinksIDMap map[string][]int64
}

func NewLinkRepoFile(filename string) (*LinkRepoFile, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}

	idToLinkMap := make(map[int64]string)
	userIDToLinksIDMap := make(map[string][]int64)

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
		userIDToLinksIDMap[URLEntry.UserID] = append(userIDToLinksIDMap[URLEntry.UserID], URLEntry.ID)
	}

	return &LinkRepoFile{
		mutex:              sync.RWMutex{},
		file:               file,
		encoder:            json.NewEncoder(file),
		idToLinkMap:        idToLinkMap,
		userIDToLinksIDMap: userIDToLinksIDMap,
	}, nil
}

func (r *LinkRepoFile) Close() error {
	return r.file.Close()
}

func (r *LinkRepoFile) GetLongByShortLink(shortURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	linkID := base62ToBase10(shortURL)

	url, exist := r.idToLinkMap[linkID]
	if !exist {
		return "", fmt.Errorf("not found shortURL %s (linkID %d)", shortURL, linkID)
	}
	return url, nil
}

func (r *LinkRepoFile) SaveLongLink(link string, userID string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	linkID := int64(len(r.idToLinkMap) + 1)

	r.idToLinkMap[linkID] = link
	r.userIDToLinksIDMap[userID] = append(r.userIDToLinksIDMap[userID], linkID)

	URLEntry := URLEntry{linkID, userID, link}
	err := r.encoder.Encode(&URLEntry)
	if err != nil {
		return "", err
	}

	shortURL := base10ToBase62(linkID)

	return shortURL, nil
}

func (r *LinkRepoFile) GetUserLinks(userID string) ([]URLPair, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userLinkIDs, exist := r.userIDToLinksIDMap[userID]
	if !exist {
		return nil, fmt.Errorf("not found URLs for userID %s", userID)
	}

	userLinks := make([]URLPair, len(userLinkIDs))

	for i, linkID := range userLinkIDs {
		shortURL := base10ToBase62(linkID)
		longURL := r.idToLinkMap[linkID]

		userLinks[i] = URLPair{shortURL, longURL}
	}

	return userLinks, nil
}
