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
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents an OpenClaw gateway client
type Client struct {
	url        string
	token      string
	conn       *websocket.Conn
	mu         sync.Mutex
	connMu     sync.Mutex
	pending    map[string]chan *Frame
	events     chan *Frame
	connected  bool
	ctx        context.Context
	cancel     context.CancelFunc
	debug      bool
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewClient creates a new OpenClaw client
func NewClient(url, token string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Generate device key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err) // This should never fail
	}
	
	return &Client{
		url:        url,
		token:      token,
		pending:    make(map[string]chan *Frame),
		events:     make(chan *Frame, 100),
		ctx:        ctx,
		cancel:     cancel,
		debug:      os.Getenv("OPENCLAW_DEBUG") != "",
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (c *Client) log(format string, args ...interface{}) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// Connect establishes a WebSocket connection to the gateway
func (c *Client) Connect() error {
	// Check if already connected (with lock)
	c.connMu.Lock()
	if c.connected {
		c.connMu.Unlock()
		return nil
	}
	c.connMu.Unlock()

	c.log("Connecting to %s", c.url)

	// Establish WebSocket connection (without lock to avoid blocking readLoop)
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
	
	c.log("WebSocket connected")

	// Read the first message synchronously (should be the challenge)
	c.log("Waiting for challenge...")
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
		c.log("Received challenge frame")
		if payload, ok := challengeFrame.Payload.(map[string]interface{}); ok {
			if n, ok := payload["nonce"].(string); ok {
				nonce = n
				c.log("Got nonce: %s", nonce)
			}
			if t, ok := payload["ts"].(float64); ok {
				ts = int64(t)
				c.log("Got ts: %d", ts)
			}
		}
	} else {
		c.log("First frame was not a challenge: type=%s, event=%s", challengeFrame.Type, challengeFrame.Event)
	}

	// Now start the readLoop for subsequent messages
	go c.readLoop()

	// Build v3 signature payload
	// v3|deviceId|clientId|clientMode|role|scopes|signedAtMs|token|nonce|platform|deviceFamily
	scopes := strings.Join(OperatorScopes, ",")
	signedAtMs := ts
	platform := getPlatform()
	deviceFamily := ""
	
	// Calculate device ID: SHA256(publicKeyRaw).hex()
	hash := sha256.Sum256(c.publicKey)
	deviceID := hex.EncodeToString(hash[:])
	c.log("Device ID: %s", deviceID)
	
	// Public key in base64url format (no padding)
	publicKeyBase64URL := base64.RawURLEncoding.EncodeToString(c.publicKey)
	c.log("Public key (base64url): %s", publicKeyBase64URL)
	
	payloadParts := []string{
		"v3",
		deviceID,
		"cli",           // clientId
		"cli",           // clientMode
		RoleOperator,    // role
		scopes,          // scopes
		fmt.Sprintf("%d", signedAtMs),
		c.token,         // token
		nonce,           // nonce
		platform,        // platform
		deviceFamily,    // deviceFamily
	}
	payload := strings.Join(payloadParts, "|")
	c.log("Signature payload: %s", payload)
	
	// Sign with Ed25519, output as base64url
	signatureBytes := ed25519.Sign(c.privateKey, []byte(payload))
	signature := base64.RawURLEncoding.EncodeToString(signatureBytes)
	c.log("Signature (base64url): %s", signature[:32]+"...")
	
	// Build connect params
	connectID := generateID()
	
	// Set up pending handlers BEFORE sending
	responseCh := make(chan *Frame, 1)
	helloCh := make(chan *Frame, 1)
	c.mu.Lock()
	c.pending[connectID] = responseCh
	c.pending["__hello__"] = helloCh
	c.mu.Unlock()
	
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
	
	c.log("Sending connect frame with id=%s", connectID)
	if err := c.sendFrame(connectFrame); err != nil {
		c.mu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.mu.Unlock()
		conn.Close()
		return fmt.Errorf("failed to send connect request: %w", err)
	}

	// Wait for connect response OR hello-ok event
	c.log("Waiting for connect response...")
	select {
	case response := <-responseCh:
		c.log("Received response: ok=%v", response.OK)
		if !response.OK {
			errMsg := "unknown error"
			if response.Error != nil {
				errMsg = response.Error.Message
				c.log("Error details: %+v", response.Error)
			}
			c.mu.Lock()
			delete(c.pending, connectID)
			delete(c.pending, "__hello__")
			c.mu.Unlock()
			conn.Close()
			return fmt.Errorf("connect rejected: %s", errMsg)
		}
		c.mu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.mu.Unlock()
	case helloFrame := <-helloCh:
		c.log("Received hello-ok event")
		c.mu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.mu.Unlock()
		_ = helloFrame // We don't need the payload for now
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.mu.Unlock()
		conn.Close()
		return fmt.Errorf("connect failed: timeout waiting for response")
	case <-c.ctx.Done():
		c.mu.Lock()
		delete(c.pending, connectID)
		delete(c.pending, "__hello__")
		c.mu.Unlock()
		conn.Close()
		return c.ctx.Err()
	}

	c.connected = true
	c.log("Connected successfully")
	return nil
}
func (c *Client) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	c.cancel()
	c.connected = false

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.connected
}

// Chat sends a message and waits for the complete response
func (c *Client) Chat(sessionKey, text string) (string, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}

	// Send chat message
	id := generateID()
	frame := NewChatSendFrame(id, sessionKey, text)

	c.log("Sending chat.send with id=%s, session=%s", id, sessionKey)
	if err := c.sendFrame(frame); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	// Wait for initial response
	response, err := c.waitForResponse(id, 30*time.Second)
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

	// Parse run ID from response
	var runID string
	if payload, ok := response.Payload.(map[string]interface{}); ok {
		if rid, ok := payload["runId"].(string); ok {
			runID = rid
			c.log("Got runId=%s", runID)
		}
	}

	// If we have a runID, use agent.wait to get the response
	if runID != "" {
		waitID := generateID()
		waitFrame := NewAgentWaitFrame(waitID, runID, 60000) // 60 second timeout
		
		c.log("Sending agent.wait with id=%s, runId=%s", waitID, runID)
		if err := c.sendFrame(waitFrame); err != nil {
			return "", fmt.Errorf("failed to send agent.wait: %w", err)
		}
		
		// Wait for agent.wait response
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
		
		// The agent.wait response contains the result
		// For now, just return a success message since the actual response
		// is in the agent's tool calls which we can observe in the main session
		return "Message sent successfully. Check the agent session for response.", nil
	}

	// Fallback: collect events until done (old behavior)
	var result string
	timeout := time.After(120 * time.Second)

	for {
		select {
		case event := <-c.events:
			c.log("Received event: %s", event.Event)
			switch event.Event {
			case "chat":
				chatEvent := parseChatEvent(event.Payload)
				if chatEvent != nil {
					if runID != "" && chatEvent.RunID != "" && chatEvent.RunID != runID {
						continue
					}
					
					// Handle content
					if chatEvent.Content != nil {
						if str, ok := chatEvent.Content.(string); ok {
							result += str
						} else if m, ok := chatEvent.Content.(map[string]interface{}); ok {
							if text, ok := m["text"].(string); ok {
								result += text
							}
						}
					}
					
					// Check for completion (state == "final" or done == true)
					if chatEvent.State == "final" || chatEvent.Done || chatEvent.Aborted {
						c.log("Chat complete: state=%s, done=%v, aborted=%v", chatEvent.State, chatEvent.Done, chatEvent.Aborted)
						// If we have content, return it; otherwise continue to wait for agent event
						if result != "" {
							return result, nil
						}
					}
				}
			case "agent":
				// Agent event may contain the response content
				if payload, ok := event.Payload.(map[string]interface{}); ok {
					c.log("Agent event payload: %+v", payload)
					// Check for text content
					if text, ok := payload["text"].(string); ok {
						result += text
					}
					// Check for content structure
					if content, ok := payload["content"].(map[string]interface{}); ok {
						if text, ok := content["text"].(string); ok {
							result += text
						}
					}
					// Check for done state
					if done, ok := payload["done"].(bool); ok && done {
						return result, nil
					}
					if state, ok := payload["state"].(string); ok && state == "final" {
						return result, nil
					}
				}
			}
		case <-timeout:
			if result != "" {
				return result, nil
			}
			return result, fmt.Errorf("timeout waiting for response")
		case <-c.ctx.Done():
			return result, c.ctx.Err()
		}
	}
}

// GetStatus gets the gateway status
func (c *Client) GetStatus() (*StatusResult, error) {
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

// GetHistory gets chat history
func (c *Client) GetHistory(sessionKey string, limit int) (interface{}, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	id := generateID()
	frame := NewChatHistoryFrame(id, sessionKey, limit)

	if err := c.sendFrame(frame); err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
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
		return nil, fmt.Errorf("history failed: %s", errMsg)
	}

	// Return the full payload which contains messages array
	return response.Payload, nil
}

// Abort aborts the current chat
func (c *Client) Abort(sessionKey string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	id := generateID()
	frame := NewAbortFrame(id, sessionKey)

	if err := c.sendFrame(frame); err != nil {
		return fmt.Errorf("failed to abort: %w", err)
	}

	return nil
}

// Events returns the event channel
func (c *Client) Events() <-chan *Frame {
	return c.events
}

// Internal methods

func (c *Client) sendFrame(frame *Frame) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.sendFrameInternal(frame)
}

// sendFrameInternal sends a frame without acquiring connMu (caller must hold it)
func (c *Client) sendFrameInternal(frame *Frame) error {
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

func (c *Client) readLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()

		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			c.log("Read error: %v", err)
			c.connMu.Lock()
			c.connected = false
			c.connMu.Unlock()
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
			c.mu.Lock()
			if ch, ok := c.pending[frame.ID]; ok {
				ch <- &frame
				delete(c.pending, frame.ID)
			} else {
				c.log("No pending handler for response id=%s", frame.ID)
			}
			c.mu.Unlock()
		case FrameTypeEvent:
			c.log("Event: %s", frame.Event)
			switch frame.Event {
			case "connect.challenge":
				c.mu.Lock()
				if ch, ok := c.pending["__challenge__"]; ok {
					ch <- &frame
				}
				c.mu.Unlock()
			case "hello-ok":
				// Connect success response
				c.mu.Lock()
				if ch, ok := c.pending["__hello__"]; ok {
					ch <- &frame
				}
				c.mu.Unlock()
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

func (c *Client) waitForResponse(id string, timeout time.Duration) (*Frame, error) {
	ch := make(chan *Frame, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
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

// waitForResponseAsync returns a channel that will receive the response
func (c *Client) waitForResponseAsync(id string) <-chan *Frame {
	ch := make(chan *Frame, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	// Return the channel - caller must handle cleanup
	return ch
}

// Helper functions

func generateID() string {
	return uuid.New().String()
}

func getPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
}

func parseChatEvent(payload interface{}) *ChatEvent {
	if payload == nil {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}

	var event ChatEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil
	}

	return &event
}
