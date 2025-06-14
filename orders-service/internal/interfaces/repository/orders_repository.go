package repository

import (
	"context"
	"database/sql"
	"orders-service/internal/domain/orders"
)

type OrdersRepository interface {
	Store(ctx context.Context, order *orders.Order) error
	StoreWithTx(ctx context.Context, tx *sql.Tx, order *orders.Order) error
	GetByID(ctx context.Context, orderID string) (*orders.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*orders.Order, error)
	Update(ctx context.Context, order *orders.Order) error
	UpdateWithTx(ctx context.Context, tx *sql.Tx, order *orders.Order) error
	UpdateStatus(ctx context.Context, orderID string, status string) error
}
