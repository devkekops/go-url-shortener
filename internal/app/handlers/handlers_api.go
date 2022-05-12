package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/devkekops/go-url-shortener/internal/app/myerrors"
	"github.com/devkekops/go-url-shortener/internal/app/storage"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
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
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			log.Println(err)
			return
		}
		originalURL := link.URL

		if err := checkLongURL(originalURL); err != nil {
			iue := err.(*myerrors.InvalidURLError)
			http.Error(w, iue.ExternalMessage, iue.StatusCode)
			log.Println(err)
			return
		}

		shortURL, err := bh.linkRepo.SaveLongLink(originalURL, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				w.WriteHeader(http.StatusBadRequest)
				log.Println(err)
				return
			}
			if !pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				w.WriteHeader(http.StatusBadRequest)
				log.Println(err)
				return
			}
		}

		r := Result{bh.baseURL + "/" + shortURL}
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(r)

		w.Header().Set("Content-Type", "application/json")
		if err == nil {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		w.Write(buf.Bytes())
	}
}

func (bh *BaseHandler) apiUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)

		userLinks, err := bh.linkRepo.GetUserLinks(userID)
		if err != nil {
			var userHasNoURLsError *myerrors.UserHasNoURLsError
			if errors.As(err, &userHasNoURLsError) {
				http.Error(w, userHasNoURLsError.ExternalMessage, userHasNoURLsError.StatusCode)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			log.Println(err)
			return
		}

		for i := range userLinks {
			userLinks[i].ShortURL = bh.baseURL + "/" + userLinks[i].ShortURL
		}

		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(userLinks)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

func (bh *BaseHandler) apiShortenBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)

		var longURLUnits []storage.LongURLUnit

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&longURLUnits)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			log.Println(err)
			return
		}

		shortURLUnits, err := bh.linkRepo.SaveLongLinks(longURLUnits, userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		for i := range shortURLUnits {
			shortURLUnits[i].ShortURL = bh.baseURL + "/" + shortURLUnits[i].ShortURL
		}

		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(shortURLUnits)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(buf.Bytes())
	}
}

func (bh *BaseHandler) apiDeleteUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)

		b, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		var shortURLs []string
		err = json.Unmarshal(b, &shortURLs)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			log.Println(err)
			return
		}

		bh.linkRepo.DeleteUserLinks(userID, shortURLs)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	}
}
