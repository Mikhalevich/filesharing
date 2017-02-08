package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
)

func storagePath(storageName string) string {
	return path.Join(rootStorageDir, storageName)
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

func sessionExpirationPeriodInSec() int64 {
	return time.Now().Unix() + SessionExpirePeriod
}

func setUserCookie(w http.ResponseWriter, sessionName, sessionId string, expires int64) {
	cookie := http.Cookie{Name: sessionName, Value: sessionId, Path: "/", Expires: time.Unix(expires, 0), HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func removeCookie(w http.ResponseWriter, sessionName string) {
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
}

func generateRandomId(size int) string {
	bytes := make([]byte, size)
	rand.Read(bytes)

	return base64.URLEncoding.EncodeToString(bytes)
}
