package youdu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/treerootboy/anyalert/pkg/notifier"
)

func TestYoudu_Name(t *testing.T) {
	y := NewYoudu()
	if y.Name() != "youdu" {
		t.Errorf("Expected name 'youdu', got '%s'", y.Name())
	}
}

func TestYoudu_Initialize(t *testing.T) {
	y := NewYoudu()

	// Test with valid config
	config := map[string]interface{}{
		"api_url": "http://localhost:8080",
		"token":   "test-token",
	}

	err := y.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Verify apiURL has trailing slash removed
	if y.apiURL != "http://localhost:8080" {
		t.Errorf("Expected apiURL 'http://localhost:8080', got '%s'", y.apiURL)
	}

	// Verify token is set
	if y.token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", y.token)
	}

	// Test with trailing slash in api_url
	configWithSlash := map[string]interface{}{
		"api_url": "http://localhost:8080/",
		"token":   "test-token",
	}

	err = y.Initialize(configWithSlash)
	if err != nil {
		t.Fatalf("Failed to initialize with trailing slash: %v", err)
	}

	if y.apiURL != "http://localhost:8080" {
		t.Errorf("Expected apiURL 'http://localhost:8080' (slash removed), got '%s'", y.apiURL)
	}

	// Test with missing api_url
	invalidConfig1 := map[string]interface{}{
		"token": "test-token",
	}

	err = y.Initialize(invalidConfig1)
	if err == nil {
		t.Fatal("Expected error when api_url is missing")
	}

	// Test with missing token
	invalidConfig2 := map[string]interface{}{
		"api_url": "http://localhost:8080",
	}

	err = y.Initialize(invalidConfig2)
	if err == nil {
		t.Fatal("Expected error when token is missing")
	}
}

func TestYoudu_Validate(t *testing.T) {
	y := NewYoudu()

	// Test with valid message
	msg := &notifier.Message{
		To:      []string{"user123"},
		Content: "Test message",
	}
	err := y.Validate(msg)
	if err != nil {
		t.Fatalf("Validation failed for valid message: %v", err)
	}

	// Test with nil message
	err = y.Validate(nil)
	if err == nil {
		t.Fatal("Expected error for nil message")
	}

	// Test with empty To
	emptyToMsg := &notifier.Message{
		Content: "Test content",
	}
	err = y.Validate(emptyToMsg)
	if err == nil {
		t.Fatal("Expected error for empty To field")
	}

	// Test with empty content
	emptyContentMsg := &notifier.Message{
		To: []string{"user123"},
	}
	err = y.Validate(emptyContentMsg)
	if err == nil {
		t.Fatal("Expected error for empty content")
	}
}

func TestYoudu_Send_Success(t *testing.T) {
	// Create a test server that mimics youdu-app-mcp API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify endpoint
		if r.URL.Path != "/api/v1/send_text_message" {
			t.Errorf("Expected path /api/v1/send_text_message, got %s", r.URL.Path)
		}

		// Verify Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify Authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Authorization 'Bearer test-token', got '%s'", auth)
		}

		// Verify request body
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if payload["to_user"] != "user123" {
			t.Errorf("Expected to_user 'user123', got '%v'", payload["to_user"])
		}

		if payload["content"] != "Test Subject\nTest Content" {
			t.Errorf("Expected content with subject and content, got '%v'", payload["content"])
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	// Initialize Youdu with test server URL
	y := NewYoudu()
	config := map[string]interface{}{
		"api_url": server.URL,
		"token":   "test-token",
	}
	err := y.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Send a message
	msg := &notifier.Message{
		To:      []string{"user123"},
		Subject: "Test Subject",
		Content: "Test Content",
	}

	resp, err := y.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=%v", resp.Success)
	}
}

func TestYoudu_Send_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   true,
			"message": "Invalid token",
		})
	}))
	defer server.Close()

	// Initialize Youdu with test server URL
	y := NewYoudu()
	config := map[string]interface{}{
		"api_url": server.URL,
		"token":   "invalid-token",
	}
	err := y.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Send a message
	msg := &notifier.Message{
		To:      []string{"user123"},
		Content: "Test Content",
	}

	resp, err := y.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected error when API returns error response")
	}

	if resp.Success {
		t.Error("Expected success=false when API returns error")
	}

	if resp.Error == "" {
		t.Error("Expected error message to be set")
	}
}

func TestYoudu_Send_WithoutSubject(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Verify content doesn't have subject prefix when subject is empty
		if payload["content"] != "Just Content" {
			t.Errorf("Expected content 'Just Content', got '%v'", payload["content"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	y := NewYoudu()
	config := map[string]interface{}{
		"api_url": server.URL,
		"token":   "test-token",
	}
	y.Initialize(config)

	// Send a message without subject
	msg := &notifier.Message{
		To:      []string{"user123"},
		Content: "Just Content",
	}

	resp, err := y.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=%v", resp.Success)
	}
}
