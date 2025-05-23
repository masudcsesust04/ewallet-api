package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/masudcsesust04/ewallet-api/internal/db"
	"github.com/masudcsesust04/ewallet-api/internal/utils"
)

type WalletDBInterface interface {
	GetWalletByID(walletID int64) (*db.Wallet, error)
	GetWalletByUserID(userID int64) (*db.Wallet, error)
	CreateWallet(userID int64) (*db.Wallet, error)
	UpdateWalletBalance(walletID int64, newBalance float64) error
	CreateTransaction(tx *db.Transaction) error
	TransferFunds(fromWalletID, toWalletID int64, amount float64) error
	GetTransactionsByWalletID(fromWalletID int64) ([]*db.Transaction, error)
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

type WithdrawRequest struct {
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
	r.HandleFunc("/wallets/new", utils.JWTMiddleware(handler.CreateNewWallet)).Methods("POST")
	r.HandleFunc("/wallets/deposit", utils.JWTMiddleware(handler.Deposit)).Methods("POST")
	r.HandleFunc("/wallets/withdraw", utils.JWTMiddleware(handler.Withdraw)).Methods("POST")
	r.HandleFunc("/wallets/transfer", utils.JWTMiddleware(handler.Transfer)).Methods("POST")
	r.HandleFunc("/wallets/balance", utils.JWTMiddleware(handler.Balance)).Methods("GET")
	r.HandleFunc("/wallets/transactions", utils.JWTMiddleware(handler.Transactions)).Methods("GET")
}

func (h *WalletHandler) CreateNewWallet(w http.ResponseWriter, r *http.Request) {
	var req WalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	wallet, err := h.DB.CreateWallet(req.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create wallet")
		return
	}

	newBalance := wallet.Balance + req.Balance
	err = h.DB.UpdateWalletBalance(wallet.ID, newBalance)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update wallet balance")
		return
	}

	wallet.Balance = newBalance
	tx := &db.Transaction{
		FromWalletID: wallet.ID,
		Type:         "deposit",
		Amount:       req.Balance,
		Status:       "completed",
		CreatedAt:    time.Now(),
	}

	err = h.DB.CreateTransaction(tx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to log transaction")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wallet)
}

func (h *WalletHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	wallet, err := h.DB.GetWalletByUserID(req.UserID)
	if err != nil {
		// If wallet not found, create one
		wallet, err = h.DB.CreateWallet(req.UserID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create wallet")
			return
		}
	}

	newBalance := wallet.Balance + req.Amount
	err = h.DB.UpdateWalletBalance(wallet.ID, newBalance)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update wallet balance")
		return
	}

	tx := &db.Transaction{
		FromWalletID: wallet.ID,
		Type:         "deposit",
		Amount:       req.Amount,
		Status:       "completed",
		CreatedAt:    time.Now(),
	}

	err = h.DB.CreateTransaction(tx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to log transaction")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Deposited successfully"})
}

func (h *WalletHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	wallet, err := h.DB.GetWalletByUserID(req.UserID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Wallet not found")
		return
	}

	if wallet.Balance < req.Amount {
		respondError(w, http.StatusBadRequest, "Insufficient funds")
		return
	}

	newBalance := wallet.Balance - req.Amount
	err = h.DB.UpdateWalletBalance(wallet.ID, newBalance)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update wallet balance")
		return
	}

	tx := &db.Transaction{
		FromWalletID: wallet.ID,
		Type:         "withdrawal",
		Amount:       req.Amount,
		Status:       "completed",
		CreatedAt:    time.Now(),
	}

	err = h.DB.CreateTransaction(tx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to log transaction")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Withdrawal successfully"})
}

func (h *WalletHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	if req.FromWalletID == req.ToWalletID {
		respondError(w, http.StatusBadRequest, "Can not transfer to the same wallet account")
		return
	}

	// Use atomic transfer function in DB layer
	err := h.DB.TransferFunds(req.FromWalletID, req.ToWalletID, req.Amount)
	if err != nil {
		if err.Error() == "insufficient funds" {
			respondError(w, http.StatusBadRequest, "insufficient funds")
		} else {
			respondError(w, http.StatusInternalServerError, "Failed to perform transfer: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Transfered successfully."})
}

func (h *WalletHandler) Balance(w http.ResponseWriter, r *http.Request) {
	userIDParam := r.URL.Query().Get("user_id")
	if userIDParam == "" {
		respondError(w, http.StatusBadRequest, "Missing user_id query parameter")
		return
	}

	userID, err := strconv.ParseInt(userIDParam, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id query parameter")
		return
	}

	wallet, err := h.DB.GetWalletByUserID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Wallet not found")
		return
	}

	json.NewEncoder(w).Encode(wallet)
}

func (h *WalletHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	walletIDParam := r.URL.Query().Get("wallet_id")
	if walletIDParam == "" {
		respondError(w, http.StatusBadRequest, "Missing wallet_id query parameter")
		return
	}

	walletID, err := strconv.ParseInt(walletIDParam, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid wallet_id query parameter")
		return
	}

	transactions, err := h.DB.GetTransactionsByWalletID(walletID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get transactions")
		return
	}

	respondJSON(w, http.StatusOK, transactions)
}
