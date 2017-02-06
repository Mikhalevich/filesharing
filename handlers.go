package main

import (
	"errors"
	"fileSharing/fileInfo"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var (
	funcs     = template.FuncMap{"increment": func(i int) int { i++; return i }}
	templates = template.Must(template.New("fileSharing").Funcs(funcs).ParseFiles("res/index.html", "res/login.html", "res/register.html"))
)

type TemplateBase struct {
	Errors map[string]string
}

func (self *TemplateBase) AddError(name string, errorValue string, params ...interface{}) {
	self.Errors[name] = fmt.Sprintf(errorValue, params...)
}

type TemplatePassword struct {
	TemplateBase
	Password string
}

func NewTemplatePassword() *TemplatePassword {
	var info TemplatePassword
	info.Errors = make(map[string]string)
	return &info
}

type TemplateRegister struct {
	TemplateBase
	StorageName string
	Password    string
}

func NewTemplateRegister() *TemplateRegister {
	return &TemplateRegister{
		TemplateBase: TemplateBase{
			Errors: make(map[string]string),
		},
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/common/", http.StatusMovedPermanently)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := NewTemplateRegister()
	renderTemplate := true

	defer func() {
		if renderTemplate {
			if err := templates.ExecuteTemplate(w, "register.html", userInfo); err != nil {
				log.Println(err)
			}
		}
	}()

	if r.Method == "POST" {
		userInfo.StorageName = r.FormValue("name")
		userInfo.Password = r.FormValue("password")

		storage := NewStorage()
		defer storage.Close()

		sessionId := generateRandomId(32)
		var sessionExpires int64 = 0

		user := &User{
			Name:           userInfo.StorageName,
			Password:       crypt(userInfo.Password),
			SessionId:      sessionId,
			SessionExpires: sessionExpires,
		}

		err := storage.AddUser(user)
		if err != nil {
			userInfo.AddError("common", "User with this name exists already")
			return
		}

		err = os.Mkdir(path.Join(rootStorageDir, userInfo.StorageName), os.ModePerm)
		if err != nil {
			userInfo.AddError("common", "Unable to create user directory")
			return
		}

		renderTemplate = false
		setUserCookie(w, sessionId)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := NewTemplatePassword()
	renderTemplate := true

	defer func() {
		if renderTemplate {
			if err := templates.ExecuteTemplate(w, "login.html", userInfo); err != nil {
				log.Println(err)
			}
		}
	}()

	if r.Method == "POST" {
		storageName := storageVar(r)
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

		storage := NewStorage()
		defer storage.Close()

		userHost := r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")]
		loginRequest, err := storage.GetRequest(storageName, userHost)
		if err == nil {
			if loginRequest.Count >= 3 {
				timeDelta := time.Now().Unix() - loginRequest.LastRequest
				allowed := timeDelta >= 60

				if allowed {
					storage.ResetRequestCounter(loginRequest)
				} else {
					userInfo.AddError("common", "Request is not allowed, please wait %d seconds", 60)
					return
				}
			}
		}

		user, err := storage.UserByNameAndPassword(storageName, crypt(userInfo.Password))
		if err != nil {
			userInfo.AddError("common", "Invalid username or password")
			err = storage.AddRequest(storageName, userHost)
			if err != nil {
				log.Println("Error in add request: ", err)
			}
			return
		}

		err = storage.RemoveRequest(storageName, userHost)
		if err != nil {
			log.Println("Unable to remove request", err)
			// continue programm execution
		}

		sessionId := generateRandomId(32)
		currentTime := time.Now().Unix()
		err = storage.UpdateLoginInfo(user.Id, sessionId, currentTime+1*60)
		if err != nil {
			userInfo.AddError("common", "Internal server error, please try again later")
			log.Println("Unable to update last login info", err)
		} else {
			renderTemplate = false
			setUserCookie(w, sessionId)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}
}

func viewStorageHandler(w http.ResponseWriter, r *http.Request) {
	sPath := storagePath(storageVar(r))

	_, err := os.Stat(sPath)
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

	fiList := fileInfo.ListDir(sPath)
	page := struct {
		Title        string
		FileInfoList []fileInfo.FileInfo
	}{title, fiList}

	err = templates.ExecuteTemplate(w, "index.html", page)
	if err != nil {
		log.Println(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST method", http.StatusMethodNotAllowed)
		return
	}

	mr, err := r.MultipartReader()
	if respondError(err, w, http.StatusInternalServerError) {
		return
	}

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

		fileName = fileInfo.UniqueName(fileName, rootStorageDir)

		err = func() error {
			f, err := os.Create(path.Join(tempDir, fileName))
			if err != nil {
				return err
			}

			defer f.Close()

			if _, err = io.Copy(f, part); err != nil {
				return err
			}

			return nil
		}()

		if respondError(err, w, http.StatusInternalServerError) {
			return
		}
	}

	fil := fileInfo.ListDir(tempDir)
	for _, fi := range fil {
		err = os.Rename(path.Join(tempDir, fi.Name()), path.Join(storagePath(storageVar(r)), fi.Name()))
		if err != nil {
			log.Println(err.Error())
		}
	}

	w.WriteHeader(http.StatusOK)
}

func removeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST method", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.FormValue("fileName")
	if len(fileName) <= 0 {
		respondError(errors.New("file name was not set"), w, http.StatusBadRequest)
		return
	}

	sPath := storagePath(mux.Vars(r)["storage"])
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

func shareTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	if len(title) <= 0 || len(body) <= 0 {
		respondError(errors.New("title or body was not set"), w, http.StatusBadRequest)
		return
	}

	sPath := storagePath(storageVar(r))
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
