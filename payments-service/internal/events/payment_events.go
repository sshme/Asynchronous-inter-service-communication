package events

import (
	"context"
	"time"

	"payments-service/config"
	"payments-service/pkg/kafka"

	"github.com/gofrs/uuid"
)

type PaymentEventService struct {
	producer    *kafka.Producer
	consumer    *kafka.Consumer
	kafkaConfig *config.KafkaConfig
}

func NewPaymentEventService(kafkaConfig *config.KafkaConfig) (*PaymentEventService, error) {
	producer, err := kafka.NewProducer(kafkaConfig.Brokers)
	if err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(
		kafkaConfig.Brokers,
		"payments-service-group",
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

	consumer.RegisterHandler("order.created", service.handleOrderCreated)

	return service, nil
}

func (s *PaymentEventService) PublishPaymentRequested(ctx context.Context, paymentData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "payment.requested",
		EventID:   generateEventID(),
		Data:      paymentData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.PaymentsEvents, event)
}

func (s *PaymentEventService) PublishPaymentCompleted(ctx context.Context, paymentData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "payment.completed",
		EventID:   generateEventID(),
		Data:      paymentData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.PaymentsEvents, event)
}

func (s *PaymentEventService) PublishPaymentFailed(ctx context.Context, paymentData map[string]interface{}) error {
	event := kafka.Event{
		EventType: "payment.failed",
		EventID:   generateEventID(),
		Data:      paymentData,
		Timestamp: time.Now().Unix(),
	}

	return s.producer.PublishEvent(ctx, s.kafkaConfig.Topics.PaymentsEvents, event)
}

func (s *PaymentEventService) handleOrderCreated(ctx context.Context, event kafka.Event) error {
	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return nil
	}

	amount, ok := event.Data["amount"].(float64)
	if !ok {
		return nil
	}

	paymentData := map[string]any{
		"order_id":   orderID,
		"amount":     amount,
		"currency":   event.Data["currency"],
		"user_id":    event.Data["user_id"],
		"payment_id": generateEventID(),
		"status":     "requested",
	}

	if err := s.PublishPaymentRequested(ctx, paymentData); err != nil {
		return err
	}

	go s.processPayment(ctx, paymentData)

	return nil
}

func (s *PaymentEventService) processPayment(ctx context.Context, paymentData map[string]interface{}) {
	time.Sleep(time.Second * 3)

	success := time.Now().UnixNano()%10 < 8

	if success {
		paymentData["status"] = "completed"
		paymentData["transaction_id"] = generateEventID()
		s.PublishPaymentCompleted(ctx, paymentData)
	} else {
		paymentData["status"] = "failed"
		paymentData["error_message"] = "Payment processing failed"
		s.PublishPaymentFailed(ctx, paymentData)
	}
}

func (s *PaymentEventService) StartConsumer(ctx context.Context) error {
	return s.consumer.Start(ctx)
}

func (s *PaymentEventService) Close() error {
	return s.producer.Close()
}

func generateEventID() string {
	v7, err := uuid.NewV7()
	if err != nil {
		return generateEventID()
	}

	return v7.String()
}
