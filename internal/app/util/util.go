package util

import (
	"math/big"
	"net/url"
	"unicode"
)

func IsLetterOrNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func IsValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

func Base10ToBase62(id int64) string {
	str := big.NewInt(id).Text(62)
	return str
}

func Base62ToBase10(str string) int64 {
	bigID := new(big.Int)
	bigID.SetString(str, 62)
	id := bigID.Int64()
	return id
}
