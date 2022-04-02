package server

import (
	"net/http"

	"github.com/devkekops/go-url-shortener/internal/app/handlers"
	"github.com/devkekops/go-url-shortener/internal/app/storage"
)

func Serve(serverAddress string, baseURL string) error {
	db := make(map[int64]string)
	linkRepo := storage.NewLinkRepo(db)

	baseHandler := handlers.NewBaseHandler(linkRepo, baseURL)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: baseHandler,
	}

	return server.ListenAndServe()
}
