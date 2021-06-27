package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ShareTextHandler crate file from share text request
func (h *Handler) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.respondWithError(errors.New("param error"), w, "ShareTextHandler", fmt.Sprintf("title or body was not set; title = %s body = %s", title, body), http.StatusBadRequest)
		return
	}

	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "ShareTextHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, title, strings.NewReader(body))
	if h.respondWithError(err, w, "ShareTextHandler", fmt.Sprintf("unable to store text file: %s for storage: %s", title, sp.StorageName), http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}
