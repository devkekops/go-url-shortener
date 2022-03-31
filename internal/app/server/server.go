package server

import (
	"net/http"

	"github.com/devkekops/go-url-shortener/internal/app/handlers"
	"github.com/devkekops/go-url-shortener/internal/app/storage"
)

func Serve(addr string) error {
	db := make(map[int64]string)
	linkRepo := storage.NewLinkRepo(db)

	origin := "http://" + addr + "/"
	baseHandler := handlers.NewBaseHandler(linkRepo, origin)
	server := &http.Server{
		Addr:    addr,
		Handler: baseHandler,
	}

	return server.ListenAndServe()
}
