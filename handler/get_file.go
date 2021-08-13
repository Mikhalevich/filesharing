package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/httpcode"
)

// GetFileHandler get single file from storage
func (h *Handler) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewWrapBadRequest(err, "invalid parameters"), w, "GetFileHandler")
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, sp.FileName, pw)
		pw.CloseWithError(err)
	}()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sp.FileName))

	_, err = io.Copy(w, pr)
	if err != nil {
		h.Error(httpcode.NewWrapInternalServerError(err, "can't open file"), w, "GetFileHandler")
		return
	}
}
