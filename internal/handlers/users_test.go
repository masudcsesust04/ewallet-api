package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/masudcsesust04/ewallet-api/internal/db"
)

type mockDB struct {
	users []*db.User
}

func (m *mockDB) GetAllUsers() ([]*db.User, error) {
	return m.users, nil
}

func (m *mockDB) GetUserByEmail(email string) (*db.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockDB) CreateUser(user *db.User) error {
	m.users = append(m.users, user)
	return nil
}

func (m *mockDB) GetUserByID(id int64) (*db.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockDB) UpdateUser(user *db.User) error {
	for i, u := range m.users {
		if u.ID == user.ID {
			m.users[i] = user
			return nil
		}
	}
	return nil
}

func (m *mockDB) DeleteUser(id int64) error {
	for i, u := range m.users {
		if u.ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockDB) CreateRefreshToken(rt *db.RefreshToken) error {
	return nil
}

func (m *mockDB) DeleteRefreshToken(userID int64) error {
	return nil
}

func (m *mockDB) GetRefreshToken(userID int64) (*db.RefreshToken, error) {
	return nil, nil
}

func TestGetUsers(t *testing.T) {
	mockUsers := []*db.User{
		{ID: 1, FirstName: "User1", LastName: "Test", Email: "user1@example.com"},
		{ID: 2, FirstName: "User2", LastName: "Test", Email: "user2@example.com"},
	}
	mockDB := &mockDB{users: mockUsers}
	handler := NewUserHandler(nil)
	handler.DB = mockDB

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()

	handler.GetUsers(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	var users []*db.User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(users) != len(mockUsers) {
		t.Fatalf("expected %d users, got %d", len(mockUsers), len(users))
	}
}
