package account

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	userID := "user-123"
	acc, err := NewAccount(userID)

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.NotEmpty(t, acc.ID)
	assert.Equal(t, userID, acc.UserID)
	assert.Equal(t, 0.0, acc.Balance)
	assert.WithinDuration(t, time.Now(), acc.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), acc.UpdatedAt, time.Second)
}

func TestAccount_Credit(t *testing.T) {
	acc, _ := NewAccount("user-123")

	err := acc.Credit(100.50)
	assert.NoError(t, err)
	assert.Equal(t, 100.50, acc.Balance)

	err = acc.Credit(50.25)
	assert.NoError(t, err)
	assert.Equal(t, 150.75, acc.Balance)

	err = acc.Credit(-10)
	assert.Error(t, err)
}

func TestAccount_Debit(t *testing.T) {
	acc, _ := NewAccount("user-123")
	_ = acc.Credit(200.0)

	err := acc.Debit(50.0)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, acc.Balance)

	err = acc.Debit(150.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, acc.Balance)

	err = acc.Debit(10.0)
	assert.Error(t, err)

	err = acc.Debit(-10.0)
	assert.Error(t, err)
}

func TestAccount_HasSufficientFunds(t *testing.T) {
	acc, _ := NewAccount("user-123")
	_ = acc.Credit(100.0)

	assert.True(t, acc.HasSufficientFunds(50.0))
	assert.True(t, acc.HasSufficientFunds(100.0))
	assert.False(t, acc.HasSufficientFunds(100.01))
}
