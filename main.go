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
	"github.com/Mikhalevich/filesharing/service"
	"github.com/Mikhalevich/filesharing/wrapper"
	"github.com/asim/go-micro/v3"

	_ "github.com/asim/go-micro/plugins/broker/nats/v3"
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
	srv, err := service.New("filesharig")
	if err != nil {
		fmt.Println(err)
		return
	}

	params, err := loadParams()
	if err != nil {
		srv.Logger().Errorf("load params error: %v", err)
		return
	}

	var r *router.Router
	if err := srv.RegisterHandler(func(srv micro.Service, s service.Servicer) error {
		fsClient := file.NewFileService(params.FileServiceName, srv.Client())
		authClient := auth.NewAuthService(params.AuthServiceName, srv.Client())

		authService, err := wrapper.NewGRPCAuthServiceClient(authClient)
		if err != nil {
			return fmt.Errorf("creating auth service client error: %v", err)
		}

		filePub := micro.NewEvent("filesharing.file.event", srv.Client())
		h := handler.NewHandler(authService, wrapper.NewGRPCFileServiceClient(fsClient), s.Logger(), filePub)

		r = router.NewRouter(true, h, s.Logger())
		return nil
	}); err != nil {
		return
	}

	srv.Logger().Infof("Running params = %v", params)

	if err = http.ListenAndServe(params.Host, r.Handler()); err != nil {
		srv.Logger().Errorf("run service error: %v", err)
	}
}
