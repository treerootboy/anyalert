package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Message represents a notification message
type Message struct {
	To       []string          `json:"to"`
	Subject  string            `json:"subject"`
	Content  string            `json:"content"`
	Priority string            `json:"priority,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SendRequest represents a request to send a notification
type SendRequest struct {
	Channel string   `json:"channel"`
	Message *Message `json:"message"`
}

// BroadcastRequest represents a request to broadcast a notification
type BroadcastRequest struct {
	Channels []string `json:"channels"`
	Message  *Message `json:"message"`
}

func main() {
	baseURL := "http://localhost:8080"

	// Example 1: Send to single channel
	fmt.Println("=== Example 1: Send to Slack ===")
	sendToSlack(baseURL)

	// Example 2: Broadcast to multiple channels
	fmt.Println("\n=== Example 2: Broadcast to multiple channels ===")
	broadcastMessage(baseURL)

	// Example 3: List available channels
	fmt.Println("\n=== Example 3: List available channels ===")
	listChannels(baseURL)
}

func sendToSlack(baseURL string) {
	req := SendRequest{
		Channel: "slack",
		Message: &Message{
			To:       []string{"#general"},
			Subject:  "系统通知",
			Content:  "这是一条来自 AnyAlert 的测试消息",
			Priority: "normal",
		},
	}

	resp, err := sendRequest(baseURL+"/api/v1/send", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp)
}

func broadcastMessage(baseURL string) {
	req := BroadcastRequest{
		Channels: []string{"slack", "sms"},
		Message: &Message{
			To:       []string{"user@example.com", "+86-13800138000"},
			Subject:  "紧急告警",
			Content:  "服务器 CPU 使用率超过 90%",
			Priority: "urgent",
		},
	}

	resp, err := sendRequest(baseURL+"/api/v1/broadcast", req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp)
}

func listChannels(baseURL string) {
	resp, err := http.Get(baseURL + "/api/v1/channels")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Available channels: %s\n", body)
}

func sendRequest(url string, data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}
