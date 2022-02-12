package service

import (
	"fmt"

	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/proto/file"
	"github.com/Mikhalevich/filesharing/pkg/service/internal/client"
)

type ClientManager struct {
	auth *client.GRPCAuthServiceClient
	file *client.GRPCFileServiceClient
}

func newClientMananger(srv micro.Service, authName, fileName string) (*ClientManager, error) {
	c := ClientManager{}

	if authName != "" {
		grpcAuth, err := client.NewGRPCAuthServiceClient(auth.NewAuthService(authName, srv.Client()))
		if err != nil {
			return nil, fmt.Errorf("auth service error: %w", err)
		}
		c.auth = grpcAuth
	}

	if fileName != "" {
		c.file = client.NewGRPCFileServiceClient(file.NewFileService(fileName, srv.Client()))
	}

	return &c, nil
}

func (cm *ClientManager) Auth() *client.GRPCAuthServiceClient {
	return cm.auth
}

func (cm *ClientManager) File() *client.GRPCFileServiceClient {
	return cm.file
}
