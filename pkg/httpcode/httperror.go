package httpcode

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	statusCode  int
	description string
}

func NewHTTPError(code int, description string) *HTTPError {
	return &HTTPError{
		statusCode:  code,
		description: description,
	}
}

func (e *HTTPError) StatusCode() int {
	return e.statusCode
}

func (e *HTTPError) Description() string {
	return e.description
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s, StatusCode = %d", e.description, e.statusCode)
}

func NewAlreadyExistError(description string) *HTTPError {
	return NewHTTPError(HTTPStatusAlreadyExist, description)
}

func NewNotExistError(description string) *HTTPError {
	return NewHTTPError(HTTPStatusNotExist, description)
}

func NewNotMatchError(description string) *HTTPError {
	return NewHTTPError(HTTPStatusNotMatch, description)
}

func NewInternalServerError(description string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, description)
}

func NewBadRequest(description string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, description)
}
