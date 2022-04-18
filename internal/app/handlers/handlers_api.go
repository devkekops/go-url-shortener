package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type URL struct {
	URL string `json:"url"`
}

type Result struct {
	Result string `json:"result"`
}

func (bh *BaseHandler) apiShorten() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)
		var link URL

		if err := json.NewDecoder(req.Body).Decode(&link); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Printf("Incorrect JSON\n")
			return
		}
		originalURL := link.URL
		if !isValidURL(originalURL) {
			http.Error(w, "URL is incorrect", http.StatusBadRequest)
			log.Printf("Incorrect URL %s\n", originalURL)
			return
		}

		shortURL, err := bh.linkRepo.SaveLongLink(originalURL, userID)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Println(err)
			return
		}

		r := Result{bh.baseURL + "/" + shortURL}
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(r)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(buf.Bytes())
	}
}

func (bh *BaseHandler) apiUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)

		userLinks, err := bh.linkRepo.GetUserLinks(userID)
		if err != nil {
			http.Error(w, "Not found URLs", http.StatusNoContent)
			log.Println(err)
			return
		}
		//fmt.Println(userLinks)

		for i := range userLinks {
			userLinks[i].ShortURL = bh.baseURL + "/" + userLinks[i].ShortURL
		}

		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(userLinks)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(buf.Bytes())
	}
}