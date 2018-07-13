package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/http"
	"time"
)

func respondError(err error, w http.ResponseWriter, httpStatusCode int) bool {
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), httpStatusCode)
		return true
	}

	return false
}

func crypt(password string) [sha1.Size]byte {
	if password != "" {
		return sha1.Sum([]byte(password))
	}
	return [sha1.Size]byte{}
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
