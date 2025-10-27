package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DB 数据库连接封装
type DB struct {
	conn *sql.DB
	path string
}

// Config 数据库配置
type Config struct {
	Path string `json:"path"` // 数据库文件路径
}

// New 创建新的数据库连接
func New(config Config) (*DB, error) {
	// 如果路径为空，使用默认路径
	if config.Path == "" {
		config.Path = "./anyalert.db"
	}

	// 确保目录存在
	dir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开数据库连接
	conn, err := sql.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	db := &DB{
		conn: conn,
		path: config.Path,
	}

	// 初始化数据库结构
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("初始化数据库结构失败: %w", err)
	}

	return db, nil
}

// initSchema 初始化数据库结构
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tokens (
		id TEXT PRIMARY KEY,
		value TEXT UNIQUE NOT NULL,
		description TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		expires_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_tokens_value ON tokens(value);
	CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		slack TEXT,
		youdu TEXT,
		phone TEXT,
		sms TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection 获取底层数据库连接
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// GetPath 获取数据库文件路径
func (db *DB) GetPath() string {
	return db.path
}
