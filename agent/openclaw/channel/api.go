package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// APIServer represents the REST API server
type APIServer struct {
	client *Client
	host   string
	port   int
	server *http.Server
	mu     sync.Mutex
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message  string `json:"message"`
	Session  string `json:"session,omitempty"`
	Timeout  int    `json:"timeout,omitempty"`  // seconds
	Stream   bool   `json:"stream,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	RunID   string `json:"runId,omitempty"`
}

// StatusResponse represents a status response
type StatusResponse struct {
	Success bool          `json:"success"`
	Status  *StatusResult `json:"status,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// NewAPIServer creates a new API server
func NewAPIServer(client *Client, host string, port int) *APIServer {
	return &APIServer{
		client: client,
		host:   host,
		port:   port,
	}
}

// RunAPIServer starts the REST API server
func RunAPIServer(client *Client, host string, port int) error {
	server := NewAPIServer(client, host, port)
	return server.Start()
}

// Start starts the API server
func (s *APIServer) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Status endpoint
	mux.HandleFunc("/status", s.handleStatus)

	// Chat endpoint
	mux.HandleFunc("/chat", s.handleChat)

	// WebSocket passthrough
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Root info
	mux.HandleFunc("/", s.handleRoot)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	log.Printf("Starting API server on %s", addr)
	log.Println("Endpoints:")
	log.Println("  GET  /health  - Health check")
	log.Println("  GET  /status  - Gateway status")
	log.Println("  POST /chat    - Send a message")
	log.Println("  GET  /ws      - WebSocket passthrough")

	return s.server.ListenAndServe()
}

func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"name":    "openclaw-channel",
		"version": "1.0.0",
		"endpoints": []map[string]string{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/status", "description": "Gateway status"},
			{"method": "POST", "path": "/chat", "description": "Send a message"},
			{"method": "GET", "path": "/ws", "description": "WebSocket passthrough"},
		},
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Ensure connection
	if !s.client.IsConnected() {
		if err := s.client.Connect(); err != nil {
			writeJSON(w, http.StatusOK, StatusResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
	}

	status, err := s.client.GetStatus()
	if err != nil {
		writeJSON(w, http.StatusOK, StatusResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, StatusResponse{
		Success: true,
		Status:  status,
	})
}

func (s *APIServer) handleChat(w http.ResponseWriter, r *http.Request) {
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

	// Ensure connection
	if !s.client.IsConnected() {
		if err := s.client.Connect(); err != nil {
			writeJSON(w, http.StatusOK, ChatResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to connect: %v", err),
			})
			return
		}
	}

	// Send message
	response, err := s.client.Chat(session, req.Message)
	if err != nil {
		writeJSON(w, http.StatusOK, ChatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, ChatResponse{
		Success: true,
		Message: response,
	})
}

func (s *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Connect to OpenClaw gateway
	if !s.client.IsConnected() {
		if err := s.client.Connect(); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
			return
		}
	}

	// Relay messages between client and gateway
	done := make(chan struct{})

	// Read from client and forward to gateway
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// Forward to gateway (would need direct connection access)
			log.Printf("Client message: %s", message)
		}
	}()

	// Read from gateway events and forward to client
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

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
