package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Mikhalevich/filesharing/templates"
	"github.com/Mikhalevich/goauth"
	"github.com/sirupsen/logrus"
)

const (
	// Title it's just title for view page
	Title = "Duplo"
)

var (
	// ErrAlreadyExist indicates that storage already exists
	ErrAlreadyExist = errors.New("alredy exist")
)

// File represents one file from storage
type File struct {
	Name    string
	Size    int64
	ModTime int64
}

// StorageChecker interface retrieve info about storage from request
type StorageChecker interface {
	Name(r *http.Request) string
	IsPermanent(r *http.Request) bool
	FileName(r *http.Request) string
	IsPublic(name string) bool
}

// Authentificator provide user auth functional
type Authentificator interface {
	GetUser(r *http.Request) (*goauth.User, error)
	AuthorizeByName(name, password, ip string) (*goauth.Session, error)
	RegisterByName(name, password string) (*goauth.Session, error)
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
	auth    Authentificator
	storage Storager
	logger  *logrus.Logger
}

// NewHandlers constructor for Handlers
func NewHandlers(checker StorageChecker, a Authentificator, s Storager, l *logrus.Logger) *Handlers {
	return &Handlers{
		sc:      checker,
		auth:    a,
		storage: s,
		logger:  l,
	}
}

func (h *Handlers) respondWithError(err error, w http.ResponseWriter, description string, httpStatusCode int) bool {
	if err != nil {
		h.logger.Error(err)
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
	if h.respondWithError(fmt.Errorf("[IndexHTMLHandler] invalid request parameters: %w", err), w, "invalid parameters", http.StatusInternalServerError) {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, "index.html", pw)
		pw.CloseWithError(err)
	}()

	w.Header().Set("Content-type", "text/html")
	_, err = io.Copy(w, pr)
	if h.respondWithError(fmt.Errorf("[IndexHTMLHandler] copy error: %w", err), w, "can't open index.html", http.StatusInternalServerError) {
		return
	}
}

// GetFileHandler get single file from storage
func (h *Handlers) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, true)
	if h.respondWithError(fmt.Errorf("[GetFileHandler] invalid request parameters: %w", err), w, "invalid parameters", http.StatusInternalServerError) {
		return
	}

	pr, pw := io.Pipe()
	go func() {
		err := h.storage.Get(sp.StorageName, sp.IsPermanent, sp.FileName, pw)
		pw.CloseWithError(err)
	}()

	_, err = io.Copy(w, pr)
	if h.respondWithError(fmt.Errorf("[GetFileHandler] copy error: %w", err), w, "can't open file", http.StatusInternalServerError) {
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

	session, err := h.auth.RegisterByName(userInfo.StorageName, userInfo.Password)
	if err == goauth.ErrAlreadyExists {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	}

	if h.respondWithError(err, w, "registration error", http.StatusInternalServerError) {
		return
	}

	checkForAlreadyExistError := false
	if session != nil {
		h.setUserCookie(w, userInfo.StorageName, session.Value, session.Expires)
	} else {
		checkForAlreadyExistError = true
	}

	err = h.storage.CreateStorage(userInfo.StorageName, true)
	if errors.Is(err, ErrAlreadyExist) {
		if checkForAlreadyExistError {
			userInfo.AddError("common", "Storage with this name already exists")
			return
		}
	} else if h.respondWithError(fmt.Errorf("[RegisterHandler] create storage error: %w", err), w, "unable to create storage", http.StatusInternalServerError) {
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

	session, err := h.auth.AuthorizeByName(storageName, userInfo.Password, r.RemoteAddr)
	if err == goauth.ErrManyRequests {
		userInfo.AddError("common", "Request is not allowed, too many requests")
		return
	} else if err == goauth.ErrNoSuchUser || err == goauth.ErrPwdNotMatch {
		userInfo.AddError("common", "Invalid user name or password")
		return
	} else if h.respondWithError(fmt.Errorf("[LoginHandler] authorization error: %w", err), w, "authorization error", http.StatusInternalServerError) {
		return
	}

	if session != nil {
		h.setUserCookie(w, storageName, session.Value, session.Expires)
	} else {
		if !h.storage.IsStorageExists(storageName) {
			userInfo.AddError("common", "Invalid user name or password")
			return
		}
	}

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
	if h.respondWithError(fmt.Errorf("[ViewHandler] invalid parameters: %w", err), w, "invalid parameters", http.StatusInternalServerError) {
		return
	}

	if !h.storage.IsStorageExists(sp.StorageName) {
		h.respondWithError(fmt.Errorf("[ViewHandler] invalid stoarge: %s", sp.StorageName), w, "invalid stoarge", http.StatusInternalServerError)
		return
	}

	files, err := h.storage.Files(sp.StorageName, sp.IsPermanent)
	if h.respondWithError(fmt.Errorf("[ViewHandler] get files error: %w", err), w, "invalid storage", http.StatusInternalServerError) {
		return
	}

	fileInfos := make([]templates.FileInfo, 0, len(files))
	for _, f := range files {
		fileInfos = append(fileInfos, *marshalFileInfo(f))
	}

	viewTemplate := templates.NewTemplateView(Title, fileInfos)

	err = viewTemplate.Execute(w)
	if err != nil {
		h.respondWithError(fmt.Errorf("[ViewHandler] template error: %w", err), w, "view error", http.StatusInternalServerError)
	}
}

// JSONViewHandler it's spike for duplo client
func (h *Handlers) JSONViewHandler(w http.ResponseWriter, r *http.Request) {
	sp, err := h.requestParameters(r, false)
	if h.respondWithError(fmt.Errorf("[JSONViewHandler] invalid parameters: %w", err), w, "invalid parameters", http.StatusInternalServerError) {
		return
	}

	if !h.storage.IsStorageExists(sp.StorageName) {
		h.respondWithError(fmt.Errorf("[JSONViewHandler] invalid stoarge: %s", sp.StorageName), w, "invalid stoarge", http.StatusInternalServerError)
		return
	}

	files, err := h.storage.Files(sp.StorageName, sp.IsPermanent)
	if h.respondWithError(fmt.Errorf("[JSONViewHandler] get files error: %w", err), w, "invalid storage", http.StatusInternalServerError) {
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
		h.respondWithError(fmt.Errorf("[JSONViewHandler] encoder error: %w", err), w, "json encoder error", http.StatusInternalServerError)
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
	if h.respondWithError(fmt.Errorf("[UploadHandler] invalid parameters: %w", err), w, "invalid parameters", http.StatusInternalServerError) {
		return
	}

	mr, err := r.MultipartReader()
	if h.respondWithError(fmt.Errorf("[UploadHandler] multipart reader error: %w", err), w, "internal server error", http.StatusInternalServerError) {
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if h.respondWithError(fmt.Errorf("[UploadHandler] read next part error: %w", err), w, "internal server error", http.StatusInternalServerError) {
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		_, err = h.storage.Upload(sp.StorageName, sp.IsPermanent, fileName, part)
		if h.respondWithError(fmt.Errorf("[UploadHandler] upload to stoarge error: %w", err), w, fmt.Sprintf("unable to store file %s", fileName), http.StatusInternalServerError) {
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
		h.respondWithError(errors.New("[RemoveHandler] file name was not set"), w, "file name was not set", http.StatusBadRequest)
		return
	}

	err := h.storage.Remove(h.sc.Name(r), h.sc.IsPermanent(r), fileName)
	// if err == fs.ErrNotExists {
	// 	h.respondWithError(fileNotExistError(fileName), w, "file name doesn't exist", http.StatusBadRequest)
	// 	return
	// }

	if h.respondWithError(fmt.Errorf("[RemoveHandler] remove file error: %w", err), w, "unable to remove file", http.StatusInternalServerError) {
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
		h.respondWithError(fmt.Errorf("[ShareTextHandler] title or body was not set; title = %s body = %s", title, body), w, "title or body was not set", http.StatusBadRequest)
		return
	}

	_, err := h.storage.Upload(h.sc.Name(r), h.sc.IsPermanent(r), title, strings.NewReader(body))
	if h.respondWithError(fmt.Errorf("[ShareTextHandler] upload error: %w", err), w, "unable to store text file", http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RecoverHandler middlewere recover for undefined panic error
func (h *Handlers) RecoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				h.respondWithError(fmt.Errorf("[RecoverHandler] someting has gone wrong: %w", e), w, "internal server error", http.StatusInternalServerError)
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
			if h.respondWithError(fmt.Errorf("[CheckAuth] create storage error: %w", err), w, "unable to create storage", http.StatusInternalServerError) {
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
		if h.respondWithError(fmt.Errorf("[CheckAuth] create storage error: %w", err), w, "unable to create storage", http.StatusInternalServerError) {
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

func (h *Handlers) setUserCookie(w http.ResponseWriter, sessionName, sessionID string, expires int64) {
	cookie := http.Cookie{Name: sessionName, Value: sessionID, Path: "/", Expires: time.Unix(expires, 0), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func (h *Handlers) removeCookie(w http.ResponseWriter, sessionName string) {
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}
