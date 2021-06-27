package handler

import (
	"fmt"
	"io"
	"net/http"
)

// GetFileHandler get single file from storage
func (h *Handler) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, true)
	if h.respondWithError(err, w, "GetFileHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, sp.FileName, pw)
		pw.CloseWithError(err)
	}()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sp.FileName))

	_, err = io.Copy(w, pr)
	if h.respondWithError(err, w, "GetFileHandler", "can't open file", http.StatusInternalServerError) {
		return
	}
}
