package main

import (
	"fileSharing/fileInfo"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	funcs     = template.FuncMap{"increment": func(i int) int { i++; return i }}
	templates = template.Must(template.New("fileSharing").Funcs(funcs).ParseFiles("res/index.html"))
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST method", http.StatusMethodNotAllowed)

		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		log.Printf(err.Error())

		http.Error(w, "reading body: "+err.Error(), http.StatusInternalServerError)

		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf(err.Error())

			http.Error(w, "reading body: "+err.Error(), http.StatusInternalServerError)

			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		fileName = fileInfo.UniqueName(fileName, storageDir)

		func() {
			f, err := os.Create(path.Join(tempDir, fileName))
			if err != nil {
				log.Println(err.Error())

				http.Error(w, "opening file: "+err.Error(), http.StatusInternalServerError)

				return
			}

			defer f.Close()

			if _, err = io.Copy(f, part); err != nil {
				log.Printf(err.Error())

				http.Error(w, "copying: "+err.Error(), http.StatusInternalServerError)

				return
			}
		}()
	}

	fil := fileInfo.ListDir(tempDir)
	for _, fi := range fil {
		err = os.Rename(fi.Path, path.Join(storageDir, fi.Name()))
		if err != nil {
			log.Println(err.Error())
		}
	}

	w.WriteHeader(http.StatusOK)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fiList := fileInfo.ListDir(storageDir)

	page := struct {
		Title        string
		FileInfoList []fileInfo.FileInfo
	}{title, fiList}

	if err := templates.ExecuteTemplate(w, "index.html", page); err != nil {
		log.Println(err)
	}
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

	fiList := fileInfo.ListDir(storageDir)

	isExist := fiList.Exist(fileName)
	if !isExist {
		err := fileName + " doesn't exist"

		log.Println(err)

		http.Error(w, err, http.StatusBadRequest)

		return
	}

	if err := os.Remove(path.Join(storageDir, fileName)); err != nil {
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

	title = fileInfo.UniqueName(title, storageDir)

	file, err := os.Create(path.Join(storageDir, title))
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
