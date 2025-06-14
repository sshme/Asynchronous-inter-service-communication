package handler

import (
	"encoding/json"
	"net/http"

	"orders-service/internal/application/service"
	"orders-service/internal/infrastructure/sse"
)

type OrdersHandler struct {
	ordersService *service.OrdersService
	sseManager    *sse.Manager
}

func NewOrdersHandler(ordersService *service.OrdersService, sseManager *sse.Manager) *OrdersHandler {
	return &OrdersHandler{
		ordersService: ordersService,
		sseManager:    sseManager,
	}
}

type CreateOrderRequest struct {
	UserID string `json:"user_id"`
}

// CreateOrder handles new order requests
// @Summary Create a new order
// @Description Create a new order with the provided details
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order creation request"
// @Success 200 {object} orders.Order
// @Router /orders [post]
func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	order, err := h.ordersService.CreateOrder(r.Context(), req.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// GetUserOrders получает список заказов пользователя
// @Summary Get user orders
// @Description Get all orders for a specific user
// @Tags Orders
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} orders.Order
// @Router /orders/user/{user_id} [get]
func (h *OrdersHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User ID is required"})
		return
	}

	orders, err := h.ordersService.GetUserOrders(r.Context(), userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get user orders"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

// GetOrderStatus получает статус заказа
// @Summary Get order status
// @Description Get the status of a specific order
// @Tags Orders
// @Produce json
// @Param order_id path string true "Order ID"
// @Success 200 {object} orders.Order
// @Router /orders/{order_id} [get]
func (h *OrdersHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Order ID is required"})
		return
	}

	order, err := h.ordersService.GetOrder(r.Context(), orderID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Order not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// StreamOrderUpdates обрабатывает SSE подключения для отслеживания обновлений заказов
// @Summary Stream order status updates
// @Description Establish SSE connection to receive real-time order status updates for a user
// @Tags Orders
// @Produce text/event-stream
// @Param user_id query string true "User ID to track orders for"
// @Success 200 {string} string "SSE stream of order updates"
// @Router /orders/stream [get]
func (h *OrdersHandler) StreamOrderUpdates(w http.ResponseWriter, r *http.Request) {
	h.sseManager.HandleSSE(w, r)
}
