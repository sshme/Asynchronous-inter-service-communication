package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"payments-service/internal/domain/account"
	"payments-service/internal/domain/inbox"
	"payments-service/internal/domain/outbox"
	"payments-service/internal/domain/payments"
)

// safeDB is a thread-safe wrapper for *sql.DB.
type safeDB struct {
	*sql.DB
	m sync.Mutex
}

func (db *safeDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	db.m.Lock()
	defer db.m.Unlock()
	return db.DB.BeginTx(ctx, opts)
}

// Mocks
type MockPaymentsRepository struct {
	mock.Mock
}

func (m *MockPaymentsRepository) Store(ctx context.Context, p *payments.Payment) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPaymentsRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, p *payments.Payment) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

func (m *MockPaymentsRepository) GetByID(ctx context.Context, id string) (*payments.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Payment), args.Error(1)
}

func (m *MockPaymentsRepository) GetByOrderID(ctx context.Context, orderID string) (*payments.Payment, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Payment), args.Error(1)
}

func (m *MockPaymentsRepository) Update(ctx context.Context, p *payments.Payment) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPaymentsRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, p *payments.Payment) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Store(ctx context.Context, a *account.Account) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAccountRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, a *account.Account) error {
	args := m.Called(ctx, tx, a)
	return args.Error(0)
}

func (m *MockAccountRepository) GetByUserID(ctx context.Context, userID string) (*account.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByUserIDWithTx(ctx context.Context, tx *sql.Tx, userID string) (*account.Account, error) {
	args := m.Called(ctx, tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(ctx context.Context, a *account.Account) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockAccountRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, a *account.Account) error {
	args := m.Called(ctx, tx, a)
	return args.Error(0)
}

type MockOutboxRepository struct {
	mock.Mock
}

func (m *MockOutboxRepository) Store(ctx context.Context, message *outbox.OutboxMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockOutboxRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, message *outbox.OutboxMessage) error {
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

func (m *MockOutboxRepository) GetFailedMessages(ctx context.Context, maxRetries, limit int) ([]*outbox.OutboxMessage, error) {
	args := m.Called(ctx, maxRetries, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*outbox.OutboxMessage), args.Error(1)
}

func (m *MockOutboxRepository) MarkAsSent(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOutboxRepository) MarkAsFailed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestPaymentsService_ProcessOrderCreated_Success(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	safeDB := &safeDB{DB: db}

	mockPaymentsRepo := new(MockPaymentsRepository)
	mockAccountRepo := new(MockAccountRepository)
	mockOutboxRepo := new(MockOutboxRepository)

	service := NewPaymentsService(safeDB, mockPaymentsRepo, mockAccountRepo, nil, mockOutboxRepo, nil)

	ctx := context.Background()
	orderEvent := inbox.OrderCreatedEvent{
		OrderID:  "order-123",
		UserID:   "user-456",
		Amount:   100.50,
		Currency: "USD",
	}
	payload, _ := json.Marshal(orderEvent)
	inboxMsg := &inbox.InboxMessage{Payload: payload}

	userAccount, _ := account.NewAccount(orderEvent.UserID)
	_ = userAccount.Credit(200.0) // Sufficient funds

	mockSQL.ExpectBegin()
	mockPaymentsRepo.On("GetByOrderID", ctx, orderEvent.OrderID).Return(nil, sql.ErrNoRows)
	mockPaymentsRepo.On("StoreWithTx", ctx, mock.Anything, mock.AnythingOfType("*payments.Payment")).Return(nil)
	mockAccountRepo.On("GetByUserID", ctx, orderEvent.UserID).Return(userAccount, nil)
	mockAccountRepo.On("UpdateWithTx", ctx, mock.Anything, userAccount).Return(nil)
	mockPaymentsRepo.On("UpdateWithTx", ctx, mock.Anything, mock.AnythingOfType("*payments.Payment")).Return(nil)
	mockOutboxRepo.On("StoreWithTx", ctx, mock.Anything, mock.AnythingOfType("*outbox.OutboxMessage")).Return(nil)
	mockSQL.ExpectCommit()

	err = service.ProcessOrderCreated(ctx, inboxMsg)
	assert.NoError(t, err)

	mockPaymentsRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	assert.NoError(t, mockSQL.ExpectationsWereMet())
}

func TestPaymentsService_ProcessOrderCreated_InsufficientFunds(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	safeDB := &safeDB{DB: db}

	mockPaymentsRepo := new(MockPaymentsRepository)
	mockAccountRepo := new(MockAccountRepository)
	mockOutboxRepo := new(MockOutboxRepository)

	service := NewPaymentsService(safeDB, mockPaymentsRepo, mockAccountRepo, nil, mockOutboxRepo, nil)

	ctx := context.Background()
	orderEvent := inbox.OrderCreatedEvent{
		OrderID:  "order-123",
		UserID:   "user-456",
		Amount:   100.50,
		Currency: "USD",
	}
	payload, _ := json.Marshal(orderEvent)
	inboxMsg := &inbox.InboxMessage{Payload: payload}

	userAccount, _ := account.NewAccount(orderEvent.UserID)
	_ = userAccount.Credit(50.0) // Insufficient funds

	mockSQL.ExpectBegin()
	mockPaymentsRepo.On("GetByOrderID", ctx, orderEvent.OrderID).Return(nil, sql.ErrNoRows)
	mockPaymentsRepo.On("StoreWithTx", ctx, mock.Anything, mock.AnythingOfType("*payments.Payment")).Return(nil)
	mockAccountRepo.On("GetByUserID", ctx, orderEvent.UserID).Return(userAccount, nil)
	mockSQL.ExpectCommit() // The transaction is committed even on retry

	err = service.ProcessOrderCreated(ctx, inboxMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")

	mockPaymentsRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	assert.NoError(t, mockSQL.ExpectationsWereMet())
}

func TestPaymentsService_ProcessOrderCreated_ExistingPayment(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	safeDB := &safeDB{DB: db}

	mockPaymentsRepo := new(MockPaymentsRepository)
	mockAccountRepo := new(MockAccountRepository)
	mockOutboxRepo := new(MockOutboxRepository)

	service := NewPaymentsService(safeDB, mockPaymentsRepo, mockAccountRepo, nil, mockOutboxRepo, nil)

	ctx := context.Background()
	orderEvent := inbox.OrderCreatedEvent{
		OrderID:  "order-123",
		UserID:   "user-456",
		Amount:   100.50,
		Currency: "USD",
	}
	payload, _ := json.Marshal(orderEvent)
	inboxMsg := &inbox.InboxMessage{Payload: payload}

	existingPayment, _ := payments.NewPayment(orderEvent.OrderID, orderEvent.UserID, orderEvent.Amount, orderEvent.Currency)
	userAccount, _ := account.NewAccount(orderEvent.UserID)
	_ = userAccount.Credit(200.0) // Sufficient funds

	mockSQL.ExpectBegin()
	mockPaymentsRepo.On("GetByOrderID", ctx, orderEvent.OrderID).Return(existingPayment, nil)
	mockAccountRepo.On("GetByUserID", ctx, orderEvent.UserID).Return(userAccount, nil)
	mockAccountRepo.On("UpdateWithTx", ctx, mock.Anything, userAccount).Return(nil)
	mockPaymentsRepo.On("UpdateWithTx", ctx, mock.Anything, existingPayment).Return(nil)
	mockOutboxRepo.On("StoreWithTx", ctx, mock.Anything, mock.AnythingOfType("*outbox.OutboxMessage")).Return(nil)
	mockSQL.ExpectCommit()

	err = service.ProcessOrderCreated(ctx, inboxMsg)
	assert.NoError(t, err)

	mockPaymentsRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)
	assert.NoError(t, mockSQL.ExpectationsWereMet())
}
