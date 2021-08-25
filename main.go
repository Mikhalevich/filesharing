package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Mikhalevich/filesharing/handler"
	"github.com/Mikhalevich/filesharing/proto/auth"
	"github.com/Mikhalevich/filesharing/proto/file"
	"github.com/Mikhalevich/filesharing/router"
	"github.com/Mikhalevich/filesharing/wrapper"
	"github.com/asim/go-micro/v3"
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
		logger.Errorln(fmt.Errorf("load params error: %w", err))
		return
	}

	microService := micro.NewService()
	microService.Init()
	fsClient := file.NewFileService(params.FileServiceName, microService.Client())
	authClient := auth.NewAuthService(params.AuthServiceName, microService.Client())

	authService, err := wrapper.NewGRPCAuthServiceClient(authClient)
	if err != nil {
		logger.Errorln(fmt.Errorf("creating auth service client error: %w", err))
		return
	}

	filePub := micro.NewEvent("filesharing.file.event", microService.Client())

	h := handler.NewHandler(authService, wrapper.NewGRPCFileServiceClient(fsClient), logger, filePub)

	r := router.NewRouter(true, h, logger)

	logger.Infof("Running params = %v", params)

	err = http.ListenAndServe(params.Host, r.Handler())
	if err != nil {
		logger.Errorln(err)
	}
}
