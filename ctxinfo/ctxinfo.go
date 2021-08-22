package ctxinfo

import (
	"context"
	"errors"
)

type contextInfoKey string

const (
	contextUserID           = contextInfoKey("contextUserID")
	contextUserName         = contextInfoKey("contextUserName")
	contextPermanentStorage = contextInfoKey("contextPermanentStorage")
	contextFileName         = contextInfoKey("contextFileName")
	contextPublicStorage    = contextInfoKey("contextPublicStorage")
)

var (
	ErrNotFound = errors.New("not found")
)

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, contextUserID, id)
}

func UserID(ctx context.Context) (int64, error) {
	v := ctx.Value(contextUserID)
	if v == nil {
		return 0, ErrNotFound
	}

	id, ok := v.(int64)
	if !ok {
		return 0, errors.New("user id is not int")
	}

	return id, nil
}

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

func WithPermanentStorage(ctx context.Context, permanent bool) context.Context {
	return context.WithValue(ctx, contextPermanentStorage, permanent)
}

func PermanentStorage(ctx context.Context) (bool, error) {
	v := ctx.Value(contextPermanentStorage)
	if v == nil {
		return false, ErrNotFound
	}

	permanent, ok := v.(bool)
	if !ok {
		return false, errors.New("permanent storage is not bool")
	}

	return permanent, nil
}

func WithFileName(ctx context.Context, fileName string) context.Context {
	return context.WithValue(ctx, contextFileName, fileName)
}

func FileName(ctx context.Context) (string, error) {
	v := ctx.Value(contextFileName)
	if v == nil {
		return "", ErrNotFound
	}

	name, ok := v.(string)
	if !ok {
		return "", errors.New("file name is not string")
	}

	return name, nil
}

func WithPublicStorage(ctx context.Context, public bool) context.Context {
	return context.WithValue(ctx, contextPublicStorage, public)
}

func PublicStorage(ctx context.Context) (bool, error) {
	v := ctx.Value(contextPublicStorage)
	if v == nil {
		return false, ErrNotFound
	}

	public, ok := v.(bool)
	if !ok {
		return false, errors.New("public storage is not bool")
	}

	return public, nil
}
