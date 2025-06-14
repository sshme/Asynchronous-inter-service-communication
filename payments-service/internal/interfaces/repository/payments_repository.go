package repository

import (
	"context"
	"database/sql"

	"payments-service/internal/domain/payments"
)

type PaymentsRepository interface {
	Store(ctx context.Context, payment *payments.Payment) error
	StoreWithTx(ctx context.Context, tx *sql.Tx, payment *payments.Payment) error
	GetByID(ctx context.Context, id string) (*payments.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*payments.Payment, error)
	Update(ctx context.Context, payment *payments.Payment) error
	UpdateWithTx(ctx context.Context, tx *sql.Tx, payment *payments.Payment) error
}
