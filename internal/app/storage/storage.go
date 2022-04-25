package storage

import (
	"math/big"
)

type URLEntry struct {
	ID     int64  `json:"id"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type URLPair struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type LongURLUnit struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortURLUnit struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type LinkRepository interface {
	GetLongByShortLink(shortURL string) (string, error)
	SaveLongLink(link string, userID string) (string, error)
	SaveLongLinks(longURLUnits []LongURLUnit, userID string) ([]ShortURLUnit, error)
	GetUserLinks(userID string) ([]URLPair, error)
	Close() error
	Ping() error
}

func base10ToBase62(id int64) string {
	str := big.NewInt(id).Text(62)
	return str
}

func base62ToBase10(str string) int64 {
	bigID := new(big.Int)
	bigID.SetString(str, 62)
	id := bigID.Int64()
	return id
}
