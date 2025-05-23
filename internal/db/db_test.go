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
