package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Mikhalevich/filesharing/template"
	"github.com/sirupsen/logrus"
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

// File represents one file from storage
type File struct {
	Name    string
	Size    int64
	ModTime int64
}

type User struct {
	Name string
	Pwd  string
}

type Token struct {
	Value string
}

// StorageChecker interface retrieve info about storage from request
type StorageChecker interface {
	Name(r *http.Request) string
	IsPermanent(r *http.Request) bool
	FileName(r *http.Request) string
	IsPublic(name string) bool
}

type Sessioner interface {
	GetToken(r *http.Request) (*Token, error)
	SetToken(w http.ResponseWriter, token *Token, name string)
	Remove(w http.ResponseWriter, name string)
}

// Authentificator provide user auth functional
type Authentificator interface {
	CreateUser(user *User) (*Token, error)
	Auth(user *User) (*Token, error)
	UserNameByToken(token string) (string, error)
}

// Storager storage communication interface
type Storager interface {
	Files(storage string, isPermanent bool) ([]*File, error)
	CreateStorage(storage string, withPermanent bool) error
	Remove(storage string, isPermanent bool, fileName string) error
	Get(storage string, isPermanent bool, fileName string, w io.Writer) error
	Upload(storage string, isPermanent bool, fileName string, r io.Reader) (*File, error)
	IsStorageExists(storage string) bool
}

// Handler represents gateway handler
type Handler struct {
	sc      StorageChecker
	session Sessioner
	auth    Authentificator
	storage Storager
	logger  *logrus.Logger
}

// NewHandler constructor for Handler
func NewHandler(checker StorageChecker, ses Sessioner, a Authentificator, s Storager, l *logrus.Logger) *Handler {
	return &Handler{
		sc:      checker,
		session: ses,
		auth:    a,
		storage: s,
		logger:  l,
	}
}

func (h *Handler) respondWithError(err error, w http.ResponseWriter, context, description string, httpStatusCode int) bool {
	if err != nil {
		h.logger.Error(fmt.Errorf("[%s] %s: %w", context, description, err))
		http.Error(w, description, httpStatusCode)
		return true
	}

	return false
}

type storageParameters struct {
	StorageName string
	IsPermanent bool
	FileName    string
}

func (h *Handler) requestParameters(r *http.Request, withFile bool) (storageParameters, error) {
	storage := h.sc.Name(r)
	if storage == "" {
		return storageParameters{}, errors.New("request storage is empty")
	}

	var file string
	if withFile {
		file = h.sc.FileName(r)
		if file == "" {
			return storageParameters{}, errors.New("request file is empty")
		}
	}

	return storageParameters{
		StorageName: storage,
		IsPermanent: h.sc.IsPermanent(r),
		FileName:    file,
	}, nil
}

func marshalFileInfo(file *File) *template.FileInfo {
	return &template.FileInfo{
		Name:    file.Name,
		Size:    file.Size,
		ModTime: file.ModTime,
	}
}

func (h *Handler) respondWithInvalidMethodError(m string, w http.ResponseWriter) bool {
	if m != http.MethodPost {
		h.logger.Errorf("invalid method %s", m)
		http.Error(w, "only POST method allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}

// RecoverMiddleware middlewere recover for undefined panic error
func (h *Handler) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				h.respondWithError(e, w, "RecoverHandler", "internal server error", http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CheckAuthMiddleware middlewere for auth
func (h *Handler) CheckAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storageName := h.sc.Name(r)
		var err error

		if storageName == "" {
			h.logger.Error(fmt.Sprintf("Storage name is empty for %s", r.URL))
			next.ServeHTTP(w, r)
			return
		}

		if h.sc.IsPublic(storageName) {
			err = h.createIfNotExist(storageName, false)
			if h.respondWithError(err, w, "CheckAuth", "unable to create public storage", http.StatusInternalServerError) {
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		token, err := h.session.GetToken(r)
		if err != nil {
			h.logger.Error(fmt.Errorf("[CheckAuth] unable to get token: %w", err))
			http.Redirect(w, r, fmt.Sprintf("/login/%s", storageName), http.StatusFound)
			return
		}

		user, err := h.auth.UserNameByToken(token.Value)
		if err != nil {
			h.logger.Error(fmt.Errorf("[CheckAuth] unable to get user by token: %w", err))
			http.Redirect(w, r, fmt.Sprintf("/login/%s", storageName), http.StatusFound)
			return
		}

		if user != storageName {
			h.logger.Error(fmt.Errorf("[CheckAuth] invalid user user = %s, storage = %s", user, storageName))
			http.Redirect(w, r, fmt.Sprintf("/login/%s", storageName), http.StatusFound)
			return
		}

		err = h.createIfNotExist(storageName, true)
		if h.respondWithError(err, w, "CheckAuth", "unable to create storage", http.StatusInternalServerError) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) createIfNotExist(name string, isPermanent bool) error {
	err := h.storage.CreateStorage(name, isPermanent)
	if errors.Is(err, ErrAlreadyExist) {
		return nil
	}
	return err
}
