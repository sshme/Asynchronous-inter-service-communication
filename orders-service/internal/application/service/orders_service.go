package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"orders-service/internal/domain/dto"
	"orders-service/internal/domain/inbox"
	"orders-service/internal/domain/orders"
	"orders-service/internal/domain/outbox"
	"orders-service/internal/infrastructure/pubsub/redis"
	"orders-service/internal/interfaces/repository"
	"orders-service/pkg/random"
)

type OrdersService struct {
	randomGenerator  random.Generator
	ordersRepository repository.OrdersRepository
	outboxRepository repository.OutboxRepository
	redisPublisher   *redis.Publisher
	db               *sql.DB
}

func NewOrdersService(
	ordersRepository repository.OrdersRepository,
	outboxRepository repository.OutboxRepository,
	randomGenerator random.Generator,
	redisPublisher *redis.Publisher,
	db *sql.DB,
) *OrdersService {
	return &OrdersService{
		ordersRepository: ordersRepository,
		outboxRepository: outboxRepository,
		randomGenerator:  randomGenerator,
		redisPublisher:   redisPublisher,
		db:               db,
	}
}

func (s *OrdersService) publishOrderUpdate(ctx context.Context, order *orders.Order) {
	sseMessage := &dto.SSEMessage{
		UserID:  order.UserID,
		Event:   "order-update",
		Payload: order,
	}
	if err := s.redisPublisher.Publish(ctx, sseMessage); err != nil {
		log.Printf("Failed to publish order update to Redis: %v", err)
	}
}

func (s *OrdersService) CreateOrder(ctx context.Context, userID string) (*orders.Order, error) {
	amount, err := s.randomGenerator.FloatWithPrecision(100, 1000)
	if err != nil {
		return nil, err
	}

	order, err := orders.NewOrder(userID, amount)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = s.ordersRepository.StoreWithTx(ctx, tx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to store order: %w", err)
	}

	orderCreatedEvent := outbox.OrderCreatedEvent{
		OrderID:  order.ID,
		UserID:   order.UserID,
		Amount:   order.Amount,
		Currency: order.Currency,
	}

	payload, err := json.Marshal(orderCreatedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order created event: %w", err)
	}

	outboxMessage, err := outbox.NewOutboxMessage("order.created", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create outbox message: %w", err)
	}

	err = s.outboxRepository.StoreMessage(ctx, tx, outboxMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to store outbox message: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Order created successfully: OrderID=%s, UserID=%s, Amount=%.2f %s",
		order.ID, order.UserID, order.Amount, order.Currency)

	s.publishOrderUpdate(ctx, order)

	return order, nil
}

func (s *OrdersService) GetOrder(ctx context.Context, orderID string) (*orders.Order, error) {
	return s.ordersRepository.GetByID(ctx, orderID)
}

func (s *OrdersService) GetUserOrders(ctx context.Context, userID string) ([]*orders.Order, error) {
	return s.ordersRepository.GetByUserID(ctx, userID)
}

func (s *OrdersService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	return s.ordersRepository.UpdateStatus(ctx, orderID, status)
}

func (s *OrdersService) ProcessPaymentCompleted(ctx context.Context, inboxMessage *inbox.InboxMessage) error {
	var paymentEvent inbox.PaymentCompletedEvent
	if err := json.Unmarshal(inboxMessage.Payload, &paymentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payment completed event: %w", err)
	}

	log.Printf("Processing payment completed event: OrderID=%s, PaymentID=%s, TransactionID=%s",
		paymentEvent.OrderID, paymentEvent.PaymentID, paymentEvent.TransactionID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	order, err := s.ordersRepository.GetByID(ctx, paymentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	order.MarkPaid(paymentEvent.PaymentID)

	if err := s.ordersRepository.UpdateWithTx(ctx, tx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	orderUpdatedEvent := outbox.OrderUpdatedEvent{
		OrderID:   order.ID,
		Status:    string(order.Status),
		PaymentID: order.PaymentID,
	}

	payload, err := json.Marshal(orderUpdatedEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal order updated event: %w", err)
	}

	outboxMessage, err := outbox.NewOutboxMessage("order.updated", payload)
	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	if err := s.outboxRepository.StoreMessage(ctx, tx, outboxMessage); err != nil {
		return fmt.Errorf("failed to store outbox message: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Order marked as paid successfully: OrderID=%s, PaymentID=%s", order.ID, order.PaymentID)

	s.publishOrderUpdate(ctx, order)

	return nil
}

func (s *OrdersService) ProcessPaymentFailed(ctx context.Context, inboxMessage *inbox.InboxMessage) error {
	var paymentEvent inbox.PaymentFailedEvent
	if err := json.Unmarshal(inboxMessage.Payload, &paymentEvent); err != nil {
		return fmt.Errorf("failed to unmarshal payment failed event: %w", err)
	}

	log.Printf("Processing payment failed event: OrderID=%s, PaymentID=%s, Error=%s",
		paymentEvent.OrderID, paymentEvent.PaymentID, paymentEvent.ErrorMessage)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	order, err := s.ordersRepository.GetByID(ctx, paymentEvent.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	order.MarkPaymentFailed(paymentEvent.ErrorMessage)

	if err := s.ordersRepository.UpdateWithTx(ctx, tx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	orderUpdatedEvent := outbox.OrderUpdatedEvent{
		OrderID: order.ID,
		Status:  string(order.Status),
		Reason:  order.ErrorReason,
	}

	payload, err := json.Marshal(orderUpdatedEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal order updated event: %w", err)
	}

	outboxMessage, err := outbox.NewOutboxMessage("order.updated", payload)
	if err != nil {
		return fmt.Errorf("failed to create outbox message: %w", err)
	}

	if err := s.outboxRepository.StoreMessage(ctx, tx, outboxMessage); err != nil {
		return fmt.Errorf("failed to store outbox message: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Order marked as payment failed: OrderID=%s, Reason=%s", order.ID, order.ErrorReason)

	s.publishOrderUpdate(ctx, order)

	return nil
}
