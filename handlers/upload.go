package handlers

import (
	"fmt"
	"io"
	"net/http"
)

// UploadHandler upload file to storage
func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "UploadHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	mr, err := r.MultipartReader()
	if h.respondWithError(err, w, "UploadHandler", "request data error", http.StatusInternalServerError) {
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if h.respondWithError(err, w, "UploadHandler", "request data error", http.StatusInternalServerError) {
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, fileName, part)
		if h.respondWithError(err, w, "UploadHandler", fmt.Sprintf("unable to store file %s", fileName), http.StatusInternalServerError) {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
