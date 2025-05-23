package db

import (
	"context"
	"os"
	"testing"
)

var testDB *DB

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		panic("TEST_DATABASE_URL environment variable is not set")
	}

	var err error
	testDB, err = NewDB(databaseURL)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	// Clean tabels before running tests
	_, err = testDB.pool.Exec(context.Background(), "TRUNCATE TABLE transactions, wallets, refresh_tokens, users RESTART IDENTITY CASCADE;")
	if err != nil {
		panic("failed to truncate tables: " + err.Error())
	}

	code := m.Run()
	testDB.Close()

	os.Exit(code)
}

// setupTestDB initializes a test database connection and returns a cleanup function
func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

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
