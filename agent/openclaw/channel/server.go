package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ReverseServer represents a WebSocket server that accepts connections from OpenClaw Gateway
type ReverseServer struct {
	config       *Config
	conn         *websocket.Conn
	mu           sync.Mutex
	connMu       sync.Mutex
	pending      map[string]chan *Frame
	events       chan *Frame
	state        atomic.Int32
	ctx          context.Context
	cancel       context.CancelFunc
	debug        bool
	privateKey   ed25519.PrivateKey
	publicKey    ed25519.PublicKey

	// Server state
	server       *http.Server
	upgrader     websocket.Upgrader

	// Statistics
	stats        ClientStats
	statsMu      sync.Mutex
}

// NewReverseServer creates a new reverse connection server
func NewReverseServer(cfg *Config) *ReverseServer {
	ctx, cancel := context.WithCancel(context.Background())

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	return &ReverseServer{
		config:     cfg,
		pending:    make(map[string]chan *Frame),
		events:     make(chan *Frame, 100),
		ctx:        ctx,
		cancel:     cancel,
		debug:      os.Getenv("OPENCLAW_DEBUG") != "",
		privateKey: privateKey,
		publicKey:  publicKey,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// GetState returns the current connection state
func (s *ReverseServer) GetState() ConnectionState {
	return ConnectionState(s.state.Load())
}

// setState updates the connection state
func (s *ReverseServer) setState(newState ConnectionState) {
	s.state.Store(int32(newState))
}

// IsConnected returns whether there's an active gateway connection
func (s *ReverseServer) IsConnected() bool {
	return s.GetState() == StateConnected
}

// GetStats returns server statistics
func (s *ReverseServer) GetStats() ClientStats {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	return s.stats
}

// updateStats updates server statistics
func (s *ReverseServer) updateStats(fn func(*ClientStats)) {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	fn(&s.stats)
}

// Start starts the reverse connection server
func (s *ReverseServer) Start() error {
	mux := http.NewServeMux()

	// WebSocket endpoint for gateway connections
	mux.HandleFunc("/gateway", s.handleGatewayConnection)

	// API endpoints (proxied to gateway)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/chat", s.handleChat)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/", s.handleRoot)

	addr := fmt.Sprintf("%s:%d", s.config.API.Host, s.config.API.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.API.ReadTimeout,
		WriteTimeout: s.config.API.WriteTimeout,
	}

	s.setState(StateDisconnected)

	log.Printf("Starting reverse connection server on %s", addr)
	log.Println("Waiting for OpenClaw Gateway to connect...")
	log.Println("API Endpoints:")
	log.Println("  GET  /health  - Health check")
	log.Println("  GET  /status  - Gateway status")
	log.Println("  POST /chat    - Send a message")
	log.Println("  GET  /stats   - Server statistics")
	log.Println("")
	log.Println("Gateway should connect to: ws://<this-server>:%d/gateway", s.config.API.Port)

	return s.server.ListenAndServe()
}

// handleGatewayConnection handles WebSocket connections from OpenClaw Gateway
func (s *ReverseServer) handleGatewayConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("Gateway connection attempt from:", r.RemoteAddr)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Check if we already have a connection
	if s.IsConnected() {
		log.Println("Rejecting connection - already have an active gateway")
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "gateway already connected"}`))
		conn.Close()
		return
	}

	s.connMu.Lock()
	s.conn = conn
	s.connMu.Unlock()

	s.setState(StateConnected)
	s.updateStats(func(st *ClientStats) {
		st.ConnectCount++
		st.LastConnectTime = time.Now()
	})

	log.Println("Gateway connected successfully!")

	// Don't send hello - wait for the gateway (via connector) to send challenge
	// The connector will transparently relay messages between the real OpenClaw Gateway
	// and this reverse server

	// Start read loop
	go s.readLoop()

	// Wait for disconnect
	<-s.ctx.Done()
}

// sendHello sends a hello message to the gateway
func (s *ReverseServer) sendHello() {
	helloFrame := &Frame{
		Type:   FrameTypeEvent,
		Event:  "hello",
		Seq:    1,
		Payload: map[string]interface{}{
			"server":   "openclaw-channel-reverse",
			"version":  "1.0.0",
			"protocol": ProtocolVersion,
		},
	}
	s.sendFrame(helloFrame)
}

// readLoop reads messages from the gateway connection
func (s *ReverseServer) readLoop() {
	defer func() {
		s.setState(StateDisconnected)
		s.updateStats(func(st *ClientStats) {
			st.DisconnectCount++
		})
		s.connMu.Lock()
		if s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
		s.connMu.Unlock()
		log.Println("Gateway disconnected")
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		s.connMu.Lock()
		conn := s.conn
		s.connMu.Unlock()

		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			s.log("Read error: %v", err)
			return
		}

		s.log("Received: %s", string(data))

		var frame Frame
		if err := json.Unmarshal(data, &frame); err != nil {
			s.log("Parse error: %v", err)
			continue
		}

		switch frame.Type {
		case FrameTypeRequest:
			// Handle requests from gateway (e.g., connect.challenge)
			s.handleRequest(&frame)
		case FrameTypeResponse:
			// Route responses to waiting handlers
			s.mu.Lock()
			if ch, ok := s.pending[frame.ID]; ok {
				ch <- &frame
				delete(s.pending, frame.ID)
			}
			s.mu.Unlock()
		case FrameTypeEvent:
			// Handle events from gateway
			switch frame.Event {
			case "connect.challenge":
				s.handleChallenge(&frame)
			default:
				select {
				case s.events <- &frame:
				default:
					s.log("Event channel full, dropping event")
				}
			}
		}
	}
}

// handleRequest handles incoming requests from the gateway
func (s *ReverseServer) handleRequest(frame *Frame) {
	s.log("Handling request: %s", frame.Method)

	switch frame.Method {
	case "connect":
		// Gateway is trying to connect to us
		s.handleGatewayConnect(frame)
	default:
		// Unknown method, send error
		response := &Frame{
			Type:  FrameTypeResponse,
			ID:    frame.ID,
			OK:    false,
			Error: &ErrorInfo{Message: fmt.Sprintf("unknown method: %s", frame.Method)},
		}
		s.sendFrame(response)
	}
}

// handleChallenge handles the connect.challenge event from gateway
func (s *ReverseServer) handleChallenge(frame *Frame) {
	s.log("Received challenge from gateway")

	var nonce string
	var ts int64
	if payload, ok := frame.Payload.(map[string]interface{}); ok {
		if n, ok := payload["nonce"].(string); ok {
			nonce = n
		}
		if t, ok := payload["ts"].(float64); ok {
			ts = int64(t)
		}
	}

	// Build v3 signature
	scopes := strings.Join(OperatorScopes, ",")
	signedAtMs := ts
	platform := getPlatform()
	deviceFamily := ""

	hash := sha256.Sum256(s.publicKey)
	deviceID := hex.EncodeToString(hash[:])

	publicKeyBase64URL := base64.RawURLEncoding.EncodeToString(s.publicKey)

	payloadParts := []string{
		"v3",
		deviceID,
		"cli",
		"cli",
		RoleOperator,
		scopes,
		fmt.Sprintf("%d", signedAtMs),
		s.config.Gateway.Token,
		nonce,
		platform,
		deviceFamily,
	}
	payload := strings.Join(payloadParts, "|")

	signatureBytes := ed25519.Sign(s.privateKey, []byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(signatureBytes)

	// Send connect response
	connectID := generateID()
	connectFrame := &Frame{
		Type:   FrameTypeRequest,
		ID:     connectID,
		Method: "connect",
		Params: ConnectParams{
			MinProtocol: ProtocolVersion,
			MaxProtocol: ProtocolVersion,
			Client: ClientInfo{
				ID:       "cli",
				Version:  "1.0.0",
				Platform: platform,
				Mode:     "cli",
			},
			Role:   RoleOperator,
			Scopes: OperatorScopes,
			Auth: &AuthInfo{
				Token: s.config.Gateway.Token,
			},
			Locale:    "en-US",
			UserAgent: "openclaw-channel-go/1.0.0",
			Device: &DeviceInfo{
				ID:        deviceID,
				PublicKey: publicKeyBase64URL,
				Signature: signature,
				SignedAt:  ts,
				Nonce:     nonce,
			},
		},
	}

	s.log("Sending connect response")
	s.sendFrame(connectFrame)
}

// handleGatewayConnect handles gateway's connect request
func (s *ReverseServer) handleGatewayConnect(frame *Frame) {
	response := &Frame{
		Type: FrameTypeResponse,
		ID:   frame.ID,
		OK:   true,
		Payload: map[string]interface{}{
			"status":   "connected",
			"server":   "openclaw-channel-reverse",
			"protocol": ProtocolVersion,
		},
	}
	s.sendFrame(response)
	s.log("Gateway connect acknowledged")
}

// Chat sends a message through the gateway connection
func (s *ReverseServer) Chat(sessionKey, text string) (string, error) {
	if !s.IsConnected() {
		return "", fmt.Errorf("gateway not connected")
	}

	id := generateID()
	frame := NewChatSendFrame(id, sessionKey, text)

	s.log("Sending chat.send with id=%s, session=%s", id, sessionKey)
	if err := s.sendFrame(frame); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	s.updateStats(func(st *ClientStats) {
		st.MessageCount++
		st.LastMessageTime = time.Now()
	})

	response, err := s.waitForResponse(id, 120*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to get response: %w", err)
	}

	if !response.OK {
		errMsg := "unknown error"
		if response.Error != nil {
			errMsg = response.Error.Message
		}
		return "", fmt.Errorf("chat.send failed: %s", errMsg)
	}

	// Parse run ID
	var runID string
	if payload, ok := response.Payload.(map[string]interface{}); ok {
		if rid, ok := payload["runId"].(string); ok {
			runID = rid
			s.log("Got runId=%s", runID)
		}
	}

	// Use agent.wait for response
	if runID != "" {
		waitID := generateID()
		waitFrame := NewAgentWaitFrame(waitID, runID, 60000)

		s.log("Sending agent.wait with id=%s, runId=%s", waitID, runID)
		if err := s.sendFrame(waitFrame); err != nil {
			return "", fmt.Errorf("failed to send agent.wait: %w", err)
		}

		waitResponse, err := s.waitForResponse(waitID, 90*time.Second)
		if err != nil {
			return "", fmt.Errorf("failed to get agent.wait response: %w", err)
		}

		if !waitResponse.OK {
			errMsg := "unknown error"
			if waitResponse.Error != nil {
				errMsg = waitResponse.Error.Message
			}
			return "", fmt.Errorf("agent.wait failed: %s", errMsg)
		}

		return "Message sent successfully. Check the agent session for response.", nil
	}

	return "Message sent.", nil
}

// GetStatus gets the gateway status
func (s *ReverseServer) GetStatus() (*StatusResult, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("gateway not connected")
	}

	id := generateID()
	frame := NewStatusFrame(id)

	if err := s.sendFrame(frame); err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	response, err := s.waitForResponse(id, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if !response.OK {
		errMsg := "unknown error"
		if response.Error != nil {
			errMsg = response.Error.Message
		}
		return nil, fmt.Errorf("status failed: %s", errMsg)
	}

	result := &StatusResult{}
	if payload, ok := response.Payload.(map[string]interface{}); ok {
		if data, err := json.Marshal(payload); err == nil {
			json.Unmarshal(data, result)
		}
	}

	return result, nil
}

// Events returns the event channel
func (s *ReverseServer) Events() <-chan *Frame {
	return s.events
}

// Close closes the server
func (s *ReverseServer) Close() error {
	s.cancel()
	s.setState(StateClosed)

	s.connMu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.connMu.Unlock()

	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// sendFrame sends a frame to the gateway
func (s *ReverseServer) sendFrame(frame *Frame) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("failed to marshal frame: %w", err)
	}

	s.log("Sending: %s", string(data))
	return s.conn.WriteMessage(websocket.TextMessage, data)
}

// waitForResponse waits for a response to a specific request
func (s *ReverseServer) waitForResponse(id string, timeout time.Duration) (*Frame, error) {
	ch := make(chan *Frame, 1)
	s.mu.Lock()
	s.pending[id] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
	}()

	select {
	case frame := <-ch:
		return frame, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for response")
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	}
}

// HTTP Handlers

func (s *ReverseServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	info := map[string]interface{}{
		"name":    "openclaw-channel-reverse",
		"version": "1.0.0",
		"mode":    "reverse-connection",
		"description": "This server waits for OpenClaw Gateway to connect to it.",
		"gateway_endpoint": fmt.Sprintf("ws://%s/gateway", r.Host),
		"endpoints": []map[string]string{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/status", "description": "Gateway status"},
			{"method": "POST", "path": "/chat", "description": "Send a message"},
			{"method": "GET", "path": "/stats", "description": "Server statistics"},
			{"method": "GET", "path": "/gateway", "description": "WebSocket endpoint for Gateway"},
		},
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *ReverseServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	state := s.GetState()
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"connected": state == StateConnected,
		"state":     state.String(),
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *ReverseServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !s.IsConnected() {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "gateway not connected",
			"state":   s.GetState().String(),
		})
		return
	}

	status, err := s.GetStatus()
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
		"state":   s.GetState().String(),
	})
}

func (s *ReverseServer) handleChat(w http.ResponseWriter, r *http.Request) {
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

	if !s.IsConnected() {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   "gateway not connected",
			"state":   s.GetState().String(),
		})
		return
	}

	response, err := s.Chat(session, req.Message)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"state":   s.GetState().String(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": response,
		"state":   s.GetState().String(),
	})
}

func (s *ReverseServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := s.GetStats()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
		"state":   s.GetState().String(),
	})
}

func (s *ReverseServer) log(format string, args ...interface{}) {
	if s.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
