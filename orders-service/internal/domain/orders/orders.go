package orders

import (
	"time"

	"github.com/gofrs/uuid"
)

type Order struct {
	ID          string
	UserID      string
	Amount      float64
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewOrder(userID, description string, amount float64) (*Order, error) {
	v7, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Order{
		ID:          v7.String(),
		UserID:      userID,
		Amount:      amount,
		Description: description,
		Status:      "NEW",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
