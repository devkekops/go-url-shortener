package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/devkekops/go-url-shortener/internal/app/server"
)

func main() {
	cfg := server.Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		SecretKey:       "asdhkhk1375jwh132",
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.Parse()

	log.Fatal(server.Serve(&cfg))
}
