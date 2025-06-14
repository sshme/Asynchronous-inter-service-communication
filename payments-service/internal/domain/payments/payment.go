package payments

import (
	"time"

	"github.com/gofrs/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID            string
	OrderID       string
	UserID        string
	Amount        float64
	Currency      string
	Status        PaymentStatus
	ErrorMessage  string
	TransactionID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewPayment(orderID, userID string, amount float64, currency string) (*Payment, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Payment{
		ID:        v7.String(),
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (p *Payment) Complete(transactionID string) {
	p.Status = PaymentStatusCompleted
	p.TransactionID = transactionID
	p.UpdatedAt = time.Now()
}

func (p *Payment) Fail(errorMessage string) {
	p.Status = PaymentStatusFailed
	p.ErrorMessage = errorMessage
	p.UpdatedAt = time.Now()
}

func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

func (p *Payment) IsTimedOut() bool {
	return time.Since(p.CreatedAt) > 15*time.Second
}
