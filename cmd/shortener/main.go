package main

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/devkekops/go-url-shortener/internal/app/server"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func main() {
	cfg := Config{"localhost:8080", "http://localhost:8080"}

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Serve(cfg.ServerAddress, cfg.BaseURL))
}
