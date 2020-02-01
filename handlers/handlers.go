package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Mikhalevich/filesharing/fs"
	"github.com/Mikhalevich/filesharing/templates"
	"github.com/Mikhalevich/goauth"
	"github.com/sirupsen/logrus"
)

const (
	Title = "Duplo"
)

type StorageChecker interface {
	Name(r *http.Request) string
	IsPermanent(r *http.Request) bool
	IsPublic(name string) bool
}

type Authentificator interface {
	GetUser(r *http.Request) (*goauth.User, error)
	AuthorizeByName(name, password, ip string) (*goauth.Session, error)
	RegisterByName(name, password string) (*goauth.Session, error)
}

type Handlers struct {
	sc                 StorageChecker
	auth               Authentificator
	fs                 *fs.FileStorage
	logger             *logrus.Logger
	rootPath           string
	permanentDirectory string
	temporaryDirectory string
}

func NewHandlers(checker StorageChecker, a Authentificator, fs *fs.FileStorage, l *logrus.Logger, root, pertament, temp string) *Handlers {
	return &Handlers{
		sc:                 checker,
		auth:               a,
		fs:                 fs,
		logger:             l,
		rootPath:           root,
		permanentDirectory: pertament,
		temporaryDirectory: temp,
	}
}

func (h *Handlers) path(name string) string {
	return path.Join(h.rootPath, name)
}

func (h *Handlers) permanentPath(name string) string {
	return path.Join(h.path(name), h.permanentDirectory)
}

func (h *Handlers) currentPath(r *http.Request) string {
	if h.sc.IsPermanent(r) {
		return h.permanentPath(h.sc.Name(r))
	}

	return h.path(h.sc.Name(r))
}

func (h *Handlers) FileServer() http.Handler {
	return http.FileServer(http.Dir(h.rootPath))
}

func (h *Handlers) respondWithError(err error, w http.ResponseWriter, description string, httpStatusCode int) bool {
	if err != nil {
		h.logger.Error(err)
		http.Error(w, description, httpStatusCode)
		return true
	}

	return false
}

func (h *Handlers) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.currentPath(r)
	indexPath := path.Join(sPath, "index.html")
	f, err := os.Open(indexPath)
	if h.respondWithError(err, w, "can't open index.html", http.StatusInternalServerError) {
		return
	}
	http.ServeContent(w, r, indexPath, time.Now(), f)
}

func (h *Handlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := templates.NewTemplateRegister()
	renderTemplate := true

	defer func() {
		if renderTemplate {
			if err := userInfo.Execute(w); err != nil {
				h.logger.Error(err)
			}
		}
	}()

	if r.Method != http.MethodPost {
		return
	}

	userInfo.StorageName = r.FormValue("name")
	userInfo.Password = r.FormValue("password")

	if userInfo.StorageName == "" {
		userInfo.AddError("name", "Please specify storage name")
		return
	}

	if h.sc.IsPublic(userInfo.StorageName) {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	}

	session, err := h.auth.RegisterByName(userInfo.StorageName, userInfo.Password)
	if err == goauth.ErrAlreadyExists {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	}

	if h.respondWithError(err, w, "registration error", http.StatusInternalServerError) {
		return
	}

	if session != nil {
		h.setUserCookie(w, userInfo.StorageName, session.Value, session.Expires)
	} else {
		if h.exists(userInfo.StorageName) {
			userInfo.AddError("common", "Storage with this name already exists")
			return
		}
	}

	err = h.createIfNotExist(userInfo.StorageName, true)
	if h.respondWithError(err, w, "unable to create storage", http.StatusInternalServerError) {
		return
	}

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", userInfo.StorageName), http.StatusFound)
}

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := templates.NewTemplatePassword()
	renderTemplate := true
	defer func() {
		if renderTemplate {
			if err := userInfo.Execute(w); err != nil {
				h.logger.Error(err)
			}
		}
	}()

	storageName := h.sc.Name(r)

	if h.sc.IsPublic(storageName) {
		userInfo.AddError("common", fmt.Sprintf("No need to login into %s", storageName))
		return
	}

	if r.Method != http.MethodPost {
		return
	}

	userInfo.Password = r.FormValue("password")

	if storageName == "" {
		userInfo.AddError("name", "Please specify storage name to login")
	}

	if userInfo.Password == "" {
		userInfo.AddError("password", "Please enter password to login")
	}

	if len(userInfo.Errors) > 0 {
		return
	}

	session, err := h.auth.AuthorizeByName(storageName, userInfo.Password, r.RemoteAddr)
	if err == goauth.ErrManyRequests {
		userInfo.AddError("common", "Request is not allowed, too many requests")
		return
	} else if err == goauth.ErrNoSuchUser || err == goauth.ErrPwdNotMatch {
		userInfo.AddError("common", "Invalid user name or password")
		return
	} else if h.respondWithError(err, w, "authorization error", http.StatusInternalServerError) {
		return
	}

	if session != nil {
		h.setUserCookie(w, storageName, session.Value, session.Expires)
	} else {
		if !h.exists(storageName) {
			userInfo.AddError("common", "Invalid user name or password")
			return
		}
	}

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", storageName), http.StatusFound)
}

func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.currentPath(r)
	_, err := os.Stat(sPath)
	if h.respondWithError(err, w, "invalid storage", http.StatusInternalServerError) {
		return
	}

	viewTemplate := templates.NewTemplateView(Title, h.fs.Files(sPath))

	err = viewTemplate.Execute(w)
	if err != nil {
		h.logger.Error(err)
	}
}

func (h *Handlers) JSONViewHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.currentPath(r)
	_, err := os.Stat(sPath)
	if h.respondWithError(err, w, "invalid storage", http.StatusInternalServerError) {
		return
	}

	list := h.fs.Files(sPath)

	type JSONInfo struct {
		Name string `json:"name"`
	}
	info := make([]JSONInfo, 0, len(list))
	for _, l := range list {
		info = append(info, JSONInfo{Name: l.Name()})
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err = enc.Encode(info)
	if err != nil {
		h.logger.Error(err)
	}
}

func (h *Handlers) respondWithInvalidMethodError(m string, w http.ResponseWriter) bool {
	if m != http.MethodPost {
		h.logger.Errorf("invalid method %s", m)
		http.Error(w, "only POST method allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}

func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	mr, err := r.MultipartReader()
	if h.respondWithError(err, w, "internal server error", http.StatusInternalServerError) {
		return
	}

	sPath := h.currentPath(r)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if h.respondWithError(err, w, "internal server error", http.StatusInternalServerError) {
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		tempFileName, err := h.fs.Store(h.temporaryDirectory, fileName, part)
		if h.respondWithError(err, w, "unable to create file", http.StatusInternalServerError) {
			return
		}

		err = h.fs.Move(path.Join(h.temporaryDirectory, tempFileName), sPath, fileName)
		if err != nil {
			h.logger.Error(err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.respondWithError(errors.New("remove handler: file name was not set"), w, "file name was not set", http.StatusBadRequest)
		return
	}

	sPath := h.currentPath(r)
	err := h.fs.Remove(sPath, fileName)
	if err == fs.ErrNotExists {
		h.respondWithError(errors.New(fileName+" doesn't exist"), w, "file name doesn't exist", http.StatusBadRequest)
		return
	}

	if h.respondWithError(err, w, "unable to remove file", http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		err := fmt.Errorf("share text: title or body was not set; title = %s body = %s", title, body)
		h.respondWithError(err, w, "title or body was not set", http.StatusBadRequest)
		return
	}

	sPath := h.currentPath(r)
	_, err := h.fs.Store(sPath, title, strings.NewReader(body))
	if h.respondWithError(err, w, "unable to store text file", http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RecoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				h.respondWithError(fmt.Errorf("someting has gone wrong: %w", e), w, "internal server error", http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) CheckAuth(next http.Handler) http.Handler {
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
			if h.respondWithError(err, w, "unable to create storage", http.StatusInternalServerError) {
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		_, err = h.auth.GetUser(r)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/login/%s", storageName), http.StatusFound)
			return
		}

		err = h.createIfNotExist(storageName, true)
		if h.respondWithError(err, w, "unable to create storage", http.StatusInternalServerError) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) createSkel(storageName string, permanent bool) error {
	createFolder := func(path string) error {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	}

	err := createFolder(h.path(storageName))
	if err != nil {
		return err
	}

	if permanent {
		err := createFolder(h.permanentPath(storageName))
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) exists(storageName string) bool {
	_, err := os.Stat(h.path(storageName))
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

func (h *Handlers) createIfNotExist(storageName string, permanent bool) error {
	if !h.exists(storageName) {
		return h.createSkel(storageName, permanent)
	}
	return nil
}

func (h *Handlers) setUserCookie(w http.ResponseWriter, sessionName, sessionId string, expires int64) {
	cookie := http.Cookie{Name: sessionName, Value: sessionId, Path: "/", Expires: time.Unix(expires, 0), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func (h *Handlers) removeCookie(w http.ResponseWriter, sessionName string) {
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}
