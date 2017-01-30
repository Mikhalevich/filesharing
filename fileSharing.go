package main

import (
	"fileSharing/fileInfo"
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

var (
	host       = flag.String("host", "127.0.0.1:8080", "listening port and hostname")
	cleanTime  = flag.String("time", "23:59", "time when storage will be clean")
	title      = "Duplo"
	storageDir = "storage"
	tempDir    = path.Join(os.TempDir(), title)
)

func init() {
	os.Mkdir(storageDir, os.ModePerm)
	os.Mkdir(tempDir, os.ModePerm)
}

func usage() {
	log.Println("usage: fileSharing -host=[host] -time [hh:mm], default host is " + *host + " and time is 23:59")

	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	router := NewRouter()

	// static resourses
	router.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("res"))))
	router.Handle("/storage/", http.StripPrefix("/storage/", http.FileServer(http.Dir(storageDir))))

	t, err := time.Parse("15:04", *cleanTime)
	if err != nil {
		log.Println(err.Error())

		usage()
	}

	now := time.Now()
	go fileInfo.CleanDir(storageDir,
		time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location()))

	log.Println("Running at " + *host)

	err = http.ListenAndServe(*host, router)
	if err != nil {
		log.Println(err.Error())
	}
}
