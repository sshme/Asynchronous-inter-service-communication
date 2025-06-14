package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"payments-service/internal/application/service"
)

type AccountsHandler struct {
	accountService *service.AccountService
}

func NewAccountsHandler(accountService *service.AccountService) *AccountsHandler {
	return &AccountsHandler{
		accountService: accountService,
	}
}

type CreateAccountResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
}

type TopUpAccountRequest struct {
	Amount float64 `json:"amount"`
}

type TopUpAccountResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Balance   float64 `json:"balance"`
	UpdatedAt string  `json:"updated_at"`
}

type AccountInfoResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateAccount создает новый счет с автогенерированным user_id
// @Summary Create new account
// @Description Create a new account with auto-generated user_id using UUIDv7
// @Tags Accounts
// @Accept json
// @Produce json
// @Success 201 {object} CreateAccountResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts [post]
func (h *AccountsHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	account, err := h.accountService.CreateAccount(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create account"})
		return
	}

	response := CreateAccountResponse{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// TopUpAccount пополняет баланс счета
// @Summary Top up account balance
// @Description Add funds to an existing account
// @Tags Accounts
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body TopUpAccountRequest true "Top up request"
// @Success 200 {object} TopUpAccountResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /accounts/{user_id}/topup [post]
func (h *AccountsHandler) TopUpAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid URL path"})
		return
	}
	userID := parts[3]

	var req TopUpAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request format"})
		return
	}

	if req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Amount must be positive"})
		return
	}

	account, err := h.accountService.TopUpAccount(r.Context(), userID, req.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to top up account"})
		return
	}

	response := TopUpAccountResponse{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		UpdatedAt: account.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAccountInfo получает информацию о счете
// @Summary Get account information
// @Description Get account details by user ID
// @Tags Accounts
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} AccountInfoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /accounts/{user_id} [get]
func (h *AccountsHandler) GetAccountInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	urlPath := r.URL.Path
	parts := strings.Split(urlPath, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid URL path"})
		return
	}
	userID := parts[3]

	account, err := h.accountService.GetAccountInfo(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Account not found"})
		return
	}

	response := AccountInfoResponse{
		ID:        account.ID,
		UserID:    account.UserID,
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: account.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
