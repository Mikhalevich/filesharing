package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/templates"
)

// LoginHandler sign in for the existing storage(user)
func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := templates.NewTemplatePassword()
	renderTemplate := true
	defer func() {
		if renderTemplate {
			if err := userInfo.Execute(w); err != nil {
				h.logger.Error(err)
			}
		}
	}()

	storageName := h.sc.Name(r)

	if h.sc.IsPublic(storageName) {
		userInfo.AddError("common", fmt.Sprintf("No need to login into %s", storageName))
		return
	}

	if r.Method != http.MethodPost {
		return
	}

	userInfo.Password = r.FormValue("password")

	if storageName == "" {
		userInfo.AddError("name", "Please specify storage name to login")
	}

	if userInfo.Password == "" {
		userInfo.AddError("password", "Please enter password to login")
	}

	if len(userInfo.Errors) > 0 {
		return
	}

	token, err := h.auth.Auth(&User{
		Name: storageName,
		Pwd:  userInfo.Password,
	})

	if errors.Is(err, ErrNotExist) {
		userInfo.AddError("name", "No such storage")
		return
	} else if errors.Is(err, ErrPwdNotMatch) {
		userInfo.AddError("password", "Password not match")
		return
	}
	if h.respondWithError(err, w, "LoginHandler", "authorization error", http.StatusInternalServerError) {
		renderTemplate = false
		return
	}

	h.session.SetToken(w, token, storageName)

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", storageName), http.StatusFound)
}
