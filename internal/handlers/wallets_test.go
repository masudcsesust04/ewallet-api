package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masudcsesust04/ewallet-api/internal/db"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type MockDB struct {
	Wallets      map[int64]*db.Wallet
	Transactions []*db.Transaction
}

var ErrNotFound = errors.New("not found")

func NewMockDB() *MockDB {
	return &MockDB{
		Wallets:      make(map[int64]*db.Wallet),
		Transactions: []*db.Transaction{},
	}
}

func generateMockJWTToken(userID int) (string, error) {
	// Replace with your JWT secret and claims structure
	secret := "your_jwt_secret"
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (m *MockDB) GetWalletByID(walletID int64) (*db.Wallet, error) {
	wallet, ok := m.Wallets[walletID]
	if !ok {
		return nil, ErrNotFound
	}
	return wallet, nil
}

func (m *MockDB) GetWalletByUserID(userID int64) (*db.Wallet, error) {
	wallet, ok := m.Wallets[userID]
	if !ok {
		return nil, ErrNotFound
	}
	return wallet, nil
}

func (m *MockDB) CreateWallet(userID int64) (*db.Wallet, error) {
	wallet := &db.Wallet{
		ID:        int64(len(m.Wallets) + 1),
		UserID:    userID,
		Balance:   0,
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.Wallets[userID] = wallet
	return wallet, nil
}

func (m *MockDB) UpdateWalletBalance(walletID int64, newBalance float64) error {
	for _, w := range m.Wallets {
		if w.ID == walletID {
			w.Balance = newBalance
			w.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockDB) CreateTransaction(txn *db.Transaction) error {
	m.Transactions = append(m.Transactions, txn)
	return nil
}

func (m *MockDB) GetTransactionsByWalletID(walletID int64) ([]*db.Transaction, error) {
	var txns []*db.Transaction
	for _, txn := range m.Transactions {
		if txn.FromWalletID == walletID {
			txns = append(txns, txn)
		}
	}
	return txns, nil
}

// TransferFunds performs an atomic transfer between two wallets
func (m *MockDB) TransferFunds(fromWalletID, toWalletID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	fromWallet, ok := m.Wallets[fromWalletID]
	if !ok {
		return fmt.Errorf("sender wallet not found")
	}

	toWallet, ok := m.Wallets[toWalletID]
	if !ok {
		return fmt.Errorf("receiver wallet not found")
	}

	if fromWallet.Balance < amount {
		return errors.New("insufficient funds")
	}

	fromWallet.Balance -= amount
	toWallet.Balance += amount

	now := time.Now()
	fromWallet.UpdatedAt = now
	toWallet.UpdatedAt = now

	m.Transactions = append(m.Transactions, &db.Transaction{
		FromWalletID: fromWalletID,
		Type:         "transfer",
		Amount:       amount,
		ToWalletID:   toWalletID,
		CreatedAt:    now,
	})
	m.Transactions = append(m.Transactions, &db.Transaction{
		FromWalletID: toWalletID,
		Type:         "transfer",
		Amount:       amount,
		ToWalletID:   fromWalletID,
		CreatedAt:    now,
	})

	return nil
}

func setupRouterWithMockDB(mockDB WalletDBInterface) *mux.Router {
	r := mux.NewRouter()
	handler := &WalletHandler{DB: mockDB}
	r.HandleFunc("/wallets/new", handler.CreateNewWallet).Methods("POST")
	r.HandleFunc("/wallets/deposit", handler.Deposit).Methods("POST")
	r.HandleFunc("/wallets/withdraw", handler.Withdraw).Methods("POST")
	r.HandleFunc("/wallets/transfer", handler.Transfer).Methods("POST")
	r.HandleFunc("/wallets/balance", handler.Balance).Methods("GET")
	r.HandleFunc("/wallets/transactions", handler.Transactions).Methods("GET")
	return r
}

func TestCreateNewWallet(t *testing.T) {
	mockDB := NewMockDB()
	r := setupRouterWithMockDB(mockDB)

	payload := map[string]interface{}{
		"user_id": 1,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/wallets/new", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	// assert.Contains(t, w.Body.String(), "Wallet created successfully")
	wallet, _ := mockDB.GetWalletByUserID(1)
	assert.NotNil(t, wallet)
}

func TestDeposit(t *testing.T) {
	mockDB := NewMockDB()
	r := setupRouterWithMockDB(mockDB)

	// Simulate user login and generate JWT token
	userID := 1
	token, err := generateMockJWTToken(userID) // Mock function to generate a JWT token
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	payload := map[string]interface{}{
		"user_id": userID,
		"amount":  100.0,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/wallets/deposit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token) // Add JWT token to the Authorization header
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Add assertions to verify the response
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Deposited successful")
	wallet, _ := mockDB.GetWalletByUserID(1)
	assert.Equal(t, 100.0, wallet.Balance)
}

func TestWithdraw(t *testing.T) {
	mockDB := NewMockDB()
	mockDB.CreateWallet(1)
	mockDB.UpdateWalletBalance(1, 100.0)
	r := setupRouterWithMockDB(mockDB)

	payload := map[string]interface{}{
		"user_id": 1,
		"amount":  50.0,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/wallets/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Withdrawal successful")
	wallet, _ := mockDB.GetWalletByUserID(1)
	assert.Equal(t, 50.0, wallet.Balance)
}

func TestTransfer(t *testing.T) {
	mockDB := NewMockDB()
	fromWallet, _ := mockDB.CreateWallet(1)
	toWallet, _ := mockDB.CreateWallet(2)
	mockDB.UpdateWalletBalance(1, 100.0)
	mockDB.UpdateWalletBalance(2, 20.0)
	r := setupRouterWithMockDB(mockDB)

	payload := map[string]interface{}{
		"from_wallet_id": fromWallet.ID,
		"to_wallet_id":   toWallet.ID,
		"amount":         30.0,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/wallets/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Transfered successfully")
	wallet1, _ := mockDB.GetWalletByUserID(1)
	wallet2, _ := mockDB.GetWalletByUserID(2)
	assert.Equal(t, 70.0, wallet1.Balance)
	assert.Equal(t, 50.0, wallet2.Balance)
}

func TestBalance(t *testing.T) {
	mockDB := NewMockDB()
	mockDB.CreateWallet(1)
	mockDB.UpdateWalletBalance(1, 150.0)
	r := setupRouterWithMockDB(mockDB)

	req := httptest.NewRequest("GET", "/wallets/balance?user_id=1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"balance":150`)
}

func TestTransactions(t *testing.T) {
	mockDB := NewMockDB()
	mockDB.CreateWallet(1)
	mockDB.UpdateWalletBalance(1, 150.0)
	mockDB.CreateTransaction(&db.Transaction{
		FromWalletID: 1,
		Type:         "deposit",
		Amount:       150.0,
		CreatedAt:    time.Now(),
	})
	mockDB.CreateTransaction(&db.Transaction{
		FromWalletID: 1,
		Type:         "withdrawal",
		Amount:       50.0,
		CreatedAt:    time.Now(),
	})

	r := setupRouterWithMockDB(mockDB)

	req := httptest.NewRequest("GET", "/wallets/transactions?wallet_id=1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"type":"deposit"`)
	assert.Contains(t, w.Body.String(), `"type":"withdrawal"`)
}
