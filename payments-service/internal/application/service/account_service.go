package service

import (
	"context"
	"log"

	"payments-service/internal/domain/account"
	"payments-service/internal/interfaces/repository"

	"github.com/gofrs/uuid"
)

type AccountService struct {
	accountRepo repository.AccountRepository
}

func NewAccountService(accountRepo repository.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context) (*account.Account, error) {
	userIDUUID, err := uuid.NewV7()
	if err != nil {
		log.Printf("Error generating user ID: %v", err)
		return nil, err
	}

	userID := userIDUUID.String()

	acc, err := account.NewAccount(userID)
	if err != nil {
		log.Printf("Error creating account: %v", err)
		return nil, err
	}

	if err := s.accountRepo.Store(ctx, acc); err != nil {
		log.Printf("Error storing account: %v", err)
		return nil, err
	}

	log.Printf("Account created successfully: ID=%s, UserID=%s", acc.ID, acc.UserID)
	return acc, nil
}

func (s *AccountService) TopUpAccount(ctx context.Context, userID string, amount float64) (*account.Account, error) {
	acc, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error getting account for user %s: %v", userID, err)
		return nil, err
	}

	if err := acc.Credit(amount); err != nil {
		log.Printf("Error crediting account %s: %v", acc.ID, err)
		return nil, err
	}

	if err := s.accountRepo.Update(ctx, acc); err != nil {
		log.Printf("Error updating account %s: %v", acc.ID, err)
		return nil, err
	}

	log.Printf("Account topped up successfully: ID=%s, Amount=%.2f, New Balance=%.2f", acc.ID, amount, acc.Balance)
	return acc, nil
}

func (s *AccountService) GetAccountInfo(ctx context.Context, userID string) (*account.Account, error) {
	acc, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error getting account for user %s: %v", userID, err)
		return nil, err
	}

	log.Printf("Account info retrieved: ID=%s, UserID=%s, Balance=%.2f", acc.ID, acc.UserID, acc.Balance)
	return acc, nil
}
