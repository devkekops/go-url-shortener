package storage

import (
	"fmt"
	"sync"
)

type LinkRepoMemory struct {
	mutex              sync.RWMutex
	idToLinkMap        map[int64]string
	userIDToLinksIDMap map[string][]int64
}

func NewLinkRepoMemory() *LinkRepoMemory {
	return &LinkRepoMemory{
		mutex:              sync.RWMutex{},
		idToLinkMap:        make(map[int64]string),
		userIDToLinksIDMap: make(map[string][]int64),
	}
}

func (r *LinkRepoMemory) GetLongByShortLink(shortURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	linkID := base62ToBase10(shortURL)

	url, exist := r.idToLinkMap[linkID]
	if !exist {
		return "", fmt.Errorf("not found shortURL %s (linkID %d)", shortURL, linkID)
	}
	return url, nil
}

func (r *LinkRepoMemory) SaveLongLink(link string, userID string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	linkID := int64(len(r.idToLinkMap) + 1)

	r.idToLinkMap[linkID] = link
	r.userIDToLinksIDMap[userID] = append(r.userIDToLinksIDMap[userID], linkID)

	shortURL := base10ToBase62(linkID)

	return shortURL, nil
}

func (r *LinkRepoMemory) GetUserLinks(userID string) ([]URLPair, error) {
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
