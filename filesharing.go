package main

import (
	"errors"
	"net/http"
	"os"

	fspb "github.com/Mikhalevich/file_service/proto"
	apb "github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/Mikhalevich/filesharing/handlers"
	"github.com/Mikhalevich/filesharing/router"
	"github.com/micro/go-micro/v2"
	"github.com/sirupsen/logrus"
)

type params struct {
	Host            string
	FileServiceName string
	AuthServiceName string
}

func loadParams() (*params, error) {
	var p params

	p.Host = os.Getenv("FS_HOST")
	if p.Host == "" {
		return nil, errors.New("host name is empty, please specify FS_HOST variable")
	}

	p.FileServiceName = os.Getenv("FS_FILE_SERVICE_NAME")
	if p.FileServiceName == "" {
		return nil, errors.New("file service name is empty, please specify FS_FILE_SERVICE_NAME variable")
	}

	p.AuthServiceName = os.Getenv("FS_AUTH_SERVICE_NAME")
	if p.AuthServiceName == "" {
		return nil, errors.New("auth service name is empty, please specify FS_AUTH_SERVICE_NAME variable")
	}

	return &p, nil
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

	storageChecker := router.NewPublicStorages()
	cookieSession := NewCookieSession(storageChecker, 1*60*60*24*30)

	microService := micro.NewService()
	microService.Init()
	fsClient := fspb.NewFileService(params.FileServiceName, microService.Client())

	authClient := apb.NewAuthService(params.AuthServiceName, microService.Client())

	h := handlers.NewHandlers(storageChecker, cookieSession, NewGRPCAuthServiceClient(authClient), NewGRPCFileServiceClient(fsClient), logger)
	r := router.NewRouter(true, h, logger)

	logger.Infof("Running params = %s", params)

	err = http.ListenAndServe(params.Host, r.Handler())
	if err != nil {
		logger.Error(err)
	}
}
