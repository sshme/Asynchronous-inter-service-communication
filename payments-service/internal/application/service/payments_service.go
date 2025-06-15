package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"

	"payments-service/internal/domain/account"
	"payments-service/internal/domain/inbox"
	"payments-service/internal/domain/outbox"
	"payments-service/internal/domain/payments"
	"payments-service/internal/interfaces/repository"
	"payments-service/pkg/random"
)

type DBTX interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type PaymentsService struct {
	db                  DBTX
	paymentsRepo        repository.PaymentsRepository
	accountRepo         repository.AccountRepository
	inboxRepo           repository.InboxRepository
	outboxRepo          repository.OutboxRepository
	randomGenerator     random.Generator
	maxRetries          int
	retryDelay          time.Duration
	outboxProcessorStop chan bool
}

func NewPaymentsService(
	db DBTX,
	paymentsRepo repository.PaymentsRepository,
	accountRepo repository.AccountRepository,
	inboxRepo repository.InboxRepository,
	outboxRepo repository.OutboxRepository,
	randomGenerator random.Generator,
) *PaymentsService {
	return &PaymentsService{
		db:                  db,
		paymentsRepo:        paymentsRepo,
		accountRepo:         accountRepo,
		inboxRepo:           inboxRepo,
		outboxRepo:          outboxRepo,
		randomGenerator:     randomGenerator,
		maxRetries:          3,
		retryDelay:          5 * time.Second,
		outboxProcessorStop: make(chan bool),
	}
}

func (s *PaymentsService) ProcessOrderCreated(ctx context.Context, inboxMessage *inbox.InboxMessage) error {
	var orderEvent inbox.OrderCreatedEvent
	if err := json.Unmarshal(inboxMessage.Payload, &orderEvent); err != nil {
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	log.Printf("Processing order created event: OrderID=%s, UserID=%s, Amount=%.2f",
		orderEvent.OrderID, orderEvent.UserID, orderEvent.Amount)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var payment *payments.Payment
	existingPayment, err := s.paymentsRepo.GetByOrderID(ctx, orderEvent.OrderID)
	if err != nil {
		if err == sql.ErrNoRows || fmt.Sprintf("%s", err) == fmt.Sprintf("payment not found for order: %s", orderEvent.OrderID) {
			payment, err = payments.NewPayment(orderEvent.OrderID, orderEvent.UserID, orderEvent.Amount, orderEvent.Currency)
			if err != nil {
				return fmt.Errorf("failed to create payment: %w", err)
			}

			if err := s.paymentsRepo.StoreWithTx(ctx, tx, payment); err != nil {
				return fmt.Errorf("failed to store payment: %w", err)
			}
			log.Printf("Created new payment: PaymentID=%s", payment.ID)
		} else {
			return fmt.Errorf("failed to check existing payment: %w", err)
		}
	} else {
		payment = existingPayment
		log.Printf("Found existing payment: PaymentID=%s, Status=%s", payment.ID, payment.Status)
	}

	success, shouldRetry, errorMessage, err := s.processPayment(ctx, tx, payment)
	if err != nil {
		return err
	}

	if shouldRetry {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		log.Printf("Payment retry scheduled: PaymentID=%s, Reason=%s", payment.ID, errorMessage)
		return fmt.Errorf("insufficient funds, will retry later: %s", errorMessage)
	}

	var outboxMessage *outbox.OutboxMessage
	if success {
		transactionID := uuid.Must(uuid.NewV7()).String()
		payment.Complete(transactionID)

		if err := s.paymentsRepo.UpdateWithTx(ctx, tx, payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		paymentEvent := outbox.PaymentCompletedEvent{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			UserID:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			TransactionID: payment.TransactionID,
		}

		payload, err := json.Marshal(paymentEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal payment completed event: %w", err)
		}

		outboxMessage, err = outbox.NewOutboxMessage("payment.completed", payload)
		if err != nil {
			return fmt.Errorf("failed to create outbox message: %w", err)
		}

		log.Printf("Payment completed successfully: PaymentID=%s, TransactionID=%s", payment.ID, transactionID)
	} else {
		payment.Fail(errorMessage)

		if err := s.paymentsRepo.UpdateWithTx(ctx, tx, payment); err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		paymentEvent := outbox.PaymentFailedEvent{
			PaymentID:    payment.ID,
			OrderID:      payment.OrderID,
			UserID:       payment.UserID,
			Amount:       payment.Amount,
			Currency:     payment.Currency,
			ErrorMessage: errorMessage,
		}

		payload, err := json.Marshal(paymentEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal payment failed event: %w", err)
		}

		outboxMessage, err = outbox.NewOutboxMessage("payment.failed", payload)
		if err != nil {
			return fmt.Errorf("failed to create outbox message: %w", err)
		}

		log.Printf("Payment failed: PaymentID=%s, Error=%s", payment.ID, errorMessage)
	}

	if err := s.outboxRepo.StoreWithTx(ctx, tx, outboxMessage); err != nil {
		return fmt.Errorf("failed to store outbox message: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully processed order created event and stored outbox message")
	return nil
}

// processPayment обрабатывает платеж со счета пользователя.
// returns: success, shouldRetry, errorMessage, error
func (s *PaymentsService) processPayment(ctx context.Context, tx *sql.Tx, payment *payments.Payment) (bool, bool, string, error) {
	if payment.IsTimedOut() {
		return false, false, "Payment timed out", nil
	}

	acc, err := s.accountRepo.GetByUserID(ctx, payment.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, false, "User account not found", nil
		}
		return false, false, fmt.Errorf("failed to get user account: %w", err).Error(), nil
	}

	if !acc.HasSufficientFunds(payment.Amount) {
		return false, true, fmt.Sprintf("Insufficient funds: balance %.2f, required %.2f", acc.Balance, payment.Amount), nil
	}

	if err := acc.Debit(payment.Amount); err != nil {
		return false, false, "", fmt.Errorf("failed to debit account: %w", err)
	}

	if err := s.accountRepo.UpdateWithTx(ctx, tx, acc); err != nil {
		return false, false, "", fmt.Errorf("failed to update account: %w", err)
	}

	log.Printf("Successfully debited %.2f from account: UserID=%s, NewBalance=%.2f",
		payment.Amount, payment.UserID, acc.Balance)

	return true, false, "", nil
}

func (s *PaymentsService) GetPaymentByOrderID(ctx context.Context, orderID string) (*payments.Payment, error) {
	return s.paymentsRepo.GetByOrderID(ctx, orderID)
}

func (s *PaymentsService) GetPaymentByID(ctx context.Context, paymentID string) (*payments.Payment, error) {
	return s.paymentsRepo.GetByID(ctx, paymentID)
}

func (s *PaymentsService) CreateAccount(ctx context.Context, userID string, initialBalance float64) (*account.Account, error) {
	acc, err := account.NewAccount(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	if initialBalance > 0 {
		if err := acc.Credit(initialBalance); err != nil {
			return nil, fmt.Errorf("failed to credit initial balance: %w", err)
		}
	}

	if err := s.accountRepo.Store(ctx, acc); err != nil {
		return nil, fmt.Errorf("failed to store account: %w", err)
	}

	return acc, nil
}

func (s *PaymentsService) GetAccountByUserID(ctx context.Context, userID string) (*account.Account, error) {
	return s.accountRepo.GetByUserID(ctx, userID)
}
