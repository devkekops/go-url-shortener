package main

import (
	"log"

	"github.com/devkekops/go-url-shortener/internal/app/server"
)

func main() {
	log.Fatal(server.Serve("localhost:8080"))
}
