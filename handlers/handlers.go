package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Mikhalevich/filesharing/templates"
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

// Handlers represents gateway handlers
type Handlers struct {
	sc      StorageChecker
	session Sessioner
	auth    Authentificator
	storage Storager
	logger  *logrus.Logger
}

// NewHandlers constructor for Handlers
func NewHandlers(checker StorageChecker, ses Sessioner, a Authentificator, s Storager, l *logrus.Logger) *Handlers {
	return &Handlers{
		sc:      checker,
		session: ses,
		auth:    a,
		storage: s,
		logger:  l,
	}
}

func (h *Handlers) respondWithError(err error, w http.ResponseWriter, context, description string, httpStatusCode int) bool {
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

func (h *Handlers) requestParameters(r *http.Request, withFile bool) (storageParameters, error) {
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

// IndexHTMLHandler process index.html file
func (h *Handlers) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "IndexHTMLHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, "index.html", pw)
		pw.CloseWithError(err)
	}()

	w.Header().Set("Content-type", "text/html")
	_, err = io.Copy(w, pr)
	if h.respondWithError(err, w, "IndexHTMLHandler", "can't open index.html", http.StatusInternalServerError) {
		return
	}
}

// GetFileHandler get single file from storage
func (h *Handlers) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, true)
	if h.respondWithError(err, w, "GetFileHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, sp.FileName, pw)
		pw.CloseWithError(err)
	}()

	_, err = io.Copy(w, pr)
	if h.respondWithError(err, w, "GetFileHandler", "can't open file", http.StatusInternalServerError) {
		return
	}
}

// RegisterHandler register a new storage(user)
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

	token, err := h.auth.CreateUser(&User{
		Name: userInfo.StorageName,
		Pwd:  userInfo.Password,
	})

	if errors.Is(err, ErrAlreadyExist) {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	}

	if h.respondWithError(err, w, "RegisterHandler", "registration error", http.StatusInternalServerError) {
		renderTemplate = false
		return
	}

	h.session.SetToken(w, token, userInfo.StorageName)

	err = h.storage.CreateStorage(userInfo.StorageName, true)
	if !errors.Is(err, ErrAlreadyExist) &&
		h.respondWithError(err, w, "RegisterHandler", "unable to create storage", http.StatusInternalServerError) {
		renderTemplate = false
		return
	}

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", userInfo.StorageName), http.StatusFound)
}

// LoginHandler sign in for the existing storage(user)
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

	token, err := h.auth.Auth(&User{
		Name: storageName,
		Pwd:  userInfo.Password,
	})

	if h.respondWithError(err, w, "LoginHandler", "authorization error", http.StatusInternalServerError) {
		renderTemplate = false
		return
	}

	h.session.SetToken(w, token, storageName)

	renderTemplate = false
	http.Redirect(w, r, fmt.Sprintf("/%s", storageName), http.StatusFound)
}

func marshalFileInfo(file *File) *templates.FileInfo {
	return &templates.FileInfo{
		Name:    file.Name,
		Size:    file.Size,
		ModTime: file.ModTime,
	}
}

// ViewHandler executes view.html template for view files in requested folder
func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
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

	fileInfos := make([]templates.FileInfo, 0, len(files))
	for _, f := range files {
		fileInfos = append(fileInfos, *marshalFileInfo(f))
	}

	viewPermanentLink := !sp.IsPermanent && !h.sc.IsPublic(sp.StorageName)
	viewTemplate := templates.NewTemplateView(Title, viewPermanentLink, fileInfos)

	err = viewTemplate.Execute(w)
	if err != nil {
		h.respondWithError(err, w, "ViewHandler", "view error", http.StatusInternalServerError)
	}
}

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

func (h *Handlers) respondWithInvalidMethodError(m string, w http.ResponseWriter) bool {
	if m != http.MethodPost {
		h.logger.Errorf("invalid method %s", m)
		http.Error(w, "only POST method allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}

// UploadHandler upload file to storage
func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "UploadHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	mr, err := r.MultipartReader()
	if h.respondWithError(err, w, "UploadHandler", "request data error", http.StatusInternalServerError) {
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if h.respondWithError(err, w, "UploadHandler", "request data error", http.StatusInternalServerError) {
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, fileName, part)
		if h.respondWithError(err, w, "UploadHandler", fmt.Sprintf("unable to store file %s", fileName), http.StatusInternalServerError) {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// RemoveHandler removes current file from storage
func (h *Handlers) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	fileName := r.FormValue("fileName")
	if fileName == "" {
		h.respondWithError(errors.New("file error"), w, "RemoveHandler", "file name was not set", http.StatusBadRequest)
		return
	}

	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "RemoveHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	err = h.storage.Remove(sp.StorageName, sp.IsPermanent, fileName)
	// if err == fs.ErrNotExists {
	// 	h.respondWithError(fileNotExistError(fileName), w, "file name doesn't exist", http.StatusBadRequest)
	// 	return
	// }

	if h.respondWithError(err, w, "RemoveHandler", fmt.Sprintf("unable to remove file: %s from storage: %s", fileName, sp.StorageName), http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ShareTextHandler crate file from share text request
func (h *Handlers) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	if h.respondWithInvalidMethodError(r.Method, w) {
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		h.respondWithError(errors.New("param error"), w, "ShareTextHandler", fmt.Sprintf("title or body was not set; title = %s body = %s", title, body), http.StatusBadRequest)
		return
	}

	sp, err := h.requestParameters(r, false)
	if h.respondWithError(err, w, "ShareTextHandler", "invalid parameters", http.StatusInternalServerError) {
		return
	}

	_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, title, strings.NewReader(body))
	if h.respondWithError(err, w, "ShareTextHandler", fmt.Sprintf("unable to store text file: %s for storage: %s", title, sp.StorageName), http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RecoverHandler middlewere recover for undefined panic error
func (h *Handlers) RecoverHandler(next http.Handler) http.Handler {
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

// CheckAuth middlewere for auth
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

func (h *Handlers) createIfNotExist(name string, isPermanent bool) error {
	err := h.storage.CreateStorage(name, isPermanent)
	if errors.Is(err, ErrAlreadyExist) {
		return nil
	}
	return err
}
