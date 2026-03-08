package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// EnhancedAPIServer represents the enhanced REST API server
type EnhancedAPIServer struct {
	client   *EnhancedClient
	config   *Config
	server   *http.Server
	upgrader websocket.Upgrader
}

// NewEnhancedAPIServer creates a new enhanced API server
func NewEnhancedAPIServer(client *EnhancedClient, config *Config) *EnhancedAPIServer {
	return &EnhancedAPIServer{
		client: client,
		config: config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// RunEnhancedAPIServer starts the enhanced REST API server
func RunEnhancedAPIServer(config *Config) error {
	client := NewEnhancedClient(config)

	// Set up event handlers
	client.OnStateChange(func(old, new ConnectionState) {
		log.Printf("[Connection] State changed: %s -> %s", old, new)
	})

	client.OnError(func(err error) {
		log.Printf("[Error] %v", err)
	})

	client.OnConnect(func() {
		log.Printf("[Connection] Connected to gateway")
	})

	client.OnDisconnect(func() {
		log.Printf("[Connection] Disconnected from gateway")
	})

	server := NewEnhancedAPIServer(client, config)
	return server.Start()
}

// Start starts the API server with graceful shutdown
func (s *EnhancedAPIServer) Start() error {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/chat", s.handleChat)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/", s.handleRoot)

	addr := fmt.Sprintf("%s:%d", s.config.API.Host, s.config.API.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.API.ReadTimeout,
		WriteTimeout: s.config.API.WriteTimeout,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting enhanced API server on %s", addr)
		log.Println("Endpoints:")
		log.Println("  GET  /health  - Health check")
		log.Println("  GET  /status  - Gateway status")
		log.Println("  POST /chat    - Send a message")
		log.Println("  GET  /stats   - Client statistics")
		log.Println("  GET  /ws      - WebSocket passthrough")

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.API.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Close client connection
	if err := s.client.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}

	log.Println("Server stopped")
	return nil
}

func (s *EnhancedAPIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"name":    "openclaw-channel-enhanced",
		"version": "1.1.0",
		"features": []string{
			"auto-reconnect",
			"heartbeat",
			"connection-monitoring",
			"graceful-shutdown",
		},
		"endpoints": []map[string]string{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/status", "description": "Gateway status"},
			{"method": "POST", "path": "/chat", "description": "Send a message"},
			{"method": "GET", "path": "/stats", "description": "Client statistics"},
			{"method": "GET", "path": "/ws", "description": "WebSocket passthrough"},
		},
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *EnhancedAPIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	state := s.client.GetState()
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"connected": state == StateConnected,
		"state":     state.String(),
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *EnhancedAPIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !s.client.IsConnected() {
		if err := s.client.Connect(); err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"state":   s.client.GetState().String(),
			})
			return
		}
	}

	status, err := s.client.GetStatus()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"status":  status,
		"state":   s.client.GetState().String(),
	})
}

func (s *EnhancedAPIServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	session := req.Session
	if session == "" {
		session = "agent:main:main"
	}

	response, err := s.client.Chat(session, req.Message)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"state":   s.client.GetState().String(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": response,
		"state":   s.client.GetState().String(),
	})
}

func (s *EnhancedAPIServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := s.client.GetStats()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
		"state":   s.client.GetState().String(),
	})
}

func (s *EnhancedAPIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if !s.client.IsConnected() {
		if err := s.client.Connect(); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
			return
		}
	}

	done := make(chan struct{})

	// Read from client
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			log.Printf("Client message: %s", message)
		}
	}()

	// Forward events to client
	for {
		select {
		case <-done:
			return
		case event := <-s.client.Events():
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}
