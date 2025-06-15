package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"payments-service/internal/domain/account"
	"payments-service/internal/domain/inbox"
	"payments-service/internal/domain/outbox"
	"payments-service/internal/domain/payments"
)

func TestPaymentsService_ProcessOrderCreated_PaymentTimeout(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	safeDB := &safeDB{DB: db}

	mockPaymentsRepo := new(MockPaymentsRepository)
	mockAccountRepo := new(MockAccountRepository)
	mockOutboxRepo := new(MockOutboxRepository)

	service := NewPaymentsService(safeDB, mockPaymentsRepo, mockAccountRepo, nil, mockOutboxRepo, nil)

	ctx := context.Background()

	orderEvent := inbox.OrderCreatedEvent{
		OrderID:  "order-timeout-123",
		UserID:   "user-789",
		Amount:   75.25,
		Currency: "USD",
	}
	payload, err := json.Marshal(orderEvent)
	require.NoError(t, err)

	inboxMsg := &inbox.InboxMessage{Payload: payload}

	oldPayment, err := payments.NewPayment(orderEvent.OrderID, orderEvent.UserID, orderEvent.Amount, orderEvent.Currency)
	require.NoError(t, err)

	oldPayment.CreatedAt = time.Now().Add(-20 * time.Second)

	userAccount, err := account.NewAccount(orderEvent.UserID)
	require.NoError(t, err)
	err = userAccount.Credit(100.0)
	require.NoError(t, err)

	mockSQL.ExpectBegin()

	mockPaymentsRepo.On("GetByOrderID", ctx, orderEvent.OrderID).Return(oldPayment, nil)

	mockPaymentsRepo.On("UpdateWithTx", ctx, mock.Anything, mock.MatchedBy(func(p *payments.Payment) bool {
		return p.IsFailed() && p.ErrorMessage == "Payment timed out"
	})).Return(nil)

	mockOutboxRepo.On("StoreWithTx", ctx, mock.Anything, mock.MatchedBy(func(msg *outbox.OutboxMessage) bool {
		return msg.EventType == "payment.failed"
	})).Return(nil)

	mockSQL.ExpectCommit()

	err = service.ProcessOrderCreated(ctx, inboxMsg)

	assert.NoError(t, err)

	mockPaymentsRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)

	mockAccountRepo.AssertNotCalled(t, "GetByUserID")

	assert.NoError(t, mockSQL.ExpectationsWereMet())
}

func TestPaymentsService_ProcessOrderCreated_PaymentNotTimedOut(t *testing.T) {
	db, mockSQL, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	safeDB := &safeDB{DB: db}

	mockPaymentsRepo := new(MockPaymentsRepository)
	mockAccountRepo := new(MockAccountRepository)
	mockOutboxRepo := new(MockOutboxRepository)

	service := NewPaymentsService(safeDB, mockPaymentsRepo, mockAccountRepo, nil, mockOutboxRepo, nil)

	ctx := context.Background()

	orderEvent := inbox.OrderCreatedEvent{
		OrderID:  "order-fresh-456",
		UserID:   "user-101",
		Amount:   50.00,
		Currency: "EUR",
	}
	payload, err := json.Marshal(orderEvent)
	require.NoError(t, err)

	inboxMsg := &inbox.InboxMessage{Payload: payload}

	freshPayment, err := payments.NewPayment(orderEvent.OrderID, orderEvent.UserID, orderEvent.Amount, orderEvent.Currency)
	require.NoError(t, err)

	freshPayment.CreatedAt = time.Now().Add(-5 * time.Second)

	userAccount, err := account.NewAccount(orderEvent.UserID)
	require.NoError(t, err)
	err = userAccount.Credit(100.0)
	require.NoError(t, err)

	mockSQL.ExpectBegin()

	mockPaymentsRepo.On("GetByOrderID", ctx, orderEvent.OrderID).Return(freshPayment, nil)

	mockAccountRepo.On("GetByUserID", ctx, orderEvent.UserID).Return(userAccount, nil)

	mockAccountRepo.On("UpdateWithTx", ctx, mock.Anything, mock.MatchedBy(func(acc *account.Account) bool {
		return acc.Balance == 50.0 // 100 - 50 = 50
	})).Return(nil)

	mockPaymentsRepo.On("UpdateWithTx", ctx, mock.Anything, mock.MatchedBy(func(p *payments.Payment) bool {
		return p.IsCompleted() && p.TransactionID != ""
	})).Return(nil)

	mockOutboxRepo.On("StoreWithTx", ctx, mock.Anything, mock.MatchedBy(func(msg *outbox.OutboxMessage) bool {
		return msg.EventType == "payment.completed"
	})).Return(nil)

	mockSQL.ExpectCommit()

	err = service.ProcessOrderCreated(ctx, inboxMsg)

	assert.NoError(t, err)

	mockPaymentsRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockOutboxRepo.AssertExpectations(t)

	assert.NoError(t, mockSQL.ExpectationsWereMet())
}

func TestPayment_IsTimedOut_EdgeCases(t *testing.T) {
	t.Run("14.9 seconds old - should not be timed out", func(t *testing.T) {
		payment, err := payments.NewPayment("order-1", "user-1", 100.0, "USD")
		require.NoError(t, err)

		payment.CreatedAt = time.Now().Add(-14*time.Second - 900*time.Millisecond)

		assert.False(t, payment.IsTimedOut())
	})

	t.Run("15.1 seconds old - should be timed out", func(t *testing.T) {
		payment, err := payments.NewPayment("order-2", "user-2", 100.0, "USD")
		require.NoError(t, err)

		payment.CreatedAt = time.Now().Add(-15*time.Second - 100*time.Millisecond)

		assert.True(t, payment.IsTimedOut())
	})

	t.Run("Just created - should not be timed out", func(t *testing.T) {
		payment, err := payments.NewPayment("order-3", "user-3", 100.0, "USD")
		require.NoError(t, err)

		assert.False(t, payment.IsTimedOut())
	})

	t.Run("Very old payment - should be timed out", func(t *testing.T) {
		payment, err := payments.NewPayment("order-4", "user-4", 100.0, "USD")
		require.NoError(t, err)

		payment.CreatedAt = time.Now().Add(-1 * time.Hour)

		assert.True(t, payment.IsTimedOut())
	})
}

func BenchmarkPayment_IsTimedOut(b *testing.B) {
	payment, _ := payments.NewPayment("order-bench", "user-bench", 100.0, "USD")
	payment.CreatedAt = time.Now().Add(-20 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = payment.IsTimedOut()
	}
}
