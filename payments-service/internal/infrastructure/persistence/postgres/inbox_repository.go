package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"payments-service/internal/domain/inbox"
	"payments-service/internal/interfaces/repository"
)

type InboxRepository struct {
	db *sql.DB
}

func NewInboxRepository(db *sql.DB) repository.InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Store(ctx context.Context, message *inbox.InboxMessage) error {
	query := `
		INSERT INTO inbox_messages (id, event_id, event_type, payload, status, processed_at, created_at, updated_at, retry_count, max_retries)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		message.ID, message.EventID, message.EventType, message.Payload, message.Status,
		message.ProcessedAt, message.CreatedAt, message.UpdatedAt, message.RetryCount, message.MaxRetries)
	if err != nil {
		return fmt.Errorf("failed to store inbox message: %w", err)
	}

	return nil
}

func (r *InboxRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, message *inbox.InboxMessage) error {
	query := `
		INSERT INTO inbox_messages (id, event_id, event_type, payload, status, processed_at, created_at, updated_at, retry_count, max_retries)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := tx.ExecContext(ctx, query,
		message.ID, message.EventID, message.EventType, message.Payload, message.Status,
		message.ProcessedAt, message.CreatedAt, message.UpdatedAt, message.RetryCount, message.MaxRetries)
	if err != nil {
		return fmt.Errorf("failed to store inbox message with tx: %w", err)
	}

	return nil
}

func (r *InboxRepository) GetByEventID(ctx context.Context, eventID string) (*inbox.InboxMessage, error) {
	query := `
		SELECT id, event_id, event_type, payload, status, processed_at, created_at, updated_at, retry_count, max_retries
		FROM inbox_messages
		WHERE event_id = $1`

	row := r.db.QueryRowContext(ctx, query, eventID)

	message := &inbox.InboxMessage{}
	err := row.Scan(&message.ID, &message.EventID, &message.EventType, &message.Payload, &message.Status,
		&message.ProcessedAt, &message.CreatedAt, &message.UpdatedAt, &message.RetryCount, &message.MaxRetries)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("inbox message not found for event: %s", eventID)
		}
		return nil, fmt.Errorf("failed to get inbox message by event ID: %w", err)
	}

	return message, nil
}

func (r *InboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]*inbox.InboxMessage, error) {
	query := `
		SELECT id, event_id, event_type, payload, status, processed_at, created_at, updated_at, retry_count, max_retries
		FROM inbox_messages
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending inbox messages: %w", err)
	}
	defer rows.Close()

	var messages []*inbox.InboxMessage
	for rows.Next() {
		message := &inbox.InboxMessage{}
		err := rows.Scan(&message.ID, &message.EventID, &message.EventType, &message.Payload, &message.Status,
			&message.ProcessedAt, &message.CreatedAt, &message.UpdatedAt, &message.RetryCount, &message.MaxRetries)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inbox message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *InboxRepository) GetFailedMessages(ctx context.Context, maxAge time.Duration, limit int) ([]*inbox.InboxMessage, error) {
	query := `
		SELECT id, event_id, event_type, payload, status, processed_at, created_at, updated_at, retry_count, max_retries
		FROM inbox_messages
		WHERE status = 'failed' AND created_at >= NOW() - INTERVAL '%d seconds'
		ORDER BY created_at ASC
		LIMIT $1`

	seconds := int(maxAge.Seconds())
	query = fmt.Sprintf(query, seconds)

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed inbox messages: %w", err)
	}
	defer rows.Close()

	var messages []*inbox.InboxMessage
	for rows.Next() {
		message := &inbox.InboxMessage{}
		err := rows.Scan(&message.ID, &message.EventID, &message.EventType, &message.Payload, &message.Status,
			&message.ProcessedAt, &message.CreatedAt, &message.UpdatedAt, &message.RetryCount, &message.MaxRetries)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inbox message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (r *InboxRepository) MarkAsProcessed(ctx context.Context, id string) error {
	query := `
		UPDATE inbox_messages
		SET status = 'processed', processed_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark inbox message as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("inbox message not found: %s", id)
	}

	return nil
}

func (r *InboxRepository) MarkAsFailed(ctx context.Context, id string) error {
	query := `
		UPDATE inbox_messages
		SET status = 'failed', retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark inbox message as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("inbox message not found: %s", id)
	}

	return nil
}

func (r *InboxRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM inbox_messages
		WHERE event_id = $1 AND status = 'processed'`

	var count int
	err := r.db.QueryRowContext(ctx, query, eventID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if event is processed: %w", err)
	}

	return count > 0, nil
}
