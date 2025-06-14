package repository

import (
	"context"
	"database/sql"

	"payments-service/internal/domain/account"
)

type AccountRepository interface {
	Store(ctx context.Context, account *account.Account) error
	StoreWithTx(ctx context.Context, tx *sql.Tx, account *account.Account) error
	GetByUserID(ctx context.Context, userID string) (*account.Account, error)
	GetByUserIDWithTx(ctx context.Context, tx *sql.Tx, userID string) (*account.Account, error)
	Update(ctx context.Context, account *account.Account) error
	UpdateWithTx(ctx context.Context, tx *sql.Tx, account *account.Account) error
}
