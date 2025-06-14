package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"payments-service/internal/domain/outbox"
	"payments-service/internal/interfaces/repository"
	"payments-service/pkg/kafka"
)

type OutboxPublisher struct {
	outboxRepo  repository.OutboxRepository
	producer    *kafka.Producer
	kafkaConfig *Config
	ticker      *time.Ticker
	done        chan bool
}

func NewOutboxPublisher(
	outboxRepo repository.OutboxRepository,
	kafkaConfig *Config,
) (*OutboxPublisher, error) {
	producer, err := kafka.NewProducer(kafkaConfig.GetBrokers())
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &OutboxPublisher{
		outboxRepo:  outboxRepo,
		producer:    producer,
		kafkaConfig: kafkaConfig,
		done:        make(chan bool),
	}, nil
}

func (p *OutboxPublisher) Start(ctx context.Context) {
	p.ticker = time.NewTicker(p.kafkaConfig.Publisher.Interval)

	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.processPendingMessages(ctx)
				p.processFailedMessages(ctx)
			case <-p.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *OutboxPublisher) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.done)
	if p.producer != nil {
		p.producer.Close()
	}
}

func (p *OutboxPublisher) processPendingMessages(ctx context.Context) {
	messages, err := p.outboxRepo.GetPendingMessages(ctx, p.kafkaConfig.Publisher.BatchSize)
	if err != nil {
		log.Printf("Error getting pending messages: %v", err)
		return
	}

	for _, message := range messages {
		err := p.publishMessage(ctx, message)
		if err != nil {
			log.Printf("Error publishing message %s: %v", message.ID, err)
			p.outboxRepo.MarkAsFailed(ctx, message.ID)
		} else {
			p.outboxRepo.MarkAsSent(ctx, message.ID)
		}
	}
}

func (p *OutboxPublisher) processFailedMessages(ctx context.Context) {
	messages, err := p.outboxRepo.GetFailedMessages(ctx, p.kafkaConfig.Publisher.MaxRetries, p.kafkaConfig.Publisher.BatchSize)
	if err != nil {
		log.Printf("Error getting failed messages: %v", err)
		return
	}

	for _, message := range messages {
		err := p.publishMessage(ctx, message)
		if err != nil {
			log.Printf("Error retrying message %s: %v", message.ID, err)
			p.outboxRepo.MarkAsFailed(ctx, message.ID)
		} else {
			p.outboxRepo.MarkAsSent(ctx, message.ID)
		}
	}
}

func (p *OutboxPublisher) publishMessage(ctx context.Context, message *outbox.OutboxMessage) error {
	var topic string
	switch message.EventType {
	case "payment.completed", "payment.failed":
		topic = p.kafkaConfig.GetPaymentsEventsTopic()
	default:
		return fmt.Errorf("unknown event type: %s", message.EventType)
	}

	var payloadMap map[string]any
	if err := json.Unmarshal(message.Payload, &payloadMap); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	kafkaEvent := kafka.Event{
		EventType: message.EventType,
		EventID:   message.ID,
		Data:      payloadMap,
		Timestamp: message.CreatedAt.Unix(),
	}

	return p.producer.PublishEvent(ctx, topic, kafkaEvent)
}
