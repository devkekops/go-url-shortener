package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

/*type OriginalUrl struct {
	Url string
}

type ShortUrl struct {
	ShortUrl string
}*/

func base10ToBase62(id int64) string {
	str := big.NewInt(id).Text(62)
	//fmt.Println(str)
	return str
}

func base62ToBase10(str string) int64 {
	bigID := new(big.Int)
	bigID.SetString(str, 62)
	id := bigID.Int64()
	//fmt.Println(id)
	return id
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		shortURL := strings.TrimPrefix(req.URL.Path, "/")
		id := base62ToBase10(shortURL)
		var url string
		err := db.QueryRow("select url from links WHERE id = ?", id).Scan(&url)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(url)

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case "POST":
		b, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println(err)
		}

		/*var originalUrl OriginalUrl
		err = json.Unmarshal(b, &originalUrl)
		if err != nil {
			fmt.Println(err)
		}
		url := originalUrl.Url
		fmt.Println(url)*/

		originalURL := string(b)

		result, err := db.Exec("insert into links (url) values (?)", originalURL)
		if err != nil {
			fmt.Println(err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(id)

		shortURL := base10ToBase62(id)
		//w.Header().Set("content-type", "application/json")
		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)

		/*shortUrl := ShortUrl{shortLink}
		resp, err := json.Marshal(shortUrl)
		if err != nil {
			fmt.Println(err)
		}*/

		w.Write([]byte(shortURL))

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	os.Remove("./links.db")

	var err error
	db, err = sql.Open("sqlite3", "./links.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	sql := "create table links (id integer PRIMARY KEY AUTOINCREMENT NOT NULL, url text)"

	_, err = db.Exec(sql)
	if err != nil {
		fmt.Printf("%q: %s\n", err, sql)
		return
	}

	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
