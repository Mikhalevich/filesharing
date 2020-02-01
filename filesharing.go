package main

import (
	"errors"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Mikhalevich/argparser"
	"github.com/Mikhalevich/filesharing/fs"
	"github.com/Mikhalevich/filesharing/handlers"
	"github.com/Mikhalevich/filesharing/router"
	"github.com/Mikhalevich/goauth"
	"github.com/Mikhalevich/goauth/db"
	"github.com/Mikhalevich/goauth/email"
	"github.com/sirupsen/logrus"
)

type dbParams struct {
	Host string `json:"host,omitempty"`
}

type params struct {
	Host         string   `json:"host,omitempty"`
	CleanTime    string   `json:"time,omitempty"`
	Title        string   `json:"title,omitempty"`
	RootStorage  string   `json:"storage,omitempty"`
	PermanentDir string   `json:"permanent_dir,omitempty"`
	TempDir      string   `json:"temp_dir,omitempty"`
	AllowPrivate bool     `json:"allow_private,omitempty"`
	DB           dbParams `json:"db,omitempty"`
}

func newParams() *params {
	return &params{
		Host:         "",
		CleanTime:    "23:59",
		RootStorage:  "storage",
		PermanentDir: "permanent",
		TempDir:      path.Join(os.TempDir(), "Duplo"),
		AllowPrivate: true,
	}
}

func loadParams() (*params, error) {
	parser := argparser.NewParser()
	host := parser.String("host", "", "listening port and hostname")

	par := newParams()
	p, err, gen := parser.Parse(par)
	if err != nil {
		return nil, err
	}

	if gen {
		return nil, errors.New("Config should be autogenerated")
	}

	par = p.(*params)

	if *host != "" {
		par.Host = *host
	}

	if par.Host == "" {
		return nil, errors.New("Invalid host name")
	}

	if par.DB.Host == "" {
		par.DB.Host = "localhost"
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

func runCleaner(cleanTime, rootPath, permanentDirectory string) error {
	t, err := time.Parse("15:04", cleanTime)
	if err != nil {
		return err
	}

	fs.PermanentDir = permanentDirectory
	cleaner := fs.NewCleaner(rootPath, permanentDirectory)
	cleaner.Run(t.Hour(), t.Minute())

	return nil
}

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	params, err := loadParams()
	if err != nil {
		logger.Error(err)
		return
	}

	err = runCleaner(params.CleanTime, params.RootStorage, params.PermanentDir)
	if err != nil {
		logger.Error(err)
		return
	}

	storageChecker := router.NewPublicStorages()

	var auth goauth.Authentifier
	if params.AllowPrivate {
		time.Sleep(time.Second * 2)
		pg, err := db.NewPostgres(params.DB.Host)
		if err != nil {
			logger.Error(err)
			return
		}
		defer pg.Close()

		es := &email.GomailSender{
			Host:     "smtp.gmail.com",
			Port:     587,
			From:     "",
			Password: "",
		}
		auth = goauth.NewAuthentificator(pg, pg, NewCookieSession(storageChecker, 1*60*60*24*30), es)
	} else {
		auth = goauth.NewNullAuthentificator()
	}

	h := handlers.NewHandlers(storageChecker, auth, fs.NewFileStorage(), logger, params.RootStorage, params.PermanentDir, params.TempDir)
	r := router.NewRouter(params.AllowPrivate, h, logger)

	logger.Infof("Running at %s", params.Host)

	err = http.ListenAndServe(params.Host, r.Handler())
	if err != nil {
		logger.Error(err)
	}
}
