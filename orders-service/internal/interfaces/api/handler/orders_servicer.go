package handler

import (
	"context"
	"orders-service/internal/domain/orders"
)

type OrdersServicer interface {
	CreateOrder(ctx context.Context, userID string) (*orders.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]*orders.Order, error)
	GetOrder(ctx context.Context, orderID string) (*orders.Order, error)
}
