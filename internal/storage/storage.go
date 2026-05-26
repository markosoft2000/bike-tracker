package storage

import (
	"context"
	"errors"
)

var (
	ErrAppPublicKeyNotFound     = errors.New("app public key not found")
	ErrUserLogoutStatusNotFound = errors.New("user logout status not found")
)

type Storage interface {
	Stop()
}

type AppPublicKeyStorage interface {
	Storage

	AppPublicKey(ctx context.Context, appID string) ([]byte, error)
	SaveAppPublicKey(ctx context.Context, appID string, pk []byte) error
	DeleteAppPublicKey(ctx context.Context, appID string) error
}

type UserLogoutStatusStorage interface {
	Storage

	UserLogoutStatus(
		ctx context.Context,
		userID string,
		appID string,
	) (bool, error)

	SaveUserLogoutStatus(
		ctx context.Context,
		userID string,
		appID string,
	) error

	DeleteUserLogoutStatus(
		ctx context.Context,
		userID string,
		appID string,
	) error
}
