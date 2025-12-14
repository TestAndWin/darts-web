package store

import (
	"os"
	"testing"
)

func TestCreateUser_DuplicateRejection(t *testing.T) {
	// Create temporary database
	dbPath := "./test_users.db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create user
	user1, err := store.CreateUser("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user1.Name != "TestPlayer" {
		t.Errorf("Expected name 'TestPlayer', got %s", user1.Name)
	}

	// Try to create same user again (should return error)
	user2, err := store.CreateUser("TestPlayer")
	if err != ErrDuplicateUsername {
		t.Errorf("Expected ErrDuplicateUsername, got %v", err)
	}

	if user2 != nil {
		t.Errorf("Expected nil user on duplicate, got %+v", user2)
	}
}

func TestListUsers(t *testing.T) {
	dbPath := "./test_list.db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create multiple users
	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		_, err := store.CreateUser(name)
		if err != nil {
			t.Fatalf("Failed to create user %s: %v", name, err)
		}
	}

	// List users
	users, err := store.ListUsers()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Users should be sorted by name
	expectedOrder := []string{"Alice", "Bob", "Charlie"}
	for i, user := range users {
		if user.Name != expectedOrder[i] {
			t.Errorf("Expected user %d to be %s, got %s", i, expectedOrder[i], user.Name)
		}
	}
}

func TestDeleteUser(t *testing.T) {
	dbPath := "./test_delete.db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create a user
	user, err := store.CreateUser("TestUser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete the user
	err = store.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Try to get the deleted user
	deletedUser, err := store.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if deletedUser != nil {
		t.Errorf("Expected user to be deleted, but got %+v", deletedUser)
	}

	// Try to delete non-existent user
	err = store.DeleteUser(9999)
	if err == nil {
		t.Errorf("Expected error when deleting non-existent user, got nil")
	}
}
