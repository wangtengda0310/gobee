package main

import "time"

// Protocol version
const ProtocolVersion = 3

// Frame types
const (
 FrameTypeRequest = "req"
 FrameTypeResponse = "res"
 FrameTypeEvent   = "event"
)

// Client roles
const (
 RoleOperator = "operator"
 RoleNode     = "node"
)

// Common scopes
var OperatorScopes = []string{
 "operator.read",
 "operator.write",
}

// Default configuration values
const (
 DefaultReconnectDelay   = 5 * time.Second
 DefaultMaxReconnect     = 10
 DefaultPingInterval     = 30 * time.Second
 DefaultTimeout          = 30 * time.Second
 DefaultReadTimeout      = 30 * time.Second
 DefaultWriteTimeout     = 120 * time.Second
 DefaultShutdownTimeout  = 10 * time.Second
)

// Frame represents a WebSocket frame
type Frame struct {
 Type    string      `json:"type"`
 ID      string      `json:"id,omitempty"`
 Method  string      `json:"method,omitempty"`
 Params  interface{} `json:"params,omitempty"`
 OK      bool        `json:"ok,omitempty"`
 Payload interface{} `json:"payload,omitempty"`
 Error   *ErrorInfo  `json:"error,omitempty"`
 Event   string      `json:"event,omitempty"`
 Seq     int64       `json:"seq,omitempty"`
}

// ErrorInfo represents error details
type ErrorInfo struct {
 Message string `json:"message"`
 Code    string `json:"code,omitempty"`
 Details interface{} `json:"details,omitempty"`
}

// ClientInfo represents client identification
type ClientInfo struct {
 ID       string `json:"id"`
 Version  string `json:"version"`
 Platform string `json:"platform"`
 Mode     string `json:"mode"`
}

// ConnectParams represents connect request parameters
type ConnectParams struct {
 MinProtocol int                `json:"minProtocol"`
 MaxProtocol int                `json:"maxProtocol"`
 Client      ClientInfo         `json:"client"`
 Role        string             `json:"role"`
 Scopes      []string           `json:"scopes"`
 Caps        []string           `json:"caps,omitempty"`
 Commands    []string           `json:"commands,omitempty"`
 Permissions map[string]bool    `json:"permissions,omitempty"`
 Auth        *AuthInfo          `json:"auth,omitempty"`
 Locale      string             `json:"locale,omitempty"`
 UserAgent   string             `json:"userAgent,omitempty"`
 Device      *DeviceInfo        `json:"device,omitempty"`
}

// AuthInfo represents authentication info
type AuthInfo struct {
 Token    string `json:"token,omitempty"`
 Password string `json:"password,omitempty"`
}

// DeviceInfo represents device identity
type DeviceInfo struct {
 ID        string `json:"id"`
 PublicKey string `json:"publicKey,omitempty"`
 Signature string `json:"signature,omitempty"`
 SignedAt  int64  `json:"signedAt,omitempty"`
 Nonce     string `json:"nonce,omitempty"`
}

// ConnectChallenge represents the pre-connect challenge
type ConnectChallenge struct {
 Nonce string `json:"nonce"`
 TS    int64  `json:"ts"`
}

// HelloOK represents successful connection response
type HelloOK struct {
 Type     string       `json:"type"`
 Protocol int          `json:"protocol"`
 Policy   *PolicyInfo  `json:"policy,omitempty"`
 Auth     *AuthPayload `json:"auth,omitempty"`
}

// PolicyInfo represents gateway policy
type PolicyInfo struct {
 TickIntervalMs int `json:"tickIntervalMs"`
}

// AuthPayload represents auth response payload
type AuthPayload struct {
 DeviceToken string   `json:"deviceToken,omitempty"`
 Role        string   `json:"role"`
 Scopes      []string `json:"scopes"`
}

// ChatSendParams represents chat.send parameters
type ChatSendParams struct {
 SessionKey    string `json:"sessionKey,omitempty"`
 Message       string `json:"message"`
 IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// ChatHistoryParams represents chat.history parameters
type ChatHistoryParams struct {
 SessionKey   string `json:"sessionKey,omitempty"`
 Limit        int    `json:"limit,omitempty"`
 IncludeTools bool   `json:"includeTools,omitempty"`
}

// ChatSendResult represents chat.send response
type ChatSendResult struct {
 RunID  string `json:"runId"`
 Status string `json:"status"`
}

// ChatEvent represents a chat event
type ChatEvent struct {
 Type       string      `json:"type"`
 SessionKey string      `json:"sessionKey,omitempty"`
 RunID      string      `json:"runId,omitempty"`
 Seq        int         `json:"seq,omitempty"`
 State      string      `json:"state,omitempty"`   // "started", "streaming", "final", etc.
 Content    interface{} `json:"content,omitempty"`
 Done       bool        `json:"done,omitempty"`
 Aborted    bool        `json:"aborted,omitempty"`
 Error      string      `json:"error,omitempty"`
}

// StatusResult represents status response
type StatusResult struct {
 Gateway  *GatewayStatus `json:"gateway,omitempty"`
 Channels []interface{}  `json:"channels,omitempty"`
 Sessions []SessionInfo  `json:"sessions,omitempty"`
}

// GatewayStatus represents gateway status
type GatewayStatus struct {
 Running   bool   `json:"running"`
 Reachable bool   `json:"reachable"`
 Host      string `json:"host,omitempty"`
 Port      int    `json:"port,omitempty"`
 Version   string `json:"version,omitempty"`
}

// SessionInfo represents session information
type SessionInfo struct {
 Key   string `json:"key"`
 Kind  string `json:"kind"`
 Model string `json:"model"`
 Age   string `json:"age"`
}

// NewConnectFrame creates a new connect request frame
func NewConnectFrame(id string, token string, nonce string, publicKey string, signature string) *Frame {
	return &Frame{
		Type:   FrameTypeRequest,
		ID:     id,
		Method: "connect",
		Params: ConnectParams{
			MinProtocol: ProtocolVersion,
			MaxProtocol: ProtocolVersion,
			Client: ClientInfo{
				ID:       "cli",
				Version:  "1.0.0",
				Platform: getPlatform(),
				Mode:     "cli",
			},
			Role:   RoleOperator,
			Scopes: OperatorScopes,
			Auth: &AuthInfo{
				Token: token,
			},
			Locale:    "en-US",
			UserAgent: "openclaw-channel-go/1.0.0",
			Device: &DeviceInfo{
				ID:        deriveDeviceID(publicKey),
				PublicKey: publicKey,
				Signature: signature,
				SignedAt:  time.Now().UnixMilli(),
				Nonce:     nonce,
			},
		},
	}
}

func deriveDeviceID(publicKey string) string {
	// Device ID is the public key itself (fingerprint = public key)
	return publicKey
}

// NewChatSendFrame creates a new chat.send request frame
func NewChatSendFrame(id, sessionKey, message string) *Frame {
 return &Frame{
  Type:   FrameTypeRequest,
  ID:     id,
  Method: "chat.send",
  Params: ChatSendParams{
   SessionKey:     sessionKey,
   Message:        message,
   IdempotencyKey: generateID(),
  },
 }
}

// NewChatHistoryFrame creates a new chat.history request frame
func NewChatHistoryFrame(id, sessionKey string, limit int) *Frame {
 return &Frame{
  Type:   FrameTypeRequest,
  ID:     id,
  Method: "chat.history",
  Params: ChatHistoryParams{
   SessionKey:   sessionKey,
   Limit:        limit,
   IncludeTools: false,
  },
 }
}

// NewStatusFrame creates a new status request frame
func NewStatusFrame(id string) *Frame {
 return &Frame{
  Type:   FrameTypeRequest,
  ID:     id,
  Method: "status",
  Params: map[string]interface{}{},
 }
}

// NewAbortFrame creates a new chat.abort request frame
func NewAbortFrame(id, sessionKey string) *Frame {
 return &Frame{
  Type:   FrameTypeRequest,
  ID:     id,
  Method: "chat.abort",
  Params: map[string]interface{}{
   "sessionKey": sessionKey,
  },
 }
}

// AgentWaitParams represents agent.wait parameters
type AgentWaitParams struct {
 RunID     string `json:"runId"`
 TimeoutMs int    `json:"timeoutMs,omitempty"`
}

// NewAgentWaitFrame creates a new agent.wait request frame
func NewAgentWaitFrame(id, runID string, timeoutMs int) *Frame {
 return &Frame{
  Type:   FrameTypeRequest,
  ID:     id,
  Method: "agent.wait",
  Params: AgentWaitParams{
   RunID:     runID,
   TimeoutMs: timeoutMs,
  },
 }
}
