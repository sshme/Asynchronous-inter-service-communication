package account

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type Account struct {
	ID        string
	UserID    string
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAccount(userID string) (*Account, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:        v7.String(),
		UserID:    userID,
		Balance:   0.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (a *Account) Credit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("credit amount must be positive")
	}

	a.Balance += amount
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Debit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("debit amount must be positive")
	}

	if a.Balance < amount {
		return fmt.Errorf("insufficient funds: balance %.2f, required %.2f", a.Balance, amount)
	}

	a.Balance -= amount
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) HasSufficientFunds(amount float64) bool {
	return a.Balance >= amount
}
