package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"orders-service/internal/domain/orders"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOrdersService struct {
	mock.Mock
}

func (m *MockOrdersService) CreateOrder(ctx context.Context, userID string) (*orders.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orders.Order), args.Error(1)
}

func (m *MockOrdersService) GetUserOrders(ctx context.Context, userID string) ([]*orders.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orders.Order), args.Error(1)
}

func (m *MockOrdersService) GetOrder(ctx context.Context, orderID string) (*orders.Order, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orders.Order), args.Error(1)
}

func TestOrdersHandler_CreateOrder(t *testing.T) {
	mockService := new(MockOrdersService)
	handler := NewOrdersHandler(mockService, nil) // sseManager is not used in this handler

	t.Run("success", func(t *testing.T) {
		userID := "user-123"
		order := &orders.Order{ID: "order-456", UserID: userID}
		mockService.On("CreateOrder", mock.Anything, userID).Return(order, nil).Once()

		reqBody, _ := json.Marshal(CreateOrderRequest{UserID: userID})
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		handler.CreateOrder(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respOrder orders.Order
		err := json.Unmarshal(rr.Body.Bytes(), &respOrder)
		assert.NoError(t, err)
		assert.Equal(t, order, &respOrder)
		mockService.AssertExpectations(t)
	})

	t.Run("bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("invalid json"))
		rr := httptest.NewRecorder()

		handler.CreateOrder(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		userID := "user-123"
		mockService.On("CreateOrder", mock.Anything, userID).Return(nil, errors.New("service error")).Once()

		reqBody, _ := json.Marshal(CreateOrderRequest{UserID: userID})
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		handler.CreateOrder(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOrdersHandler_GetUserOrders(t *testing.T) {
	mockService := new(MockOrdersService)
	handler := NewOrdersHandler(mockService, nil)

	t.Run("success", func(t *testing.T) {
		userID := "user-123"
		expectedOrders := []*orders.Order{{ID: "order-1"}, {ID: "order-2"}}
		mockService.On("GetUserOrders", mock.Anything, userID).Return(expectedOrders, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/orders/user/"+userID, nil)
		req.SetPathValue("id", userID)
		rr := httptest.NewRecorder()

		handler.GetUserOrders(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respOrders []*orders.Order
		err := json.Unmarshal(rr.Body.Bytes(), &respOrders)
		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, respOrders)
		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		userID := "user-123"
		mockService.On("GetUserOrders", mock.Anything, userID).Return(nil, errors.New("service error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/orders/user/"+userID, nil)
		req.SetPathValue("id", userID)
		rr := httptest.NewRecorder()

		handler.GetUserOrders(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("success with no orders", func(t *testing.T) {
		userID := "user-789"
		expectedOrders := []*orders.Order{}
		mockService.On("GetUserOrders", mock.Anything, userID).Return(expectedOrders, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/orders/user/"+userID, nil)
		req.SetPathValue("id", userID)
		rr := httptest.NewRecorder()

		handler.GetUserOrders(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respOrders []*orders.Order
		err := json.Unmarshal(rr.Body.Bytes(), &respOrders)
		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, respOrders)
		mockService.AssertExpectations(t)
	})
}

func TestOrdersHandler_GetOrderStatus(t *testing.T) {
	mockService := new(MockOrdersService)
	handler := NewOrdersHandler(mockService, nil)

	t.Run("success", func(t *testing.T) {
		orderID := "order-456"
		expectedOrder := &orders.Order{ID: orderID}
		mockService.On("GetOrder", mock.Anything, orderID).Return(expectedOrder, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID, nil)
		req.SetPathValue("id", orderID)
		rr := httptest.NewRecorder()

		handler.GetOrderStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respOrder *orders.Order
		err := json.Unmarshal(rr.Body.Bytes(), &respOrder)
		assert.NoError(t, err)
		assert.Equal(t, expectedOrder, respOrder)
		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		orderID := "order-456"
		mockService.On("GetOrder", mock.Anything, orderID).Return(nil, errors.New("not found")).Once()

		req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID, nil)
		req.SetPathValue("id", orderID)
		rr := httptest.NewRecorder()

		handler.GetOrderStatus(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}
