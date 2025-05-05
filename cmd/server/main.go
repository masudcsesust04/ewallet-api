package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/masudcsesust04/ewallet-api/config"
	"github.com/masudcsesust04/ewallet-api/internal/db"
	"github.com/masudcsesust04/ewallet-api/internal/handlers"
	"github.com/masudcsesust04/ewallet-api/internal/utils"
)

func main() {
	// Load environment variables
	loadEnv()

	// Initialize database connection
	dbConn, err := db.NewDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %w", err)
	}
	defer dbConn.Close()

	// Initialize user handler
	userHandler := handlers.NewUserHandler(dbConn)

	// Setup router

	router := mux.NewRouter()
	// user routes
	router.HandleFunc("/users", utils.JWTMiddleware(userHandler.GetUsers)).Methods("GET")
	router.HandleFunc("/users", userHandler.CreateUsers).Methods("POST")
	router.HandleFunc("/users/{id}", utils.JWTMiddleware(userHandler.GetUser)).Methods("GET")
	router.HandleFunc("/users/{id}", utils.JWTMiddleware(userHandler.UpdateUser)).Methods("PUT")
	router.HandleFunc("/users/{id}", utils.JWTMiddleware(userHandler.DeleteUser)).Methods("DELETE")

	// Auth routes
	router.HandleFunc("/login", userHandler.Login).Methods("POST")
	router.HandleFunc("/logout", utils.JWTMiddleware(userHandler.Logout)).Methods("POST")
	router.HandleFunc("/token/refresh", userHandler.RefreshToken).Methods("POST")

	// register wallet handler
	handlers.RegisterWalletRoutes(router, dbConn)

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadEnv() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	config.JWTSecretKey = []byte(jwtSecretKey)
	config.DatabaseURL = databaseURL
}
