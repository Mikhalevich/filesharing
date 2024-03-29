package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/proto/event"
)

// RemoveHandler removes current file from storage
func (h *Handler) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.Error(httperror.NewInvalidParams("file name was not set"), w, "RemoveHandler")
		return
	}

	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "RemoveHandler")
		return
	}

	err = h.file.Remove(sp.StorageName, sp.IsPermanent, fileName)
	// if err == fs.ErrNotExists {
	// 	h.respondWithError(fileNotExistError(fileName), w, "file name doesn't exist", http.StatusBadRequest)
	// 	return
	// }

	if err != nil {
		h.Error(httperror.NewInternalError(fmt.Sprintf("unable to remove file: %s from storage: %s", fileName, sp.StorageName)).WithError(err), w, "RemoveHandler")
		return
	}

	go func() {
		h.filePub.Publish(context.Background(), &event.FileEvent{
			UserID:   sp.UserID,
			UserName: sp.StorageName,
			FileName: fileName,
			Time:     time.Now().Unix(),
			Action:   event.Action_Remove,
		})
	}()

	w.WriteHeader(http.StatusOK)
}
