package db

import (
	"testing"
	"time"
)

func TestCreateAndGetUser(t *testing.T) {
	user := &User{
		FirstName:   "Test",
		LastName:    "User",
		PhoneNumber: "1234567890",
		Email:       "testuser@example.com",
		Status:      "active",
		Password:    "password123",
	}

	err := testDB.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Use GetUserByEmail instead of GetUserByUsername
	gotUser, err := testDB.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if gotUser == nil || gotUser.Email != user.Email {
		t.Fatalf("GetUserByEmail returned wrong user")
	}
}

func TestUpdateUser(t *testing.T) {
	user := &User{
		FirstName:   "Update",
		LastName:    "User",
		PhoneNumber: "0987654321",
		Email:       "updateuser@example.com",
		Status:      "active",
		Password:    "password123",
	}

	err := testDB.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	user.Email = "newemail@example.com"
	err = testDB.UpdateUser(user)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	updatedUser, err := testDB.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if updatedUser.Email != "newemail@example.com" {
		t.Fatalf("UpdateUser did not update email")
	}
}

func TestDeleteUser(t *testing.T) {
	user := &User{
		FirstName:   "Delete",
		LastName:    "User",
		PhoneNumber: "1112223333",
		Email:       "deleteuser@example.com",
		Status:      "active",
		Password:    "password123",
	}

	err := testDB.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = testDB.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	deletedUser, err := testDB.GetUserByID(user.ID)
	if err == nil && deletedUser != nil {
		t.Fatalf("DeleteUser did not delete user")
	}
}

func TestCreateAndDeleteRefreshToken(t *testing.T) {
	user := &User{
		FirstName:   "Token",
		LastName:    "User",
		PhoneNumber: "4445556666",
		Email:       "tokenuser@example.com",
		Status:      "active",
		Password:    "password123",
	}

	err := testDB.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	rt := &RefreshToken{
		UserID:    user.ID,
		Token:     "testtoken",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err = testDB.CreateRefreshToken(rt)
	if err != nil {
		t.Fatalf("CreateRefreshToken failed: %v", err)
	}

	gotRT, err := testDB.GetRefreshToken(rt.UserID)
	if err != nil {
		t.Fatalf("GetRefreshToken failed: %v", err)
	}
	if gotRT == nil || gotRT.Token != rt.Token {
		t.Fatalf("GetRefreshToken returned wrong token")
	}

	err = testDB.DeleteRefreshToken(rt.UserID)
	if err != nil {
		t.Fatalf("DeleteRefreshToken failed: %v", err)
	}

	deletedRT, err := testDB.GetRefreshToken(rt.UserID)
	if err == nil && deletedRT != nil {
		t.Fatalf("DeleteRefreshToken did not delete token")
	}
}

func TestGetAllUsers(t *testing.T) {
	users, err := testDB.GetAllUsers()
	if err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("GetAllUsers returned empty list")
	}
}
