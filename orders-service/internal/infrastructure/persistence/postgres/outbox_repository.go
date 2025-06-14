package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"orders-service/internal/domain/outbox"
	"orders-service/internal/interfaces/repository"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) repository.OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) StoreMessage(ctx context.Context, tx *sql.Tx, message *outbox.OutboxMessage) error {
	query := `
		INSERT INTO outbox_messages (id, event_type, payload, status, sent_at, created_at, updated_at, retry_count, max_retries)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := tx.ExecContext(ctx, query,
		message.ID,
		message.EventType,
		message.Payload,
		string(message.Status),
		message.SentAt,
		message.CreatedAt,
		message.UpdatedAt,
		message.RetryCount,
		message.MaxRetries,
	)

	if err != nil {
		return fmt.Errorf("failed to store outbox message: %w", err)
	}

	return nil
}

func (r *OutboxRepository) GetPendingMessages(ctx context.Context, limit int) ([]*outbox.OutboxMessage, error) {
	query := `
		SELECT id, event_type, payload, status, sent_at, created_at, updated_at, retry_count, max_retries
		FROM outbox_messages
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending outbox messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

func (r *OutboxRepository) MarkAsSent(ctx context.Context, messageID string) error {
	query := `
		UPDATE outbox_messages
		SET status = 'sent', sent_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as sent: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outbox message not found: %s", messageID)
	}

	return nil
}

func (r *OutboxRepository) MarkAsFailed(ctx context.Context, messageID string) error {
	query := `
		UPDATE outbox_messages
		SET status = 'failed', retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox message as failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("outbox message not found: %s", messageID)
	}

	return nil
}

func (r *OutboxRepository) GetFailedMessages(ctx context.Context, maxRetries int, limit int) ([]*outbox.OutboxMessage, error) {
	query := `
		SELECT id, event_type, payload, status, sent_at, created_at, updated_at, retry_count, max_retries
		FROM outbox_messages
		WHERE status = 'failed' AND retry_count < $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed outbox messages: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

func (r *OutboxRepository) scanMessages(rows *sql.Rows) ([]*outbox.OutboxMessage, error) {
	var messages []*outbox.OutboxMessage

	for rows.Next() {
		var message outbox.OutboxMessage
		var status string
		var sentAt sql.NullTime

		err := rows.Scan(
			&message.ID,
			&message.EventType,
			&message.Payload,
			&status,
			&sentAt,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.RetryCount,
			&message.MaxRetries,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox message: %w", err)
		}

		message.Status = outbox.OutboxMessageStatus(status)
		if sentAt.Valid {
			message.SentAt = &sentAt.Time
		}

		messages = append(messages, &message)
	}

	return messages, rows.Err()
}
