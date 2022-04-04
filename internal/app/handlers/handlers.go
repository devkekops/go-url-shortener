package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"unicode"

	"github.com/devkekops/go-url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type URL struct {
	URL string `json:"url"`
}

type Result struct {
	Result string `json:"result"`
}

type BaseHandler struct {
	*chi.Mux
	linkRepo storage.LinkRepository
	baseURL  string
}

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

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}
		next.ServeHTTP(w, r)
	})
}

func NewBaseHandler(linkRepo storage.LinkRepository, baseURL string) *BaseHandler {
	bh := &BaseHandler{
		Mux:      chi.NewMux(),
		linkRepo: linkRepo,
		baseURL:  baseURL,
	}

	bh.Use(middleware.RequestID)
	bh.Use(middleware.RealIP)
	bh.Use(middleware.Logger)
	bh.Use(middleware.Recoverer)

	bh.Use(middleware.Compress(5))
	bh.Use(gzipHandle)

	bh.Post("/", bh.shortenLink())
	bh.Get("/{id}", bh.expandLink())
	bh.Post("/api/shorten", bh.apiShorten())

	return bh
}

func (bh *BaseHandler) shortenLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		id, err := bh.linkRepo.Save(originalURL)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Println(err)
			return
		}
		shortURL := base10ToBase62(id)

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

		id := base62ToBase10(shortURL)
		url, err := bh.linkRepo.FindByID(id)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			log.Printf("Not found row number %d\n", id)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (bh *BaseHandler) apiShorten() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var link URL
		if err := json.NewDecoder(req.Body).Decode(&link); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Printf("Incorrect JSON\n")
			return
		}
		originalURL := link.URL
		if !isValidURL(originalURL) {
			http.Error(w, "URL is incorrect", http.StatusBadRequest)
			log.Printf("Incorrect URL %s\n", originalURL)
			return
		}

		id, err := bh.linkRepo.Save(originalURL)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Println(err)
			return
		}
		shortURL := base10ToBase62(id)

		r := Result{bh.baseURL + "/" + shortURL}
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(r)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(buf.Bytes())
	}
}
