package util

import (
	"math/big"
)

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
