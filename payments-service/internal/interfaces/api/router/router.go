package router

import (
	"net/http"
	"payments-service/internal/interfaces/api/handler"
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

	mux.HandleFunc("GET /payments-api/info", r.infoHandler.HealthCheck)

	mux.HandleFunc("GET /payments-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /payments-api/docs/swagger.json", r.docsHandler.Swagger)

	return mux
}
