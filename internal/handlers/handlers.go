package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/devkekops/go-url-shortener/internal/storage"
	"github.com/devkekops/go-url-shortener/internal/util"
)

type Server struct {
	linkRepo storage.LinkRepository
}

func NewServer(linkRepo storage.LinkRepository) *Server {
	return &Server{
		linkRepo: linkRepo,
	}
}

func (s *Server) RootHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		shortURL := strings.TrimPrefix(req.URL.Path, "/")
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

	case "POST":
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

	default:
		//w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "Only GET and POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
}
