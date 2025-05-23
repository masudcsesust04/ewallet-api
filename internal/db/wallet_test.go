package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndGetWallet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create user first
	user := &User{
		FirstName:   "Test",
		LastName:    "User",
		PhoneNumber: "1234567890",
		Email:       "testuser@example.com",
		Status:      "active",
		Password:    "password123",
	}
	err := db.CreateUser(user)
	assert.NoError(t, err)

	userID := int64(user.ID)

	// Create wallet
	wallet, err := db.CreateWallet(userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, wallet.UserID)
	assert.Equal(t, float64(0), wallet.Balance)

	// Get wallet
	gotWallet, err := db.GetWalletByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, wallet.ID, gotWallet.ID)
	assert.Equal(t, wallet.UserID, gotWallet.UserID)
}

func TestUpdateWalletBalance(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create user first
	user := &User{
		FirstName:   "Update",
		LastName:    "User",
		PhoneNumber: "0987654321",
		Email:       "updateuser@example.com",
		Status:      "active",
		Password:    "password123",
	}
	err := db.CreateUser(user)
	assert.NoError(t, err)

	userID := int64(user.ID)

	wallet, err := db.CreateWallet(userID)
	assert.NoError(t, err)

	newBalance := 123.45
	err = db.UpdateWalletBalance(wallet.ID, newBalance)
	assert.NoError(t, err)

	updatedWallet, err := db.GetWalletByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, newBalance, updatedWallet.Balance)
}

func TestCreateTransaction(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create user first
	user := &User{
		FirstName:   "Delete",
		LastName:    "User",
		PhoneNumber: "1112223333",
		Email:       "deleteuser@example.com",
		Status:      "active",
		Password:    "password123",
	}
	err := db.CreateUser(user)
	assert.NoError(t, err)

	userID := int64(user.ID)

	wallet, err := db.CreateWallet(userID)
	assert.NoError(t, err)

	txn := &Transaction{
		FromWalletID: wallet.ID,
		Type:         "deposit",
		Amount:       100.0,
		Status:       "completed",
		CreatedAt:    time.Now(),
	}

	err = db.CreateTransaction(txn)
	assert.NoError(t, err)
	assert.NotZero(t, txn.ID)
}

// setupTestDB initializes a test database connection and returns a cleanup function
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()
	// Use environment variable or test config for database URL
	// databaseURL := "postgresql://golangjwt:golangjwt@127.0.0.1:5435/golangjwt_test?sslmode=disable"

	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		panic("TEST_DATABASE_URL environment variable is not set")
	}

	db, err := NewDB(databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up tables before and after tests
	cleanup := func() {
		db.pool.Exec(context.Background(), "TRUNCATE TABLE transactions, wallets, users, refresh_tokens RESTART IDENTITY CASCADE")
		// Do not close db here to keep connection alive for tests
	}

	// Clean before starting
	cleanup()

	return db, cleanup
}
