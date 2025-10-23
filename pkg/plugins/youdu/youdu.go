package youdu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

// Youdu implements the Notifier interface for Youdu IM
type Youdu struct {
	apiURL string
	buin   int
	appID  string
	apiKey string
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
		y.apiURL = apiURL
	} else {
		return fmt.Errorf("api_url is required")
	}

	if buin, ok := config["buin"].(int); ok {
		y.buin = buin
	} else if buinFloat, ok := config["buin"].(float64); ok {
		y.buin = int(buinFloat)
	} else {
		return fmt.Errorf("buin is required")
	}

	if appID, ok := config["app_id"].(string); ok {
		y.appID = appID
	} else {
		return fmt.Errorf("app_id is required")
	}

	if apiKey, ok := config["api_key"].(string); ok {
		y.apiKey = apiKey
	} else {
		return fmt.Errorf("api_key is required")
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

// Send sends a message via Youdu IM
func (y *Youdu) Send(ctx context.Context, msg *notifier.Message) (*notifier.Response, error) {
	// Build Youdu message payload
	payload := map[string]interface{}{
		"buin":    y.buin,
		"appId":   y.appID,
		"toUser":  msg.To,
		"msgType": "text",
		"text": map[string]string{
			"content": msg.Content,
		},
	}

	if msg.Subject != "" {
		payload["text"] = map[string]string{
			"content": fmt.Sprintf("%s\n%s", msg.Subject, msg.Content),
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal payload: %v", err),
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", y.apiURL+"/cgi/msg/send", bytes.NewBuffer(jsonData))
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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to decode response: %v", err),
		}, err
	}

	if errCode, ok := result["errcode"].(float64); ok && errCode != 0 {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("youdu returned error code: %v", errCode),
		}, fmt.Errorf("youdu returned error code: %v", errCode)
	}

	return &notifier.Response{
		Success: true,
		Details: result,
	}, nil
}
