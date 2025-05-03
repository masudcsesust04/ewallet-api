package db

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Wallet represents a user's wallet
type Wallet struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	FromWalletID int64     `json:"from_wallet_id"`
	ToWalletID   int64     `json:"to_wallet_id"`
	Amount       float64   `json:"amount"`
	Fee          float64   `json:"fee"`
	Note         string    `json:"note"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func (db *DB) GetWalletByID(id int64) (*Wallet, error) {
	query := `SELECT * FROM  wallets WHERE id = $1`
	wallet := &Wallet{}

	err := db.pool.QueryRow(context.Background(), query, id).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet by id: %w", err)
	}

	return wallet, nil
}

func (db *DB) GetWalletByUserID(userID int64) (*Wallet, error) {
	query := `SELECT * FROM  wallets WHERE user_id = $1`
	wallet := &Wallet{}

	err := db.pool.QueryRow(context.Background(), query, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet by user id: %w", err)
	}

	return wallet, nil
}

func (db *DB) CreateWallet(userID int64) (*Wallet, error) {
	query := `INSERT INTO wallets (user_id, balance, currency) VALUES ($1, $2, $3) RETURNING id, user_id, balance, currency, created_at, updated_at`
	wallet := &Wallet{}

	err := db.pool.QueryRow(context.Background(), query, userID, 0.0, "USD").Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to creaqte wallet by user id: %w", err)
	}

	return wallet, nil
}

func (db *DB) UpdateWalletBalance(walletID int64, newBalance float64) error {
	query := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`

	_, err := db.pool.Exec(context.Background(), query, newBalance, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return nil
}

func (db *DB) TransferFunds(fromWalletID, toWalletID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive number.")
	}

	ctx := context.Background()
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	var fromBalance float64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, fromWalletID).Scan(&fromBalance)
	if err != nil {
		return fmt.Errorf("failed to get sender wallet balance: %w", err)
	}

	if fromBalance < amount {
		return errors.New("insufficient account balance")
	}

	var toBalance float64
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, toWalletID).Scan(&toBalance)
	if err != nil {
		return fmt.Errorf("failed to get receiver wallet balance: %w", err)
	}

	// update balance
	_, err = tx.Exec(ctx, `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`, fromBalance-amount, fromWalletID)
	if err != nil {
		return fmt.Errorf("failed to debit sender wallet: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`, toBalance+amount, toWalletID)
	if err != nil {
		return fmt.Errorf("failed to credit receiver wallet: %w", err)
	}

	// Save transactions
	_, err = tx.Exec(ctx, `INSERT INTO transactions (type, from_wallet_id, to_wallet_id, amount, status) VALUES ('send', $1, $2, $3, $4)`, fromWalletID, toWalletID, amount, "completed")
	if err != nil {
		return fmt.Errorf("failed to log sender transaction: %w", err)
	}

	_, err = tx.Exec(ctx, `INSERT INTO transactions (type, from_wallet_id, to_wallet_id, amount, status) VALUES ('receive', $1, $2, $3, $4)`, toWalletID, fromWalletID, amount, "completed")
	if err != nil {
		return fmt.Errorf("failed to log receiver transaction: %w", err)
	}

	return nil
}

// CreateTransaction inserts a new tranaction into the database
func (db *DB) CreateTransaction(tx *Transaction) error {
	query := `INSERT INTO transactions (type, from_wallet_id, to_wallet_id, amount, fee, note, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING ID`
	err := db.pool.QueryRow(context.Background(), query, tx.Type, tx.FromWalletID, tx.ToWalletID, tx.Amount, tx.Fee, tx.Note, tx.Status).Scan(&tx.ID)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (db *DB) GetTransactionsByWalletID(fromWalletID int64) ([]*Transaction, error) {
	query := `SELECT id, type, from_wallet_id, to_wallet_id, amount, fee, status, created_at FROM transactions WHERE from_wallet_id = $1 ORDER BY created_at DESC`
	rows, err := db.pool.Query(context.Background(), query, fromWalletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(&tx.ID, &tx.Type, &tx.FromWalletID, &tx.ToWalletID, &tx.Amount, &tx.Fee, &tx.Status, &tx.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return transactions, nil
}
