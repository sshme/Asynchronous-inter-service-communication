package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"payments-service/internal/domain/inbox"
	"payments-service/internal/interfaces/repository"
	"payments-service/pkg/kafka"
)

type InboxProcessor struct {
	inboxRepo   repository.InboxRepository
	consumer    *kafka.Consumer
	kafkaConfig *Config
	ticker      *time.Ticker
	done        chan bool
	handlers    map[string]func(context.Context, *inbox.InboxMessage) error
}

func NewInboxProcessor(
	inboxRepo repository.InboxRepository,
	kafkaConfig *Config,
) (*InboxProcessor, error) {
	consumer, err := kafka.NewConsumer(
		kafkaConfig.GetBrokers(),
		kafkaConfig.Consumer.GroupID,
		[]string{kafkaConfig.GetOrdersEventsTopic()},
	)
	if err != nil {
		return nil, err
	}

	processor := &InboxProcessor{
		inboxRepo:   inboxRepo,
		consumer:    consumer,
		kafkaConfig: kafkaConfig,
		done:        make(chan bool),
		handlers:    make(map[string]func(context.Context, *inbox.InboxMessage) error),
	}

	consumer.RegisterHandler("order.created", processor.handleKafkaEvent)

	return processor, nil
}

func (p *InboxProcessor) RegisterHandler(eventType string, handler func(context.Context, *inbox.InboxMessage) error) {
	p.handlers[eventType] = handler
}

func (p *InboxProcessor) Start(ctx context.Context) {
	go func() {
		if err := p.consumer.Start(ctx); err != nil {
			log.Printf("Error starting Kafka consumer: %v", err)
		}
	}()

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

func (p *InboxProcessor) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.done)
}

func (p *InboxProcessor) handleKafkaEvent(ctx context.Context, event kafka.Event) error {
	processed, err := p.inboxRepo.IsEventProcessed(ctx, event.EventID)
	if err != nil {
		log.Printf("Error checking if event is processed: %v", err)
		return err
	}

	if processed {
		log.Printf("Event %s already processed, skipping", event.EventID)
		return nil
	}

	payload, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return err
	}

	inboxMessage, err := inbox.NewInboxMessage(event.EventID, event.EventType, payload)
	if err != nil {
		log.Printf("Error creating inbox message: %v", err)
		return err
	}

	if err := p.inboxRepo.Store(ctx, inboxMessage); err != nil {
		log.Printf("Error storing inbox message: %v", err)
		return err
	}

	log.Printf("Stored inbox message for event %s", event.EventID)
	return nil
}

func (p *InboxProcessor) processPendingMessages(ctx context.Context) {
	messages, err := p.inboxRepo.GetPendingMessages(ctx, p.kafkaConfig.Publisher.BatchSize)
	if err != nil {
		log.Printf("Error getting pending inbox messages: %v", err)
		return
	}

	for _, message := range messages {
		p.processMessage(ctx, message)
	}
}

func (p *InboxProcessor) processFailedMessages(ctx context.Context) {
	maxAge := 120 * time.Second
	messages, err := p.inboxRepo.GetFailedMessages(ctx, maxAge, p.kafkaConfig.Publisher.BatchSize)
	if err != nil {
		log.Printf("Error getting failed inbox messages: %v", err)
		return
	}

	for _, message := range messages {
		p.processMessage(ctx, message)
	}
}

func (p *InboxProcessor) processMessage(ctx context.Context, message *inbox.InboxMessage) {
	handler, exists := p.handlers[message.EventType]
	if !exists {
		log.Printf("No handler found for event type: %s", message.EventType)
		p.inboxRepo.MarkAsFailed(ctx, message.ID)
		return
	}

	err := handler(ctx, message)
	if err != nil {
		log.Printf("Error processing inbox message %s: %v", message.ID, err)
		p.inboxRepo.MarkAsFailed(ctx, message.ID)
	} else {
		p.inboxRepo.MarkAsProcessed(ctx, message.ID)
		log.Printf("Successfully processed inbox message %s", message.ID)
	}
}
