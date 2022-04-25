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
	DatabaseDSN     string `env:"DATABASE_DSN"`
	SecretKey       string
}

func Serve(cfg *Config) error {

	var baseHandler *handlers.BaseHandler
	var linkRepo storage.LinkRepository

	if cfg.DatabaseDSN != "" {
		var err error
		linkRepo, err = storage.NewLinkRepoDB(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfg.FileStoragePath != "" {
		var err error
		linkRepo, err = storage.NewLinkRepoFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		linkRepo = storage.NewLinkRepoMemory()
	}
	defer linkRepo.Close()

	baseHandler = handlers.NewBaseHandler(linkRepo, cfg.BaseURL, cfg.SecretKey)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: baseHandler,
	}

	return server.ListenAndServe()
}
