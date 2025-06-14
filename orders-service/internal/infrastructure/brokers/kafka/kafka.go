package kafka

import (
	"context"
	"orders-service/internal/infrastructure/config"
	"orders-service/pkg/kafka"
	"time"

	"github.com/gofrs/uuid"
)

type Config struct {
	Brokers   []string
	Topics    Topics
	Publisher Publisher
	Consumer  Consumer
}

type Publisher struct {
	Interval   time.Duration
	BatchSize  int
	MaxRetries int
}

type Consumer struct {
	GroupID string
}

type Topics struct {
	OrdersEvents   string
	PaymentsEvents string
}

func (c *Config) GetBrokers() []string {
	return c.Brokers
}

func (c *Config) GetOrdersEventsTopic() string {
	return c.Topics.OrdersEvents
}

func (c *Config) GetPaymentsEventsTopic() string {
	return c.Topics.PaymentsEvents
}

func NewConfig(mainConfig *config.Config) *Config {
	return &Config{
		Brokers: mainConfig.Kafka.Brokers,
		Topics: Topics{
			OrdersEvents:   "orders-events",
			PaymentsEvents: "payments-events",
		},
		Publisher: Publisher{
			Interval:   mainConfig.GetPublisherInterval(),
			BatchSize:  mainConfig.GetPublisherBatchSize(),
			MaxRetries: mainConfig.GetPublisherMaxRetries(),
		},
		Consumer: Consumer{
			GroupID: "orders-service-group",
		},
	}
}

type OrderEventService struct {
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	kafkaConfig *Config
}

func NewOrderEventService(kafkaConfig *Config) (*OrderEventService, error) {
	producer, err := kafka.NewProducer(kafkaConfig.Brokers)
	if err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(
		kafkaConfig.Brokers,
		kafkaConfig.Consumer.GroupID,
		[]string{kafkaConfig.Topics.PaymentsEvents},
	)
	if err != nil {
		return nil, err
	}

	service := &OrderEventService{
		producer:    producer,
		consumer:    consumer,
		kafkaConfig: kafkaConfig,
	}

	consumer.RegisterHandler("payment.completed", service.handlePaymentCompleted)
	consumer.RegisterHandler("payment.failed", service.handlePaymentFailed)

	return service, nil
}

func (s *OrderEventService) PublishOrderCreated(ctx context.Context, orderData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "order.created",
		EventID:   uuid.Must(uuid.NewV7()).String(),
		Data:      orderData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.OrdersEvents, event)
}

func (s *OrderEventService) PublishOrderUpdated(ctx context.Context, orderData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "order.updated",
		EventID:   uuid.Must(uuid.NewV7()).String(),
		Data:      orderData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.OrdersEvents, event)
}

func (s *OrderEventService) PublishOrderCompleted(ctx context.Context, orderData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "order.completed",
		EventID:   uuid.Must(uuid.NewV7()).String(),
		Data:      orderData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.OrdersEvents, event)
}

func (s *OrderEventService) handlePaymentCompleted(ctx context.Context, event kafka.Event) error {
	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return nil
	}

	orderData := map[string]any{
		"order_id":   orderID,
		"status":     "paid",
		"payment_id": event.Data["payment_id"],
	}

	return s.PublishOrderUpdated(ctx, orderData)
}

func (s *OrderEventService) handlePaymentFailed(ctx context.Context, event kafka.Event) error {
	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return nil
	}

	orderData := map[string]any{
		"order_id": orderID,
		"status":   "payment_failed",
		"reason":   event.Data["error_message"],
	}

	return s.PublishOrderUpdated(ctx, orderData)
}

func (s *OrderEventService) StartConsumer(ctx context.Context) error {
	return s.consumer.Start(ctx)
}

func (s *OrderEventService) Close() error {
	return s.producer.Close()
}
