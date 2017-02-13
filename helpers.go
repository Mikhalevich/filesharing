package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
)

func storagePath(storageName string) string {
	return path.Join(rootStorageDir, storageName)
}

func checkStorage(storageName string) error {
	_, err := os.Stat(storagePath(storageName))
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(storagePath(storageName), os.ModePerm)
		}
	}

	return err
}

func respondError(err error, w http.ResponseWriter, httpStatusCode int) bool {
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), httpStatusCode)
		return true
	}

	return false
}

func storageVar(r *http.Request) string {
	return mux.Vars(r)["storage"]
}

func crypt(password string) [sha1.Size]byte {
	return sha1.Sum([]byte(password))
}

func newSessionParams() (string, int64) {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	id := base64.URLEncoding.EncodeToString(bytes)

	expire := time.Now().Unix() + SessionExpirePeriod

	return id, expire
}

func isExpired(sessionTime int64) bool {
	return sessionTime < time.Now().Unix()
}

func setUserCookie(w http.ResponseWriter, sessionName, sessionId string, expires int64) {
	cookie := http.Cookie{Name: sessionName, Value: sessionId, Path: "/", Expires: time.Unix(expires, 0), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func removeCookie(w http.ResponseWriter, sessionName string) {
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}
