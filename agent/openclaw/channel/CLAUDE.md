# OpenClaw Channel - Developer Guide

This document is for developers working on the OpenClaw Channel codebase.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        OpenClaw Channel                          │
├─────────────────────────────────────────────────────────────────┤
│  main.go (CLI)                                                  │
│    ├── interactive.go   - Interactive chat mode                 │
│    ├── batch.go         - Batch message processing              │
│    ├── api.go           - Basic REST API server                 │
│    ├── api_enhanced.go  - Enhanced API with Web UI              │
│    ├── server.go        - Reverse connection server             │
│    ├── connector.go     - Bridge local gateway to remote        │
│    └── mcp.go           - MCP server for AI assistants          │
├─────────────────────────────────────────────────────────────────┤
│  Core Components                                                │
│    ├── client.go / client_enhanced.go  - WebSocket client       │
│    ├── protocol.go                      - Protocol v3 types     │
│    └── config.go                        - Configuration mgmt    │
├─────────────────────────────────────────────────────────────────┤
│  Static Assets (embedded)                                       │
│    └── static/index.html  - Web UI (embedded via go:embed)     │
└─────────────────────────────────────────────────────────────────┘
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point with cobra commands |
| `server.go` | ReverseServer - accepts connections from connector |
| `api_enhanced.go` | EnhancedAPIServer - connects to gateway |
| `connector.go` | Connector - bridges local gateway ↔ remote server |
| `client_enhanced.go` | EnhancedClient - auto-reconnect, heartbeat |
| `protocol.go` | Frame types, protocol constants, v3 auth |
| `config.go` | ConfigManager - file/env loading |

## Reverse Connection Architecture

```
[External Clients]
       │
       ▼
┌──────────────────┐     WebSocket      ┌──────────────────┐
│  Reverse Server  │◄──────────────────►│    Connector     │
│  (public server) │                    │  (local machine) │
│  server.go       │                    │  connector.go    │
└──────────────────┘                    └────────┬─────────┘
       │                                         │
       │ HTTP API                                │ WebSocket
       ▼                                         ▼
  /health, /chat, /stats              ┌──────────────────┐
                                      │  OpenClaw Gateway │
                                      │  (localhost)      │
                                      └──────────────────┘
```

## Protocol v3 Authentication

The channel implements OpenClaw protocol v3 with Ed25519 device authentication:

```go
// Device ID: SHA256 hash of public key
deviceID := hex.EncodeToString(sha256.Sum256(publicKey)[:])

// Signature payload (pipe-separated)
payload := strings.Join([]string{
    "v3",
    deviceID,
    "cli",           // clientId
    "cli",           // clientMode
    "operator",      // role
    "chat,agent",    // scopes
    fmt.Sprintf("%d", signedAtMs),
    token,
    nonce,
    platform,
    deviceFamily,
}, "|")

// Ed25519 signature
signature := ed25519.Sign(privateKey, []byte(payload))
```

## Frame Types

```go
type Frame struct {
    Type   string      // "req", "res", "event"
    ID     string      // Unique ID for req/res correlation
    Method string      // Method name (for requests)
    Event  string      // Event name (for events)
    Params interface{} // Request parameters
    Payload interface{}// Response/event payload
    OK     bool        // Response success
    Error  *ErrorInfo  // Error details
}
```

### Key Methods
- `chat.send` - Send a message
- `chat.history` - Get message history
- `agent.wait` - Wait for agent completion
- `status` - Get gateway status

### Key Events
- `connect.challenge` - Authentication challenge
- `hello` / `hello-ok` - Connection handshake
- `chat` - Chat state updates (delta/final)
- `agent` - Agent stream events
- `health` - Health status broadcast

## Static Files (Embedded)

The Web UI is embedded using Go's `embed` package:

```go
//go:embed static/*
var staticFS embed.FS
```

This means:
- No external file dependencies
- Single binary deployment
- Web UI available at `/`

## Configuration

### Environment Variables
```bash
OPENCLAW_GATEWAY_URL      # Gateway WebSocket URL
OPENCLAW_GATEWAY_TOKEN    # Auth token
OPENCLAW_API_HOST         # API server host
OPENCLAW_API_PORT         # API server port
OPENCLAW_REMOTE_URL       # Remote server URL (connector)
OPENCLAW_LOG_LEVEL        # debug, info, warn, error
OPENCLAW_DEBUG            # Set to "1" for debug logs
```

### Config File (YAML/JSON)
See `config.example.yaml` for structure.

## Development

### Building
```bash
go build -o openclaw-channel .

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o openclaw-channel-linux .
```

### Testing
```bash
# Interactive mode
./openclaw-channel interactive

# API mode
./openclaw-channel api-enhanced --port 8080

# Reverse mode
./openclaw-channel reverse --port 18770 --token TOKEN
```

### Debug Mode
```bash
OPENCLAW_DEBUG=1 ./openclaw-channel batch "test"
```

## Deployment

### itsnot.fun (Production)
```bash
# On server
go install github.com/wangtengda0310/gobee/agent/openclaw/channel@latest
~/go/bin/channel reverse --port 18770 --token TOKEN

# On local machine
./openclaw-channel connector --remote ws://itsnot.fun:18770/gateway --token TOKEN
```

## CORS

CORS middleware is enabled for browser access:
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
```

## Common Tasks

### Add a new API endpoint
1. Add handler in `server.go` (ReverseServer) or `api_enhanced.go` (EnhancedAPIServer)
2. Register route in `Start()` method
3. Update Web UI in `static/index.html` if needed

### Add a new CLI command
1. Create command in `main.go` using cobra
2. Implement logic in separate file or existing module
3. Add flags and help text

### Modify protocol
1. Update types in `protocol.go`
2. Handle new frames in `readLoop()` of client/server
3. Update authentication if needed

## Dependencies

| Package | Usage |
|---------|-------|
| `github.com/gorilla/websocket` | WebSocket client/server |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/google/uuid` | UUID generation |
| `github.com/mark3labs/mcp-go` | MCP server |
| `gopkg.in/yaml.v3` | YAML config |
