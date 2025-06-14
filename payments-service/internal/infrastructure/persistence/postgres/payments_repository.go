package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"payments-service/internal/domain/payments"
	"payments-service/internal/interfaces/repository"
)

type PaymentsRepository struct {
	db *sql.DB
}

func NewPaymentsRepository(db *sql.DB) repository.PaymentsRepository {
	return &PaymentsRepository{db: db}
}

func (r *PaymentsRepository) Store(ctx context.Context, payment *payments.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, user_id, amount, currency, status, error_message, transaction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.UserID, payment.Amount, payment.Currency,
		payment.Status, payment.ErrorMessage, payment.TransactionID, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store payment: %w", err)
	}

	return nil
}

func (r *PaymentsRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, payment *payments.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, user_id, amount, currency, status, error_message, transaction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := tx.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.UserID, payment.Amount, payment.Currency,
		payment.Status, payment.ErrorMessage, payment.TransactionID, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store payment with tx: %w", err)
	}

	return nil
}

func (r *PaymentsRepository) GetByID(ctx context.Context, id string) (*payments.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, currency, status, error_message, transaction_id, created_at, updated_at
		FROM payments
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	payment := &payments.Payment{}
	err := row.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.ErrorMessage, &payment.TransactionID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get payment by ID: %w", err)
	}

	return payment, nil
}

func (r *PaymentsRepository) GetByOrderID(ctx context.Context, orderID string) (*payments.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, currency, status, error_message, transaction_id, created_at, updated_at
		FROM payments
		WHERE order_id = $1`

	row := r.db.QueryRowContext(ctx, query, orderID)

	payment := &payments.Payment{}
	err := row.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.ErrorMessage, &payment.TransactionID, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found for order: %s", orderID)
		}
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	return payment, nil
}

func (r *PaymentsRepository) Update(ctx context.Context, payment *payments.Payment) error {
	query := `
		UPDATE payments
		SET status = $2, error_message = $3, transaction_id = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.Status, payment.ErrorMessage, payment.TransactionID, payment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", payment.ID)
	}

	return nil
}

func (r *PaymentsRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, payment *payments.Payment) error {
	query := `
		UPDATE payments
		SET status = $2, error_message = $3, transaction_id = $4, updated_at = $5
		WHERE id = $1`

	result, err := tx.ExecContext(ctx, query,
		payment.ID, payment.Status, payment.ErrorMessage, payment.TransactionID, payment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update payment with tx: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", payment.ID)
	}

	return nil
}
