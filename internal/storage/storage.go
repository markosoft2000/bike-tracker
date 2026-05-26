package storage

import (
	"context"
	"errors"
)

var (
	ErrAppPublicKeyNotFound = errors.New("app public key not found")
)

type AppPublicKeyStorage interface {
	Stop()

	AppPublicKey(ctx context.Context, appID string) ([]byte, error)
	SaveAppPublicKey(ctx context.Context, appID string, pk []byte) error
	DeleteApp(ctx context.Context, appID string) error
}
