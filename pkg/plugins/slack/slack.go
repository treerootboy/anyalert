package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/treerootboy/anyalert/pkg/notifier"
)

// Slack implements the Notifier interface for Slack
type Slack struct {
	// Webhook mode
	webhookURL string
	
	// Bot token mode
	botToken  string
	client    *slack.Client
	
	// Common settings
	channel   string
	username  string
	iconEmoji string
	iconURL   string
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
	// Check for bot_token (new Bot User OAuth Token mode)
	if botToken, ok := config["bot_token"].(string); ok && botToken != "" {
		s.botToken = botToken
		s.client = slack.New(botToken)
	} else if webhookURL, ok := config["webhook_url"].(string); ok && webhookURL != "" {
		// Fallback to webhook mode for backward compatibility
		s.webhookURL = webhookURL
	} else {
		return fmt.Errorf("either bot_token or webhook_url is required")
	}

	// Common configuration
	if channel, ok := config["channel"].(string); ok {
		s.channel = channel
	}

	if username, ok := config["username"].(string); ok {
		s.username = username
	} else {
		s.username = "AnyAlert"
	}

	if iconEmoji, ok := config["icon_emoji"].(string); ok {
		s.iconEmoji = iconEmoji
	}

	if iconURL, ok := config["icon_url"].(string); ok {
		s.iconURL = iconURL
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
	// Determine target channel/user
	target := s.channel
	
	// Override with message metadata if provided
	if ch, ok := msg.Metadata["channel"]; ok {
		target = ch
	}
	
	// Support specifying target in the "To" field
	if len(msg.To) > 0 && msg.To[0] != "" {
		target = msg.To[0]
	}
	
	// Format message text
	text := msg.Content
	if msg.Subject != "" {
		text = fmt.Sprintf("*%s*\n%s", msg.Subject, msg.Content)
	}
	
	// Use Bot Token mode if available
	if s.client != nil {
		return s.sendWithBotToken(ctx, target, text)
	}
	
	// Fallback to webhook mode
	return s.sendWithWebhook(ctx, target, text)
}

// sendWithBotToken sends message using Bot User OAuth Token
func (s *Slack) sendWithBotToken(ctx context.Context, target, text string) (*notifier.Response, error) {
	// Build message options
	options := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionAsUser(false),
	}
	
	if s.username != "" {
		options = append(options, slack.MsgOptionUsername(s.username))
	}
	
	if s.iconEmoji != "" {
		options = append(options, slack.MsgOptionIconEmoji(s.iconEmoji))
	} else if s.iconURL != "" {
		options = append(options, slack.MsgOptionIconURL(s.iconURL))
	}
	
	// Send message
	channel, timestamp, err := s.client.PostMessageContext(ctx, target, options...)
	if err != nil {
		return &notifier.Response{
			Success: false,
			Error:   fmt.Sprintf("failed to send message: %v", err),
		}, err
	}
	
	return &notifier.Response{
		Success:   true,
		MessageID: timestamp,
		Details: map[string]interface{}{
			"channel":   channel,
			"timestamp": timestamp,
		},
	}, nil
}

// sendWithWebhook sends message using webhook (backward compatibility)
func (s *Slack) sendWithWebhook(ctx context.Context, target, text string) (*notifier.Response, error) {
	payload := map[string]interface{}{
		"text":     text,
		"username": s.username,
	}

	if target != "" {
		payload["channel"] = target
	}
	
	if s.iconEmoji != "" {
		payload["icon_emoji"] = s.iconEmoji
	} else if s.iconURL != "" {
		payload["icon_url"] = s.iconURL
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
