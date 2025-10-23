package sms

import (
	"context"
	"fmt"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

// SMS implements the Notifier interface for SMS
type SMS struct {
	provider  string
	apiKey    string
	apiSecret string
	from      string
}

// NewSMS creates a new SMS notifier
func NewSMS() *SMS {
	return &SMS{}
}

// Name returns the name of this notifier
func (s *SMS) Name() string {
	return "sms"
}

// Initialize sets up the SMS notifier with configuration
func (s *SMS) Initialize(config map[string]interface{}) error {
	if provider, ok := config["provider"].(string); ok {
		s.provider = provider
	} else {
		s.provider = "default"
	}

	if apiKey, ok := config["api_key"].(string); ok {
		s.apiKey = apiKey
	} else {
		return fmt.Errorf("api_key is required")
	}

	if apiSecret, ok := config["api_secret"].(string); ok {
		s.apiSecret = apiSecret
	}

	if from, ok := config["from"].(string); ok {
		s.from = from
	}

	return nil
}

// Validate checks if the message is valid for SMS
func (s *SMS) Validate(msg *notifier.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if msg.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	return nil
}

// Send sends an SMS message
func (s *SMS) Send(ctx context.Context, msg *notifier.Message) (*notifier.Response, error) {
	// This is a mock implementation. In production, you would integrate with
	// actual SMS providers like Twilio, Nexmo, Aliyun, etc.

	// Simulate sending SMS
	messageID := fmt.Sprintf("sms-%s-%d", s.provider, len(msg.Content))

	return &notifier.Response{
		Success:   true,
		MessageID: messageID,
		Details: map[string]interface{}{
			"provider":   s.provider,
			"recipients": msg.To,
			"from":       s.from,
		},
	}, nil
}
