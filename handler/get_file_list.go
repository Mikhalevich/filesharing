package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// JSONViewHandler it's spike for duplo client
func (h *Handler) GetFileList(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "GetFileList", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	if !h.storage.IsStorageExists(sp.StorageName) {
		h.respondWithError(errors.New("invalid storage"), w, "GetFileList", fmt.Sprintf("storage does not exist: %s", sp.StorageName), http.StatusInternalServerError)
		return
	}

	files, err := h.storage.Files(sp.StorageName, sp.IsPermanent)
	if h.respondWithError(err, w, "GetFileList", fmt.Sprintf("unable to get files from storage: %s", sp.StorageName), http.StatusInternalServerError) {
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
	encoder := json.NewEncoder(w)
	err = encoder.Encode(info)
	if err != nil {
		h.respondWithError(err, w, "GetFileList", "json encoder error", http.StatusInternalServerError)
	}
}
