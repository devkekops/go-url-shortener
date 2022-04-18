package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"unicode"

	"github.com/go-chi/chi/v5"
)

func isLetterOrNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func isValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

func (bh *BaseHandler) shortenLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)
		fmt.Printf("from shortenLink %s\n", userID)

		b, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Println(err)
			return
		}
		originalURL := string(b)
		if !isValidURL(originalURL) {
			http.Error(w, "URL is incorrect", http.StatusBadRequest)
			log.Printf("Incorrect URL %s\n", originalURL)
			return
		}

		shortURL, err := bh.linkRepo.SaveLongLink(originalURL, userID)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bh.baseURL + "/" + shortURL))
	}
}

func (bh *BaseHandler) expandLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		shortURL := chi.URLParam(req, "id")
		if !isLetterOrNumber(shortURL) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Printf("Incorrect URL %s\n", shortURL)
			return
		}

		url, err := bh.linkRepo.GetLongByShortLink(shortURL)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Println(err)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
