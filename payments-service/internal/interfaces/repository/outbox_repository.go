package repository

import (
	"context"
	"database/sql"

	"payments-service/internal/domain/outbox"
)

type OutboxRepository interface {
	Store(ctx context.Context, message *outbox.OutboxMessage) error
	StoreWithTx(ctx context.Context, tx *sql.Tx, message *outbox.OutboxMessage) error
	GetPendingMessages(ctx context.Context, limit int) ([]*outbox.OutboxMessage, error)
	GetFailedMessages(ctx context.Context, maxRetries, limit int) ([]*outbox.OutboxMessage, error)
	MarkAsSent(ctx context.Context, id string) error
	MarkAsFailed(ctx context.Context, id string) error
}
