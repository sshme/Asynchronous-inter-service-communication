package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"payments-service/internal/domain/account"
)

func TestAccountService_CreateAccount(t *testing.T) {
	mockAccountRepo := new(MockAccountRepository)
	service := NewAccountService(mockAccountRepo)
	ctx := context.Background()

	mockAccountRepo.On("Store", ctx, mock.AnythingOfType("*account.Account")).Return(nil)

	acc, err := service.CreateAccount(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	mockAccountRepo.AssertExpectations(t)
}

func TestAccountService_TopUpAccount(t *testing.T) {
	mockAccountRepo := new(MockAccountRepository)
	service := NewAccountService(mockAccountRepo)
	ctx := context.Background()
	userID := "user-123"
	amount := 100.50

	existingAccount, _ := account.NewAccount(userID)
	mockAccountRepo.On("GetByUserID", ctx, userID).Return(existingAccount, nil)
	mockAccountRepo.On("Update", ctx, existingAccount).Return(nil)

	acc, err := service.TopUpAccount(ctx, userID, amount)

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, amount, acc.Balance)
	mockAccountRepo.AssertExpectations(t)
}

func TestAccountService_GetAccountInfo(t *testing.T) {
	mockAccountRepo := new(MockAccountRepository)
	service := NewAccountService(mockAccountRepo)
	ctx := context.Background()
	userID := "user-123"

	existingAccount, _ := account.NewAccount(userID)
	mockAccountRepo.On("GetByUserID", ctx, userID).Return(existingAccount, nil)

	acc, err := service.GetAccountInfo(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, existingAccount.ID, acc.ID)
	mockAccountRepo.AssertExpectations(t)
}

func TestAccountService_CreateAccount_StoreError(t *testing.T) {
	mockAccountRepo := new(MockAccountRepository)
	service := NewAccountService(mockAccountRepo)
	ctx := context.Background()

	expectedErr := errors.New("store error")
	mockAccountRepo.On("Store", ctx, mock.AnythingOfType("*account.Account")).Return(expectedErr)

	acc, err := service.CreateAccount(ctx)

	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Equal(t, expectedErr, err)
	mockAccountRepo.AssertExpectations(t)
}
