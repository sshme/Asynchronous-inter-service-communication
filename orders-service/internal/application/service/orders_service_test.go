package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"orders-service/internal/domain/inbox"
	"orders-service/internal/domain/orders"
	"orders-service/internal/domain/outbox"
	"orders-service/internal/infrastructure/pubsub/redis"
)

// Mocks
type MockOrdersRepository struct {
	mock.Mock
}

func (m *MockOrdersRepository) Store(ctx context.Context, order *orders.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrdersRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, order *orders.Order) error {
	args := m.Called(ctx, tx, order)
	return args.Error(0)
}

func (m *MockOrdersRepository) GetByID(ctx context.Context, orderID string) (*orders.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orders.Order), args.Error(1)
}

func (m *MockOrdersRepository) GetByUserID(ctx context.Context, userID string) ([]*orders.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orders.Order), args.Error(1)
}

func (m *MockOrdersRepository) Update(ctx context.Context, order *orders.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrdersRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, order *orders.Order) error {
	args := m.Called(ctx, tx, order)
	return args.Error(0)
}

func (m *MockOrdersRepository) UpdateStatus(ctx context.Context, orderID string, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) StoreMessage(ctx context.Context, tx *sql.Tx, message *outbox.OutboxMessage) error {
	args := m.Called(ctx, tx, message)
	return args.Error(0)
}

func (m *MockOutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]*outbox.OutboxMessage, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbox.OutboxMessage), args.Error(1)
}

func (m *MockOutboxRepository) MarkAsSent(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockOutboxRepository) MarkAsFailed(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockOutboxRepository) GetFailedMessages(ctx context.Context, maxRetries int, limit int) ([]*outbox.OutboxMessage, error) {
	args := m.Called(ctx, maxRetries, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbox.OutboxMessage), args.Error(1)
}

type MockRandomGenerator struct {
	mock.Mock
}

func (m *MockRandomGenerator) FloatWithPrecision(min, max float64) (float64, error) {
	args := m.Called(min, max)
	return args.Get(0).(float64), args.Error(1)
}

func TestOrdersService_CreateOrder(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisClient, redisMock := redismock.NewClientMock()

	mockOrdersRepo := new(MockOrdersRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	mockRandomGen := new(MockRandomGenerator)
	redisPublisher := redis.NewPublisher(redisClient, &redis.Config{Channel: "test"})

	service := NewOrdersService(mockOrdersRepo, mockOutboxRepo, mockRandomGen, redisPublisher, db)

	userID := "test-user"
	expectedAmount := 123.45
	ctx := context.Background()

	mockRandomGen.On("FloatWithPrecision", 100.0, 1000.0).Return(expectedAmount, nil)
	mockSQL.ExpectBegin()
	mockOrdersRepo.On("StoreWithTx", ctx, mock.Anything, mock.AnythingOfType("*orders.Order")).Return(nil)
	mockOutboxRepo.On("StoreMessage", ctx, mock.Anything, mock.AnythingOfType("*outbox.OutboxMessage")).Return(nil)
	mockSQL.ExpectCommit()
	redisMock.ExpectPublish("test", mock.Anything).SetVal(0)

	order, err := service.CreateOrder(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, expectedAmount, order.Amount)
	mockOrdersRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	mockRandomGen.AssertExpectations(t)
	assert.NoError(t, mockSQL.ExpectationsWereMet())
}

func TestOrdersService_ProcessPaymentCompleted(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisClient, redisMock := redismock.NewClientMock()

	mockOrdersRepo := new(MockOrdersRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	redisPublisher := redis.NewPublisher(redisClient, &redis.Config{Channel: "test"})

	service := NewOrdersService(mockOrdersRepo, mockOutboxRepo, nil, redisPublisher, db)

	ctx := context.Background()
	orderID := "test-order-id"
	paymentID := "test-payment-id"
	transactionID := "test-transaction-id"

	paymentEvent := inbox.PaymentCompletedEvent{
		OrderID:       orderID,
		PaymentID:     paymentID,
		TransactionID: transactionID,
	}
	payload, _ := json.Marshal(paymentEvent)
	inboxMsg := &inbox.InboxMessage{
		ID:      "test-inbox-id",
		Payload: payload,
	}

	order, _ := orders.NewOrder("test-user", 100)
	order.ID = orderID

	mockSQL.ExpectBegin()
	mockOrdersRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockOrdersRepo.On("UpdateWithTx", ctx, mock.Anything, mock.AnythingOfType("*orders.Order")).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*orders.Order)
		assert.Equal(t, orders.OrderStatusPaid, arg.Status)
		assert.Equal(t, paymentID, arg.PaymentID)
	}).Return(nil)
	mockOutboxRepo.On("StoreMessage", ctx, mock.Anything, mock.AnythingOfType("*outbox.OutboxMessage")).Return(nil)
	mockSQL.ExpectCommit()
	redisMock.ExpectPublish("test", mock.Anything).SetVal(0)

	err = service.ProcessPaymentCompleted(ctx, inboxMsg)

	assert.NoError(t, err)
	mockOrdersRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	assert.NoError(t, mockSQL.ExpectationsWereMet())
}

func TestOrdersService_ProcessPaymentFailed(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisClient, redisMock := redismock.NewClientMock()

	mockOrdersRepo := new(MockOrdersRepository)
	mockOutboxRepo := new(MockOutboxRepository)
	redisPublisher := redis.NewPublisher(redisClient, &redis.Config{Channel: "test"})

	service := NewOrdersService(mockOrdersRepo, mockOutboxRepo, nil, redisPublisher, db)

	ctx := context.Background()
	orderID := "test-order-id"
	paymentID := "test-payment-id"
	errorMessage := "payment provider error"

	paymentEvent := inbox.PaymentFailedEvent{
		OrderID:      orderID,
		PaymentID:    paymentID,
		ErrorMessage: errorMessage,
	}
	payload, _ := json.Marshal(paymentEvent)
	inboxMsg := &inbox.InboxMessage{
		ID:      "test-inbox-id",
		Payload: payload,
	}

	order, _ := orders.NewOrder("test-user", 100)
	order.ID = orderID

	mockSQL.ExpectBegin()
	mockOrdersRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockOrdersRepo.On("UpdateWithTx", ctx, mock.Anything, mock.AnythingOfType("*orders.Order")).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*orders.Order)
		assert.Equal(t, orders.OrderStatusPaymentFailed, arg.Status)
		assert.NotNil(t, arg.ErrorReason)
		assert.Equal(t, errorMessage, arg.ErrorReason)
	}).Return(nil)
	mockOutboxRepo.On("StoreMessage", ctx, mock.Anything, mock.AnythingOfType("*outbox.OutboxMessage")).Return(nil)
	mockSQL.ExpectCommit()
	redisMock.ExpectPublish("test", mock.Anything).SetVal(0)

	err = service.ProcessPaymentFailed(ctx, inboxMsg)

	assert.NoError(t, err)
	mockOrdersRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	assert.NoError(t, mockSQL.ExpectationsWereMet())
}
