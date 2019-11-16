package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Mikhalevich/filesharing/fs"
	"github.com/Mikhalevich/filesharing/templates"
	"github.com/Mikhalevich/goauth"
	"github.com/gorilla/context"
)

const (
	SessionExpirePeriod       = 1 * 60 * 60 * 24 * 30 // sec
	LoginRequestMaxCount      = 3
	LoginRequestWaitingPeriod = 60 // sec
	Title                     = "Duplo"
	ContextStoragePath        = "storagePath"
)

type StorageChecker interface {
	Name(r *http.Request) string
	IsPublic(name string) bool
	Path(name string) string
	PermanentPath(name string) string
}

type Authentificator interface {
	GetUser(r *http.Request) (*goauth.User, error)
	AuthorizeByName(name, password, ip string) (*goauth.Session, error)
	RegisterByName(name, password string) (*goauth.Session, error)
}

type Handlers struct {
	sc                 StorageChecker
	auth               Authentificator
	temporaryDirectory string
}

func NewHandlers(checker StorageChecker, a Authentificator, tempDir string) *Handlers {
	return &Handlers{
		sc:                 checker,
		auth:               a,
		temporaryDirectory: tempDir,
	}
}

func (h *Handlers) RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/common/", http.StatusMovedPermanently)
}

func (h *Handlers) IndexHTMLHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.contextStorage(r)
	indexPath := path.Join(sPath, "index.html")
	f, err := os.Open(indexPath)
	if err != nil {
		http.Error(w, "Can't open index.html", http.StatusInternalServerError)
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
				log.Println(err)
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

	session, err := h.auth.RegisterByName(userInfo.Name, userInfo.Password)
	if err != nil {
		if err == goauth.ErrAlreadyExists {
			userInfo.AddError("common", "Storage with this name already exists")
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate = false
	h.setUserCookie(w, session.Name, session.Value, session.Expires)
	http.Redirect(w, r, fmt.Sprintf("/%s", userInfo.StorageName), http.StatusFound)
}

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := templates.NewTemplatePassword()
	renderTemplate := true
	defer func() {
		if renderTemplate {
			if err := userInfo.Execute(w); err != nil {
				log.Println(err)
			}
		}
	}()

	storageName := h.sc.Name(r)

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

	session, err := h.auth.AuthorizeByName(userInfo.Name, userInfo.Password, r.RemoteAddr)
	if err == goauth.ErrManyRequests {
		userInfo.AddError("common", "Request is not allowed, too many requests")
		return
	} else if err == goauth.ErrNoSuchUser || err == goauth.ErrPwdNotMatch {
		userInfo.AddError("common", "Invalid user name or password")
		return
	} else if err != nil {
		userInfo.AddError("name", err.Error())
		return
	}

	renderTemplate = false
	h.setUserCookie(w, storageName, session.Value, session.Expires)
	http.Redirect(w, r, fmt.Sprintf("/%s", storageName), http.StatusFound)
}

func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.contextStorage(r)
	_, err := os.Stat(sPath)
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	viewTemplate := templates.NewTemplateView(Title, fs.NewDirectory(sPath).List())

	err = viewTemplate.Execute(w)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handlers) JSONViewHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.contextStorage(r)
	_, err := os.Stat(sPath)
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	list := fs.NewDirectory(sPath).List()

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
		log.Println(err)
	}
}

func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	mr, err := r.MultipartReader()
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	files := []string{}
	sPath := h.contextStorage(r)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if respondError(err, w, http.StatusInternalServerError) {
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		fn, err := h.storeTempFile(fileName, part)
		if respondError(err, w, http.StatusInternalServerError) {
			return
		}
		files = append(files, fn)
	}

	for _, fi := range files {
		fn := fs.NewDirectory(sPath).UniqueName(fi)
		err = os.Rename(path.Join(h.temporaryDirectory, fi), path.Join(sPath, fn))
		if err != nil {
			log.Println(err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) storeTempFile(fileName string, part *multipart.Part) (string, error) {
	fileName = fs.NewDirectory(h.temporaryDirectory).UniqueName(fileName)
	f, err := os.Create(path.Join(h.temporaryDirectory, fileName))
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err = io.Copy(f, part); err != nil {
		return "", err
	}

	return fileName, nil
}

func (h *Handlers) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.FormValue("fileName")
	if fileName == "" {
		respondError(errors.New("file name was not set"), w, http.StatusBadRequest)
		return
	}

	sPath := h.contextStorage(r)
	fiList := fs.NewDirectory(sPath).List()
	if !fiList.Exist(fileName) {
		respondError(errors.New(fileName+" doesn't exist"), w, http.StatusBadRequest)
		return
	}

	err := os.Remove(path.Join(sPath, fileName))
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) ShareTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	if title == "" || body == "" {
		respondError(errors.New("title or body was not set"), w, http.StatusBadRequest)
		return
	}

	sPath := h.contextStorage(r)
	title = fs.NewDirectory(sPath).UniqueName(title)

	file, err := os.Create(path.Join(sPath, title))
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}
	defer file.Close()

	_, err = file.WriteString(body)
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RecoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
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
			log.Println(fmt.Sprintf("Storage name is empty for %s", r.URL))
			next.ServeHTTP(w, r)
			return
		}

		if h.sc.IsPublic(storageName) {
			err = h.checkStorage(storageName, false)
			if err != nil {
				respondError(err, w, http.StatusInternalServerError)
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

		err = h.checkStorage(storageName, true)
		if err != nil {
			respondError(err, w, http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) StorePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storage := h.sc.Name(r)
		if storage == "" {
			log.Printf("Invalid storage request, url = %s", r.URL)
		} else {
			context.Set(r, ContextStoragePath, h.sc.Path(storage))
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) StorePermanentPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storage := h.sc.Name(r)
		if storage == "" {
			log.Printf("Invalid storage request, url = %s", r.URL)
		} else {
			context.Set(r, ContextStoragePath, h.sc.PermanentPath(storage))
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) contextStorage(r *http.Request) string {
	return context.Get(r, ContextStoragePath).(string)
}

func (h *Handlers) createSkel(storageName string, permanent bool) error {
	createFolder := func(path string) error {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	}

	err := createFolder(h.sc.Path(storageName))
	if err != nil {
		return err
	}

	if permanent {
		err := createFolder(h.sc.PermanentPath(storageName))
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) checkStorage(storageName string, permanent bool) error {
	_, err := os.Stat(h.sc.Path(storageName))
	if err != nil {
		if os.IsNotExist(err) {
			err = h.createSkel(storageName, permanent)
		}
	}

	return err
}

func (h *Handlers) setUserCookie(w http.ResponseWriter, sessionName, sessionId string, expires int64) {
	cookie := http.Cookie{Name: sessionName, Value: sessionId, Path: "/", Expires: time.Unix(expires, 0), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func (h *Handlers) removeCookie(w http.ResponseWriter, sessionName string) {
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}

func respondError(err error, w http.ResponseWriter, httpStatusCode int) bool {
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), httpStatusCode)
		return true
	}

	return false
}
