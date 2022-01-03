package main

import (
	"errors"
	"fmt"

	_ "github.com/asim/go-micro/plugins/broker/nats/v3"
	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing/internal/handler"
	"github.com/Mikhalevich/filesharing/internal/router"
	"github.com/Mikhalevich/filesharing/internal/wrapper"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/proto/file"
	"github.com/Mikhalevich/filesharing/pkg/service"
)

type config struct {
	service.Config  `yaml:"service"`
	FileServiceName string `yaml:"file_service_name"`
	AuthServiceName string `yaml:"auth_service_name"`
}

func (c *config) Service() service.Config {
	return c.Config
}

func (c *config) Validate() error {
	if c.FileServiceName == "" {
		return errors.New("file_service_name is required")
	}

	if c.AuthServiceName == "" {
		return errors.New("auth_service_name is required")
	}

	return nil
}

func main() {
	var cfg config
	service.Run("filesharig", &cfg, func(srv micro.Service, s service.Servicer) error {
		fsClient := file.NewFileService(cfg.FileServiceName, srv.Client())
		authClient := auth.NewAuthService(cfg.AuthServiceName, srv.Client())

		authService, err := wrapper.NewGRPCAuthServiceClient(authClient)
		if err != nil {
			return fmt.Errorf("creating auth service client error: %v", err)
		}

		filePub := micro.NewEvent("filesharing.file.event", srv.Client())
		h := handler.NewHandler(authService, wrapper.NewGRPCFileServiceClient(fsClient), s.Logger(), filePub)

		router.MakeRoutes(s.Router(), true, h, s.Logger())

		return nil
	})
}
