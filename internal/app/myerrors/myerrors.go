package myerrors

import (
	"fmt"
	"net/http"
)

type NotFoundURLError struct {
	StatusCode      int
	ExternalMessage string
	URL             string
}

type UserHasNoURLsError struct {
	StatusCode      int
	ExternalMessage string
	UserID          string
}

type DeletedURLError struct {
	StatusCode      int
	ExternalMessage string
	URL             string
}

type InvalidURLError struct {
	StatusCode      int
	ExternalMessage string
	URL             string
}

func NewNotFoundURLError(URL string) error {
	return &NotFoundURLError{
		StatusCode:      http.StatusNotFound,
		ExternalMessage: "Not found URL",
		URL:             URL,
	}
}

func NewUserHasNoURLsError(userID string) error {
	return &UserHasNoURLsError{
		StatusCode:      http.StatusNoContent,
		ExternalMessage: "User has no URLs",
		UserID:          userID,
	}
}

func NewDeletedURLError(URL string) error {
	return &DeletedURLError{
		StatusCode:      http.StatusGone,
		ExternalMessage: "URL deleted",
		URL:             URL,
	}
}

func NewInvalidURLError(URL string) error {
	return &InvalidURLError{
		StatusCode:      http.StatusBadRequest,
		ExternalMessage: "Invalid URL",
		URL:             URL,
	}
}

func (e *NotFoundURLError) Error() string {
	return fmt.Sprintf("Not found shortURL %s", e.URL)
}

func (e *UserHasNoURLsError) Error() string {
	return fmt.Sprintf("User %s has no URLs", e.UserID)
}

func (e *DeletedURLError) Error() string {
	return fmt.Sprintf("URL %s deleted", e.URL)
}

func (e *InvalidURLError) Error() string {
	return fmt.Sprintf("Invalid URL %s", e.URL)
}
