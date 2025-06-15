package payments

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPayment(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"
	amount := 100.50
	currency := "USD"

	p, err := NewPayment(orderID, userID, amount, currency)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, orderID, p.OrderID)
	assert.Equal(t, userID, p.UserID)
	assert.Equal(t, amount, p.Amount)
	assert.Equal(t, currency, p.Currency)
	assert.Equal(t, PaymentStatusPending, p.Status)
	assert.Empty(t, p.TransactionID)
	assert.Empty(t, p.ErrorMessage)
	assert.WithinDuration(t, time.Now(), p.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), p.UpdatedAt, time.Second)
}

func TestPayment_Complete(t *testing.T) {
	p, _ := NewPayment("order-123", "user-456", 100.50, "USD")
	transactionID := "txn-789"

	p.Complete(transactionID)

	assert.Equal(t, PaymentStatusCompleted, p.Status)
	assert.Equal(t, transactionID, p.TransactionID)
	assert.True(t, p.UpdatedAt.After(p.CreatedAt))
}

func TestPayment_Fail(t *testing.T) {
	p, _ := NewPayment("order-123", "user-456", 100.50, "USD")
	failureReason := "Insufficient funds"

	p.Fail(failureReason)

	assert.Equal(t, PaymentStatusFailed, p.Status)
	assert.Equal(t, failureReason, p.ErrorMessage)
	assert.True(t, p.UpdatedAt.After(p.CreatedAt))
}

func TestPayment_IsStatus(t *testing.T) {
	p, _ := NewPayment("order-123", "user-456", 100.50, "USD")
	assert.True(t, p.IsPending())
	assert.False(t, p.IsCompleted())
	assert.False(t, p.IsFailed())

	p.Complete("txn-1")
	assert.False(t, p.IsPending())
	assert.True(t, p.IsCompleted())
	assert.False(t, p.IsFailed())

	p, _ = NewPayment("order-123", "user-456", 100.50, "USD")
	p.Fail("failed")
	assert.False(t, p.IsPending())
	assert.False(t, p.IsCompleted())
	assert.True(t, p.IsFailed())
}

func TestPayment_IsTimedOut(t *testing.T) {
	p, _ := NewPayment("order-123", "user-456", 100.50, "USD")
	assert.False(t, p.IsTimedOut())

	p.CreatedAt = time.Now().Add(-16 * time.Minute)
	assert.True(t, p.IsTimedOut())
}
