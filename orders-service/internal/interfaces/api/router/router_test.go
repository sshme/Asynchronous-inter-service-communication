package router

import (
	"context"
	"io"
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

func TestRouter_SetupRoutes(t *testing.T) {
	mockOrdersService := new(MockOrdersService)

	router := NewRouter(mockOrdersService, nil)
	server := httptest.NewServer(router.SetupRoutes())
	defer server.Close()

	testCases := []struct {
		name       string
		method     string
		path       string
		statusCode int
	}{
		{"HealthCheck", http.MethodGet, "/orders-api/info", http.StatusOK},
		{"Docs", http.MethodGet, "/orders-api/docs/", http.StatusOK},
		{"Swagger", http.MethodGet, "/orders-api/docs/swagger.json/", http.StatusOK},
		{"CreateOrder", http.MethodPost, "/orders-api/orders", http.StatusBadRequest},
		{"GetOrderStatus", http.MethodGet, "/orders-api/orders/some-id", http.StatusNotFound},
		{"GetUserOrders", http.MethodGet, "/orders-api/orders/user/some-id", http.StatusInternalServerError},
	}

	mockOrdersService.On("GetOrder", mock.Anything, "some-id").Return(nil, assert.AnError)
	mockOrdersService.On("GetUserOrders", mock.Anything, "some-id").Return(nil, assert.AnError)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, server.URL+tc.path, nil)
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tc.statusCode, resp.StatusCode)
			_, _ = io.Copy(io.Discard, resp.Body)
		})
	}
}
