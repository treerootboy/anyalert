package notifier

import "context"

// Message represents a notification message
type Message struct {
	// To is the recipient(s) of the message
	To []string `json:"to"`
	// Subject is the title or subject of the message
	Subject string `json:"subject"`
	// Content is the main body of the message
	Content string `json:"content"`
	// Priority indicates the urgency level (low, normal, high, urgent)
	Priority string `json:"priority,omitempty"`
	// Metadata for additional channel-specific options
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Response represents the result of a notification
type Response struct {
	// Success indicates if the notification was sent successfully
	Success bool `json:"success"`
	// MessageID is the unique identifier for the sent message
	MessageID string `json:"message_id,omitempty"`
	// Error contains error information if the notification failed
	Error string `json:"error,omitempty"`
	// Details contains additional channel-specific response information
	Details map[string]interface{} `json:"details,omitempty"`
}

// Notifier is the interface that all notification plugins must implement
type Notifier interface {
	// Name returns the name of the notification channel
	Name() string
	// Send sends a notification message
	Send(ctx context.Context, msg *Message) (*Response, error)
	// Validate checks if the message is valid for this channel
	Validate(msg *Message) error
	// Initialize sets up the notifier with configuration
	Initialize(config map[string]interface{}) error
}
