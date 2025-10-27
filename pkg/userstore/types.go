package userstore

import "time"

// User represents a user with their notification channel identifiers
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slack     string    `json:"slack,omitempty"` // Slack email or username
	Youdu     string    `json:"youdu,omitempty"` // Youdu account ID
	Phone     string    `json:"phone,omitempty"` // Phone number
	SMS       string    `json:"sms,omitempty"`   // SMS phone number
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store defines the interface for user storage operations
type Store interface {
	// Create adds a new user to the store
	Create(user *User) error
	// Get retrieves a user by ID
	Get(id int64) (*User, error)
	// GetByName retrieves a user by name
	GetByName(name string) (*User, error)
	// List retrieves all users
	List() ([]*User, error)
	// Update updates an existing user
	Update(user *User) error
	// Delete removes a user by ID
	Delete(id int64) error
	// Close closes the store
	Close() error
}
