package userstore

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-based user store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return store, nil
}

// initialize creates the users table if it doesn't exist
func (s *SQLiteStore) initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		slack TEXT,
		youdu TEXT,
		phone TEXT,
		sms TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
	`

	_, err := s.db.Exec(query)
	return err
}

// Create adds a new user to the store
func (s *SQLiteStore) Create(user *User) error {
	query := `
	INSERT INTO users (name, slack, youdu, phone, sms, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := s.db.Exec(query,
		user.Name,
		user.Slack,
		user.Youdu,
		user.Phone,
		user.SMS,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// Get retrieves a user by ID
func (s *SQLiteStore) Get(id int64) (*User, error) {
	query := `
	SELECT id, name, slack, youdu, phone, sms, created_at, updated_at
	FROM users
	WHERE id = ?
	`

	user := &User{}
	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Slack,
		&user.Youdu,
		&user.Phone,
		&user.SMS,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByName retrieves a user by name
func (s *SQLiteStore) GetByName(name string) (*User, error) {
	query := `
	SELECT id, name, slack, youdu, phone, sms, created_at, updated_at
	FROM users
	WHERE name = ?
	`

	user := &User{}
	err := s.db.QueryRow(query, name).Scan(
		&user.ID,
		&user.Name,
		&user.Slack,
		&user.Youdu,
		&user.Phone,
		&user.SMS,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// List retrieves all users
func (s *SQLiteStore) List() ([]*User, error) {
	query := `
	SELECT id, name, slack, youdu, phone, sms, created_at, updated_at
	FROM users
	ORDER BY name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Slack,
			&user.Youdu,
			&user.Phone,
			&user.SMS,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Update updates an existing user
func (s *SQLiteStore) Update(user *User) error {
	query := `
	UPDATE users
	SET name = ?, slack = ?, youdu = ?, phone = ?, sms = ?, updated_at = ?
	WHERE id = ?
	`

	now := time.Now()
	result, err := s.db.Exec(query,
		user.Name,
		user.Slack,
		user.Youdu,
		user.Phone,
		user.SMS,
		now,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	user.UpdatedAt = now
	return nil
}

// Delete removes a user by ID
func (s *SQLiteStore) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
