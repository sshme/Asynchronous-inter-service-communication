package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer      sarama.ConsumerGroup
	topics        []string
	eventHandlers map[string]EventHandler
}

type EventHandler func(ctx context.Context, event Event) error

type ConsumerGroupHandler struct {
	eventHandlers map[string]EventHandler
}

func NewConsumer(brokers []string, groupID string, topics []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second   // 10 seconds
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second // 3 seconds

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		consumer:      consumer,
		topics:        topics,
		eventHandlers: make(map[string]EventHandler),
	}, nil
}

func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) {
	c.eventHandlers[eventType] = handler
}

func (c *Consumer) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	handler := &ConsumerGroupHandler{
		eventHandlers: c.eventHandlers,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := c.consumer.Consume(ctx, c.topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Handle graceful shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Consumer context cancelled")
	case <-sigterm:
		log.Println("Termination signal received")
		cancel()
	}

	wg.Wait()
	return c.consumer.Close()
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var event Event
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			if handler, exists := h.eventHandlers[event.EventType]; exists {
				if err := handler(context.Background(), event); err != nil {
					log.Printf("Failed to handle event %s: %v", event.EventType, err)
				} else {
					log.Printf("Successfully processed event %s with ID %s", event.EventType, event.EventID)
				}
			} else {
				log.Printf("No handler registered for event type: %s", event.EventType)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
