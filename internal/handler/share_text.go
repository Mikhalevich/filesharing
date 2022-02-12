package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// ShareTextHandler crate file from share text request
func (h *Handler) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.Error(httperror.NewInvalidParams(fmt.Sprintf("title or body was not set; title = %s body = %s", title, body)), w, "ShareTextHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "ShareTextHandler")
		return
	}

	_, err = h.file.Upload(sp.StorageName, sp.IsPermanent, title, strings.NewReader(body))
	if err != nil {
		h.Error(httperror.NewInternalError(fmt.Sprintf("unable to store text file: %s for storage: %s", title, sp.StorageName)).WithError(err), w, "ShareTextHandler")
	}

	w.WriteHeader(http.StatusOK)
}
