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

type key string

var (
	cookieName     = "token"
	cookiePath     = "/"
	userIDKey  key = "userID"
)

type Token struct {
	userID      string
	cookieValue string
}

func createToken(secretKey []byte) Token {
	id := uuid.New()
	key := sha256.Sum256(secretKey)

	h := hmac.New(sha256.New, key[:])
	h.Write(id[:])
	dst := h.Sum(nil)

	cookieValueBytes := append(id[:], dst[:]...)
	cookieValue := hex.EncodeToString(cookieValueBytes)

	return Token{id.String(), cookieValue}
}

func checkSignature(cookieValue string, secretKey []byte) (string, error) {
	token, err := hex.DecodeString(cookieValue)
	if err != nil {
		return "", err
	}

	id, err := uuid.FromBytes(token[:16])
	if err != nil {
		return "", err
	}
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

func authHandle(secretKey string) (ah func(http.Handler) http.Handler) {
	ah = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token Token
			tokenCookie, err := r.Cookie(cookieName)
			secretKeyByte := []byte(secretKey)

			if err != nil {
				if err == http.ErrNoCookie {
					token = createToken(secretKeyByte)

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
				id, err := checkSignature(cookieValue, secretKeyByte)
				if err != nil {
					token = createToken(secretKeyByte)

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
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	return
}
