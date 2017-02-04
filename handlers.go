package main

import (
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
	storage := mux.Vars(r)["storage"]

	sPath := storagePath(storage)

	err := os.Mkdir(sPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	if err != nil {
		log.Println(err)
		http.Error(w, "reading body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println(err)
			http.Error(w, "reading body: "+err.Error(), http.StatusInternalServerError)
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

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fil := fileInfo.ListDir(tempDir)
	for _, fi := range fil {
		err = os.Rename(path.Join(tempDir, fi.Name()), path.Join(rootStorageDir, mux.Vars(r)["storage"], fi.Name()))
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
	if len(fileName) == 0 {
		err := "file name was not set"

		log.Println(err)

		http.Error(w, err, http.StatusBadRequest)

		return

	}

	fiList := fileInfo.ListDir(rootStorageDir)

	isExist := fiList.Exist(fileName)
	if !isExist {
		err := fileName + " doesn't exist"

		log.Println(err)

		http.Error(w, err, http.StatusBadRequest)

		return
	}

	if err := os.Remove(path.Join(rootStorageDir, fileName)); err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func shareTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST method", http.StatusMethodNotAllowed)

		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	if len(title) == 0 || len(body) == 0 {
		err := "title or body was not set"

		log.Println(err)

		http.Error(w, err, http.StatusBadRequest)

		return
	}

	title = fileInfo.UniqueName(title, rootStorageDir)

	file, err := os.Create(path.Join(rootStorageDir, title))
	if err != nil {
		log.Println(err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	defer file.Close()

	_, err = file.WriteString(body)
	if err != nil {
		log.Println(err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
