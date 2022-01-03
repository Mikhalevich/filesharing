package handler

import (
	"errors"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httpcode"
	"github.com/Mikhalevich/filesharing/pkg/proto/types"
)

// RegisterHandler register a new storage(user)
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	storageName := r.FormValue("name")
	password := r.FormValue("password")

	if storageName == "" {
		h.Error(httpcode.NewInvalidParams("invalid storage name"), w, "RegisterHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewInvalidParams(err.Error()).WithError(err), w, "RegisterHandler")
		return
	}

	if sp.IsPublic {
		h.Error(httpcode.NewAlreadyExistError("storage with this name already exists"), w, "RegisterHandler")
		return
	}

	token, err := h.auth.CreateUser(&types.User{
		Name:     storageName,
		Password: password,
	})

	if err != nil {
		if errors.Is(err, ErrAlreadyExist) {
			h.Error(httpcode.NewAlreadyExistError("storage with this name already exists"), w, "RegisterHandler")
		}
		h.Error(httpcode.NewInternalError("registration error").WithError(err), w, "RegisterHandler")
		return
	}

	w.Write([]byte(token.Value))

	err = h.storage.CreateStorage(storageName, true)
	if err != nil {
		if !errors.Is(err, ErrAlreadyExist) {
			h.Error(httpcode.NewInternalError("unable to create storage"), w, "RegisterHandler")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
