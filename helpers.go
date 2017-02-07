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

func setUserCookie(w http.ResponseWriter, sessionId string) {
	expire := time.Now().Add(SessionExpirePeriod * time.Second)
	cookie := http.Cookie{Name: SessionName, Value: sessionId, Expires: expire, HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func generateRandomId(size int) string {
	bytes := make([]byte, size)
	rand.Read(bytes)

	return base64.URLEncoding.EncodeToString(bytes)
}
