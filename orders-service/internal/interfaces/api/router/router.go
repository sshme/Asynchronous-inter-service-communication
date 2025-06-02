package router

import (
	"net/http"
	"orders-service/internal/interfaces/api/handler"
)

type Router struct {
	infoHandler *handler.InfoHandler
	docsHandler *handler.DocsHandler
}

func NewRouter() *Router {
	return &Router{
		infoHandler: handler.NewInfoHandler(),
		docsHandler: handler.NewDocsHandler(),
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /orders-api/info", r.infoHandler.HealthCheck)

	mux.HandleFunc("GET /orders-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /orders-api/docs/swagger.json", r.docsHandler.Swagger)

	return mux
}
