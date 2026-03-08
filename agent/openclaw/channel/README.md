# OpenClaw Channel (Go)

A Go implementation of an OpenClaw channel that supports multiple interaction modes with enhanced features.

## Status

✅ **完全可用** - 所有模式都已实现并测试通过。
✅ **反向连接** - 支持公网部署，实现内网穿透。

## Features

### Core Features
- **Interactive Mode**: Real-time command-line chat with OpenClaw
- **Batch Mode**: Process messages from stdin or command line arguments
- **RESTful API Mode**: HTTP endpoints for integration with other systems
- **MCP Mode**: Model Context Protocol server for AI assistant integration
- **Reverse Mode**: Reverse connection server for public deployment (内网穿透)

### Enhanced Features
- **Auto-reconnect**: Automatic reconnection on disconnection
- **Heartbeat/Keep-alive**: Connection health monitoring
- **Connection State Tracking**: Real-time state monitoring
- **Client Statistics**: Message counts, error tracking, timing
- **Graceful Shutdown**: Clean shutdown with timeout
- **Configuration Files**: YAML/JSON config support
- **Environment Variables**: Override settings via env vars

## Project Structure

```
agent/openclaw/channel/
├── main.go              # CLI entry point and commands
├── client.go            # WebSocket client (original)
├── client_enhanced.go   # Enhanced client with reconnection
├── config.go            # Configuration management
├── protocol.go          # Protocol types and frames (v3)
├── batch.go             # Batch mode implementation
├── interactive.go       # Interactive mode implementation
├── api.go               # REST API server (original)
├── api_enhanced.go      # Enhanced API server with monitoring
├── server.go            # Reverse connection server
├── connector.go         # Connector for bridging local gateway
├── mcp.go               # MCP server implementation
├── go.mod               # Go module definition
├── go.sum               # Go dependencies checksum
├── config.example.yaml  # Example configuration file
└── README.md            # This file
```

## Building

```bash
cd E:\github.com\gobee\agent\openclaw\channel
go mod tidy
go build -o openclaw-channel.exe .
```

## Usage

### Global Flags

```
  -g, --gateway string   OpenClaw gateway WebSocket URL (default "ws://127.0.0.1:18789")
  -t, --token string     OpenClaw gateway auth token
  -c, --config string    Configuration file path (YAML or JSON)
```

### Interactive Mode

```bash
./openclaw-channel interactive
```

Commands:
- `/quit`, `/exit`, `/q` - Exit
- `/status` - Show gateway status
- `/history` - Show recent messages
- `/session <key>` - Switch session
- `/help` - Show help

### Batch Mode

```bash
./openclaw-channel batch "Hello, how are you?"
echo "What time is it?" | ./openclaw-channel batch
```

### RESTful API Mode

```bash
./openclaw-channel api --port 8080
```

Endpoints:
- `GET /health` - Health check
- `GET /status` - Gateway status
- `POST /chat` - Send a message
- `GET /ws` - WebSocket passthrough

### Enhanced API Mode (Recommended)

```bash
# Using command-line flags
./openclaw-channel api-enhanced --port 8080 --gateway "ws://127.0.0.1:18789" --token "YOUR_TOKEN"

# Using configuration file
./openclaw-channel api-enhanced --config config.yaml

# Using environment variables
export OPENCLAW_GATEWAY_URL="ws://127.0.0.1:18789"
export OPENCLAW_GATEWAY_TOKEN="YOUR_TOKEN"
export OPENCLAW_API_HOST="0.0.0.0"
export OPENCLAW_API_PORT="8080"
./openclaw-channel api-enhanced
```

---

## Reverse Connection Mode (内网穿透)

This mode allows external clients to communicate with a local OpenClaw Gateway through a public server.

### Architecture

```
[External Client] --> [Reverse Server (Public)] <-- [Connector] <-- [Local OpenClaw Gateway]
```

- **Reverse Server**: Deployed on a public server, waits for Gateway connections
- **Connector**: Runs on the local machine, bridges local Gateway to remote Reverse Server
- **OpenClaw Gateway**: The local OpenClaw instance (no configuration changes needed)

### How OpenClaw Gateway Connects

The OpenClaw Gateway does NOT need any special configuration. The connector handles the connection:

1. **Connector** connects to local OpenClaw Gateway (`ws://127.0.0.1:18789`)
2. **Connector** connects to remote Reverse Server (`ws://your-server:port/gateway`)
3. **Connector** transparently bridges all WebSocket frames between them

The Gateway sees the connector as a regular client and responds to challenges normally.

### Setup Instructions

#### Step 1: Start Reverse Server (on public server)

```bash
# On itsnot.fun or any public server
./openclaw-channel reverse --port 18770 --token "YOUR_GATEWAY_TOKEN"

# Or with config file
./openclaw-channel reverse --config config.yaml
```

#### Step 2: Start Connector (on local machine with OpenClaw Gateway)

```bash
# The connector connects BOTH to local gateway AND remote reverse server
./openclaw-channel connector --remote ws://your-server:18770/gateway --token "YOUR_GATEWAY_TOKEN"

# Or with config file
./openclaw-channel connector --config config.yaml
```

#### Step 3: Test the Connection

```bash
# Check if gateway is connected
curl http://your-server:18770/health
# Expected: {"connected":true,"state":"connected","status":"ok",...}

# Send a message through the reverse server
curl -X POST -H "Content-Type: application/json" \
  -d '{"message":"Hello from remote","session":"agent:main:main"}' \
  http://your-server:18770/chat
```

### Reverse Server Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check with connection state |
| `/status` | GET | Gateway status (requires connected gateway) |
| `/chat` | POST | Send a message to OpenClaw |
| `/stats` | GET | Server statistics |
| `/gateway` | GET | WebSocket endpoint for Gateway/Connector |

### Configuration for Reverse Mode

```yaml
# config.yaml
gateway:
  url: ws://127.0.0.1:18789  # Local gateway URL (for connector)
  token: your-gateway-token-here

api:
  host: 0.0.0.0
  port: 18770
  externalAddress: your-server.com:18770  # For connector to connect
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENCLAW_GATEWAY_URL` | Local gateway WebSocket URL |
| `OPENCLAW_GATEWAY_TOKEN` | Gateway auth token |
| `OPENCLAW_REMOTE_URL` | Remote reverse server URL (for connector) |
| `OPENCLAW_API_HOST` | API server host |
| `OPENCLAW_API_PORT` | API server port |

---

## MCP Mode

```bash
./openclaw-channel mcp --transport stdio
```

Available tools:
- `openclaw_chat` - Send a message to OpenClaw
- `openclaw_status` - Get gateway status
- `openclaw_history` - Get chat history

## Protocol Implementation

This channel implements the OpenClaw WebSocket protocol v3:

### Device Authentication

1. **Device ID**: SHA256(publicKey).hex()
2. **Public Key**: base64url(publicKey)
3. **Signature Payload (v3)**: `v3|deviceId|clientId|clientMode|role|scopes|signedAtMs|token|nonce|platform|deviceFamily`
4. **Signature**: base64url(Ed25519.Sign(payload))

### Message Flow

1. Connect to WebSocket
2. Receive challenge event
3. Send connect request with device signature
4. Receive hello-ok event
5. Send chat.send request
6. Use agent.wait to wait for completion
7. Get response from chat.history

## Configuration

### Configuration File (config.yaml)

```yaml
# Gateway connection settings
gateway:
  url: ws://127.0.0.1:18789
  token: your-gateway-token-here
  reconnectDelay: 5s
  maxReconnect: 10
  pingInterval: 30s
  timeout: 30s
  enableHeartbeat: true

# API server settings
api:
  host: 0.0.0.0
  port: 8080
  externalAddress: ""  # For reverse mode
  readTimeout: 30s
  writeTimeout: 120s
  shutdownTimeout: 10s

# Logging settings
log:
  level: info  # debug, info, warn, error
  format: text  # text, json
  output: stdout
```

## Debug Mode

Set `OPENCLAW_DEBUG=1` to enable debug logging:

```bash
export OPENCLAW_DEBUG=1
./openclaw-channel batch "Hello"
```

## Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/gorilla/websocket` - WebSocket client
- `github.com/mark3labs/mcp-go` - MCP server implementation
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/cast` - Type conversion
- `gopkg.in/yaml.v3` - YAML parsing

## License

MIT
