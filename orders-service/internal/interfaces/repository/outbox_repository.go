package repository

import (
	"context"
	"database/sql"
	"orders-service/internal/domain/outbox"
)

type OutboxRepository interface {
	StoreMessage(ctx context.Context, tx *sql.Tx, message *outbox.OutboxMessage) error
	GetPendingMessages(ctx context.Context, limit int) ([]*outbox.OutboxMessage, error)
	MarkAsSent(ctx context.Context, messageID string) error
	MarkAsFailed(ctx context.Context, messageID string) error
	GetFailedMessages(ctx context.Context, maxRetries int, limit int) ([]*outbox.OutboxMessage, error)
}
