package db

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represent a user in the system
type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
	Status       string    `json:"status"`
	Password     string    `json:"password,omitempty"` // plain password, not stored in DB
	PasswordHash string    `json:"passwrod_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"created_at"`
	CreatedAt time.Time `json:"updated_at"`
}

// GetUserByEmail retr4ieves a user by email
func (db *DB) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, first_name, last_name, phone_number, email, password_hash, status, created_at, updated_at FROM  users WHERE email = $1`
	user := &User{}

	err := db.pool.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// CreateUser inserts a new user into the database with password hashing
func (db *DB) CreateUser(user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)

	query := `INSERT INTO users (first_name, last_name, phone_number, email, status, password_hash) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	err = db.pool.QueryRow(context.Background(), query, user.FirstName, user.LastName, user.PhoneNumber, user.Email, user.Status, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrives a user by ID
func (db *DB) GetUserByID(id int64) (*User, error) {
	query := `SELECT id, first_name, last_name, phone_number, email, status, created_at, updated_at FROM  users WHERE id = $1`
	user := &User{}

	err := db.pool.QueryRow(context.Background(), query, id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetAllUsers retrives all users from the database
func (db *DB) GetAllUsers() ([]*User, error) {
	query := `SELECT id, first_name, last_name, phone_number, email, status, created_at, updated_at FROM  users`

	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// UpdateUser updates an existing users' information
func (db *DB) UpdateUser(user *User) error {
	query := `UPDATE users SET first_name = $1, last_name = $2, phone_number = $3, email = $4, status= $5 WHERE id = $6`
	_, err := db.pool.Exec(context.Background(), query, user.FirstName, user.LastName, user.PhoneNumber, user.Email, user.Status, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (db *DB) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE ID = $1`
	_, err := db.pool.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// CreateRefreshToken inserts a new refresh token into the database
func (db *DB) CreateRefreshToken(rt *RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.pool.QueryRow(context.Background(), query, rt.UserID, rt.Token, rt.ExpiresAt, rt.CreatedAt).Scan(&rt.ID)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken inserts a new refresh token into the database
func (db *DB) GetRefreshToken(userID int64) (*RefreshToken, error) {
	query := `SELECT * FROM refresh_tokens WHERE user_id = $1 AND expires_at > NOW() ORDER BY id DESC LIMIT 1`
	rt := &RefreshToken{}
	err := db.pool.QueryRow(context.Background(), query, userID).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return rt, nil
}

// DeleteRefreshToken delete refresh token by user_id
func (db *DB) DeleteRefreshToken(userId int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := db.pool.Exec(context.Background(), query, userId)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	return nil
}
