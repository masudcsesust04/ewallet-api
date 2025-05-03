package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masudcsesust04/ewallet-api/internal/db"
)

var (
	JWTSecretKey = []byte(os.Getenv("JWT_SECRET"))
)

// JWTMiddleware is a middleware to validate JWT token in Authorization header
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if string(JWTSecretKey) == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(JWTSecretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func GenerateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecretKey)
}

func GenerateRefreshToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		// fallback to less secure random string if needed
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}

func ValidateRefreshToken(refreshToken *db.RefreshToken) error {
	exp := refreshToken.ExpiresAt.Unix()

	if exp < time.Now().Unix() {
		return fmt.Errorf("refresh token has expired")
	}

	return nil
}
