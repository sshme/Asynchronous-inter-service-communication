package handler

import (
	"net/http"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
)

type DocsHandler struct {
}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) Docs(w http.ResponseWriter, r *http.Request) {
	httpSwagger.Handler(httpSwagger.URL("/orders-api/docs/swagger.json"))(w, r)
}

func (h *DocsHandler) Swagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if doc, err := swag.ReadDoc("swagger"); err == nil {
		w.Write([]byte(doc))
		return
	}

	http.ServeFile(w, r, "./docs/swagger.json")
}

func (h *DocsHandler) ScalarReference(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecURL:  "./docs/swagger.json",
		DarkMode: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(htmlContent))
}
