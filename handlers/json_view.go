package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// JSONViewHandler it's spike for duplo client
func (h *Handlers) JSONViewHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "JSONViewHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	if !h.storage.IsStorageExists(sp.StorageName) {
		h.respondWithError(errors.New("invalid storage"), w, "JSONViewHandler", fmt.Sprintf("storage does not exist: %s", sp.StorageName), http.StatusInternalServerError)
		return
	}

	files, err := h.storage.Files(sp.StorageName, sp.IsPermanent)
	if h.respondWithError(err, w, "JSONViewHandler", fmt.Sprintf("unable to get files from storage: %s", sp.StorageName), http.StatusInternalServerError) {
		return
	}

	type JSONInfo struct {
		Name string `json:"name"`
	}
	info := make([]JSONInfo, 0, len(files))
	for _, f := range files {
		info = append(info, JSONInfo{Name: f.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(info)
	if err != nil {
		h.respondWithError(err, w, "JSONViewHandler", "json encoder error", http.StatusInternalServerError)
	}
}
