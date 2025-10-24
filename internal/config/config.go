package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig             `json:"server"`
	Channels map[string]ChannelConfig `json:"channels"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	HTTP HTTPConfig `json:"http"`
	GRPC GRPCConfig `json:"grpc"`
}

// HTTPConfig represents HTTP server configuration
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Host    string `json:"host"`
}

// GRPCConfig represents gRPC server configuration
type GRPCConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Host    string `json:"host"`
}

// ChannelConfig represents configuration for a notification channel
type ChannelConfig struct {
	Enabled bool                   `json:"enabled"`
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
}

// LoadConfig loads configuration from a file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Enabled: true,
				Port:    8080,
				Host:    "0.0.0.0",
			},
			GRPC: GRPCConfig{
				Enabled: true,
				Port:    9090,
				Host:    "0.0.0.0",
			},
		},
		Channels: make(map[string]ChannelConfig),
	}
}
