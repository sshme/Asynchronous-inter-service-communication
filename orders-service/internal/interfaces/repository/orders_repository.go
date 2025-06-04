package repository

import (
	"context"
	"orders-service/internal/domain/orders"
)

type OrdersRepository interface {
	Store(ctx context.Context, order *orders.Order) error
}
