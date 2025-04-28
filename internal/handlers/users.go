package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/masudcsesust04/ewallet-api/internal/db"
)

type UserDBInterface interface {
	GetAllUsers() ([]*db.User, error)
	GetUserByID(id int) (*db.User, error)
	GetUserByEmail(emaio string) (*db.User, error)
	CreateUser(user *db.User) error
	UpdateUser(user *db.User) error
	DeleteUser(id int) error
	CreateRefreshToken(refreshToken *db.RefreshToken) error
	DeleteRefreshToken(token string) error
}

type UserHandler struct {
	DB UserDBInterface
}

func NewUserHandler(db *db.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.DB.GetAllUsers()
	if err != nil {
		http.Error(w, "failed to get users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) CreateUsers(w http.ResponseWriter, r *http.Request) {
	var user db.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
		return
	}

	err := h.DB.CreateUser(&user)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUser handle GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	id, err := strconv.Atoi(userIdStr)

	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
	}

	user, err := h.DB.GetUserByID(id)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	id, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
	}

	var user db.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user.ID = id
	err = h.DB.UpdateUser(&user)
	if err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	id, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
	}

	err = h.DB.DeleteUser(id)
	if err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
