package repository

import (
	"context"
	"database/sql"

	"orders-service/internal/domain/inbox"
)

type InboxRepository interface {
	Store(ctx context.Context, message *inbox.InboxMessage) error
	StoreWithTx(ctx context.Context, tx *sql.Tx, message *inbox.InboxMessage) error
	GetByEventID(ctx context.Context, eventID string) (*inbox.InboxMessage, error)
	GetPendingMessages(ctx context.Context, limit int) ([]*inbox.InboxMessage, error)
	GetFailedMessages(ctx context.Context, maxRetries, limit int) ([]*inbox.InboxMessage, error)
	MarkAsProcessed(ctx context.Context, id string) error
	MarkAsFailed(ctx context.Context, id string) error
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
}
