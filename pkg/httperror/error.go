package httperror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Error struct {
	Code        Code   `json:"code"`
	Description string `json:"description"`
	err         error
}

func New(code Code, description string) *Error {
	return &Error{
		Code:        code,
		Description: description,
	}
}

func (e *Error) WithError(err error) *Error {
	e.err = err

	if e.Code == CodeInvalidParams {
		e.Description = fmt.Sprintf("%s: %s", e.Description, err.Error())
	}

	return e
}

func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf(`description: "%s", code = %d, origin err = "%v"`, e.Description, e.Code, e.err)
	}

	return fmt.Sprintf(`description: "%s", code = %d`, e.Description, e.Code)
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) WriteJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(e)
}

func NewInternalError(description string) *Error {
	return New(CodeInternalError, description)
}

func NewInvalidParams(description string) *Error {
	return New(CodeInvalidParams, description)
}

func NewUnauthorized(description string) *Error {
	return New(CodeUnauthorized, description)
}

func NewAlreadyExistError(description string) *Error {
	return New(CodeAlreadyExist, description)
}

func NewNotExistError(description string) *Error {
	return New(CodeNotExist, description)
}

func NewNotMatchError(description string) *Error {
	return New(CodeNotMatch, description)
}
