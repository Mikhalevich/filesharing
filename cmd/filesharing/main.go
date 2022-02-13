package main

import (
	"errors"

	_ "github.com/asim/go-micro/plugins/broker/nats/v3"
	"github.com/asim/go-micro/v3/server"

	"github.com/Mikhalevich/filesharing/internal/handler"
	"github.com/Mikhalevich/filesharing/internal/router"
	"github.com/Mikhalevich/filesharing/pkg/service"
)

type config struct {
	service.Config `yaml:"service"`
}

func (c *config) Service() service.Config {
	return c.Config
}

func (c *config) Validate() error {
	if c.Config.FileServiceName == "" {
		return errors.New("file_service_name is required")
	}

	if c.Config.AuthServiceName == "" {
		return errors.New("auth_service_name is required")
	}

	return nil
}

func main() {
	var cfg config
	service.Run("filesharig", &cfg, func(srv server.Server, s service.Servicer) error {
		filePub := s.Publisher().New("filesharing.file.event")
		h := handler.NewHandler(s.ClientManager().Auth(), s.ClientManager().File(), s.Logger(), filePub)

		router.MakeRoutes(s.Router(), true, h, s.Logger())

		return nil
	})
}
