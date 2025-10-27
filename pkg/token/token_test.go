package token

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Initialize schema
	schema := `
	CREATE TABLE IF NOT EXISTS tokens (
		id TEXT PRIMARY KEY,
		value TEXT UNIQUE NOT NULL,
		description TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		expires_at DATETIME
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to initialize database schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestNewManager(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.db != db {
		t.Error("Manager database connection not set correctly")
	}
}

func TestManager_Generate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Generate token without expiration
	token, err := mgr.Generate("Test token", nil)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token.ID == "" {
		t.Error("Token ID is empty")
	}
	if token.Value == "" {
		t.Error("Token value is empty")
	}
	if token.Description != "Test token" {
		t.Errorf("Expected description 'Test token', got '%s'", token.Description)
	}
	if token.ExpiresAt != nil {
		t.Error("Token should not have expiration")
	}

	// Generate token with expiration
	duration := 24 * time.Hour
	token2, err := mgr.Generate("Expiring token", &duration)
	if err != nil {
		t.Fatalf("Failed to generate token with expiration: %v", err)
	}

	if token2.ExpiresAt == nil {
		t.Error("Token should have expiration")
	}
}

func TestManager_Add(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	token := &Token{
		Value:       "test-token-value",
		Description: "Added token",
		CreatedAt:   time.Now(),
	}

	err := mgr.Add(token)
	if err != nil {
		t.Fatalf("Failed to add token: %v", err)
	}

	if token.ID == "" {
		t.Error("Token ID should be generated")
	}

	// Verify token was saved
	savedToken, found := mgr.Get(token.Value)
	if !found {
		t.Error("Token not found after adding")
	}
	if savedToken.Description != "Added token" {
		t.Errorf("Expected description 'Added token', got '%s'", savedToken.Description)
	}
}

func TestManager_Validate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Add a valid token
	validToken := &Token{
		Value:       "valid-token",
		Description: "Valid token",
		CreatedAt:   time.Now(),
	}
	mgr.Add(validToken)

	// Test valid token
	if !mgr.Validate("valid-token") {
		t.Error("Valid token should be validated")
	}

	// Test invalid token
	if mgr.Validate("invalid-token") {
		t.Error("Invalid token should not be validated")
	}

	// Add an expired token
	pastTime := time.Now().Add(-1 * time.Hour)
	expiredToken := &Token{
		Value:       "expired-token",
		Description: "Expired token",
		CreatedAt:   time.Now(),
		ExpiresAt:   &pastTime,
	}
	mgr.Add(expiredToken)

	// Test expired token
	if mgr.Validate("expired-token") {
		t.Error("Expired token should not be validated")
	}

	// Add a future expiring token
	futureTime := time.Now().Add(1 * time.Hour)
	futureToken := &Token{
		Value:       "future-token",
		Description: "Future expiring token",
		CreatedAt:   time.Now(),
		ExpiresAt:   &futureTime,
	}
	mgr.Add(futureToken)

	// Test future expiring token
	if !mgr.Validate("future-token") {
		t.Error("Future expiring token should be validated")
	}
}

func TestManager_Revoke(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Add a token
	token := &Token{
		Value:       "revoke-me",
		Description: "Token to revoke",
		CreatedAt:   time.Now(),
	}
	mgr.Add(token)

	// Verify it exists
	if !mgr.Validate("revoke-me") {
		t.Error("Token should be valid before revoking")
	}

	// Revoke it
	err := mgr.Revoke("revoke-me")
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Verify it's gone
	if mgr.Validate("revoke-me") {
		t.Error("Token should not be valid after revoking")
	}

	// Try to revoke non-existent token
	err = mgr.Revoke("non-existent")
	if err == nil {
		t.Error("Revoking non-existent token should return an error")
	}
}

func TestManager_RevokeByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Add a token
	token := &Token{
		ID:          "test-id",
		Value:       "test-value",
		Description: "Test token",
		CreatedAt:   time.Now(),
	}
	mgr.Add(token)

	// Revoke by ID
	err := mgr.RevokeByID("test-id")
	if err != nil {
		t.Fatalf("Failed to revoke token by ID: %v", err)
	}

	// Verify it's gone
	if mgr.Validate("test-value") {
		t.Error("Token should not be valid after revoking by ID")
	}
}

func TestManager_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Initially should be empty
	tokens := mgr.List()
	if len(tokens) != 0 {
		t.Errorf("Expected 0 tokens, got %d", len(tokens))
	}

	// Add some tokens
	mgr.Generate("Token 1", nil)
	mgr.Generate("Token 2", nil)
	duration := 24 * time.Hour
	mgr.Generate("Token 3", &duration)

	// List should return 3 tokens
	tokens = mgr.List()
	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}
}

func TestManager_Get(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Add a token
	original := &Token{
		Value:       "get-me",
		Description: "Token to get",
		CreatedAt:   time.Now(),
	}
	mgr.Add(original)

	// Get it back
	retrieved, found := mgr.Get("get-me")
	if !found {
		t.Fatal("Token not found")
	}

	if retrieved.Value != "get-me" {
		t.Errorf("Expected value 'get-me', got '%s'", retrieved.Value)
	}
	if retrieved.Description != "Token to get" {
		t.Errorf("Expected description 'Token to get', got '%s'", retrieved.Description)
	}

	// Try to get non-existent token
	_, found = mgr.Get("non-existent")
	if found {
		t.Error("Should not find non-existent token")
	}
}

func TestManager_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Add a token
	original := &Token{
		ID:          "my-id",
		Value:       "my-value",
		Description: "My token",
		CreatedAt:   time.Now(),
	}
	mgr.Add(original)

	// Get by ID
	retrieved, found := mgr.GetByID("my-id")
	if !found {
		t.Fatal("Token not found by ID")
	}

	if retrieved.ID != "my-id" {
		t.Errorf("Expected ID 'my-id', got '%s'", retrieved.ID)
	}
	if retrieved.Value != "my-value" {
		t.Errorf("Expected value 'my-value', got '%s'", retrieved.Value)
	}
}

func TestManager_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Initially should be 0
	count := mgr.Count()
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Add some tokens
	mgr.Generate("Token 1", nil)
	mgr.Generate("Token 2", nil)

	// Count should be 2
	count = mgr.Count()
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestManager_Clear(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)

	// Add some tokens
	mgr.Generate("Token 1", nil)
	mgr.Generate("Token 2", nil)

	// Verify they exist
	if mgr.Count() != 2 {
		t.Error("Expected 2 tokens before clear")
	}

	// Clear
	mgr.Clear()

	// Verify they're gone
	if mgr.Count() != 0 {
		t.Error("Expected 0 tokens after clear")
	}
}
