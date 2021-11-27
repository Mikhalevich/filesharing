package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Mikhalevich/filesharing/pkg/httpcode"
)

// ShareTextHandler crate file from share text request
func (h *Handler) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.Error(httpcode.NewBadRequest(fmt.Sprintf("title or body was not set; title = %s body = %s", title, body)), w, "ShareTextHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapBadRequest(err, "invalid parameters"), w, "ShareTextHandler")
		return
	}

	_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, title, strings.NewReader(body))
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, fmt.Sprintf("unable to store text file: %s for storage: %s", title, sp.StorageName)), w, "ShareTextHandler")
	}

	w.WriteHeader(http.StatusOK)
}
