package redis

import (
	"context"
	"fmt"

	"github.com/markosoft2000/bike-tracker/internal/storage"
	"github.com/redis/rueidis"
)

const (
	userStatusPattern = "user_logout_status:user_id:%s:app_id:%s"
)

func getUserStatusKey(userID string, appID string) string {
	return fmt.Sprintf(userStatusPattern, userID, appID)
}

func (s *Storage) UserLogoutStatus(
	ctx context.Context,
	userID string,
	appID string,
) (bool, error) {
	const op = "storage.redis.UserLogoutStatus"

	ctxOp, OpCancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer OpCancel()

	key := getUserStatusKey(userID, appID)

	resp := s.client.Do(ctxOp, s.client.B().Get().Key(key).Build())

	// Handle "Key Not Found" specifically
	if rueidis.IsRedisNil(resp.Error()) {
		return false, fmt.Errorf("%s: app not found: %w", op, storage.ErrUserLogoutStatusNotFound)
	}

	if err := resp.Error(); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) SaveUserLogoutStatus(
	ctx context.Context,
	userID string,
	appID string,
) error {
	const op = "storage.redis.SaveUserLogoutStatus"

	ctxOp, OpCancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer OpCancel()

	key := getUserStatusKey(userID, appID)

	err := s.client.Do(ctxOp, s.client.B().Set().
		Key(key).
		Value(string("")).
		Ex(s.cfg.TokenTTL).
		Build()).Error()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteUserLogoutStatus(
	ctx context.Context,
	userID string,
	appID string,
) error {
	const op = "storage.redis.DeleteUserLogoutStatus"

	ctxOp, OpCancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer OpCancel()

	key := getUserStatusKey(userID, appID)

	resp := s.client.Do(ctxOp, s.client.B().Del().Key(key).Build())
	if err := resp.Error(); err != nil {
		return fmt.Errorf("%s: internal failure: %w", op, err)
	}

	return nil
}
