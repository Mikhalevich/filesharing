package handler

import (
	"errors"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
)

// RegisterHandler register a new storage(user)
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	storageName := r.FormValue("name")
	password := r.FormValue("password")

	if storageName == "" {
		h.Error(httperror.NewInvalidParams("invalid storage name"), w, "RegisterHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "RegisterHandler")
		return
	}

	if sp.IsPublic {
		h.Error(httperror.NewAlreadyExistError("storage with this name already exists"), w, "RegisterHandler")
		return
	}

	token, err := h.auth.CreateUser(&auth.User{
		Name:     storageName,
		Password: password,
	})

	if err != nil {
		if errors.Is(err, ErrAlreadyExist) {
			h.Error(httperror.NewAlreadyExistError("storage with this name already exists"), w, "RegisterHandler")
		}
		h.Error(httperror.NewInternalError("registration error").WithError(err), w, "RegisterHandler")
		return
	}

	w.Write([]byte(token.Value))

	err = h.storage.CreateStorage(storageName, true)
	if err != nil {
		if !errors.Is(err, ErrAlreadyExist) {
			h.Error(httperror.NewInternalError("unable to create storage"), w, "RegisterHandler")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
