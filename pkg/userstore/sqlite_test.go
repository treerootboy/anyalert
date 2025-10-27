package userstore

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) (*SQLiteStore, func()) {
	tmpfile, err := os.CreateTemp("", "test_users_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	store, err := NewSQLiteStore(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(tmpfile.Name())
	}

	return store, cleanup
}

func TestSQLiteStore_Create(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		Name:  "user1",
		Slack: "user1@example.com",
		Youdu: "10232",
		Phone: "+8613800138000",
		SMS:   "+8613800138000",
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}

	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestSQLiteStore_Get(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		Name:  "user1",
		Slack: "user1@example.com",
		Youdu: "10232",
		Phone: "+8613800138000",
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := store.Get(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, retrieved.Name)
	}

	if retrieved.Slack != user.Slack {
		t.Errorf("Expected slack %s, got %s", user.Slack, retrieved.Slack)
	}

	if retrieved.Youdu != user.Youdu {
		t.Errorf("Expected youdu %s, got %s", user.Youdu, retrieved.Youdu)
	}
}

func TestSQLiteStore_GetByName(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		Name:  "user1",
		Slack: "user1@example.com",
		Youdu: "10232",
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := store.GetByName("user1")
	if err != nil {
		t.Fatalf("Failed to get user by name: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, retrieved.ID)
	}
}

func TestSQLiteStore_List(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	users := []*User{
		{Name: "user1", Slack: "user1@example.com"},
		{Name: "user2", Youdu: "10233"},
		{Name: "user3", Phone: "+8613800138001"},
	}

	for _, u := range users {
		err := store.Create(u)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	retrieved, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(retrieved) != len(users) {
		t.Errorf("Expected %d users, got %d", len(users), len(retrieved))
	}
}

func TestSQLiteStore_Update(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		Name:  "user1",
		Slack: "user1@example.com",
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user.Slack = "updated@example.com"
	user.Phone = "+8613800138000"

	err = store.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, err := store.Get(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.Slack != "updated@example.com" {
		t.Errorf("Expected updated slack, got %s", retrieved.Slack)
	}

	if retrieved.Phone != "+8613800138000" {
		t.Errorf("Expected updated phone, got %s", retrieved.Phone)
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	user := &User{
		Name:  "user1",
		Slack: "user1@example.com",
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = store.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = store.Get(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestSQLiteStore_CreateDuplicateName(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	user1 := &User{Name: "user1", Slack: "user1@example.com"}
	err := store.Create(user1)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	user2 := &User{Name: "user1", Slack: "other@example.com"}
	err = store.Create(user2)
	if err == nil {
		t.Error("Expected error when creating user with duplicate name")
	}
}
