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

type BaseHandler struct {
	linkRepo storage.LinkRepository
}

// NewBaseHandler returns a new BaseHandler
func NewBaseHandler(linkRepo storage.LinkRepository) *BaseHandler {
	return &BaseHandler{
		linkRepo: linkRepo,
	}
}

func (h *BaseHandler) RootHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		shortURL := strings.TrimPrefix(req.URL.Path, "/")
		id := util.Base62ToBase10(shortURL)
		url, err := h.linkRepo.FindByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "not found", http.StatusNotFound)
				fmt.Println(err)
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
		id, err := h.linkRepo.Save(originalURL)
		if err != nil {
			fmt.Println(err)
		}
		shortURL := util.Base10ToBase62(id)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + shortURL))

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
