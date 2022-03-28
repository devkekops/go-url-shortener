package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/devkekops/go-url-shortener/internal/app/storage"
	"github.com/devkekops/go-url-shortener/internal/app/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BaseHandler struct {
	*chi.Mux
	linkRepo storage.LinkRepository
	origin   string
}

func NewBaseHandler(linkRepo storage.LinkRepository, origin string) *BaseHandler {
	bh := &BaseHandler{
		Mux:      chi.NewMux(),
		linkRepo: linkRepo,
		origin:   origin,
	}

	bh.Use(middleware.RequestID)
	bh.Use(middleware.RealIP)
	bh.Use(middleware.Logger)
	bh.Use(middleware.Recoverer)

	bh.Post("/", bh.shortenLink())
	bh.Get("/{id}", bh.expandLink())

	return bh
}

func (bh *BaseHandler) shortenLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		originalURL := string(b)

		if !util.IsValidURL(originalURL) {
			http.Error(w, "URL is incorrect", http.StatusBadRequest)
			fmt.Printf("Incorrect URL %s\n", originalURL)
			return
		}

		id := bh.linkRepo.Save(originalURL)
		shortURL := util.Base10ToBase62(id)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(bh.origin + shortURL))
	}
}

func (bh *BaseHandler) expandLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		shortURL := chi.URLParam(req, "id")
		if !util.IsLetterOrNumber(shortURL) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			fmt.Printf("Incorrect URL %s\n", shortURL)
			return
		}

		id := util.Base62ToBase10(shortURL)
		url, err := bh.linkRepo.FindByID(id)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			fmt.Printf("Not found row number %d\n", id)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
