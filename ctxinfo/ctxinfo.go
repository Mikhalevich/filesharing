package ctxinfo

import (
	"context"
	"errors"
)

type contextInfoKey string

const (
	contextUserName = contextInfoKey("contextUserName")
)

var (
	ErrNotFound = errors.New("not found")
)

func WithUserName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, contextUserName, name)
}

func UserName(ctx context.Context) (string, error) {
	v := ctx.Value(contextUserName)
	if v == nil {
		return "", ErrNotFound
	}

	name, ok := v.(string)
	if !ok {
		return "", errors.New("user name is not string")
	}

	return name, nil
}
