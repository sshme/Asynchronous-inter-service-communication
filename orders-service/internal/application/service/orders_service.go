package service

import (
	"context"
	"orders-service/internal/domain/orders"
	"orders-service/internal/interfaces/repository"
)

type OrdersService struct {
	ordersRepository repository.OrdersRepository
}

func NewOrdersService(ordersRepository repository.OrdersRepository) *OrdersService {
	return &OrdersService{ordersRepository: ordersRepository}
}

func (s *OrdersService) CreateOrder(ctx context.Context) (*orders.Order, error) {
	order, err := orders.NewOrder("1", "Test order", 100)
	if err != nil {
		return nil, err
	}

	err = s.ordersRepository.Store(ctx, order)
	if err != nil {
		return nil, err
	}

	return order, nil
}
