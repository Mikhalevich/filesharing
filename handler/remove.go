package handler

import (
	"errors"
	"fmt"
	"net/http"
)

// RemoveHandler removes current file from storage
func (h *Handler) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.respondWithError(errors.New("file error"), w, "RemoveHandler", "file name was not set", http.StatusBadRequest)
		return
	}

	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "RemoveHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	err = h.storage.Remove(sp.StorageName, sp.IsPermanent, fileName)
	// if err == fs.ErrNotExists {
	// 	h.respondWithError(fileNotExistError(fileName), w, "file name doesn't exist", http.StatusBadRequest)
	// 	return
	// }

	if h.respondWithError(err, w, "RemoveHandler", fmt.Sprintf("unable to remove file: %s from storage: %s", fileName, sp.StorageName), http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}
