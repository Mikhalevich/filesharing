package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/proto/types"
)

// LoginHandler sign in for the existing storage(user)
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "LoginHandler")
		return
	}

	if sp.IsPublic {
		h.Error(httperror.NewInvalidParams(fmt.Sprintf("no need to login into %s", sp.StorageName)), w, "LoginHandler")
		return
	}

	if sp.StorageName == "" {
		h.Error(httperror.NewInvalidParams("invalid storage name"), w, "LoginHandler")
		return
	}

	password := r.FormValue("password")
	if password == "" {
		h.Error(httperror.NewInvalidParams("invalid password"), w, "LoginHandler")
		return
	}

	token, err := h.auth.Auth(&types.User{
		Name:     sp.StorageName,
		Password: password,
	})

	if errors.Is(err, ErrNotExist) {
		h.Error(httperror.NewNotExistError("no such storage"), w, "LoginHandler")
		return
	} else if errors.Is(err, ErrPwdNotMatch) {
		h.Error(httperror.NewNotMatchError("not match"), w, "LoginHandler")
		return
	}

	w.Write([]byte(token.Value))
	w.WriteHeader(http.StatusOK)
}
