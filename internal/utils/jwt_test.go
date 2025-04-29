package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func TestJWTMiddleware(t *testing.T) {
	// Set the JWT_SECRET environment variable for testing
	os.Setenv("JWT_SECRET", "testsecretkey")
	defer os.Unsetenv("JWT_SECRET")

	// A simple handler to test middleware passing
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap the test handler with JWTMiddleware
	handler := JWTMiddleware(testHandler)

	// Test cases
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Authorization header format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalidtoken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid token",
			authHeader:     "Bearer " + generateTestToken(t, "testsecretkey"),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		if tc.authHeader != "" {
			req.Header.Set("Authorization", tc.authHeader)
		}
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != tc.expectedStatus {
			t.Errorf("%s: expected status %d, got %d", tc.name, tc.expectedStatus, rr.Code)
		}
	}
}

func generateTestToken(t *testing.T, secret string) string {
	t.Helper()
	claims := &Claims{
		UserID:           123,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return tokenString
}
