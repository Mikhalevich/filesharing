package main

import (
	"errors"
	"fmt"
	"os"

	_ "github.com/asim/go-micro/plugins/broker/nats/v3"
	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing/internal/handler"
	"github.com/Mikhalevich/filesharing/internal/router"
	"github.com/Mikhalevich/filesharing/internal/wrapper"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/proto/file"
	"github.com/Mikhalevich/filesharing/pkg/service"
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
	params, err := loadParams()
	if err != nil {
		fmt.Printf("failed to load parameters: %v\n", err)
		return
	}

	service.Run("filesharig", params.Host, func(srv micro.Service, s service.Servicer) error {
		fsClient := file.NewFileService(params.FileServiceName, srv.Client())
		authClient := auth.NewAuthService(params.AuthServiceName, srv.Client())

		authService, err := wrapper.NewGRPCAuthServiceClient(authClient)
		if err != nil {
			return fmt.Errorf("creating auth service client error: %v", err)
		}

		filePub := micro.NewEvent("filesharing.file.event", srv.Client())
		h := handler.NewHandler(authService, wrapper.NewGRPCFileServiceClient(fsClient), s.Logger(), filePub)

		router.NewRouter(true, h, s.Logger()).MakeRoutes(s.Router())

		return nil
	})
}
