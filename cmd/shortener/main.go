package main

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/devkekops/go-url-shortener/internal/app/server"
)

func main() {
	cfg := server.Config{
		"localhost:8080",
		"http://localhost:8080",
		"",
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Serve(&cfg))
}
