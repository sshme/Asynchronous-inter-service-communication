package router

import (
	"net/http"
	"orders-service/internal/application/service"
	"orders-service/internal/interfaces/api/handler"
)

type Router struct {
	infoHandler   *handler.InfoHandler
	docsHandler   *handler.DocsHandler
	ordersHandler *handler.OrdersHandler
}

func NewRouter(ordersService *service.OrdersService) *Router {
	return &Router{
		infoHandler:   handler.NewInfoHandler(),
		docsHandler:   handler.NewDocsHandler(),
		ordersHandler: handler.NewOrdersHandler(ordersService),
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /orders-api/info", r.infoHandler.HealthCheck)

	mux.HandleFunc("GET /orders-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /orders-api/docs/swagger.json", r.docsHandler.Swagger)

	mux.HandleFunc("POST /orders-api/orders", r.ordersHandler.CreateOrder)

	return mux
}
