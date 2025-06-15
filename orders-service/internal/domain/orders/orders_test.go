package orders

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrder_StateTransitions(t *testing.T) {
	order, err := NewOrder("user-123", 100.50)
	assert.NoError(t, err)
	assert.True(t, order.IsCreated())
	assert.Equal(t, OrderStatusCreated, order.Status)

	order.MarkPaymentPending()
	assert.True(t, order.IsPaymentPending())
	assert.Equal(t, OrderStatusPaymentPending, order.Status)

	order.MarkPaid("payment-abc")
	assert.True(t, order.IsPaid())
	assert.Equal(t, OrderStatusPaid, order.Status)
	assert.Equal(t, "payment-abc", order.PaymentID)

	order.MarkCompleted()
	assert.True(t, order.IsCompleted())
	assert.Equal(t, OrderStatusCompleted, order.Status)
}

func TestOrder_FailureAndCancellation(t *testing.T) {
	order, err := NewOrder("user-123", 100.50)
	assert.NoError(t, err)

	failReason := "insufficient funds"
	order.MarkPaymentFailed(failReason)
	assert.True(t, order.IsPaymentFailed())
	assert.Equal(t, OrderStatusPaymentFailed, order.Status)
	assert.Equal(t, failReason, order.ErrorReason)

	cancelReason := "user cancelled"
	order.MarkCancelled(cancelReason)
	assert.True(t, order.IsCancelled())
	assert.Equal(t, OrderStatusCancelled, order.Status)
	assert.Equal(t, cancelReason, order.ErrorReason)
}

func TestNewOrder(t *testing.T) {
	userID := "user-456"
	amount := 250.75
	order, err := NewOrder(userID, amount)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.NotEmpty(t, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, amount, order.Amount)
	assert.Equal(t, "USD", order.Currency)
	assert.Equal(t, OrderStatusCreated, order.Status)
	assert.WithinDuration(t, time.Now(), order.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), order.UpdatedAt, time.Second)
}
