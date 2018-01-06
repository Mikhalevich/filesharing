package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Mikhalevich/argparser"
	"github.com/Mikhalevich/filesharing/fileInfo"
)

const (
	SessionExpirePeriod       = 1 * 60 * 60 * 24 * 30 // sec
	LoginRequestMaxCount      = 3
	LoginRequestWaitingPeriod = 60 // sec
	Title                     = "Duplo"
)

var (
	params *Params
)

type Params struct {
	Host         string `json:"host,omitempty"`
	CleanTime    string `json:"time,omitempty"`
	Title        string `json:"title,omitempty"`
	RootStorage  string `json:"storage,omitempty"`
	PermanentDir string `json:"permanent_dir,omitempty"`
	TempDir      string `json:"temp_dir,omitempty"`
	AllowPrivate bool   `json:"allow_private,omitempty"`
}

func NewParams() *Params {
	return &Params{
		Host:         "",
		CleanTime:    "23:59",
		RootStorage:  "storage",
		PermanentDir: "permanent",
		TempDir:      path.Join(os.TempDir(), Title),
		AllowPrivate: true,
	}
}

func loadParams() (*Params, error) {
	host := argparser.String("host", "", "listening port and hostname")

	par := NewParams()
	p, err, gen := argparser.Parse(par)
	if err != nil {
		return nil, err
	}

	if gen {
		return nil, errors.New("Config should be autogenerated")
	}

	par = p.(*Params)

	if *host != "" {
		par.Host = *host
	}

	if par.Host == "" {
		return nil, errors.New("Invalid host name")
	}

	err = os.MkdirAll(par.RootStorage, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(par.TempDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return par, nil
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var err error
	params, err = loadParams()
	if err != nil {
		log.Println(err)
		return
	}

	t, err := time.Parse("15:04", params.CleanTime)
	if err != nil {
		log.Println(err)
		return
	}

	fileInfo.PermanentDir = params.PermanentDir
	now := time.Now()
	go fileInfo.CleanDir(params.RootStorage, params.PermanentDir,
		time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location()))

	log.Printf("Running at %s\n", params.Host)

	router := NewRouter(params.AllowPrivate)

	err = http.ListenAndServe(params.Host, router)
	if err != nil {
		log.Println(err)
	}
}
