package kafka

//go:generate go run github.com/mailru/easyjson/easyjson -all .

import (
	"context"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/markosoft2000/bike-tracker/internal/config"
	"github.com/markosoft2000/bike-tracker/internal/storage"
)

const (
	userStatusLogin  = "login"
	userStatusLogout = "logout"
)

//easyjson:json
type UserActivityEvent struct {
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	AppID     string    `json:"app_id"`
}

func RunTokenBlackListConsumer(
	ctx context.Context,
	log *slog.Logger,
	cfg config.KafkaConfig,
	userLogoutStatusStorage storage.UserLogoutStatusStorage,
) {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
		"group.id":          cfg.UserActivityGroupID,
		"auto.offset.reset": cfg.UserActivityAutoOffsetReset,
	})

	if err != nil {
		log.Error("failed to create kafka consumer", slog.Any("error", err))
		return
	}

	err = c.SubscribeTopics([]string{cfg.UserActivityTopic}, nil)
	if err != nil {
		log.Error("failed to subscribe to kafka topic", slog.Any("error", err))
		return
	}

	defer c.Close()

	run := true

	for run {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			log.Debug("received kafka message",
				slog.String("topic_partition", msg.TopicPartition.String()),
				slog.Int("payload_len", len(msg.Value)),
				slog.String("payload", string(msg.Value)),
			)

			event := UserActivityEvent{}
			if err := event.UnmarshalJSON(msg.Value); err != nil {
				log.Error("failed to unmarshal kafka message", slog.Any("error", err))
				continue
			}

			switch event.EventType {
			case userStatusLogin:
				if err := userLogoutStatusStorage.DeleteUserLogoutStatus(
					ctx,
					event.UserID,
					event.AppID,
				); err != nil && err != storage.ErrUserLogoutStatusNotFound {
					log.Error("failed to delete user logout status", slog.Any("error", err))
				}

			case userStatusLogout:
				if err := userLogoutStatusStorage.SaveUserLogoutStatus(
					ctx,
					event.UserID,
					event.AppID,
				); err != nil {
					log.Error("failed to save user logout status", slog.Any("error", err))
				}

			default:
				log.Warn("unknown user activity event type", slog.String("event_type", event.EventType))
			}
		} else if kerr, ok := err.(kafka.Error); ok && !kerr.IsTimeout() {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			log.Error("kafka consumer error", slog.Any("error", err))
		}
	}
}
