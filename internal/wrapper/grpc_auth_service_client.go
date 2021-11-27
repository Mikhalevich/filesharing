package wrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/Mikhalevich/filesharing/internal/handler"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/proto/types"
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

func (c *GRPCAuthServiceClient) AuthPublicUser(name string) (*types.Token, error) {
	rsp, err := c.client.AuthPublicUser(context.Background(), &auth.AuthPublicUserRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	return rsp.GetToken(), nil
}

func (c *GRPCAuthServiceClient) UserByToken(tokenString string) (*types.User, error) {
	claims, err := c.decoder.Decode(tokenString)
	if err != nil {
		return nil, err
	}

	return &types.User{
		Id:     claims.User.ID,
		Name:   claims.User.Name,
		Public: claims.User.Public,
	}, nil
}
