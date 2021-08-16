package httpcode

import (
	"net/http"
)

type HTTPRedirectError struct {
	*HTTPError
	location string
}

func (e *HTTPRedirectError) Location() string {
	return e.location
}

func NewHTTPRedirectFoundError(location string, description string) *HTTPRedirectError {
	return &HTTPRedirectError{
		HTTPError: &HTTPError{
			statusCode:  http.StatusFound,
			description: description,
		},
		location: location,
	}
}
