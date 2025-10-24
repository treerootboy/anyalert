package youdu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

// Youdu implements the Notifier interface for Youdu IM
// It uses the youdu-app-mcp HTTP API interface with token authentication
type Youdu struct {
	apiURL string
	token  string
}

// NewYoudu creates a new Youdu notifier
func NewYoudu() *Youdu {
	return &Youdu{}
}

// Name returns the name of this notifier
func (y *Youdu) Name() string {
	return "youdu"
}

// Initialize sets up the Youdu notifier with configuration
func (y *Youdu) Initialize(config map[string]interface{}) error {
	if apiURL, ok := config["api_url"].(string); ok {
		// Remove trailing slash if present
		y.apiURL = strings.TrimRight(apiURL, "/")
	} else {
		return fmt.Errorf("api_url is required")
	}

	if token, ok := config["token"].(string); ok {
		y.token = token
	} else {
		return fmt.Errorf("token is required")
	}

	return nil
}

// Validate checks if the message is valid for Youdu
func (y *Youdu) Validate(msg *notifier.Message) error {
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

// Send sends a message via Youdu IM using the youdu-app-mcp HTTP API
func (y *Youdu) Send(ctx context.Context, msg *notifier.Message) (*notifier.Response, error) {
	// The youdu-app-mcp API uses send_text_message endpoint
	// Format: POST /api/v1/send_text_message
	// Body: {"to_user": "user123", "content": "message"}

	// Build request payload
	// For multiple recipients, we'll send to each one
	// Or if the first recipient contains the user ID format, use it directly
	var toUser string
	if len(msg.To) > 0 {
		toUser = msg.To[0] // Use first recipient as the user ID
	}

	// Combine subject and content if subject exists
	content := msg.Content
	if msg.Subject != "" {
		content = fmt.Sprintf("%s\n%s", msg.Subject, msg.Content)
	}

	payload := map[string]interface{}{
		"to_user": toUser,
		"content": content,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal payload: %v", err),
		}, err
	}

	// Create HTTP request to the youdu-app-mcp API
	apiEndpoint := fmt.Sprintf("%s/api/v1/send_text_message", y.apiURL)
	req, err := http.NewRequestWithContext(ctx, "POST", apiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	// Add token authentication - support both Bearer format and direct token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", y.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to send request: %v", err),
		}, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to decode response: %v", err),
		}, err
	}

	// Check for error in response
	if errVal, ok := result["error"]; ok {
		if errBool, ok := errVal.(bool); ok && errBool {
			errMsg := "unknown error"
			if msg, ok := result["message"].(string); ok {
				errMsg = msg
			}
			return &notifier.Response{
				Success: false,
				Error:   fmt.Sprintf("youdu-app-mcp returned error: %s", errMsg),
			}, fmt.Errorf("youdu-app-mcp returned error: %s", errMsg)
		}
	}

	return &notifier.Response{
		Success: true,
		Details: result,
	}, nil
}
