package handlers

import (
	"io"
	"net/http"
)

// IndexHTMLHandler process index.html file
func (h *Handlers) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "IndexHTMLHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, "index.html", pw)
		pw.CloseWithError(err)
	}()

	w.Header().Set("Content-type", "text/html")
	_, err = io.Copy(w, pr)
	if h.respondWithError(err, w, "IndexHTMLHandler", "can't open index.html", http.StatusInternalServerError) {
		return
	}
}
