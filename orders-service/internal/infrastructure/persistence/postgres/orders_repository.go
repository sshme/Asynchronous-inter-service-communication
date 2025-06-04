package postgres

import (
	"context"
	"database/sql"
	"orders-service/internal/domain/orders"
)

type OrdersRepository struct {
	db *sql.DB
}

func NewOrdersRepository(db *sql.DB) *OrdersRepository {
	return &OrdersRepository{db: db}
}

func (r *OrdersRepository) Store(ctx context.Context, order *orders.Order) error {
	return nil
}
