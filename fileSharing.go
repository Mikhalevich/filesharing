package main

import (
	"bytes"
	"fileSharing/fileInfo"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	host       = flag.String("host", "127.0.0.1:8080", "listening port and hostname")
	storageDir = "storage"
	title      = "File sharing"
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
	mr, err := r.MultipartReader()
	if err != nil {
		log.Printf(err.Error())

		http.Redirect(w, r, "/", 303)

		return
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			http.Error(w, "reading body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fileName := part.FileName()
		if fileName == "" {
			continue
		}

		if ld := fileInfo.ListDir("storage"); ld.IsExist(fileName) {
			ext := filepath.Ext(fileName)
			fileNameTpl := strings.TrimSuffix(fileName, ext) + "_%d" + ext

			count := 1
			var f func()
			f = func() {
				fileName = fmt.Sprintf(fileNameTpl, count)
				if ld.IsExist(fileName) {
					count++
					f()
				}
			}
			f()
		}

		buf := bytes.NewBuffer(make([]byte, 0))
		if _, err = io.Copy(buf, part); err != nil {
			http.Error(w, "copying: "+err.Error(), http.StatusInternalServerError)
			return
		}

		f, err := os.Create(path.Join(storageDir, fileName))
		if err != nil {
			http.Error(w, "opening file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		if _, err = buf.WriteTo(f); err != nil {
			http.Error(w, "writing: "+err.Error(), http.StatusInternalServerError)
			return
		}

		break
	}

	http.Redirect(w, r, "/", 303)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fiList := fileInfo.ListDir("storage")
	if fiList == nil {
		return
	}

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

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	http.HandleFunc("/", makeHandler(rootHandler))
	http.HandleFunc("/upload/", makeHandler(uploadHandler))

	// static resourses
	http.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("res"))))
	http.Handle("/storage/", http.StripPrefix("/storage/", http.FileServer(http.Dir("storage"))))

	go fileInfo.CleanDir("storage")

	log.Println("Running at " + *host)

	err := http.ListenAndServe(*host, nil)
	if err != nil {
		log.Println(err.Error())
	}
}
