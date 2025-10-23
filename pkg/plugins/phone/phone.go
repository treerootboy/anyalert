package phone

import (
	"context"
	"fmt"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

// Phone implements the Notifier interface for phone calls
type Phone struct {
	provider  string
	apiKey    string
	apiSecret string
	from      string
}

// NewPhone creates a new Phone notifier
func NewPhone() *Phone {
	return &Phone{}
}

// Name returns the name of this notifier
func (p *Phone) Name() string {
	return "phone"
}

// Initialize sets up the Phone notifier with configuration
func (p *Phone) Initialize(config map[string]interface{}) error {
	if provider, ok := config["provider"].(string); ok {
		p.provider = provider
	} else {
		p.provider = "default"
	}

	if apiKey, ok := config["api_key"].(string); ok {
		p.apiKey = apiKey
	} else {
		return fmt.Errorf("api_key is required")
	}

	if apiSecret, ok := config["api_secret"].(string); ok {
		p.apiSecret = apiSecret
	}

	if from, ok := config["from"].(string); ok {
		p.from = from
	}

	return nil
}

// Validate checks if the message is valid for Phone
func (p *Phone) Validate(msg *notifier.Message) error {
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

// Send initiates a phone call
func (p *Phone) Send(ctx context.Context, msg *notifier.Message) (*notifier.Response, error) {
	// This is a mock implementation. In production, you would integrate with
	// actual voice call providers like Twilio Voice, Plivo, etc.

	// Simulate initiating phone call
	messageID := fmt.Sprintf("call-%s-%d", p.provider, len(msg.Content))

	return &notifier.Response{
		Success:   true,
		MessageID: messageID,
		Details: map[string]interface{}{
			"provider":   p.provider,
			"recipients": msg.To,
			"from":       p.from,
			"type":       "voice_call",
		},
	}, nil
}
