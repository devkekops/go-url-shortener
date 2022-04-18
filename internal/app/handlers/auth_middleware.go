package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var (
	secretKey  = []byte("asdhkhk1375jwh132")
	cookieName = "token"
	cookiePath = "/"
	userIDKey  = "userID"
)

type Token struct {
	userID      string
	cookieValue string
}

func createToken() Token {
	id := uuid.New()
	key := sha256.Sum256(secretKey)
	//fmt.Println(hex.EncodeToString(id[:]))

	h := hmac.New(sha256.New, key[:])
	h.Write(id[:])
	dst := h.Sum(nil)
	//fmt.Println(hex.EncodeToString(dst))

	cookieValueBytes := append(id[:], dst[:]...)
	//fmt.Println(cookieValueBytes)

	cookieValue := hex.EncodeToString(cookieValueBytes)
	//fmt.Println(cookieValue)

	return Token{id.String(), cookieValue}
}

func checkSignature(cookieValue string) (string, error) {
	token, err := hex.DecodeString(cookieValue)
	if err != nil {
		return "", err
	}

	id, err := uuid.FromBytes(token[:16])
	//fmt.Println(id)

	key := sha256.Sum256(secretKey)

	h := hmac.New(sha256.New, key[:])
	h.Write(id[:])
	sign := h.Sum(nil)

	if hmac.Equal(sign, token[16:]) {
		return id.String(), nil
	} else {
		return "", fmt.Errorf("invalid signature")
	}
}

func authHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token Token
		tokenCookie, err := r.Cookie(cookieName)

		if err != nil {
			if err == http.ErrNoCookie {
				token = createToken()

				cookie := &http.Cookie{
					Name:  cookieName,
					Value: token.cookieValue,
					Path:  cookiePath,
				}
				http.SetCookie(w, cookie)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				log.Println(err)
				return
			}
		} else {
			cookieValue := tokenCookie.Value
			id, err := checkSignature(cookieValue)
			if err != nil {
				token = createToken()

				cookie := &http.Cookie{
					Name:  cookieName,
					Value: token.cookieValue,
					Path:  cookiePath,
				}
				http.SetCookie(w, cookie)
			} else {
				token = Token{id, cookieValue}
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, token.userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
