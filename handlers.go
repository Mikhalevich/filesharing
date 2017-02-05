package main

import (
	"errors"
	"fileSharing/fileInfo"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
)

var (
	funcs     = template.FuncMap{"increment": func(i int) int { i++; return i }}
	templates = template.Must(template.New("fileSharing").Funcs(funcs).ParseFiles("res/index.html"))
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/common/", http.StatusMovedPermanently)
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
