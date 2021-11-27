package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httpcode"
	"github.com/Mikhalevich/filesharing/pkg/proto/types"
)

// LoginHandler sign in for the existing storage(user)
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapBadRequest(err, "invalid parameters"), w, "LoginHandler")
		return
	}

	if sp.IsPublic {
		h.Error(httpcode.NewBadRequest(fmt.Sprintf("No need to login into %s", sp.StorageName)), w, "LoginHandler")
		return
	}

	if sp.StorageName == "" {
		h.Error(httpcode.NewBadRequest("invalid storage name"), w, "LoginHandler")
		return
	}

	password := r.FormValue("password")
	if password == "" {
		h.Error(httpcode.NewBadRequest("invalid password"), w, "LoginHandler")
		return
	}

	token, err := h.auth.Auth(&types.User{
		Name:     sp.StorageName,
		Password: password,
	})

	if errors.Is(err, ErrNotExist) {
		h.Error(httpcode.NewNotExistError("no such storage"), w, "LoginHandler")
		return
	} else if errors.Is(err, ErrPwdNotMatch) {
		h.Error(httpcode.NewNotMatchError("not match"), w, "LoginHandler")
		return
	}

	w.Write([]byte(token.Value))
	w.WriteHeader(http.StatusOK)
}
