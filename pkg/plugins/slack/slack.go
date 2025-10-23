package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

// Slack implements the Notifier interface for Slack
type Slack struct {
	webhookURL string
	channel    string
	username   string
}

// NewSlack creates a new Slack notifier
func NewSlack() *Slack {
	return &Slack{}
}

// Name returns the name of this notifier
func (s *Slack) Name() string {
	return "slack"
}

// Initialize sets up the Slack notifier with configuration
func (s *Slack) Initialize(config map[string]interface{}) error {
	if webhookURL, ok := config["webhook_url"].(string); ok {
		s.webhookURL = webhookURL
	} else {
		return fmt.Errorf("webhook_url is required")
	}

	if channel, ok := config["channel"].(string); ok {
		s.channel = channel
	}

	if username, ok := config["username"].(string); ok {
		s.username = username
	} else {
		s.username = "AnyAlert"
	}

	return nil
}

// Validate checks if the message is valid for Slack
func (s *Slack) Validate(msg *notifier.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if msg.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}
	return nil
}

// Send sends a message to Slack
func (s *Slack) Send(ctx context.Context, msg *notifier.Message) (*notifier.Response, error) {
	payload := map[string]interface{}{
		"text":     msg.Content,
		"username": s.username,
	}

	if msg.Subject != "" {
		payload["text"] = fmt.Sprintf("*%s*\n%s", msg.Subject, msg.Content)
	}

	if s.channel != "" {
		payload["channel"] = s.channel
	}

	// Override channel if specified in message metadata
	if ch, ok := msg.Metadata["channel"]; ok {
		payload["channel"] = ch
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal payload: %v", err),
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to send request: %v", err),
		}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("slack returned status %d", resp.StatusCode),
		}, fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return &notifier.Response{
		Success: true,
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
		},
	}, nil
}
