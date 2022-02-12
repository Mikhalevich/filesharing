package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing/pkg/ctxinfo"
	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/proto/file"
	"github.com/Mikhalevich/filesharing/pkg/service"
)

const (
	// Title it's just title for view page
	Title = "Duplo"
)

var (
	// ErrAlreadyExist indicates that storage already exists
	ErrAlreadyExist = errors.New("alredy exist")
	ErrNotExist     = errors.New("not exist")
	ErrPwdNotMatch  = errors.New("password not match")
	ErrExpired      = errors.New("session is expired")
)

type Auther interface {
	Create(user *auth.User) (*auth.Token, error)
	Auth(user *auth.User) (*auth.Token, error)
	AuthPublicUser(name string) (*auth.Token, error)
	UserByToken(token string) (*auth.User, error)
}

type Filer interface {
	Files(storage string, isPermanent bool) ([]*file.File, error)
	Create(storage string, withPermanent bool) error
	Remove(storage string, isPermanent bool, fileName string) error
	Get(storage string, isPermanent bool, fileName string, w io.Writer) error
	Upload(storage string, isPermanent bool, fileName string, r io.Reader) (*file.File, error)
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	WithField(key string, value interface{}) service.Logger
	WithError(err error) service.Logger
}

// Handler represents gateway handler
type Handler struct {
	auth    Auther
	file    Filer
	logger  Logger
	filePub micro.Event
}

// NewHandler constructor for Handler
func NewHandler(a Auther, f Filer, l Logger, filePub micro.Event) *Handler {
	return &Handler{
		auth:    a,
		file:    f,
		logger:  l,
		filePub: filePub,
	}
}

func (h *Handler) Error(err *httperror.Error, w http.ResponseWriter, handler string) {
	h.logger.WithError(err).
		WithField("handler", handler).
		Error("handler error")
	err.WriteJSON(w)
}

type storageParameters struct {
	UserID      int64
	StorageName string
	IsPublic    bool
	IsPermanent bool
	FileName    string
}

func (h *Handler) requestParameters(r *http.Request) (storageParameters, error) {
	ctx := r.Context()
	userID, err := ctxinfo.UserID(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		userID = 0
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get user id: %w", err)
	}

	storage, err := ctxinfo.UserName(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		storage = ""
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get storage name: %w", err)
	}

	isPublic, err := ctxinfo.PublicStorage(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		isPublic = false
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get public storage: %w", err)
	}

	isPermanent, err := ctxinfo.PermanentStorage(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		isPermanent = false
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get permanent storage: %w", err)
	}

	fileName, err := ctxinfo.FileName(ctx)
	if errors.Is(err, ctxinfo.ErrNotFound) {
		fileName = ""
	} else if err != nil {
		return storageParameters{}, fmt.Errorf("unable to get file name: %w", err)
	}

	return storageParameters{
		UserID:      userID,
		StorageName: storage,
		IsPublic:    isPublic,
		IsPermanent: isPermanent,
		FileName:    fileName,
	}, nil
}

// func (h *Handler) respondWithInvalidMethodError(m string, w http.ResponseWriter) bool {
// 	if m != http.MethodPost {
// 		h.logger.Errorf("invalid method %s", m)
// 		http.Error(w, "only POST method allowed", http.StatusMethodNotAllowed)
// 		return true
// 	}
// 	return false
// }

// RecoverMiddleware middlewere recover for undefined panic error
func (h *Handler) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				h.Error(httperror.NewInternalError("panic recovered").WithError(e), w, "RecoverHandler")
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken == "" {
		return ""
	}

	args := strings.Split(bearerToken, " ")
	if len(args) < 2 {
		return ""
	}

	return args[1]
}

// CheckAuthMiddleware middlewere for auth
func (h *Handler) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := h.requestParameters(r)
		if err != nil {
			h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "CheckAuthMiddleware")
			return
		}

		if p.IsPublic {
			next.ServeHTTP(w, r)
			return
		}

		if p.StorageName == "" {
			h.Error(httperror.NewInvalidParams("storage name is empty"), w, "CheckAuthMiddleware")
			return
		}

		token := extractToken(r)
		if token == "" {
			t, err := h.auth.AuthPublicUser(p.StorageName)
			if err != nil {
				h.Error(httperror.NewUnauthorized("unable to get token").WithError(err), w, "CheckAuthMiddleware")
				return
			}
			token = t.GetValue()
			w.Header().Set("X-Token", token)
		}

		user, err := h.auth.UserByToken(token)
		if err != nil {
			t, err := h.auth.AuthPublicUser(p.StorageName)
			if err != nil {
				h.Error(httperror.NewUnauthorized("unable to get user by token").WithError(err), w, "CheckAuthMiddleware")
				return
			}
			token = t.GetValue()
			w.Header().Set("X-Token", token)
		}

		if user.Name != p.StorageName {
			h.Error(httperror.NewNotMatchError(fmt.Sprintf("invalid request user = %s, storage = %s", user, p.StorageName)), w, "CheckAuthMiddleware")
			return
		}

		ctx := ctxinfo.WithUserID(r.Context(), user.Id)
		if !p.IsPublic && user.Public {
			ctx = ctxinfo.WithPublicStorage(ctx, true)
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// CreateStorageMiddleware middleware check storage for existence
func (h *Handler) CreateStorageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := h.requestParameters(r)
		if err != nil {
			h.Error(httperror.NewInvalidParams("request params").WithError(err), w, "CreateStorageMiddleware")
			return
		}
		err = h.createIfNotExist(p.StorageName, true)
		if err != nil {
			h.Error(httperror.NewInternalError("unable to create storage").WithError(err), w, "CreateStorageMiddleware")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) createIfNotExist(name string, isPermanent bool) error {
	err := h.file.Create(name, isPermanent)
	if errors.Is(err, ErrAlreadyExist) {
		return nil
	}
	return err
}
