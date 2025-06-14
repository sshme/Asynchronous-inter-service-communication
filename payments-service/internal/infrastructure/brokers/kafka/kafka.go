package kafka

import (
	"context"
	"payments-service/internal/infrastructure/config"
	"payments-service/pkg/kafka"
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

func (c *Config) GetPaymentsEventsTopic() string {
	return c.Topics.PaymentsEvents
}

func (c *Config) GetOrdersEventsTopic() string {
	return c.Topics.OrdersEvents
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
			GroupID: "payments-service-group",
		},
	}
}

type PaymentEventService struct {
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	kafkaConfig *Config
}

func NewPaymentEventService(kafkaConfig *Config) (*PaymentEventService, error) {
	producer, err := kafka.NewProducer(kafkaConfig.Brokers)
	if err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(
		kafkaConfig.Brokers,
		kafkaConfig.Consumer.GroupID,
		[]string{kafkaConfig.Topics.OrdersEvents},
	)
	if err != nil {
		return nil, err
	}

	service := &PaymentEventService{
		producer:    producer,
		consumer:    consumer,
		kafkaConfig: kafkaConfig,
	}

	return service, nil
}

func (s *PaymentEventService) PublishPaymentCompleted(ctx context.Context, paymentData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "payment.completed",
		EventID:   uuid.Must(uuid.NewV7()).String(),
		Data:      paymentData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.PaymentsEvents, event)
}

func (s *PaymentEventService) PublishPaymentFailed(ctx context.Context, paymentData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "payment.failed",
		EventID:   uuid.Must(uuid.NewV7()).String(),
		Data:      paymentData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.PaymentsEvents, event)
}

func (s *PaymentEventService) StartConsumer(ctx context.Context) error {
	return s.consumer.Start(ctx)
}

func (s *PaymentEventService) RegisterHandler(eventType string, handler func(context.Context, kafka.Event) error) {
	s.consumer.RegisterHandler(eventType, handler)
}

func (s *PaymentEventService) Close() error {
	return s.producer.Close()
}
