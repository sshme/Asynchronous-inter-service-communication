package outbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOutboxMessage(t *testing.T) {
	eventType := "order.created"
	payload := json.RawMessage(`{"order_id": "123"}`)
	msg, err := NewOutboxMessage(eventType, payload)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, eventType, msg.EventType)
	assert.Equal(t, payload, msg.Payload)
	assert.True(t, msg.IsPending())
	assert.Equal(t, 0, msg.RetryCount)
	assert.Equal(t, 3, msg.MaxRetries)
	assert.WithinDuration(t, time.Now(), msg.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, time.Second)
	assert.Nil(t, msg.SentAt)
}

func TestOutboxMessage_StateTransitions(t *testing.T) {
	msg, _ := NewOutboxMessage("order.updated", json.RawMessage(`{}`))

	msg.MarkAsSent()
	assert.True(t, msg.IsSent())
	assert.False(t, msg.IsPending())
	assert.False(t, msg.IsFailed())
	assert.NotNil(t, msg.SentAt)
	assert.WithinDuration(t, time.Now(), *msg.SentAt, time.Second)
}

func TestOutboxMessage_Retries(t *testing.T) {
	msg, _ := NewOutboxMessage("order.failed", json.RawMessage(`{}`))
	msg.MaxRetries = 2

	assert.True(t, msg.CanRetry())

	msg.MarkAsFailed()
	assert.True(t, msg.IsFailed())
	assert.Equal(t, 1, msg.RetryCount)
	assert.True(t, msg.CanRetry())

	msg.MarkAsFailed()
	assert.Equal(t, 2, msg.RetryCount)
	assert.False(t, msg.CanRetry())
}
