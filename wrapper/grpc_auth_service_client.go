package wrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/Mikhalevich/filesharing/handler"
)

type GRPCAuthServiceClient struct {
	client  proto.AuthService
	decoder token.Decoder
}

func NewGRPCAuthServiceClient(c proto.AuthService) (*GRPCAuthServiceClient, error) {
	dec, err := token.NewRSADecoder()
	if err != nil {
		return nil, fmt.Errorf("unable to crate rsa decoder: %w", err)
	}

	return &GRPCAuthServiceClient{
		client:  c,
		decoder: dec,
	}, nil
}

func marshalUser(user *handler.User) *proto.User {
	return &proto.User{
		Name:     user.Name,
		Password: user.Pwd,
	}
}

func (c *GRPCAuthServiceClient) CreateUser(user *handler.User) (*handler.Token, error) {
	r, err := c.client.Create(context.Background(), marshalUser(user))
	if err != nil {
		return nil, err
	}

	switch r.GetStatus() {
	case proto.Status_Ok:
		// break
	case proto.Status_AlreadyExist:
		return nil, handler.ErrAlreadyExist
	default:
		return nil, errors.New("invalid response")
	}

	return &handler.Token{
		Value: r.GetToken(),
	}, nil
}

func (c *GRPCAuthServiceClient) Auth(user *handler.User) (*handler.Token, error) {
	r, err := c.client.Auth(context.Background(), marshalUser(user))
	if err != nil {
		return nil, err
	}

	switch r.GetStatus() {
	case proto.Status_Ok:
		// break
	case proto.Status_PwdNotMatch:
		return nil, handler.ErrPwdNotMatch
	case proto.Status_NotExist:
		return nil, handler.ErrNotExist
	default:
		return nil, errors.New("invalid response")
	}
	return &handler.Token{
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
