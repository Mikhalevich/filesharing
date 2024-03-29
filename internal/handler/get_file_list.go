package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
)

// GetFileList returns json encoded file list
func (h *Handler) GetFileList(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "GetFileList")
		return
	}

	files, err := h.file.Files(sp.StorageName, sp.IsPermanent)
	if err != nil {
		h.Error(httperror.NewInternalError(fmt.Sprintf("unable to get files from storage: %s", sp.StorageName)).WithError(err), w, "GetFileList")
		return
	}

	type JSONInfo struct {
		Name    string `json:"name"`
		Size    int64  `json:"size"`
		ModTime int64  `json:"mod_time"`
	}
	info := make([]JSONInfo, 0, len(files))
	for _, f := range files {
		info = append(info, JSONInfo{
			Name:    f.Name,
			Size:    f.Size,
			ModTime: f.ModTime,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		h.Error(httperror.NewInternalError("json encoder error").WithError(err), w, "GetFileList")
		return
	}
}
