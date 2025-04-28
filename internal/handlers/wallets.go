package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/masudcsesust04/ewallet-api/internal/db"
)

type WalletDBInterface interface {
	GetWalletByID(walletID int64) (*db.Wallet, error)
	GetWalletByUserID(userID int64) (*db.Wallet, error)
	CreateWallet(userID int64) (*db.Wallet, error)
	UpdateWalletBalance(walletID int64, newBalance float64) error
	CreateTransaction(tx *db.Transaction) error
	TransferFunds(fromWalletID, toWalletID int64, amount float64) error
	GetTransactions(fromWalletID int64) ([]*db.Transaction, error)
}

type WalletHandler struct {
	DB WalletDBInterface
}

func NewWalletHandler(db *db.DB) *WalletHandler {
	return &WalletHandler{DB: db}
}

type WalletRequest struct {
	UserID   int64   `json:"user_id"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"curency"`
}

type DepositRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type TransferRequest struct {
	FromWalletID int64   `json:"from_wallet_id"`
	ToWalletID   int64   `json:"to_wallet_id"`
	Amount       float64 `json:"amount"`
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func RegisterWalletRoutes(r *mux.Router, db *db.DB) {
	handler := &WalletHandler{DB: db}
	r.HandleFunc("/wallets/new", handler.CreateNewWallet).Methods("POST")
	r.HandleFunc("/wallets/deposit", handler.Deposit).Methods("POST")
	r.HandleFunc("/wallets/withdraw", handler.Withdraw).Methods("POST")
	r.HandleFunc("/wallets/transfer", handler.Transfer).Methods("POST")
	r.HandleFunc("/wallets/balance", handler.Balance).Methods("GET")
	r.HandleFunc("/wallets/transactions/{walletID}", handler.Transactions).Methods("GET")
}

func (h *WalletHandler) CreateNewWallet(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *WalletHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *WalletHandler) Balance(w http.ResponseWriter, r *http.Request) {
	return
}

func (h *WalletHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	return
}
