package client

import (
	"context"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/pkg/token"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
)

type GRPCAuthServiceClient struct {
	client  auth.AuthService
	decoder token.Decoder
}

func NewGRPCAuthServiceClient(c auth.AuthService) (*GRPCAuthServiceClient, error) {
	dec, err := token.NewRSADecoder()
	if err != nil {
		return nil, fmt.Errorf("unable to crate rsa decoder: %w", err)
	}

	return &GRPCAuthServiceClient{
		client:  c,
		decoder: dec,
	}, nil
}

func (c *GRPCAuthServiceClient) Create(user *auth.User) (*auth.Token, error) {
	rsp, err := c.client.Create(context.Background(), &auth.CreateUserRequest{
		User: user,
	})
	if err != nil {
		return nil, err
	}
	return rsp.GetToken(), nil
}

func (c *GRPCAuthServiceClient) Auth(user *auth.User) (*auth.Token, error) {
	rsp, err := c.client.Auth(context.Background(), &auth.AuthUserRequest{
		User: user,
	})
	if err != nil {
		return nil, err
	}
	return rsp.GetToken(), nil
}

func (c *GRPCAuthServiceClient) AuthPublicUser(name string) (*auth.Token, error) {
	rsp, err := c.client.AuthPublicUser(context.Background(), &auth.AuthPublicUserRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	return rsp.GetToken(), nil
}

func (c *GRPCAuthServiceClient) UserByToken(tokenString string) (*auth.User, error) {
	claims, err := c.decoder.Decode(tokenString)
	if err != nil {
		return nil, err
	}

	return &auth.User{
		Id:     claims.User.ID,
		Name:   claims.User.Name,
		Public: claims.User.Public,
	}, nil
}
