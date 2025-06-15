package redis

import (
	"context"
	"fmt"
	"orders-service/internal/domain/dto"

	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	client  *redis.Client
	channel string
}

func NewPublisher(client *redis.Client, cfg *Config) *Publisher {
	return &Publisher{
		client:  client,
		channel: cfg.Channel,
	}
}

func (p *Publisher) Publish(ctx context.Context, message *dto.SSEMessage) error {
	payload, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal sse message to json: %w", err)
	}

	return p.client.Publish(ctx, p.channel, payload).Err()
}
