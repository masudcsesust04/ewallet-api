package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masudcsesust04/ewallet-api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var jwtKey = []byte(os.Getenv("JWT_SECRET")) // Replace with your secret key

// Claims represents the JWT claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// generateRandomToken generates a secure random string for refresh tokens
func generateRandomToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// fallback to less secure random string if needed
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// Login handles POST /login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Println(req)

	user, err := h.DB.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Println(user)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	fmt.Println(err)

	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshTokenString := generateRandomToken()
	refreshToken := &db.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	fmt.Printf("Creating refresh token: %+v\n", refreshToken)
	err = h.DB.CreateRefreshToken(refreshToken)
	if err != nil {
		fmt.Printf("Error creating refresh token: %v\n", err)
		http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
		return
	}

	resp := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}

	json.NewEncoder(w).Encode(resp)
}

// Logout handles POST /logout
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// For logout, typically the client deletes the tokens.
	// Here, we expect the refresh token in the request body to delete it from DB.

	type LogoutRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := h.DB.DeleteRefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Failed to logout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
