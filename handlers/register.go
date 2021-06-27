package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/templates"
)

// RegisterHandler register a new storage(user)
func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := templates.NewTemplateRegister()
	renderTemplate := true

	defer func() {
		if renderTemplate {
			if err := userInfo.Execute(w); err != nil {
				h.logger.Error(err)
			}
		}
	}()

	if r.Method != http.MethodPost {
		return
	}

	userInfo.StorageName = r.FormValue("name")
	userInfo.Password = r.FormValue("password")

	if userInfo.StorageName == "" {
		userInfo.AddError("name", "Please specify storage name")
		return
	}

	if h.sc.IsPublic(userInfo.StorageName) {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	}

	token, err := h.auth.CreateUser(&User{
		Name: userInfo.StorageName,
		Pwd:  userInfo.Password,
	})

	if errors.Is(err, ErrAlreadyExist) {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	} else if h.respondWithError(err, w, "RegisterHandler", "registration error", http.StatusInternalServerError) {
		renderTemplate = false
		return
	}

	h.session.SetToken(w, token, userInfo.StorageName)

	err = h.storage.CreateStorage(userInfo.StorageName, true)
	if !errors.Is(err, ErrAlreadyExist) &&
		h.respondWithError(err, w, "RegisterHandler", "unable to create storage", http.StatusInternalServerError) {
		renderTemplate = false
		return
	}

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", userInfo.StorageName), http.StatusFound)
}
