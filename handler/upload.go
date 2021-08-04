package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// UploadHandler upload file to storage
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapBadRequest(err, "invalid parameters"), w, "UploadHandler")
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "request data error"), w, "UploadHandler")
		return
	}

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			h.Error(httpcode.NewWrapInternalServerError(err, "request data error"), w, "UploadHandler")
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, fileName, part)
		if err != nil {
			h.Error(httpcode.NewWrapInternalServerError(err, fmt.Sprintf("unable to store file %s", fileName)), w, "UploadHandler")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
