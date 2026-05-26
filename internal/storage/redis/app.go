package redis

import (
	"context"
	"fmt"

	"github.com/markosoft2000/bike-tracker/internal/storage"
	"github.com/redis/rueidis"
)

const (
	appPrefix = "app:"
)

func getAppKey(appID string) string {
	return appPrefix + appID // TODO 0 alloc
}

// AppPublicKey provides app public key
func (s *Storage) AppPublicKey(ctx context.Context, appID string) ([]byte, error) {
	const op = "storage.redis.AppPublicKey"

	ctxOp, OpCancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer OpCancel()

	key := getAppKey(appID)

	resp := s.client.Do(ctxOp, s.client.B().Get().Key(key).Build())

	// Handle "Key Not Found" specifically
	if rueidis.IsRedisNil(resp.Error()) {
		return nil, fmt.Errorf("%s: app not found: %w", op, storage.ErrAppPublicKeyNotFound)
	}

	if err := resp.Error(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data, err := resp.AsBytes()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return data, nil
}

// SaveAppPublicKey saves app public key
func (s *Storage) SaveAppPublicKey(ctx context.Context, appID string, pk []byte) error {
	const op = "storage.redis.SaveAppPublicKey"

	ctxOp, OpCancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer OpCancel()

	key := getAppKey(appID)

	err := s.client.Do(ctxOp, s.client.B().Set().
		Key(key).
		Value(string(pk)).
		Build()).Error()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteAppPublicKey deletes app public key
func (s *Storage) DeleteAppPublicKey(ctx context.Context, appID string) error {
	const op = "storage.redis.DeleteAppPublicKey"

	ctxOp, OpCancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer OpCancel()

	key := getAppKey(appID)

	resp := s.client.Do(ctxOp, s.client.B().Del().Key(key).Build())
	if err := resp.Error(); err != nil {
		return fmt.Errorf("%s: internal failure: %w", op, err)
	}

	return nil
}
