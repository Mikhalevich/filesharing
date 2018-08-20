package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Mikhalevich/filesharing/db"
	"github.com/Mikhalevich/filesharing/fileInfo"
	"github.com/Mikhalevich/filesharing/templates"
	"github.com/gorilla/context"
)

const (
	contextStoragePath = "storagePath"
)

type Handlers struct {
	sc StorageChecker
}

func NewHandlers(checker StorageChecker) *Handlers {
	return &Handlers{
		sc: checker,
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

	storage := db.NewStorage()
	defer storage.Close()

	sessionId, sessionExpires := newSessionParams()

	user := &db.User{
		Name:     userInfo.StorageName,
		Password: crypt(userInfo.Password),
		Sessions: []db.Session{
			db.Session{
				Id:      sessionId,
				Expires: sessionExpires,
			},
		},
	}

	err := storage.AddUser(user)
	if err != nil {
		userInfo.AddError("common", "Storage with this name already exists")
		return
	}

	err = h.checkStorage(userInfo.StorageName, true)
	if err != nil {
		userInfo.AddError("common", "Unable to create storage directory")
		return
	}

	renderTemplate = false
	h.setUserCookie(w, userInfo.StorageName, sessionId, sessionExpires)
	http.Redirect(w, r, "/"+userInfo.StorageName, http.StatusFound)
}

func (h *Handlers) isAlreadyRegistered(w http.ResponseWriter, cookies []*http.Cookie, name string) bool {
	for _, cook := range cookies {
		if cook.Name == name {
			storage := db.NewStorage()
			defer storage.Close()

			clearCook := true
			defer func() {
				if clearCook {
					h.removeCookie(w, name)
				}
			}()

			user, err := storage.UserByName(name)
			if err != nil {
				break
			}

			session, err := user.SessionById(cook.Value)
			if err != nil {
				break
			}

			if !isExpired(session.Expires) {
				clearCook = false
				return true
			}

			break
		}
	}

	return false
}

func (h *Handlers) isRequestAllowed(storageName string, host string) (bool, int64) {
	storage := db.NewStorage()
	defer storage.Close()

	loginRequest, err := storage.GetRequest(storageName, host)
	if err != nil {
		return true, 0
	}

	if loginRequest.Count >= LoginRequestMaxCount {
		timeDelta := time.Now().Unix() - loginRequest.LastRequest
		allowed := timeDelta >= LoginRequestWaitingPeriod

		if !allowed {
			return false, LoginRequestWaitingPeriod - timeDelta
		}

		storage.ResetRequestCounter(loginRequest)
	}

	return true, 0
}

func (h *Handlers) generateSession(storageName string, pwd string, userHost string) (string, int64, error) {
	storage := db.NewStorage()
	defer storage.Close()

	user, err := storage.UserByNameAndPassword(storageName, crypt(pwd))
	if err != nil {
		err = storage.AddRequest(storageName, userHost)
		if err != nil {
			log.Println("Error in add request: ", err)
		}
		return "", 0, errors.New("Invalid storage name or password")
	}

	err = storage.RemoveRequest(storageName, userHost)
	if err != nil {
		log.Println("Unable to remove request:", err)
	}

	err = storage.RemoveExpiredSessions(user.Id, time.Now().Unix())
	if err != nil {
		log.Println("Unable to remove expired sessions: ", err)
	}

	sessionId, sessionExpires := newSessionParams()
	err = storage.AddSession(user.Id, sessionId, sessionExpires)
	if err != nil {
		log.Println("Unable to update last login info", err)
		return "", 0, errors.New("Internal server error, please try again later")
	}

	return sessionId, sessionExpires, nil
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

	if h.isAlreadyRegistered(w, r.Cookies(), storageName) {
		userInfo.AddError("common", "Already registered")
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

	userHost := r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")]
	if allowed, waitPeriod := h.isRequestAllowed(storageName, userHost); !allowed {
		userInfo.AddError("common", "Request is not allowed, please wait %d seconds", waitPeriod)
		return
	}

	sessionId, sessionExpires, err := h.generateSession(storageName, userInfo.Password, userHost)
	if err != nil {
		userInfo.AddError("common", err.Error())
		return
	}

	renderTemplate = false
	h.setUserCookie(w, storageName, sessionId, sessionExpires)
	http.Redirect(w, r, fmt.Sprintf("/%s", storageName), http.StatusFound)
}

func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
	sPath := h.contextStorage(r)
	_, err := os.Stat(sPath)
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	viewTemplate := templates.NewTemplateView(Title, fileInfo.ListDir(sPath))

	err = viewTemplate.Execute(w)
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
		fn := fileInfo.UniqueName(fi, sPath)
		err = os.Rename(path.Join(params.TempDir, fi), path.Join(sPath, fn))
		if err != nil {
			log.Println(err.Error())
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) storeTempFile(fileName string, part *multipart.Part) (string, error) {
	fileName = fileInfo.UniqueName(fileName, params.TempDir)
	f, err := os.Create(path.Join(params.TempDir, fileName))
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
	fiList := fileInfo.ListDir(sPath)
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
	title = fileInfo.UniqueName(title, sPath)

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

		storage := db.NewStorage()
		defer storage.Close()

		user, err := storage.UserByName(storageName)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		err = h.checkStorage(storageName, true)
		if err != nil {
			respondError(err, w, http.StatusInternalServerError)
			return
		}

		if user.Password.IsEmpty() {
			next.ServeHTTP(w, r)
			return
		}

		authorized := false
		defer func() {
			if authorized {
				next.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, fmt.Sprintf("/login/%s", storageName), http.StatusFound)
			}
		}()

		authorized = h.checkSessionCookie(r, w, storageName, user)
	})
}

func (h *Handlers) checkSessionCookie(r *http.Request, w http.ResponseWriter, storageName string, user db.User) bool {
	cookies := r.Cookies()
	for _, cook := range cookies {
		if cook.Name == storageName {
			session, err := user.SessionById(cook.Value)
			if err != nil {
				h.removeCookie(w, storageName)
				return false
			}

			if isExpired(session.Expires) {
				h.removeCookie(w, storageName)
				return false
			}

			return true
		}
	}
	return false
}

func (h *Handlers) StorePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storage := h.sc.Name(r)
		if storage == "" {
			log.Printf("Invalid storage request, url = %s", r.URL)
		} else {
			context.Set(r, contextStoragePath, h.sc.Path(storage))
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
			context.Set(r, contextStoragePath, h.sc.PermanentPath(storage))
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) contextStorage(r *http.Request) string {
	return context.Get(r, contextStoragePath).(string)
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
