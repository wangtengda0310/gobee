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
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionState represents the current connection state
type ConnectionState int32

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateClosed
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateReconnecting:
		return "reconnecting"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// ClientStats holds client statistics
type ClientStats struct {
	ConnectCount      int64
	DisconnectCount   int64
	MessageCount      int64
	ErrorCount        int64
	ReconnectCount    int64
	LastConnectTime   time.Time
	LastMessageTime   time.Time
	LastErrorTime     time.Time
	LastError         string
}

// EnhancedClient represents an enhanced OpenClaw gateway client with reconnection support
type EnhancedClient struct {
	url        string
	token      string
	conn       *websocket.Conn
	mu         sync.Mutex
	connMu     sync.Mutex
	pendingMu  sync.Mutex // Separate lock for pending map to avoid deadlock
	pending    map[string]chan *Frame
	events     chan *Frame
	state      atomic.Int32
	ctx        context.Context
	cancel     context.CancelFunc
	debug      bool
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey

	// Enhanced features
	config          *Config
	stats           ClientStats
	statsMu         sync.Mutex
	reconnectCount  int
	onStateChange   func(old, new ConnectionState)
	onError         func(error)
	onConnect       func()
	onDisconnect    func()

	// Heartbeat
	heartbeatTicker *time.Ticker
	heartbeatStop   chan struct{}

	// HTTP client for alternative communication
	httpClient *http.Client
}

// NewEnhancedClient creates a new enhanced OpenClaw client
func NewEnhancedClient(cfg *Config) *EnhancedClient {
	ctx, cancel := context.WithCancel(context.Background())

	// Generate device key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	client := &EnhancedClient{
		url:        cfg.Gateway.URL,
		token:      cfg.Gateway.Token,
		pending:    make(map[string]chan *Frame),
		events:     make(chan *Frame, 100),
		ctx:        ctx,
		cancel:     cancel,
		debug:      os.Getenv("OPENCLAW_DEBUG") != "",
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     cfg,
		heartbeatStop: make(chan struct{}),
	}

	// Create HTTP client with timeout
	client.httpClient = &http.Client{
		Timeout: cfg.Gateway.Timeout,
	}

	client.state.Store(int32(StateDisconnected))

	return client
}

// GetState returns the current connection state
func (c *EnhancedClient) GetState() ConnectionState {
	return ConnectionState(c.state.Load())
}

// setState updates the connection state and notifies listeners
func (c *EnhancedClient) setState(newState ConnectionState) {
	oldState := ConnectionState(c.state.Swap(int32(newState)))
	if oldState != newState && c.onStateChange != nil {
		c.onStateChange(oldState, newState)
	}
}

// OnStateChange sets the state change callback
func (c *EnhancedClient) OnStateChange(fn func(old, new ConnectionState)) {
	c.onStateChange = fn
}

// OnError sets the error callback
func (c *EnhancedClient) OnError(fn func(error)) {
	c.onError = fn
}

// OnConnect sets the connect callback
func (c *EnhancedClient) OnConnect(fn func()) {
	c.onConnect = fn
}

// OnDisconnect sets the disconnect callback
func (c *EnhancedClient) OnDisconnect(fn func()) {
	c.onDisconnect = fn
}

// GetStats returns client statistics
func (c *EnhancedClient) GetStats() ClientStats {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	return c.stats
}

// updateStats updates client statistics
func (c *EnhancedClient) updateStats(fn func(*ClientStats)) {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	fn(&c.stats)
}

// Connect establishes a WebSocket connection to the gateway with auto-reconnect
func (c *EnhancedClient) Connect() error {
	return c.connectWithRetry(false)
}

// connectWithRetry attempts to connect with retry logic
func (c *EnhancedClient) connectWithRetry(isReconnect bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentState := c.GetState()
	if currentState == StateConnected {
		return nil
	}

	maxAttempts := 1
	if isReconnect {
		maxAttempts = c.config.Gateway.MaxReconnect
		if maxAttempts <= 0 {
			maxAttempts = 10
		}
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			c.setState(StateReconnecting)
			c.log("Reconnect attempt %d/%d after %v", attempt, maxAttempts, c.config.Gateway.ReconnectDelay)
			time.Sleep(c.config.Gateway.ReconnectDelay)
		} else {
			c.setState(StateConnecting)
		}

		err := c.doConnect()
		if err == nil {
			c.setState(StateConnected)
			c.updateStats(func(s *ClientStats) {
				s.ConnectCount++
				s.LastConnectTime = time.Now()
			})
			c.reconnectCount = 0

			if c.onConnect != nil {
				c.onConnect()
			}

			// Start heartbeat if enabled
			if c.config.Gateway.EnableHeartbeat {
				go c.startHeartbeat()
			}

			return nil
		}

		lastErr = err
		c.updateStats(func(s *ClientStats) {
			s.ErrorCount++
			s.LastErrorTime = time.Now()
			s.LastError = err.Error()
		})

		if c.onError != nil {
			c.onError(err)
		}

		c.log("Connection attempt %d failed: %v", attempt+1, err)
	}

	c.setState(StateDisconnected)
	return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

// doConnect performs the actual connection
func (c *EnhancedClient) doConnect() error {
	c.log("Connecting to %s", c.url)

	// Establish WebSocket connection
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = c.config.Gateway.Timeout

	conn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	c.log("WebSocket connected")

	// Read the first message (challenge)
	_, data, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read challenge: %w", err)
	}
	c.log("Received raw: %s", string(data))

	var challengeFrame Frame
	if err := json.Unmarshal(data, &challengeFrame); err != nil {
		conn.Close()
		return fmt.Errorf("failed to parse challenge: %w", err)
	}

	var nonce string
	var ts int64
	if challengeFrame.Type == FrameTypeEvent && challengeFrame.Event == "connect.challenge" {
		if payload, ok := challengeFrame.Payload.(map[string]interface{}); ok {
			if n, ok := payload["nonce"].(string); ok {
				nonce = n
			}
			if t, ok := payload["ts"].(float64); ok {
				ts = int64(t)
			}
		}
	}

	// Start read loop BEFORE sending connect frame to receive the response
	go c.readLoop()

	// Build v3 signature
	scopes := strings.Join(OperatorScopes, ",")
	signedAtMs := ts
	platform := getPlatform()
	deviceFamily := ""

	hash := sha256.Sum256(c.publicKey)
	deviceID := hex.EncodeToString(hash[:])
	c.log("Device ID: %s", deviceID)

	publicKeyBase64URL := base64.RawURLEncoding.EncodeToString(c.publicKey)

	payloadParts := []string{
		"v3",
		deviceID,
		"cli",
		"cli",
		RoleOperator,
		scopes,
		fmt.Sprintf("%d", signedAtMs),
		c.token,
		nonce,
		platform,
		deviceFamily,
	}
	payload := strings.Join(payloadParts, "|")
	c.log("Signature payload: %s", payload)

	signatureBytes := ed25519.Sign(c.privateKey, []byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(signatureBytes)
	c.log("Signature: %s...", signature[:32])

	// Build connect frame
	connectID := generateID()
	c.log("Connect ID: %s", connectID)

	responseCh := make(chan *Frame, 1)
	helloCh := make(chan *Frame, 1)
	c.log("Created response channels, acquiring pendingMu lock...")
	c.pendingMu.Lock()
	c.pending[connectID] = responseCh
	c.pending["__hello__"] = helloCh
	c.pendingMu.Unlock()
	c.log("Lock released, building connect frame...")

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
				Token: c.token,
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

	c.log("Connect frame created, preparing to send...")
	c.log("Sending connect frame with id=%s", connectID)
	c.log("Connect frame: Type=%s, Method=%s", connectFrame.Type, connectFrame.Method)
	if err := c.sendFrame(connectFrame); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.pendingMu.Unlock()
		conn.Close()
		return fmt.Errorf("failed to send connect request: %w", err)
	}
	c.log("Connect frame sent successfully, waiting for response...")

	// Wait for response
	select {
	case response := <-responseCh:
		if !response.OK {
			errMsg := "unknown error"
			if response.Error != nil {
				errMsg = response.Error.Message
			}
			c.pendingMu.Lock()
			delete(c.pending, connectID)
			delete(c.pending, "__hello__")
			c.pendingMu.Unlock()
			conn.Close()
			return fmt.Errorf("connect rejected: %s", errMsg)
		}
		c.pendingMu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.pendingMu.Unlock()
		c.log("Connect successful!")
	case <-helloCh:
		c.pendingMu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.pendingMu.Unlock()
		c.log("Connect successful (hello-ok)!")
	case <-time.After(30 * time.Second):
		c.pendingMu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.pendingMu.Unlock()
		conn.Close()
		return fmt.Errorf("connect failed: timeout waiting for response")
	case <-c.ctx.Done():
		c.pendingMu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.pendingMu.Unlock()
		conn.Close()
		return c.ctx.Err()
	}

	c.log("Connected successfully")
	return nil
}

// startHeartbeat starts the heartbeat ticker
func (c *EnhancedClient) startHeartbeat() {
	c.heartbeatTicker = time.NewTicker(c.config.Gateway.PingInterval)
	defer c.heartbeatTicker.Stop()

	for {
		select {
		case <-c.heartbeatTicker.C:
			if err := c.sendPing(); err != nil {
				c.log("Heartbeat failed: %v", err)
				c.handleDisconnect()
			}
		case <-c.heartbeatStop:
			return
		case <-c.ctx.Done():
			return
		}
	}
}

// sendPing sends a ping frame to keep the connection alive
func (c *EnhancedClient) sendPing() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.WriteMessage(websocket.PingMessage, []byte{})
}

// handleDisconnect handles disconnection events
func (c *EnhancedClient) handleDisconnect() {
	c.setState(StateDisconnected)
	c.updateStats(func(s *ClientStats) {
		s.DisconnectCount++
	})

	if c.onDisconnect != nil {
		c.onDisconnect()
	}

	// Attempt reconnect
	if c.config.Gateway.MaxReconnect > 0 {
		go c.connectWithRetry(true)
	}
}

// Close closes the client connection
func (c *EnhancedClient) Close() error {
	c.setState(StateClosed)
	c.cancel()

	close(c.heartbeatStop)

	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *EnhancedClient) IsConnected() bool {
	return c.GetState() == StateConnected
}

// Chat sends a message and waits for the complete response
func (c *EnhancedClient) Chat(sessionKey, text string) (string, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}

	id := generateID()
	frame := NewChatSendFrame(id, sessionKey, text)

	c.log("Sending chat.send with id=%s, session=%s", id, sessionKey)
	if err := c.sendFrame(frame); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	c.updateStats(func(s *ClientStats) {
		s.MessageCount++
		s.LastMessageTime = time.Now()
	})

	response, err := c.waitForResponse(id, 120*time.Second)
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
			c.log("Got runId=%s", runID)
		}
	}

	// Use agent.wait for response
	if runID != "" {
		waitID := generateID()
		waitFrame := NewAgentWaitFrame(waitID, runID, 60000)

		c.log("Sending agent.wait with id=%s, runId=%s", waitID, runID)
		if err := c.sendFrame(waitFrame); err != nil {
			return "", fmt.Errorf("failed to send agent.wait: %w", err)
		}

		waitResponse, err := c.waitForResponse(waitID, 90*time.Second)
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
func (c *EnhancedClient) GetStatus() (*StatusResult, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	id := generateID()
	frame := NewStatusFrame(id)

	if err := c.sendFrame(frame); err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	response, err := c.waitForResponse(id, 10*time.Second)
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
func (c *EnhancedClient) Events() <-chan *Frame {
	return c.events
}

// Internal methods

func (c *EnhancedClient) sendFrame(frame *Frame) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.sendFrameInternal(frame)
}

func (c *EnhancedClient) sendFrameInternal(frame *Frame) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("failed to marshal frame: %w", err)
	}

	c.log("Sending: %s", string(data))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *EnhancedClient) readLoop() {
	c.log("readLoop started")
	for {
		select {
		case <-c.ctx.Done():
			c.log("readLoop: context done")
			return
		default:
		}

		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()

		if conn == nil {
			c.log("readLoop: connection is nil, exiting")
			return
		}

		c.log("readLoop: waiting for message...")
		_, data, err := conn.ReadMessage()
		if err != nil {
			c.log("Read error: %v", err)
			if c.GetState() != StateClosed {
				c.handleDisconnect()
			}
			return
		}

		c.log("Received: %s", string(data))

		var frame Frame
		if err := json.Unmarshal(data, &frame); err != nil {
			c.log("Parse error: %v", err)
			continue
		}

		switch frame.Type {
		case FrameTypeResponse:
			c.pendingMu.Lock()
			if ch, ok := c.pending[frame.ID]; ok {
				ch <- &frame
				delete(c.pending, frame.ID)
			}
			c.pendingMu.Unlock()
		case FrameTypeEvent:
			switch frame.Event {
			case "connect.challenge":
				c.pendingMu.Lock()
				if ch, ok := c.pending["__challenge__"]; ok {
					ch <- &frame
				}
				c.pendingMu.Unlock()
			case "hello-ok":
				c.pendingMu.Lock()
				if ch, ok := c.pending["__hello__"]; ok {
					ch <- &frame
				}
				c.pendingMu.Unlock()
			default:
				select {
				case c.events <- &frame:
				default:
					c.log("Event channel full, dropping event")
				}
			}
		}
	}
}

func (c *EnhancedClient) waitForResponse(id string, timeout time.Duration) (*Frame, error) {
	ch := make(chan *Frame, 1)
	c.pendingMu.Lock()
	c.pending[id] = ch
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()

	select {
	case frame := <-ch:
		return frame, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for response")
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
}

func (c *EnhancedClient) log(format string, args ...interface{}) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
