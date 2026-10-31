package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerService struct {
	db *pgxpool.Pool
	logger *slog.Logger
}

func NewLedgerService(pool *pgxpool.Pool,logger *slog.Logger) *LedgerService {
	return &LedgerService{
		db: pool,
		logger: logger,
	}

}
func (s *LedgerService) TransferFunds(ctx context.Context, fromAccountID, toAccountID uuid.UUID, amount int64) error {
	log := s.logger.With(
		slog.String("from_account", fromAccountID.String()),
		slog.String("to_account", toAccountID.String()),
		slog.Int64("amount", amount),
	)
	log.Info("initiating funds transfer")
	if amount <= 0 {
		log.Warn("rejected: invalid amount")
		return errors.New("transfer amount must be strictly positive")
	}
	if fromAccountID == toAccountID {
		log.Warn("rejected: self transfer attempted")
		return errors.New("cannot transfer to the same account")
	}
	// 1. The Transaction Block
	// We use an explicit transaction block so everything succeeds or fails together
	tx, err := s.db.Begin(ctx)
	if err != nil {
		log.Error("failed to open database transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Deterministic Lock Ordering (To prevent Deadlocks)
	firstID, secondID := fromAccountID, toAccountID
	if fromAccountID.String() > toAccountID.String() {
		firstID, secondID = toAccountID, fromAccountID
	}
	// Lock and read the cache in one move
	var firstBal, secondBal int64
	err = tx.QueryRow(ctx, "SELECT cached_balance FROM accounts WHERE id = $1 FOR UPDATE", firstID).Scan(&firstBal)
	if err != nil {
		log.Error("failed to lock first account", "error", err, "account_id", firstID)
		return fmt.Errorf("failed to lock first account: %w", err)
	}
	err = tx.QueryRow(ctx, "SELECT cached_balance FROM accounts WHERE id = $1 FOR UPDATE", secondID).Scan(&secondBal)
	if err != nil {
		log.Error("failed to lock second account", "error", err, "account_id", secondID)
		return fmt.Errorf("failed to lock second account: %w", err)
	}

	// 4. The Balance Check
	// Determine which balance belongs to the sender
	senderBalance := firstBal
	if fromAccountID == secondID {
		senderBalance = secondBal
	}

	if senderBalance < amount {
		return errors.New("insufficient funds")
	}
	// 5. The Ledger Inserts (The Append-Only Mutation)
	transactionID := uuid.New() // Generate the envelope ID in Go
	// Create the envelope
	_, err = tx.Exec(ctx, "INSERT into transactions (id,reference) VALUES ($1,$2)", transactionID, uuid.New().String())
	if err != nil {
		
		return fmt.Errorf("failed to create transaction envelope: %w", err)
	}
	// Insert the Debit (-amount)
	_, err = tx.Exec(ctx, "INSERT into ledger (id,transaction_id,account_id,amount) VALUES ($1,$2,$3,$4)", uuid.New(), transactionID, fromAccountID, -amount)
	if err != nil {
		return fmt.Errorf("failed to insert debit: %w", err)
	}
	// Insert the Credit (+amount)
	_, err = tx.Exec(ctx, "INSERT into ledger (id,transaction_id,account_id,amount) VALUES ($1,$2,$3,$4)", uuid.New(), transactionID, toAccountID, amount)
	if err != nil {
		return fmt.Errorf("failed to insert credit: %w", err)
	}
	// update the cached balance status for both accounts
	_, err = tx.Exec(ctx, "UPDATE accounts SET cached_balance = cached_balance - $1 where id = $2", amount, fromAccountID)
	if err != nil {
		return fmt.Errorf("failed to deduct from sender cache: %w", err)
	}
	_, err = tx.Exec(ctx, "UPDATE accounts SET cached_balance = cached_balance + $1 where id = $2", amount, toAccountID)
	if err != nil {
		return fmt.Errorf("failed to add to receiver cache: %w", err)
	}
	// 6. The Commit (The Database Bouncer wakes up here)
	if err := tx.Commit(ctx); err != nil {
		log.Error("transaction commit failed (trigger violation)", "error", err)
		return fmt.Errorf("transaction commit failed (trigger might have caught a zero-sum violation): %w", err)
	}
	log.Info("transfer committed successfully", "transaction_id", transactionID.String())
	return nil

}
