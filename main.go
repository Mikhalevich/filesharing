package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	apb "github.com/Mikhalevich/filesharing-auth-service/proto"
	fspb "github.com/Mikhalevich/filesharing-file-service/proto"
	"github.com/Mikhalevich/filesharing/handler"
	"github.com/Mikhalevich/filesharing/router"
	"github.com/Mikhalevich/filesharing/wrapper"
	"github.com/micro/go-micro/v2"
	"github.com/sirupsen/logrus"
)

type params struct {
	Host                     string
	FileServiceName          string
	AuthServiceName          string
	SessionExpirePeriodInSec int
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

	p.SessionExpirePeriodInSec = 60 * 60 * 24
	expirePeriodEnv := os.Getenv("FS_SESSION_EXPIRE_PERIOD_SEC")
	if expirePeriodEnv != "" {
		period, err := strconv.Atoi(expirePeriodEnv)
		if err != nil {
			return nil, fmt.Errorf("unable to convert expire session period to integer value expirePeriod: %s, error: %w", expirePeriodEnv, err)
		}
		p.SessionExpirePeriodInSec = period
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

	storageChecker := router.NewPublicStorages()
	cookieSession := wrapper.NewCookieSession(storageChecker, int64(params.SessionExpirePeriodInSec))

	microService := micro.NewService()
	microService.Init()
	fsClient := fspb.NewFileService(params.FileServiceName, microService.Client())
	authClient := apb.NewAuthService(params.AuthServiceName, microService.Client())

	authService, err := wrapper.NewGRPCAuthServiceClient(authClient)
	if err != nil {
		logger.Errorln(fmt.Errorf("creating auth service client error: %w", err))
		return
	}

	h := handler.NewHandler(storageChecker, cookieSession, authService, wrapper.NewGRPCFileServiceClient(fsClient), logger)
	r := router.NewRouter(true, h, logger)

	logger.Infof("Running params = %v", params)

	err = http.ListenAndServe(params.Host, r.Handler())
	if err != nil {
		logger.Errorln(err)
	}
}
