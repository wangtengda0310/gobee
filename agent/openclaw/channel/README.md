# OpenClaw Channel (Go)

A Go implementation of an OpenClaw channel that supports multiple interaction modes with enhanced features.

## Status

✅ **完全可用** - 所有模式都已实现并测试通过。
✅ **反向连接** - 支持公网部署，实现内网穿透。
✅ **Web UI** - 内置现代化聊天界面。

## Features

- **Web UI**: 现代化聊天界面，深色主题，响应式设计
- **Interactive Mode**: 实时命令行聊天
- **Batch Mode**: 批量消息处理
- **RESTful API**: HTTP 接口，支持 CORS
- **Reverse Mode**: 反向连接，公网部署
- **MCP Mode**: Model Context Protocol 服务器

## Quick Start

### 安装

```bash
# 从源码编译
git clone https://github.com/wangtengda0310/gobee.git
cd gobee/agent/openclaw/channel
go build -o openclaw-channel .

# 或直接安装
go install github.com/wangtengda0310/gobee/agent/openclaw/channel@latest
```

### 本地使用

```bash
# 启动 API 服务器（带 Web UI）
./openclaw-channel api-enhanced --port 8080 --token YOUR_TOKEN

# 访问 http://localhost:8080/ 即可使用 Web UI
```

### 公网部署（反向连接）

**服务端（公网服务器）**:
```bash
./openclaw-channel reverse --port 18770 --token YOUR_TOKEN
```

**本地（连接内网 Gateway）**:
```bash
./openclaw-channel connector --remote ws://your-server:18770/gateway --token YOUR_TOKEN
```

然后访问 `http://your-server:18770/` 即可使用 Web UI。

## Web UI

访问根路径 `/` 即可使用内置的聊天界面：

- 连接状态指示器
- 会话选择（Main/Dev/Test Agent）
- 消息发送和接收
- 统计信息面板
- 快捷操作按钮

## API Endpoints

| 端点 | 方法 | 说明 |
|------|------|------|
| `/` | GET | Web UI (HTML) |
| `/health` | GET | 健康检查 |
| `/status` | GET | Gateway 状态 |
| `/chat` | POST | 发送消息 |
| `/stats` | GET | 统计信息 |
| `/gateway` | WS | WebSocket 端点 |

### 发送消息示例

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello!","session":"agent:main:main"}'
```

## CLI Commands

```bash
# 交互模式
./openclaw-channel interactive

# 批量模式
./openclaw-channel batch "Hello"

# API 服务器
./openclaw-channel api-enhanced --port 8080

# 反向服务器
./openclaw-channel reverse --port 18770 --token TOKEN

# 连接器
./openclaw-channel connector --remote ws://server:18770/gateway --token TOKEN

# MCP 服务器
./openclaw-channel mcp
```

## Global Flags

```
-g, --gateway string   Gateway URL (default "ws://127.0.0.1:18789")
-t, --token string     Auth token
-c, --config string    Config file path
```

## Configuration

### 环境变量

```bash
OPENCLAW_GATEWAY_URL    # Gateway URL
OPENCLAW_GATEWAY_TOKEN  # Auth token
OPENCLAW_API_PORT       # API 端口
OPENCLAW_DEBUG          # 调试模式 (1)
```

### 配置文件 (config.yaml)

```yaml
gateway:
  url: ws://127.0.0.1:18789
  token: your-token

api:
  host: 0.0.0.0
  port: 8080
```

## Project Structure

```
channel/
├── main.go              # CLI 入口
├── server.go            # 反向服务器 + Web UI
├── api_enhanced.go      # 增强 API 服务器
├── connector.go         # 连接器
├── client_enhanced.go   # WebSocket 客户端
├── protocol.go          # 协议 v3
├── config.go            # 配置管理
├── static/index.html    # Web UI (嵌入)
├── CLAUDE.md            # 开发者文档
└── README.md            # 本文件
```

## Live Demo

- Web UI: http://itsnot.fun:18770/
- Health: http://itsnot.fun:18770/health

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
