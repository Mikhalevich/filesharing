package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Mikhalevich/filesharing/pkg/httpcode"
	"github.com/Mikhalevich/filesharing/pkg/proto/event"
)

// UploadHandler upload file to storage
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r)
	if err != nil {
		h.Error(httpcode.NewInvalidParams(err.Error()).WithError(err), w, "UploadHandler")
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		h.Error(httpcode.NewInternalError("request data error").WithError(fmt.Errorf("multipart reader: %w", err)), w, "UploadHandler")
		return
	}

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			h.Error(httpcode.NewInternalError("request data error").WithError(fmt.Errorf("next part: %w", err)), w, "UploadHandler")
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, fileName, part)
		if err != nil {
			h.Error(httpcode.NewInternalError(fmt.Sprintf("unable to store file %s", fileName)).WithError(fmt.Errorf("upload: %w", err)), w, "UploadHandler")
			return
		}

		go func() {
			h.filePub.Publish(context.Background(), &event.FileEvent{
				UserID:   sp.UserID,
				UserName: sp.StorageName,
				FileName: fileName,
				Time:     time.Now().Unix(),
				Action:   event.Action_Add,
			})
		}()
	}

	w.WriteHeader(http.StatusOK)
}
