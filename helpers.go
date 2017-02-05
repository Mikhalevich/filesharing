package main

import (
	"log"
	"net/http"
	"path"

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
