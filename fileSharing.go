package main

import (
	"bytes"
	"fileSharing/fileInfo"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
)

var (
	host       = flag.String("host", "127.0.0.1:8080", "listening port and hostname")
	storageDir = "storage"
	title      = "Duplo"
)

func usage() {
	log.Println("usage: fileSharing -host=[host], default host is " + *host)

	os.Exit(1)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)

				return
			}
		}()

		fn(w, r)
	}
}

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

		buf := bytes.NewBuffer(make([]byte, 0))
		if _, err = io.Copy(buf, part); err != nil {
			log.Printf(err.Error())

			http.Error(w, "copying: "+err.Error(), http.StatusInternalServerError)

			return
		}

		f, err := os.Create(path.Join(storageDir, fileName))
		if err != nil {
			log.Println(err.Error())

			http.Error(w, "opening file: "+err.Error(), http.StatusInternalServerError)

			return
		}
		defer f.Close()

		if _, err = buf.WriteTo(f); err != nil {
			log.Printf(err.Error())

			http.Error(w, "writing: "+err.Error(), http.StatusInternalServerError)

			return
		}

		break
	}

	w.WriteHeader(http.StatusOK)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fiList := fileInfo.ListDir(storageDir)

	funcs := template.FuncMap{"increment": func(i int) int { i++; return i }}

	t, err := template.New("index.html").Funcs(funcs).ParseFiles("res/index.html")
	if err != nil {
		log.Println(err.Error())
	}

	page := struct {
		Title        string
		FileInfoList []fileInfo.FileInfo
	}{title, fiList}

	t.Execute(w, page)
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

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	http.HandleFunc("/", makeHandler(rootHandler))
	http.HandleFunc("/upload/", makeHandler(uploadHandler))
	http.HandleFunc("/remove/", makeHandler(removeHandler))
	http.HandleFunc("/shareText/", makeHandler(shareTextHandler))

	// static resourses
	http.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("res"))))
	http.Handle("/storage/", http.StripPrefix("/storage/", http.FileServer(http.Dir(storageDir))))

	go fileInfo.CleanDir(storageDir)

	log.Println("Running at " + *host)

	err := http.ListenAndServe(*host, nil)
	if err != nil {
		log.Println(err.Error())
	}
}
