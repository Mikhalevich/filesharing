package httpcode

import (
	"fmt"
	"net/http"
)

type HTTPWrapError struct {
	*HTTPError
	err error
}

func NewHTTPWrapError(e error, code int, description string) *HTTPWrapError {
	return &HTTPWrapError{
		HTTPError: &HTTPError{
			statusCode:  code,
			description: description,
		},
		err: e,
	}
}

func (e *HTTPWrapError) Error() string {
	return fmt.Sprintf("%s, WrappedError = %v", e.HTTPError.Error(), e.err)
}

func NewWrapBadRequest(err error, description string) *HTTPWrapError {
	return NewHTTPWrapError(err, http.StatusBadRequest, description)
}

func NewWrapInternalServerError(err error, description string) *HTTPWrapError {
	return NewHTTPWrapError(err, http.StatusInternalServerError, description)
}
