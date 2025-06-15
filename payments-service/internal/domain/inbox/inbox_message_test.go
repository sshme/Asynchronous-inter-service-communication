package inbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInboxMessage(t *testing.T) {
	eventID := "event-123"
	eventType := "order.created"
	payload := json.RawMessage(`{"order_id":"order-456"}`)

	msg, err := NewInboxMessage(eventID, eventType, payload)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, eventID, msg.EventID)
	assert.Equal(t, eventType, msg.EventType)
	assert.Equal(t, payload, msg.Payload)
	assert.Equal(t, InboxMessageStatusPending, msg.Status)
	assert.Equal(t, 0, msg.RetryCount)
	assert.WithinDuration(t, time.Now(), msg.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, time.Second)
}

func TestInboxMessage_MarkAsProcessed(t *testing.T) {
	msg, _ := NewInboxMessage("event-123", "order.created", json.RawMessage(`{}`))
	msg.MarkAsProcessed()

	assert.Equal(t, InboxMessageStatusProcessed, msg.Status)
	assert.True(t, msg.UpdatedAt.After(msg.CreatedAt))
}

func TestInboxMessage_MarkAsFailed(t *testing.T) {
	msg, _ := NewInboxMessage("event-123", "order.created", json.RawMessage(`{}`))
	msg.MarkAsFailed()

	assert.Equal(t, InboxMessageStatusFailed, msg.Status)
	assert.Equal(t, 1, msg.RetryCount)
	assert.True(t, msg.UpdatedAt.After(msg.CreatedAt))
}

func TestInboxMessage_CanRetry(t *testing.T) {
	msg, _ := NewInboxMessage("event-123", "order.created", json.RawMessage(`{}`))
	assert.True(t, msg.CanRetry())

	msg.RetryCount = 3
	assert.False(t, msg.CanRetry())
}

func TestInboxMessage_IsStatus(t *testing.T) {
	msg, _ := NewInboxMessage("event-123", "order.created", json.RawMessage(`{}`))
	assert.True(t, msg.IsPending())
	assert.False(t, msg.IsProcessed())
	assert.False(t, msg.IsFailed())

	msg.MarkAsProcessed()
	assert.False(t, msg.IsPending())
	assert.True(t, msg.IsProcessed())
	assert.False(t, msg.IsFailed())

	msg, _ = NewInboxMessage("event-123", "order.created", json.RawMessage(`{}`))
	msg.MarkAsFailed()
	assert.False(t, msg.IsPending())
	assert.False(t, msg.IsProcessed())
	assert.True(t, msg.IsFailed())
}
