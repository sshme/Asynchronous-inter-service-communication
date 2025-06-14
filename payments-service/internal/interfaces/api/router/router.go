package router

import (
	"net/http"
	"payments-service/internal/interfaces/api/handler"
)

type Router struct {
	infoHandler     *handler.InfoHandler
	docsHandler     *handler.DocsHandler
	accountsHandler *handler.AccountsHandler
}

func NewRouter(accountsHandler *handler.AccountsHandler) *Router {
	return &Router{
		infoHandler:     handler.NewInfoHandler(),
		docsHandler:     handler.NewDocsHandler(),
		accountsHandler: accountsHandler,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /payments-api/info", r.infoHandler.HealthCheck)

	mux.HandleFunc("GET /payments-api/docs/", r.docsHandler.Docs)
	mux.HandleFunc("GET /payments-api/docs/swagger.json", r.docsHandler.Swagger)

	mux.HandleFunc("POST /payments-api/accounts", r.accountsHandler.CreateAccount)
	mux.HandleFunc("GET /payments-api/accounts/", r.accountsHandler.GetAccountInfo) // /accounts/{user_id}
	mux.HandleFunc("POST /payments-api/accounts/", r.accountsHandler.TopUpAccount)  // /accounts/{user_id}/topup

	return mux
}
