package outbox

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

type OutboxMessageStatus string

const (
	OutboxMessageStatusPending OutboxMessageStatus = "pending"
	OutboxMessageStatusSent    OutboxMessageStatus = "sent"
	OutboxMessageStatusFailed  OutboxMessageStatus = "failed"
)

type OutboxMessage struct {
	ID         string
	EventType  string
	Payload    json.RawMessage
	Status     OutboxMessageStatus
	SentAt     *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	RetryCount int
	MaxRetries int
}

func NewOutboxMessage(eventType string, payload json.RawMessage) (*OutboxMessage, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &OutboxMessage{
		ID:         v7.String(),
		EventType:  eventType,
		Payload:    payload,
		Status:     OutboxMessageStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		RetryCount: 0,
		MaxRetries: 3,
	}, nil
}

func (m *OutboxMessage) MarkAsSent() {
	now := time.Now()
	m.Status = OutboxMessageStatusSent
	m.SentAt = &now
	m.UpdatedAt = now
}

func (m *OutboxMessage) MarkAsFailed() {
	m.Status = OutboxMessageStatusFailed
	m.RetryCount++
	m.UpdatedAt = time.Now()
}

func (m *OutboxMessage) CanRetry() bool {
	return m.RetryCount < m.MaxRetries
}

func (m *OutboxMessage) IsSent() bool {
	return m.Status == OutboxMessageStatusSent
}

func (m *OutboxMessage) IsFailed() bool {
	return m.Status == OutboxMessageStatusFailed
}

func (m *OutboxMessage) IsPending() bool {
	return m.Status == OutboxMessageStatusPending
}

type PaymentCompletedEvent struct {
	PaymentID     string  `json:"payment_id"`
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	TransactionID string  `json:"transaction_id"`
}

type PaymentFailedEvent struct {
	PaymentID    string  `json:"payment_id"`
	OrderID      string  `json:"order_id"`
	UserID       string  `json:"user_id"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	ErrorMessage string  `json:"error_message"`
}
