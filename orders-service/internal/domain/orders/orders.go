package orders

import (
	"time"

	"github.com/gofrs/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated        OrderStatus = "created"
	OrderStatusPaymentPending OrderStatus = "payment_pending"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusPaymentFailed  OrderStatus = "payment_failed"
	OrderStatusCompleted      OrderStatus = "completed"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"userID"`
	Amount      float64     `json:"amount"`
	Currency    string      `json:"currency"`
	Status      OrderStatus `json:"status"`
	PaymentID   string      `json:"paymentID"`
	ErrorReason string      `json:"errorReason"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func NewOrder(userID string, amount float64) (*Order, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Order{
		ID:        v7.String(),
		UserID:    userID,
		Amount:    amount,
		Currency:  "USD",
		Status:    OrderStatusCreated,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (o *Order) MarkPaymentPending() {
	o.Status = OrderStatusPaymentPending
	o.UpdatedAt = time.Now()
}

func (o *Order) MarkPaid(paymentID string) {
	o.Status = OrderStatusPaid
	o.PaymentID = paymentID
	o.UpdatedAt = time.Now()
}

func (o *Order) MarkPaymentFailed(reason string) {
	o.Status = OrderStatusPaymentFailed
	o.ErrorReason = reason
	o.UpdatedAt = time.Now()
}

func (o *Order) MarkCompleted() {
	o.Status = OrderStatusCompleted
	o.UpdatedAt = time.Now()
}

func (o *Order) MarkCancelled(reason string) {
	o.Status = OrderStatusCancelled
	o.ErrorReason = reason
	o.UpdatedAt = time.Now()
}

func (o *Order) IsCreated() bool {
	return o.Status == OrderStatusCreated
}

func (o *Order) IsPaymentPending() bool {
	return o.Status == OrderStatusPaymentPending
}

func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid
}

func (o *Order) IsPaymentFailed() bool {
	return o.Status == OrderStatusPaymentFailed
}

func (o *Order) IsCompleted() bool {
	return o.Status == OrderStatusCompleted
}

func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}
