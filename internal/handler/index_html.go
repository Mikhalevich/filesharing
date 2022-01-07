package handler

import (
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// IndexHTMLHandler process index.html file
func (h *Handler) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "IndexHTMLHandler")
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, "index.html", pw)
		pw.CloseWithError(err)
	}()

	w.Header().Set("Content-type", "text/html")
	_, err = io.Copy(w, pr)
	if err != nil {
		h.Error(httperror.NewInternalError("can't open index.html").WithError(err), w, "IndexHTMLHandler")
		return
	}
}
