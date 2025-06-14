package inbox

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

type InboxMessageStatus string

const (
	InboxMessageStatusPending   InboxMessageStatus = "pending"
	InboxMessageStatusProcessed InboxMessageStatus = "processed"
	InboxMessageStatusFailed    InboxMessageStatus = "failed"
)

type InboxMessage struct {
	ID          string
	EventID     string
	EventType   string
	Payload     json.RawMessage
	Status      InboxMessageStatus
	ProcessedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RetryCount  int
	MaxRetries  int
}

func NewInboxMessage(eventID, eventType string, payload json.RawMessage) (*InboxMessage, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &InboxMessage{
		ID:         v7.String(),
		EventID:    eventID,
		EventType:  eventType,
		Payload:    payload,
		Status:     InboxMessageStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		RetryCount: 0,
		MaxRetries: 3,
	}, nil
}

func (m *InboxMessage) MarkAsProcessed() {
	now := time.Now()
	m.Status = InboxMessageStatusProcessed
	m.ProcessedAt = &now
	m.UpdatedAt = now
}

func (m *InboxMessage) MarkAsFailed() {
	m.Status = InboxMessageStatusFailed
	m.RetryCount++
	m.UpdatedAt = time.Now()
}

func (m *InboxMessage) CanRetry() bool {
	return m.RetryCount < m.MaxRetries
}

func (m *InboxMessage) IsProcessed() bool {
	return m.Status == InboxMessageStatusProcessed
}

func (m *InboxMessage) IsFailed() bool {
	return m.Status == InboxMessageStatusFailed
}

func (m *InboxMessage) IsPending() bool {
	return m.Status == InboxMessageStatusPending
}

type OrderCreatedEvent struct {
	OrderID  string  `json:"order_id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
