package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"orders-service/internal/domain/orders"
	"orders-service/internal/interfaces/repository"
)

type OrdersRepository struct {
	db *sql.DB
}

func NewOrdersRepository(db *sql.DB) repository.OrdersRepository {
	return &OrdersRepository{db: db}
}

func (r *OrdersRepository) Store(ctx context.Context, order *orders.Order) error {
	query := `
		INSERT INTO orders (id, user_id, amount, currency, status, payment_id, error_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.Amount,
		order.Currency,
		order.Status,
		order.PaymentID,
		order.ErrorReason,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store order: %w", err)
	}

	return nil
}

func (r *OrdersRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, order *orders.Order) error {
	query := `
		INSERT INTO orders (id, user_id, amount, currency, status, payment_id, error_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := tx.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.Amount,
		order.Currency,
		order.Status,
		order.PaymentID,
		order.ErrorReason,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store order with tx: %w", err)
	}

	return nil
}

func (r *OrdersRepository) GetByID(ctx context.Context, orderID string) (*orders.Order, error) {
	query := `
		SELECT id, user_id, amount, currency, status, payment_id, error_reason, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, orderID)

	var order orders.Order
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.Currency,
		&order.Status,
		&order.PaymentID,
		&order.ErrorReason,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	return &order, nil
}

func (r *OrdersRepository) GetByUserID(ctx context.Context, userID string) ([]*orders.Order, error) {
	query := `
		SELECT id, user_id, amount, currency, status, payment_id, error_reason, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user ID: %w", err)
	}
	defer rows.Close()

	var ordersList []*orders.Order

	for rows.Next() {
		var order orders.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Amount,
			&order.Currency,
			&order.Status,
			&order.PaymentID,
			&order.ErrorReason,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		ordersList = append(ordersList, &order)
	}

	return ordersList, rows.Err()
}

func (r *OrdersRepository) Update(ctx context.Context, order *orders.Order) error {
	query := `
		UPDATE orders
		SET status = $2, payment_id = $3, error_reason = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.Status,
		order.PaymentID,
		order.ErrorReason,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", order.ID)
	}

	return nil
}

func (r *OrdersRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, order *orders.Order) error {
	query := `
		UPDATE orders
		SET status = $2, payment_id = $3, error_reason = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query,
		order.ID,
		order.Status,
		order.PaymentID,
		order.ErrorReason,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update order with tx: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", order.ID)
	}

	return nil
}

func (r *OrdersRepository) UpdateStatus(ctx context.Context, orderID string, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", orderID)
	}

	return nil
}
