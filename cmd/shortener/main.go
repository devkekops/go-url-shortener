package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/devkekops/go-url-shortener/internal/handlers"
	"github.com/devkekops/go-url-shortener/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

const dbFileName = "links.db"

func main() {
	os.Remove(dbFileName)

	var err error
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	linkRepo := storage.NewLinkRepo(db)
	if err := linkRepo.Migrate(); err != nil {
		log.Fatal(err)
	}

	h := handlers.NewBaseHandler(linkRepo)

	http.HandleFunc("/", h.RootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
