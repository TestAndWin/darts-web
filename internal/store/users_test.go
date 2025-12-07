package store

import (
	"os"
	"testing"
)

func TestCreateUser_Upsert(t *testing.T) {
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

	// Try to create same user again (should return existing)
	user2, err := store.CreateUser("TestPlayer")
	if err != nil {
		t.Fatalf("Failed to create/get user: %v", err)
	}

	if user1.ID != user2.ID {
		t.Errorf("Expected same user ID, got %d and %d", user1.ID, user2.ID)
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
