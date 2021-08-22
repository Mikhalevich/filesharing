package wrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/Mikhalevich/filesharing/handler"
	"github.com/Mikhalevich/filesharing/proto/auth"
	"github.com/Mikhalevich/filesharing/proto/types"
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

func (c *GRPCAuthServiceClient) CreateUser(user *types.User) (*types.Token, error) {
	rsp, err := c.client.Create(context.Background(), &auth.CreateUserRequest{
		User: user,
	})
	if err != nil {
		return nil, err
	}

	switch rsp.GetStatus() {
	case auth.Status_Ok:
		// break
	case auth.Status_AlreadyExist:
		return nil, handler.ErrAlreadyExist
	default:
		return nil, errors.New("invalid response")
	}

	return rsp.GetToken(), nil
}

func (c *GRPCAuthServiceClient) Auth(user *types.User) (*types.Token, error) {
	rsp, err := c.client.Auth(context.Background(), &auth.AuthUserRequest{
		User: user,
	})
	if err != nil {
		return nil, err
	}

	switch rsp.GetStatus() {
	case auth.Status_Ok:
		// break
	case auth.Status_PwdNotMatch:
		return nil, handler.ErrPwdNotMatch
	case auth.Status_NotExist:
		return nil, handler.ErrNotExist
	default:
		return nil, errors.New("invalid response")
	}
	return rsp.GetToken(), nil
}

func (c *GRPCAuthServiceClient) UserNameByToken(tokenString string) (string, error) {
	claims, err := c.decoder.Decode(tokenString)
	if err != nil {
		return "", err
	}

	return claims.User.Name, nil
}
