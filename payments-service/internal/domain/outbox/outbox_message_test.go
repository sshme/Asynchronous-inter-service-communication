package outbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOutboxMessage(t *testing.T) {
	eventType := "test.topic"
	payload := json.RawMessage(`{"key":"value"}`)

	msg, err := NewOutboxMessage(eventType, payload)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, eventType, msg.EventType)
	assert.Equal(t, payload, msg.Payload)
	assert.Equal(t, OutboxMessageStatusPending, msg.Status)
	assert.Equal(t, 0, msg.RetryCount)
	assert.WithinDuration(t, time.Now(), msg.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, time.Second)
}

func TestOutboxMessage_MarkAsSent(t *testing.T) {
	msg, _ := NewOutboxMessage("test.topic", json.RawMessage(`{}`))
	msg.MarkAsSent()

	assert.Equal(t, OutboxMessageStatusSent, msg.Status)
	assert.NotNil(t, msg.SentAt)
	assert.True(t, msg.UpdatedAt.After(msg.CreatedAt))
}

func TestOutboxMessage_MarkAsFailed(t *testing.T) {
	msg, _ := NewOutboxMessage("test.topic", json.RawMessage(`{}`))
	msg.MarkAsFailed()

	assert.Equal(t, OutboxMessageStatusFailed, msg.Status)
	assert.Equal(t, 1, msg.RetryCount)
	assert.True(t, msg.UpdatedAt.After(msg.CreatedAt))
}

func TestOutboxMessage_CanRetry(t *testing.T) {
	msg, _ := NewOutboxMessage("test.topic", json.RawMessage(`{}`))
	assert.True(t, msg.CanRetry())

	msg.RetryCount = 3
	assert.False(t, msg.CanRetry())
}

func TestOutboxMessage_IsStatus(t *testing.T) {
	msg, _ := NewOutboxMessage("test.topic", json.RawMessage(`{}`))
	assert.True(t, msg.IsPending())
	assert.False(t, msg.IsSent())
	assert.False(t, msg.IsFailed())

	msg.MarkAsSent()
	assert.False(t, msg.IsPending())
	assert.True(t, msg.IsSent())
	assert.False(t, msg.IsFailed())

	msg, _ = NewOutboxMessage("test.topic", json.RawMessage(`{}`))
	msg.MarkAsFailed()
	assert.False(t, msg.IsPending())
	assert.False(t, msg.IsSent())
	assert.True(t, msg.IsFailed())
}
