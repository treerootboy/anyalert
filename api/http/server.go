package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/treerootboy/anyalert/internal/config"
	"github.com/treerootboy/anyalert/pkg/notifier"
)

// Server represents the HTTP server
type Server struct {
	manager      *notifier.Manager
	addr         string
	config       *config.Config
	tokenEnabled bool
}

// NewServer creates a new HTTP server
func NewServer(manager *notifier.Manager, cfg *config.Config, host string, port int) *Server {
	tokenEnabled := cfg.Token.Enabled && cfg.TokenManager != nil && cfg.TokenManager.Count() > 0
	return &Server{
		manager:      manager,
		addr:         fmt.Sprintf("%s:%d", host, port),
		config:       cfg,
		tokenEnabled: tokenEnabled,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register endpoints
	mux.HandleFunc("/api/v1/send", s.handleSend)
	mux.HandleFunc("/api/v1/broadcast", s.handleBroadcast)
	mux.HandleFunc("/api/v1/channels", s.handleListChannels)
	mux.HandleFunc("/health", s.handleHealth)

	log.Printf("Starting HTTP server on %s", s.addr)
	if s.tokenEnabled {
		log.Printf("Token authentication: enabled (%d tokens)", s.config.TokenManager.Count())
	} else {
		log.Printf("Token authentication: disabled")
	}

	// 应用中间件
	handler := s.corsMiddleware(mux)
	if s.tokenEnabled {
		handler = s.tokenAuthMiddleware(handler)
	}

	return http.ListenAndServe(s.addr, handler)
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleSend handles sending a notification to a single channel
func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Channel string            `json:"channel"`
		Message *notifier.Message `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.Channel == "" {
		http.Error(w, "Channel is required", http.StatusBadRequest)
		return
	}

	resp, err := s.manager.Send(r.Context(), req.Channel, req.Message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleBroadcast handles broadcasting a notification to multiple channels
func (s *Server) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Channels []string          `json:"channels"`
		Message  *notifier.Message `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.Channels) == 0 {
		http.Error(w, "At least one channel is required", http.StatusBadRequest)
		return
	}

	results := s.manager.Broadcast(r.Context(), req.Channels, req.Message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// handleListChannels returns the list of available channels
func (s *Server) handleListChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channels := s.manager.ListChannels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"channels": channels,
	})
}

// handleHealth returns the health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
	})
}

// tokenAuthMiddleware 验证 token
func (s *Server) tokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 跳过健康检查和 channels 列表
		if r.URL.Path == "/health" || r.URL.Path == "/api/v1/channels" {
			next.ServeHTTP(w, r)
			return
		}

		// 从 Authorization header 获取 token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Missing Authorization header",
			})
			return
		}

		// 支持两种格式: "Bearer <token>" 或 直接 "<token>"
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 验证 token
		if !s.config.TokenManager.Validate(token) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid token",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
