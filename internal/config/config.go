package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/treerootboy/anyalert/pkg/database"
	"github.com/treerootboy/anyalert/pkg/token"
)

// Config represents the application configuration
type Config struct {
	Server       ServerConfig             `json:"server"`
	Channels     map[string]ChannelConfig `json:"channels"`
	Database     DatabaseConfig           `json:"database"`
	Token        TokenConfig              `json:"token"`
	DB           *database.DB             `json:"-"` // 数据库连接（不序列化）
	TokenManager *token.Manager           `json:"-"` // Token 管理器（不序列化）
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

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `json:"path"`
}

// TokenConfig represents token authentication configuration
type TokenConfig struct {
	Enabled bool `json:"enabled"`
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

	// 初始化数据库连接
	if err := config.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化 token 管理器
	if err := config.initTokenManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize token manager: %w", err)
	}

	return &config, nil
}

// initDatabase 初始化数据库连接
func (c *Config) initDatabase() error {
	// 如果没有配置数据库路径，使用默认值
	if c.Database.Path == "" {
		c.Database.Path = "./anyalert.db"
	}

	dbConfig := database.Config{
		Path: c.Database.Path,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return err
	}

	c.DB = db
	return nil
}

// initTokenManager 初始化 token 管理器
func (c *Config) initTokenManager() error {
	// 如果数据库未初始化，返回错误
	if c.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// 创建 token 管理器
	c.TokenManager = token.NewManager(c.DB.GetConnection())

	return nil
}

// Close 关闭配置相关资源
func (c *Config) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	cfg := &Config{
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
		Database: DatabaseConfig{
			Path: "./anyalert.db",
		},
		Token: TokenConfig{
			Enabled: false,
		},
	}

	// 初始化数据库和 token 管理器
	cfg.initDatabase()
	cfg.initTokenManager()

	return cfg
}
