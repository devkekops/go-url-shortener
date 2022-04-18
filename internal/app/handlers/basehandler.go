package handlers

import (
	"github.com/devkekops/go-url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BaseHandler struct {
	*chi.Mux
	linkRepo  storage.LinkRepository
	baseURL   string
	secretKey string
}

func NewBaseHandler(linkRepo storage.LinkRepository, baseURL string, secretKey string) *BaseHandler {
	bh := &BaseHandler{
		Mux:       chi.NewMux(),
		linkRepo:  linkRepo,
		baseURL:   baseURL,
		secretKey: secretKey,
	}

	bh.Use(middleware.RequestID)
	bh.Use(middleware.RealIP)
	bh.Use(middleware.Logger)
	bh.Use(middleware.Recoverer)

	bh.Use(middleware.Compress(5))
	bh.Use(gzipHandle)
	bh.Use(authHandle(bh.secretKey))

	bh.Post("/", bh.shortenLink())
	bh.Get("/{id}", bh.expandLink())
	bh.Post("/api/shorten", bh.apiShorten())
	bh.Get("/api/user/urls", bh.apiUserURLs())

	return bh
}
