package inbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInboxMessage(t *testing.T) {
	eventID := "event-123"
	eventType := "payment.completed"
	payload := json.RawMessage(`{"payment_id": "p-123"}`)
	msg, err := NewInboxMessage(eventID, eventType, payload)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, eventID, msg.EventID)
	assert.Equal(t, eventType, msg.EventType)
	assert.Equal(t, payload, msg.Payload)
	assert.True(t, msg.IsPending())
	assert.Equal(t, 0, msg.RetryCount)
	assert.Equal(t, 3, msg.MaxRetries)
	assert.WithinDuration(t, time.Now(), msg.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), msg.UpdatedAt, time.Second)
	assert.Nil(t, msg.ProcessedAt)
}

func TestInboxMessage_StateTransitions(t *testing.T) {
	msg, _ := NewInboxMessage("event-456", "payment.failed", json.RawMessage(`{}`))

	msg.MarkAsProcessed()
	assert.True(t, msg.IsProcessed())
	assert.False(t, msg.IsPending())
	assert.False(t, msg.IsFailed())
	assert.NotNil(t, msg.ProcessedAt)
	assert.WithinDuration(t, time.Now(), *msg.ProcessedAt, time.Second)

	msg.MarkAsFailed()
	assert.True(t, msg.IsFailed())
	assert.Equal(t, 1, msg.RetryCount)
}

func TestInboxMessage_CanRetry(t *testing.T) {
	msg, _ := NewInboxMessage("event-789", "payment.whatever", json.RawMessage(`{}`))

	assert.True(t, msg.CanRetry())

	// Simulate time passing
	msg.CreatedAt = time.Now().Add(-20 * time.Minute)
	assert.False(t, msg.CanRetry())
}
