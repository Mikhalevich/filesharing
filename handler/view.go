package handler

import (
	"fmt"
	"net/http"

	"github.com/Mikhalevich/filesharing/template"
)

// ViewHandler executes view.html template for view files in requested folder
func (h *Handler) ViewHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "ViewHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	if !h.storage.IsStorageExists(sp.StorageName) {
		h.respondWithError(fmt.Errorf("invalid storage: %s", sp.StorageName), w, "ViewHandler", "storage does not exist", http.StatusInternalServerError)
		return
	}

	files, err := h.storage.Files(sp.StorageName, sp.IsPermanent)
	if h.respondWithError(err, w, "ViewHandler", fmt.Sprintf("unable to get files from storage: %s", sp.StorageName), http.StatusInternalServerError) {
		return
	}

	fileInfos := make([]template.FileInfo, 0, len(files))
	for _, f := range files {
		fileInfos = append(fileInfos, *marshalFileInfo(f))
	}

	viewPermanentLink := !sp.IsPermanent && !h.sc.IsPublic(sp.StorageName)
	viewTemplate := template.NewTemplateView(Title, viewPermanentLink, fileInfos)

	err = viewTemplate.Execute(w)
	if err != nil {
		h.respondWithError(err, w, "ViewHandler", "view error", http.StatusInternalServerError)
	}
}
