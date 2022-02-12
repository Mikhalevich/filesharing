package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
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

	token, err := h.auth.Auth(&auth.User{
		Name:     sp.StorageName,
		Password: password,
	})

	if err != nil {
		var httpErr *httperror.Error
		if errors.As(err, &httpErr) {
			switch httpErr.Code {
			case httperror.CodeNotExist:
				h.Error(httperror.NewNotExistError("no such storage"), w, "LoginHandler")

			case httperror.CodeNotMatch:
				h.Error(httperror.NewNotMatchError("not match"), w, "LoginHandler")

			default:
				h.Error(httpErr, w, "LoginHandler")
			}
			return
		}

		h.Error(httperror.NewInternalError("auth error").WithError(err), w, "LoginHandler")
		return
	}

	w.Write([]byte(token.Value))
	w.WriteHeader(http.StatusOK)
}
