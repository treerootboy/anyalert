package slack

import (
	"context"
	"testing"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

func TestSlack_Name(t *testing.T) {
	s := NewSlack()
	if s.Name() != "slack" {
		t.Errorf("Expected name 'slack', got '%s'", s.Name())
	}
}

func TestSlack_Initialize(t *testing.T) {
	s := NewSlack()

	// Test with valid webhook config (backward compatibility)
	webhookConfig := map[string]interface{}{
		"webhook_url": "https://hooks.slack.com/test",
		"username":    "TestBot",
		"channel":     "#test",
	}

	err := s.Initialize(webhookConfig)
	if err != nil {
		t.Fatalf("Failed to initialize with webhook: %v", err)
	}

	// Test with valid bot token config
	s2 := NewSlack()
	botConfig := map[string]interface{}{
		"bot_token":  "xoxb-test-token",
		"username":   "TestBot",
		"channel":    "#test",
		"icon_emoji": ":robot_face:",
	}

	err = s2.Initialize(botConfig)
	if err != nil {
		t.Fatalf("Failed to initialize with bot token: %v", err)
	}

	// Verify bot token was set
	if s2.botToken != "xoxb-test-token" {
		t.Errorf("Expected bot_token 'xoxb-test-token', got '%s'", s2.botToken)
	}

	// Verify icon_emoji was set
	if s2.iconEmoji != ":robot_face:" {
		t.Errorf("Expected icon_emoji ':robot_face:', got '%s'", s2.iconEmoji)
	}

	// Test with icon_url instead of icon_emoji
	s3 := NewSlack()
	iconURLConfig := map[string]interface{}{
		"bot_token": "xoxb-test-token",
		"username":  "TestBot",
		"icon_url":  "https://example.com/icon.png",
	}

	err = s3.Initialize(iconURLConfig)
	if err != nil {
		t.Fatalf("Failed to initialize with icon_url: %v", err)
	}

	if s3.iconURL != "https://example.com/icon.png" {
		t.Errorf("Expected icon_url 'https://example.com/icon.png', got '%s'", s3.iconURL)
	}

	// Test with missing both webhook_url and bot_token
	s4 := NewSlack()
	invalidConfig := map[string]interface{}{
		"username": "TestBot",
	}

	err = s4.Initialize(invalidConfig)
	if err == nil {
		t.Fatal("Expected error when both webhook_url and bot_token are missing")
	}
}

func TestSlack_Validate(t *testing.T) {
	s := NewSlack()

	// Test with valid message
	msg := &notifier.Message{
		Content: "Test message",
	}
	err := s.Validate(msg)
	if err != nil {
		t.Fatalf("Validation failed for valid message: %v", err)
	}

	// Test with nil message
	err = s.Validate(nil)
	if err == nil {
		t.Fatal("Expected error for nil message")
	}

	// Test with empty content
	emptyMsg := &notifier.Message{}
	err = s.Validate(emptyMsg)
	if err == nil {
		t.Fatal("Expected error for empty content")
	}
}

func TestSlack_Send_WithWebhook(t *testing.T) {
	s := NewSlack()

	// Initialize with a test webhook URL (this will fail but we can test the logic)
	config := map[string]interface{}{
		"webhook_url": "https://hooks.slack.com/test",
		"username":    "TestBot",
	}
	s.Initialize(config)

	msg := &notifier.Message{
		Subject: "Test Subject",
		Content: "Test Content",
	}

	// This will fail because it's not a real webhook, but we can test that it doesn't panic
	_, err := s.Send(context.Background(), msg)
	// We expect an error because it's not a real webhook
	if err == nil {
		t.Log("Note: Send succeeded, which is unexpected with a test URL")
	}
}

func TestSlack_Send_WithBotToken(t *testing.T) {
	s := NewSlack()

	// Initialize with a test bot token (this will fail but we can test the logic)
	config := map[string]interface{}{
		"bot_token":  "xoxb-test-token",
		"username":   "TestBot",
		"icon_emoji": ":robot_face:",
		"channel":    "#general",
	}
	s.Initialize(config)

	// Test sending to default channel
	msg := &notifier.Message{
		Subject: "Test Subject",
		Content: "Test Content",
	}

	// This will fail because it's not a real token, but we can test that it doesn't panic
	_, err := s.Send(context.Background(), msg)
	// We expect an error because it's not a real token
	if err == nil {
		t.Log("Note: Send succeeded, which is unexpected with a test token")
	}

	// Test sending to specific user
	msgToUser := &notifier.Message{
		To:      []string{"@username"},
		Subject: "Test Subject",
		Content: "Test Content",
	}

	_, err = s.Send(context.Background(), msgToUser)
	if err == nil {
		t.Log("Note: Send succeeded, which is unexpected with a test token")
	}

	// Test sending to specific channel via metadata
	msgToChannel := &notifier.Message{
		Subject:  "Test Subject",
		Content:  "Test Content",
		Metadata: map[string]string{"channel": "#alerts"},
	}

	_, err = s.Send(context.Background(), msgToChannel)
	if err == nil {
		t.Log("Note: Send succeeded, which is unexpected with a test token")
	}
}
