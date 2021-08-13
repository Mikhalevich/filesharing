package handler

import (
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// RemoveHandler removes current file from storage
func (h *Handler) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.Error(httpcode.NewBadRequest("file name was not set"), w, "RemoveHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapBadRequest(err, "invalid parameters"), w, "RemoveHandler")
		return
	}

	err = h.storage.Remove(sp.StorageName, sp.IsPermanent, fileName)
	// if err == fs.ErrNotExists {
	// 	h.respondWithError(fileNotExistError(fileName), w, "file name doesn't exist", http.StatusBadRequest)
	// 	return
	// }

	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, fmt.Sprintf("unable to remove file: %s from storage: %s", fileName, sp.StorageName)), w, "RemoveHandler")
		return
	}

	w.WriteHeader(http.StatusOK)
}
