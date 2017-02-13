package main

import (
	"errors"
	"fileSharing/fileInfo"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

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

		if userInfo.StorageName == "" {
			userInfo.AddError("name", "Please specify storage name")
		}

		if userInfo.Password == "" {
			userInfo.AddError("password", "Please enter password")
		}

		if len(userInfo.Errors) > 0 {
			return
		}

		storage := NewStorage()
		defer storage.Close()

		sessionId, sessionExpires := newSessionParams()

		user := &User{
			Name:     userInfo.StorageName,
			Password: crypt(userInfo.Password),
			Sessions: []Session{
				Session{
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

		err = os.Mkdir(path.Join(rootStorageDir, userInfo.StorageName), os.ModePerm)
		if err != nil {
			if !os.IsExist(err) {
				userInfo.AddError("common", "Unable to create storage directory")
				return
			}
		}

		renderTemplate = false
		setUserCookie(w, userInfo.StorageName, sessionId, sessionExpires)
		http.Redirect(w, r, "/"+userInfo.StorageName, http.StatusFound)
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

	storageName := storageVar(r)

	cookies := r.Cookies()
	for _, cook := range cookies {
		if cook.Name == storageName {

			storage := NewStorage()
			defer storage.Close()

			user, err := storage.UserByName(storageName)
			if err != nil {
				removeCookie(w, storageName)
				break
			}
			session, err := user.SessionById(cook.Value)
			if err != nil {
				removeCookie(w, storageName)
				break
			}

			if !isExpired(session.Expires) {
				userInfo.AddError("common", "Already registered")
				return
			}
		}
	}

	if r.Method == "POST" {
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
			if loginRequest.Count >= LoginRequestMaxCount {
				timeDelta := time.Now().Unix() - loginRequest.LastRequest
				allowed := timeDelta >= LoginRequestWaitingPeriod

				if allowed {
					storage.ResetRequestCounter(loginRequest)
				} else {
					userInfo.AddError("common", "Request is not allowed, please wait %d seconds", LoginRequestWaitingPeriod-timeDelta)
					return
				}
			}
		}

		user, err := storage.UserByNameAndPassword(storageName, crypt(userInfo.Password))
		if err != nil {
			userInfo.AddError("common", "Invalid storage name or password")
			err = storage.AddRequest(storageName, userHost)
			if err != nil {
				log.Println("Error in add request: ", err)
			}
			return
		}

		err = storage.RemoveRequest(storageName, userHost)
		if err != nil {
			log.Println("Unable to remove request:", err)
			// continue programm execution
		}

		err = storage.RemoveExpiredSessions(user.Id, time.Now().Unix())
		if err != nil {
			log.Println("Unable to remove expired sessions: ", err)
		}

		sessionId, sessionExpires := newSessionParams()
		err = storage.AddSession(user.Id, sessionId, sessionExpires)
		if err != nil {
			userInfo.AddError("common", "Internal server error, please try again later")
			log.Println("Unable to update last login info", err)
		} else {
			renderTemplate = false
			setUserCookie(w, storageName, sessionId, sessionExpires)
			http.Redirect(w, r, "/"+storageName, http.StatusFound)
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

	err = templates.ExecuteTemplate(w, "view.html", page)
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
