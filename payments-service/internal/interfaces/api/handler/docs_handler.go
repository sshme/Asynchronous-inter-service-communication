package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
)

type DocsHandler struct {
}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) Docs(w http.ResponseWriter, r *http.Request) {
	httpSwagger.Handler(httpSwagger.URL("/payments-api/docs/swagger.json"))(w, r)
}

func (h *DocsHandler) Swagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if doc, err := swag.ReadDoc("swagger"); err == nil {
		w.Write([]byte(doc))
		return
	}

	http.ServeFile(w, r, "./docs/swagger.json")
}
