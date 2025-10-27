package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := Config{
		Path: dbPath,
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Database is nil")
	}

	if db.GetPath() != dbPath {
		t.Errorf("Expected path %s, got %s", dbPath, db.GetPath())
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNew_DefaultPath(t *testing.T) {
	// Clean up any existing default database
	defaultPath := "./anyalert.db"
	os.Remove(defaultPath)
	defer os.Remove(defaultPath)

	config := Config{}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database with default path: %v", err)
	}
	defer db.Close()

	if db.GetPath() != defaultPath {
		t.Errorf("Expected default path %s, got %s", defaultPath, db.GetPath())
	}

	// Verify database file was created
	if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
		t.Error("Default database file was not created")
	}
}

func TestNew_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "nested", "test.db")

	config := Config{
		Path: dbPath,
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database with nested path: %v", err)
	}
	defer db.Close()

	// Verify directory was created
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Database directory was not created")
	}
}

func TestDB_GetConnection(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := Config{
		Path: dbPath,
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	conn := db.GetConnection()
	if conn == nil {
		t.Fatal("Connection is nil")
	}

	// Test connection is usable
	err = conn.Ping()
	if err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestDB_InitSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := Config{
		Path: dbPath,
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify tokens table exists
	var tableName string
	err = db.GetConnection().QueryRow(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name='tokens'
	`).Scan(&tableName)
	if err != nil {
		t.Errorf("Tokens table not created: %v", err)
	}

	// Verify users table exists
	err = db.GetConnection().QueryRow(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name='users'
	`).Scan(&tableName)
	if err != nil {
		t.Errorf("Users table not created: %v", err)
	}

	// Verify tokens indexes exist
	var indexName string
	err = db.GetConnection().QueryRow(`
		SELECT name FROM sqlite_master 
		WHERE type='index' AND name='idx_tokens_value'
	`).Scan(&indexName)
	if err != nil {
		t.Errorf("Tokens value index not created: %v", err)
	}

	// Verify users indexes exist
	err = db.GetConnection().QueryRow(`
		SELECT name FROM sqlite_master 
		WHERE type='index' AND name='idx_users_name'
	`).Scan(&indexName)
	if err != nil {
		t.Errorf("Users name index not created: %v", err)
	}
}

func TestDB_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := Config{
		Path: dbPath,
	}

	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Verify connection is closed
	conn := db.GetConnection()
	err = conn.Ping()
	if err == nil {
		t.Error("Expected error when pinging closed connection")
	}
}
