package router

import (
	"net/http"
	"orders-service/internal/infrastructure/sse"
	"orders-service/internal/interfaces/api/handler"
)

type Router struct {
	infoHandler   *handler.InfoHandler
	docsHandler   *handler.DocsHandler
	ordersHandler *handler.OrdersHandler
}

func NewRouter(ordersService handler.OrdersServicer, sseManager *sse.Manager) *Router {
	return &Router{
		infoHandler:   handler.NewInfoHandler(),
		docsHandler:   handler.NewDocsHandler(),
		ordersHandler: handler.NewOrdersHandler(ordersService, sseManager),
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /orders-api/info", r.infoHandler.HealthCheck)

	mux.HandleFunc("GET /orders-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /orders-api/docs/swagger.json", r.docsHandler.Swagger)
	mux.HandleFunc("GET /orders-api/scalar", r.docsHandler.ScalarReference)

	mux.HandleFunc("POST /orders-api/orders", r.ordersHandler.CreateOrder)

	mux.HandleFunc("GET /orders-api/orders/{id}", r.ordersHandler.GetOrderStatus)
	mux.HandleFunc("GET /orders-api/orders/user/{id}", r.ordersHandler.GetUserOrders)

	// SSE endpoint for real-time order updates
	mux.HandleFunc("GET /orders-api/orders/stream", r.ordersHandler.StreamOrderUpdates)

	return mux
}
