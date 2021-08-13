package handler

import (
	"errors"
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// RegisterHandler register a new storage(user)
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	storageName := r.FormValue("name")
	password := r.FormValue("password")

	if storageName == "" {
		h.Error(httpcode.NewBadRequest("invalid storage name"), w, "RegisterHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapBadRequest(err, "invalid parameters"), w, "RegisterHandler")
		return
	}

	if sp.IsPublic {
		h.Error(httpcode.NewAlreadyExistError("storage with this name already exists"), w, "RegisterHandler")
		return
	}

	token, err := h.auth.CreateUser(&User{
		Name: storageName,
		Pwd:  password,
	})

	if err != nil {
		if errors.Is(err, ErrAlreadyExist) {
			h.Error(httpcode.NewAlreadyExistError("storage with this name already exists"), w, "RegisterHandler")
		}
		h.Error(httpcode.NewInternalServerError("registration error"), w, "RegisterHandler")
		return
	}

	w.Write([]byte(token.Value))

	err = h.storage.CreateStorage(storageName, true)
	if err != nil {
		if !errors.Is(err, ErrAlreadyExist) {
			h.Error(httpcode.NewInternalServerError("unable to create storage"), w, "RegisterHandler")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
