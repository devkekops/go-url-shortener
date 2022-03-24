package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/devkekops/go-url-shortener/internal/storage"
	"github.com/devkekops/go-url-shortener/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	*chi.Mux
	linkRepo storage.LinkRepository
}

func NewServer(linkRepo storage.LinkRepository) *Server {
	s := &Server{
		Mux:      chi.NewMux(),
		linkRepo: linkRepo,
	}

	s.Use(middleware.RequestID)
	s.Use(middleware.RealIP)
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)

	s.Post("/", s.shortenLink())
	s.Get("/{id}", s.expandLink())

	return s
}

func (s *Server) shortenLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println(err)
		}
		originalURL := string(b)

		if !util.IsValidURL(originalURL) {
			http.Error(w, "URL is incorrect", http.StatusBadRequest)
			fmt.Printf("Incorrect URL %s\n", originalURL)
			return
		}

		id, err := s.linkRepo.Save(originalURL)
		if err != nil {
			fmt.Println(err)
		}
		shortURL := util.Base10ToBase62(id)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + shortURL))
	}
}

func (s *Server) expandLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		shortURL := chi.URLParam(req, "id")
		if !util.IsLetterOrNumber(shortURL) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		id := util.Base62ToBase10(shortURL)
		url, err := s.linkRepo.FindByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Not found", http.StatusNotFound)
				fmt.Printf("Not found row number %d\n", id)
				return
			}
			fmt.Println(err)
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
