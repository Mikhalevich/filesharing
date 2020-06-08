package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/Mikhalevich/filesharing/handlers"
)

type GRPCAuthServiceClient struct {
	client  proto.AuthService
	decoder token.Decoder
}

func NewGRPCAuthServiceClient(c proto.AuthService, publicCert string) (*GRPCAuthServiceClient, error) {
	publicKey, err := token.LoadCertFromFile(publicCert)
	if err != nil {
		return nil, fmt.Errorf("unable to load public certificate: %w", err)
	}

	dec, err := token.NewRSADecoder(publicKey)
	if err != nil {
		return nil, fmt.Errorf("unable to crate rsa decoder: %w", err)
	}

	return &GRPCAuthServiceClient{
		client:  c,
		decoder: dec,
	}, nil
}

func marshalUser(user *handlers.User) *proto.User {
	return &proto.User{
		Name:     user.Name,
		Password: user.Pwd,
	}
}

func (c *GRPCAuthServiceClient) CreateUser(user *handlers.User) (*handlers.Token, error) {
	r, err := c.client.Create(context.Background(), marshalUser(user))
	if err != nil {
		return nil, err
	}

	if r.GetStatus() != proto.Status_Ok {
		return nil, errors.New("invalid response")
	}

	return &handlers.Token{
		Value: r.GetToken(),
	}, nil
}

func (c *GRPCAuthServiceClient) Auth(user *handlers.User) (*handlers.Token, error) {
	r, err := c.client.Auth(context.Background(), marshalUser(user))
	if err != nil {
		return nil, err
	}

	if r.GetStatus() != proto.Status_Ok {
		return nil, errors.New("invalid response")
	}

	return &handlers.Token{
		Value: r.GetToken(),
	}, nil
}

func (c *GRPCAuthServiceClient) UserNameByToken(tokenString string) (string, error) {
	claims, err := c.decoder.Decode(tokenString)
	if err != nil {
		return "", err
	}

	return claims.User.Name, nil
}
