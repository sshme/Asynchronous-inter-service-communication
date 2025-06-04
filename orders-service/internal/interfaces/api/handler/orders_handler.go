package handler

import (
	"encoding/json"
	"net/http"
	"orders-service/internal/application/service"
)

type OrdersHandler struct {
	ordersService *service.OrdersService
}

func NewOrdersHandler(ordersService *service.OrdersService) *OrdersHandler {
	return &OrdersHandler{ordersService: ordersService}
}

// CreateOrder handles new order requests
// @Summary Create a new order
// @Description Create a new order with the provided details
// @Tags orders
// @Router /orders [post]
func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.ordersService.CreateOrder(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
}
