package notifier

import (
	"context"
	"testing"
)

// mockNotifier is a mock implementation for testing
type mockNotifier struct {
	name        string
	initialized bool
	sendError   error
}

func (m *mockNotifier) Name() string {
	return m.name
}

func (m *mockNotifier) Send(ctx context.Context, msg *Message) (*Response, error) {
	if m.sendError != nil {
		return &Response{Success: false, Error: m.sendError.Error()}, m.sendError
	}
	return &Response{Success: true, MessageID: "test-123"}, nil
}

func (m *mockNotifier) Validate(msg *Message) error {
	if msg == nil {
		return nil
	}
	return nil
}

func (m *mockNotifier) Initialize(config map[string]interface{}) error {
	m.initialized = true
	return nil
}

func TestManager_Register(t *testing.T) {
	manager := NewManager()
	notifier := &mockNotifier{name: "test"}

	err := manager.Register(notifier)
	if err != nil {
		t.Fatalf("Failed to register notifier: %v", err)
	}

	// Test duplicate registration
	err = manager.Register(notifier)
	if err == nil {
		t.Fatal("Expected error when registering duplicate notifier")
	}
}

func TestManager_Send(t *testing.T) {
	manager := NewManager()
	notifier := &mockNotifier{name: "test"}
	manager.Register(notifier)

	msg := &Message{
		To:      []string{"test@example.com"},
		Subject: "Test",
		Content: "Test message",
	}

	resp, err := manager.Send(context.Background(), "test", msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if !resp.Success {
		t.Fatal("Expected successful response")
	}

	// Test sending to non-existent channel
	_, err = manager.Send(context.Background(), "nonexistent", msg)
	if err == nil {
		t.Fatal("Expected error when sending to non-existent channel")
	}
}

func TestManager_Broadcast(t *testing.T) {
	manager := NewManager()
	notifier1 := &mockNotifier{name: "test1"}
	notifier2 := &mockNotifier{name: "test2"}
	manager.Register(notifier1)
	manager.Register(notifier2)

	msg := &Message{
		To:      []string{"test@example.com"},
		Subject: "Test",
		Content: "Test message",
	}

	results := manager.Broadcast(context.Background(), []string{"test1", "test2"}, msg)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for channel, resp := range results {
		if !resp.Success {
			t.Errorf("Channel %s failed: %v", channel, resp.Error)
		}
	}
}

func TestManager_ListChannels(t *testing.T) {
	manager := NewManager()
	notifier1 := &mockNotifier{name: "test1"}
	notifier2 := &mockNotifier{name: "test2"}
	manager.Register(notifier1)
	manager.Register(notifier2)

	channels := manager.ListChannels()
	if len(channels) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(channels))
	}
}

func TestManager_Get(t *testing.T) {
	manager := NewManager()
	notifier := &mockNotifier{name: "test"}
	manager.Register(notifier)

	n, err := manager.Get("test")
	if err != nil {
		t.Fatalf("Failed to get notifier: %v", err)
	}

	if n.Name() != "test" {
		t.Errorf("Expected notifier name 'test', got '%s'", n.Name())
	}

	// Test getting non-existent channel
	_, err = manager.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent channel")
	}
}
