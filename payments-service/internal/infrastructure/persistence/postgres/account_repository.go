package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"payments-service/internal/domain/account"
	"payments-service/internal/interfaces/repository"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) repository.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Store(ctx context.Context, acc *account.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		acc.ID, acc.UserID, acc.Balance, acc.CreatedAt, acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store account: %w", err)
	}

	return nil
}

func (r *AccountRepository) StoreWithTx(ctx context.Context, tx *sql.Tx, acc *account.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := tx.ExecContext(ctx, query,
		acc.ID, acc.UserID, acc.Balance, acc.CreatedAt, acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store account with tx: %w", err)
	}

	return nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID string) (*account.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1`

	row := r.db.QueryRowContext(ctx, query, userID)

	acc := &account.Account{}
	err := row.Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get account by user ID: %w", err)
	}

	return acc, nil
}

func (r *AccountRepository) GetByUserIDWithTx(ctx context.Context, tx *sql.Tx, userID string) (*account.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		FOR UPDATE`

	row := tx.QueryRowContext(ctx, query, userID)

	acc := &account.Account{}
	err := row.Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get account by user ID with tx: %w", err)
	}

	return acc, nil
}

func (r *AccountRepository) Update(ctx context.Context, acc *account.Account) error {
	query := `
		UPDATE accounts
		SET balance = $2, updated_at = $3
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, acc.ID, acc.Balance, acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found: %s", acc.ID)
	}

	return nil
}

func (r *AccountRepository) UpdateWithTx(ctx context.Context, tx *sql.Tx, acc *account.Account) error {
	query := `
		UPDATE accounts
		SET balance = $2, updated_at = $3
		WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, acc.ID, acc.Balance, acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update account with tx: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found: %s", acc.ID)
	}

	return nil
}
