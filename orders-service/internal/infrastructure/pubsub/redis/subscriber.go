package redis

import (
	"context"
	"log"
	"orders-service/internal/domain/dto"

	"github.com/redis/go-redis/v9"
)

type MessageHandler func(message *dto.SSEMessage)

type Subscriber struct {
	client  *redis.Client
	channel string
}

func NewSubscriber(client *redis.Client, cfg *Config) *Subscriber {
	return &Subscriber{
		client:  client,
		channel: cfg.Channel,
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, handler MessageHandler) {
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			sseMsg, err := dto.FromJSON([]byte(msg.Payload))
			if err != nil {
				log.Printf("Failed to unmarshal SSE message from Redis: %v", err)
				continue
			}
			handler(sseMsg)
		}
	}
}
