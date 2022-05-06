package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/devkekops/go-url-shortener/internal/app/myerrors"
)

func checkShortURL(s string) error {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return myerrors.NewInvalidURLError(s)
		}
	}
	return nil
}

func checkLongURL(s string) error {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return myerrors.NewInvalidURLError(s)
	} else {
		return nil
	}
}

func (bh *BaseHandler) shortenLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userIDctx := req.Context().Value(userIDKey)
		userID := userIDctx.(string)

		b, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}
		originalURL := string(b)

		if err = checkLongURL(originalURL); err != nil {
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

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if err == nil {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		w.Write([]byte(bh.baseURL + "/" + shortURL))
	}
}

func (bh *BaseHandler) expandLink() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		shortURL := chi.URLParam(req, "id")

		if err := checkShortURL(shortURL); err != nil {
			iue := err.(*myerrors.InvalidURLError)
			http.Error(w, iue.ExternalMessage, iue.StatusCode)
			log.Println(err)
			return
		}

		url, err := bh.linkRepo.GetLongByShortLink(shortURL)
		if err != nil {
			var notFoundURLError *myerrors.NotFoundURLError
			var deletedURLError *myerrors.DeletedURLError

			if errors.As(err, &notFoundURLError) {
				http.Error(w, notFoundURLError.ExternalMessage, notFoundURLError.StatusCode)
			} else if errors.As(err, &deletedURLError) {
				http.Error(w, deletedURLError.ExternalMessage, deletedURLError.StatusCode)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			log.Println(err)
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (bh *BaseHandler) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := bh.linkRepo.Ping()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
