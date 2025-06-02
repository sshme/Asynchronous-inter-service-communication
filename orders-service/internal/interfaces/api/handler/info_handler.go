package handler

import (
	"encoding/json"
	"net/http"
)

type InfoHandler struct {
}

func NewInfoHandler() *InfoHandler {
	return &InfoHandler{}
}

// HealthCheck handles the health check endpoint
// @Summary Health check endpoint
// @Description Check if the service is up and running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any "OK"
// @Router /info [get]
func (h *InfoHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"status": "OK",
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}
