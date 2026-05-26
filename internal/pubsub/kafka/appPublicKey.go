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
	appKeyAdded   = "app_key_added"
	appKeyRemoved = "app_key_removed"
)

//easyjson:json
type AppKeyEvent struct {
	EventType string `json:"event_type"`
	PublicKey string `json:"public_key"`
	AppID     string `json:"app_id"`
}

func RunAppPublicKeyConsumer(
	ctx context.Context,
	log *slog.Logger,
	cfg config.KafkaConfig,
	appKeyStorage storage.AppPublicKeyStorage,
) {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
		"group.id":          cfg.AppPublicKeyGroupID,
		"auto.offset.reset": cfg.AppPublicKeyAutoOffsetReset,
	})

	if err != nil {
		log.Error("failed to create kafka consumer", slog.Any("error", err))
		return
	}

	err = c.SubscribeTopics([]string{cfg.AppPublicKeyTopic}, nil)
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

			event := AppKeyEvent{}

			if err := event.UnmarshalJSON(msg.Value); err != nil {
				log.Error("failed to unmarshal kafka message", slog.Any("error", err))
				continue
			}

			switch event.EventType {
			case appKeyAdded:
				if err := appKeyStorage.SaveAppPublicKey(
					ctx,
					event.AppID,
					[]byte(event.PublicKey),
				); err != nil {
					log.Error("failed to save app public key", slog.Any("error", err))
				}

			case appKeyRemoved:
				if err := appKeyStorage.DeleteAppPublicKey(
					ctx,
					event.AppID,
				); err != nil {
					log.Error("failed to delete app public key", slog.Any("error", err))
				}

			default:
				log.Warn("unknown app-key event type", slog.String("event_type", event.EventType))
			}
		} else if kerr, ok := err.(kafka.Error); ok && !kerr.IsTimeout() {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			log.Error("kafka consumer error", slog.Any("error", err))
		}
	}
}
