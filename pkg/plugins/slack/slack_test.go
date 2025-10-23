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

	// Test with valid config
	config := map[string]interface{}{
		"webhook_url": "https://hooks.slack.com/test",
		"username":    "TestBot",
		"channel":     "#test",
	}

	err := s.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test with missing webhook_url
	invalidConfig := map[string]interface{}{
		"username": "TestBot",
	}

	err = s.Initialize(invalidConfig)
	if err == nil {
		t.Fatal("Expected error when webhook_url is missing")
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

func TestSlack_Send(t *testing.T) {
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
