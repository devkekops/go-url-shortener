package server

import (
	"log"
	"net/http"

	"github.com/devkekops/go-url-shortener/internal/app/handlers"
	"github.com/devkekops/go-url-shortener/internal/app/storage"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func Serve(cfg *Config) error {

	var baseHandler *handlers.BaseHandler

	if cfg.FileStoragePath != "" {
		linkRepo, err := storage.NewLinkRepoFile(cfg.FileStoragePath)
		if err != nil {
			log.Println(err)
		}
		defer linkRepo.Close()
		baseHandler = handlers.NewBaseHandler(linkRepo, cfg.BaseURL)

	} else {
		linkRepo := storage.NewLinkRepoMemory()
		baseHandler = handlers.NewBaseHandler(linkRepo, cfg.BaseURL)
	}

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: baseHandler,
	}

	return server.ListenAndServe()
}
