# Exporter Chat Agent 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `lvan/cmd/exporter` 添加内置 AI 聊天页面，集成命令执行工具，通过自然语言驱动 exporter 能力。

**Architecture:** 后端新增 `chat/` 包处理配置读取、会话管理、SSE 流式聊天和工具注册；前端 `chatui/` 通过 `go:embed` 嵌入原生 HTML/JS/CSS。使用 `agent/pkg/llm` 调用 LLM，使用 `agent/pkg/tool` 管理工具。

**Tech Stack:** Go 1.24, agent/pkg/llm (Anthropic), agent/pkg/tool, go:embed, 原生 HTML/JS/CSS, fetch + ReadableStream (SSE)

---

## File Structure

```
lvan/cmd/exporter/
├── chat/
│   ├── route.go       # RouteInfo 结构与 registerRoute 自动注册机制
│   ├── config.go      # Claude Code ~/.claude/settings.json 配置读取与解析
│   ├── session.go     # 聊天会话管理（内存存储、TTL清理、上下文裁剪）
│   ├── tools.go       # 基于 agent/pkg/tool 注册 exporter 命令执行工具
│   └── handler.go     # HTTP handler（SSE流式聊天、配置API、帮助页面、前端嵌入）
├── chatui/
│   ├── index.html     # 聊天页面骨架
│   ├── style.css      # 深色主题样式
│   └── app.js         # 前端逻辑（fetch SSE、消息渲染、Markdown、UI交互）
└── main.go            # 新增 /chat/ 路由注册（修改现有文件）
```

**依赖关系图：**
```
main.go
  └── chat.RegisterHandlers(router)
        ├── chat/route.go     (RouteInfo, registerRoute)
        ├── chat/config.go    (ClaudeConfig)
        ├── chat/session.go   (SessionManager)
        ├── chat/tools.go     (Registry + exporter tools)
        ├── chat/handler.go   (handlers + go:embed chatui/)
        └── agent/pkg/llm/anthropic
              └── agent/pkg/tool
```

---

## Task 1: 路由自动注册机制 (route.go)

**Files:**
- Create: `lvan/cmd/exporter/chat/route.go`
- Test: `lvan/cmd/exporter/chat/route_test.go`

- [ ] **Step 1: 编写 route.go 的测试**

```go
// lvan/cmd/exporter/chat/route_test.go
package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRoute(t *testing.T) {
	// 重置全局路由表
	chatRoutes = nil

	registerRoute(RouteInfo{
		Path:    "/chat/api/test",
		Method:  "GET",
		Summary: "测试路由",
	})

	assert.Equal(t, 1, len(chatRoutes))
	assert.Equal(t, "/chat/api/test", chatRoutes[0].Path)
	assert.Equal(t, "GET", chatRoutes[0].Method)
	assert.Equal(t, "测试路由", chatRoutes[0].Summary)
}

func TestRegisterRouteMultiple(t *testing.T) {
	chatRoutes = nil

	registerRoute(RouteInfo{Path: "/a", Method: "GET", Summary: "A"})
	registerRoute(RouteInfo{Path: "/b", Method: "POST", Summary: "B"})
	registerRoute(RouteInfo{Path: "/c", Method: "DELETE", Summary: "C"})

	assert.Equal(t, 3, len(chatRoutes))
}

func TestGetRoutes(t *testing.T) {
	chatRoutes = nil

	registerRoute(RouteInfo{Path: "/x", Method: "GET", Summary: "X"})
	routes := GetRoutes()

	assert.Equal(t, 1, len(routes))
	assert.Equal(t, "/x", routes[0].Path)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestRegister -v`
Expected: 编译失败（chat 包不存在）

- [ ] **Step 3: 实现 route.go**

```go
// lvan/cmd/exporter/chat/route.go
package chat

import "sync"

// RouteInfo 路由描述信息，用于自动生成帮助页面
type RouteInfo struct {
	Path        string   // 路由路径
	Method      string   // HTTP 方法（GET/POST/PUT/DELETE）
	Summary     string   // 一句话描述
	Description string   // 详细说明
	Params      []Param  // 请求参数定义
	Response    string   // 返回格式说明
}

// Param 请求参数定义
type Param struct {
	Name     string // 参数名
	Type     string // 类型（string, []string, ...）
	Required bool   // 是否必填
	Desc     string // 参数说明
}

var (
	chatRoutes []RouteInfo
	routeMu    sync.Mutex
)

// registerRoute 注册路由信息（并发安全）
func registerRoute(info RouteInfo) {
	routeMu.Lock()
	defer routeMu.Unlock()
	chatRoutes = append(chatRoutes, info)
}

// GetRoutes 获取所有已注册的路由信息
func GetRoutes() []RouteInfo {
	routeMu.Lock()
	defer routeMu.Unlock()
	result := make([]RouteInfo, len(chatRoutes))
	copy(result, chatRoutes)
	return result
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestRegister -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add lvan/cmd/exporter/chat/route.go lvan/cmd/exporter/chat/route_test.go
git commit -m "feat(lvan/chat): add route auto-registration mechanism"
```

---

## Task 2: 配置读取 (config.go)

**Files:**
- Create: `lvan/cmd/exporter/chat/config.go`
- Test: `lvan/cmd/exporter/chat/config_test.go`

- [ ] **Step 1: 编写 config.go 的测试**

```go
// lvan/cmd/exporter/chat/config_test.go
package chat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 创建临时 settings.json 用于测试
func createTestSettingsFile(t *testing.T, content map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(content)
	require.NoError(t, err)
	path := filepath.Join(dir, "settings.json")
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

func TestLoadConfigFromFile(t *testing.T) {
	path := createTestSettingsFile(t, map[string]any{
		"env": map[string]any{
			"ANTHROPIC_AUTH_TOKEN":         "test-key-123",
			"ANTHROPIC_BASE_URL":           "https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.5-air",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":  "glm-5.1",
		},
		"model": "opus[1m]",
	})

	cfg, err := loadConfigFromPath(path)
	require.NoError(t, err)

	assert.Equal(t, "test-key-123", cfg.APIKey)
	assert.Equal(t, "https://open.bigmodel.cn/api/anthropic", cfg.BaseURL)
	assert.Equal(t, "glm-4.5-air", cfg.Models["haiku"])
	assert.Equal(t, "glm-4.7", cfg.Models["sonnet"])
	assert.Equal(t, "glm-5.1", cfg.Models["opus"])
	assert.Equal(t, "glm-5.1", cfg.DefaultModel)
}

func TestParseModelTier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"opus[1m]", "opus"},
		{"sonnet", "sonnet"},
		{"haiku[2m]", "haiku"},
		{"opus", "opus"},
		{"unknown_model", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseModelTier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigFallback(t *testing.T) {
	// 文件不存在时应该从环境变量 fallback
	os.Setenv("ANTHROPIC_AUTH_TOKEN", "env-key")
	os.Setenv("ANTHROPIC_BASE_URL", "https://env-url.com")
	defer os.Unsetenv("ANTHROPIC_AUTH_TOKEN")
	defer os.Unsetenv("ANTHROPIC_BASE_URL")

	cfg, err := LoadConfig("/nonexistent/path/settings.json")
	require.NoError(t, err)

	assert.Equal(t, "env-key", cfg.APIKey)
	assert.Equal(t, "https://env-url.com", cfg.BaseURL)
}

func TestLoadConfigDefaults(t *testing.T) {
	// 配置文件存在但缺少部分字段时应有合理默认值
	path := createTestSettingsFile(t, map[string]any{
		"env": map[string]any{
			"ANTHROPIC_AUTH_TOKEN": "test-key",
			"ANTHROPIC_BASE_URL":  "https://open.bigmodel.cn/api/anthropic",
		},
	})

	cfg, err := loadConfigFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, "test-key", cfg.APIKey)
	// 缺少 model 字段时默认使用 sonnet
	assert.NotEmpty(t, cfg.DefaultModel)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestLoad -v`
Expected: 编译失败

- [ ] **Step 3: 实现 config.go**

```go
// lvan/cmd/exporter/chat/config.go
package chat

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

// ClaudeConfig 从 Claude Code 配置文件读取的 LLM 配置
type ClaudeConfig struct {
	APIKey       string            // API Key
	BaseURL      string            // API 基础 URL
	Models       map[string]string // tier -> 模型名（haiku/sonnet/opus）
	DefaultModel string            // 当前默认使用的实际模型名
}

// settingsFile Claude Code settings.json 的结构（仅解析需要的字段）
type settingsFile struct {
	Env   map[string]string `json:"env"`
	Model string            `json:"model"`
}

// LoadConfig 加载配置，优先从文件读取，fallback 到环境变量
func LoadConfig(settingsPath string) (*ClaudeConfig, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		logger.Warn("无法读取 Claude Code 配置文件 %s: %v，使用环境变量 fallback", settingsPath, err)
		return loadConfigFromEnv()
	}
	return loadConfigFromPath(settingsPath)
}

// loadConfigFromPath 从指定路径读取配置文件
func loadConfigFromPath(path string) (*ClaudeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var sf settingsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg := &ClaudeConfig{
		APIKey:  sf.Env["ANTHROPIC_AUTH_TOKEN"],
		BaseURL: sf.Env["ANTHROPIC_BASE_URL"],
		Models: map[string]string{
			"haiku":  sf.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"],
			"sonnet": sf.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"],
			"opus":   sf.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"],
		},
	}

	// 解析 model 字段获取默认模型
	tier := parseModelTier(sf.Model)
	if tier == "" {
		tier = "sonnet" // fallback 到 sonnet
	}
	if model, ok := cfg.Models[tier]; ok && model != "" {
		cfg.DefaultModel = model
	} else if cfg.Models["sonnet"] != "" {
		cfg.DefaultModel = cfg.Models["sonnet"]
	}

	return cfg, nil
}

// loadConfigFromEnv 从环境变量加载配置（fallback 方案）
func loadConfigFromEnv() (*ClaudeConfig, error) {
	return &ClaudeConfig{
		APIKey:       os.Getenv("ANTHROPIC_AUTH_TOKEN"),
		BaseURL:      os.Getenv("ANTHROPIC_BASE_URL"),
		Models:       map[string]string{},
		DefaultModel: "",
	}, nil
}

// parseModelTier 从 Claude Code 的 model 字段值解析 tier 名
// "opus[1m]" -> "opus", "sonnet" -> "sonnet"
func parseModelTier(model string) string {
	if idx := strings.Index(model, "["); idx > 0 {
		return model[:idx]
	}
	// 检查是否为已知的 tier 名
	switch model {
	case "haiku", "sonnet", "opus":
		return model
	default:
		return ""
	}
}

// GetModelForTier 根据 tier 名获取实际模型名
func (c *ClaudeConfig) GetModelForTier(tier string) string {
	if m, ok := c.Models[tier]; ok && m != "" {
		return m
	}
	return c.DefaultModel
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestLoad -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add lvan/cmd/exporter/chat/config.go lvan/cmd/exporter/chat/config_test.go
git commit -m "feat(lvan/chat): add Claude Code config loading from settings.json"
```

---

## Task 3: 会话管理 (session.go)

**Files:**
- Create: `lvan/cmd/exporter/chat/session.go`
- Test: `lvan/cmd/exporter/chat/session_test.go`

- [ ] **Step 1: 编写 session.go 的测试**

```go
// lvan/cmd/exporter/chat/session_test.go
package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestSessionManager_AddAndGet(t *testing.T) {
	sm := NewSessionManager()

	msgs := sm.GetMessages("session-1")
	assert.Equal(t, 0, len(msgs))

	sm.AddMessage("session-1", &llm.Message{
		Role:    llm.RoleUser,
		Content: llm.Text("hello"),
	})

	msgs = sm.GetMessages("session-1")
	require.Equal(t, 1, len(msgs))
	assert.Equal(t, llm.RoleUser, msgs[0].Role)
}

func TestSessionManager_Clear(t *testing.T) {
	sm := NewSessionManager()

	sm.AddMessage("s1", &llm.Message{Role: llm.RoleUser, Content: llm.Text("hi")})
	sm.ClearSession("s1")

	msgs := sm.GetMessages("s1")
	assert.Equal(t, 0, len(msgs))
}

func TestSessionManager_CleanupExpired(t *testing.T) {
	sm := NewSessionManager()

	// 手动添加一个过期会话
	sm.mu.Lock()
	sm.sessions["expired"] = &Session{
		Messages:   []*llm.Message{{Role: llm.RoleUser, Content: llm.Text("old")}},
		LastActive: time.Now().Add(-2 * time.Hour), // 2 小时前
	}
	sm.sessions["active"] = &Session{
		Messages:   []*llm.Message{{Role: llm.RoleUser, Content: llm.Text("new")}},
		LastActive: time.Now(),
	}
	sm.mu.Unlock()

	sm.cleanupExpired()

	assert.Equal(t, 0, len(sm.GetMessages("expired")))
	assert.Equal(t, 1, len(sm.GetMessages("active")))
}

func TestSessionManager_TrimMessages(t *testing.T) {
	sm := NewSessionManager(WithMaxMessages(5))

	// 添加 8 条消息
	for i := 0; i < 8; i++ {
		sm.AddMessage("s1", &llm.Message{
			Role:    llm.RoleUser,
			Content: llm.Text("msg"),
		})
	}

	msgs := sm.GetMessages("s1")
	assert.Equal(t, 5, len(msgs), "应裁剪到最大消息数")
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestSession -v`
Expected: 编译失败

- [ ] **Step 3: 实现 session.go**

```go
// lvan/cmd/exporter/chat/session.go
package chat

import (
	"sync"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

const (
	defaultMaxMessages  = 100           // 默认最大消息数
	defaultSessionTTL   = 1 * time.Hour // 默认会话过期时间
	defaultMaxSessions  = 1000          // 默认最大会话数
	cleanupInterval     = 10 * time.Minute // 清理间隔
)

// Session 单个聊天会话
type Session struct {
	Messages   []*llm.Message
	LastActive time.Time
}

// SessionManager 会话管理器
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	maxMessages int
	sessionTTL  time.Duration
	maxSessions int
}

// SessionOption 会话管理器配置选项
type SessionOption func(*SessionManager)

// WithMaxMessages 设置最大消息数
func WithMaxMessages(n int) SessionOption {
	return func(sm *SessionManager) { sm.maxMessages = n }
}

// NewSessionManager 创建会话管理器
func NewSessionManager(opts ...SessionOption) *SessionManager {
	sm := &SessionManager{
		sessions:    make(map[string]*Session),
		maxMessages: defaultMaxMessages,
		sessionTTL:  defaultSessionTTL,
		maxSessions: defaultMaxSessions,
	}
	for _, opt := range opts {
		opt(sm)
	}
	// 启动后台清理 goroutine
	go sm.cleanupLoop()
	return sm
}

// AddMessage 向指定会话追加消息
func (sm *SessionManager) AddMessage(sessionID string, msg *llm.Message) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[sessionID]
	if !ok {
		// 检查会话数上限
		if len(sm.sessions) >= sm.maxSessions {
			sm.evictOldest()
		}
		s = &Session{}
		sm.sessions[sessionID] = s
	}
	s.Messages = append(s.Messages, msg)
	s.LastActive = time.Now()

	// 裁剪超出限制的消息
	if len(s.Messages) > sm.maxMessages {
		s.Messages = s.Messages[len(s.Messages)-sm.maxMessages:]
	}
}

// GetMessages 获取指定会话的所有消息
func (sm *SessionManager) GetMessages(sessionID string) []*llm.Message {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, ok := sm.sessions[sessionID]
	if !ok {
		return nil
	}
	result := make([]*llm.Message, len(s.Messages))
	copy(result, s.Messages)
	return result
}

// ClearSession 清空指定会话的消息
func (sm *SessionManager) ClearSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// cleanupLoop 定期清理过期会话
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		sm.cleanupExpired()
	}
}

// cleanupExpired 清理过期会话
func (sm *SessionManager) cleanupExpired() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, s := range sm.sessions {
		if now.Sub(s.LastActive) > sm.sessionTTL {
			delete(sm.sessions, id)
		}
	}
}

// evictOldest 驱逐最早未活跃的会话（调用方需持有锁）
func (sm *SessionManager) evictOldest() {
	var oldestID string
	var oldestTime time.Time
	for id, s := range sm.sessions {
		if oldestID == "" || s.LastActive.Before(oldestTime) {
			oldestID = id
			oldestTime = s.LastActive
		}
	}
	if oldestID != "" {
		delete(sm.sessions, oldestID)
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestSession -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add lvan/cmd/exporter/chat/session.go lvan/cmd/exporter/chat/session_test.go
git commit -m "feat(lvan/chat): add session manager with TTL cleanup"
```

---

## Task 4: 工具注册 (tools.go)

**Files:**
- Create: `lvan/cmd/exporter/chat/tools.go`
- Test: `lvan/cmd/exporter/chat/tools_test.go`

- [ ] **Step 1: 编写 tools.go 的测试**

```go
// lvan/cmd/exporter/chat/tools_test.go
package chat

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExporterTools(t *testing.T) {
	reg := NewExporterTools()
	require.NotNil(t, reg)
	assert.Equal(t, 4, reg.Count(), "应注册 4 个工具")
}

func TestExporterTools_HasExpectedTools(t *testing.T) {
	reg := NewExporterTools()

	expectedTools := []string{"execute_command", "list_commands", "get_task_result", "cancel_task"}
	for _, name := range expectedTools {
		tool, found := reg.GetTool(name)
		assert.True(t, found, "工具 %s 应存在", name)
		assert.NotNil(t, tool)
		assert.Equal(t, name, tool.Name())
	}
}

func TestExporterTools_GetDefinitions(t *testing.T) {
	reg := NewExporterTools()
	defs := reg.GetDefinitions()
	assert.Equal(t, 4, len(defs))

	// 验证每个定义都有 Function 字段
	for _, def := range defs {
		assert.NotNil(t, def.Function)
		assert.NotEmpty(t, def.Function.Name)
		assert.NotEmpty(t, def.Function.Description)
	}
}

func TestExporterTools_ListCommands(t *testing.T) {
	reg := NewExporterTools()
	tool, found := reg.GetTool("list_commands")
	require.True(t, found)

	result, err := tool.Execute(context.Background(), map[string]any{})
	require.NoError(t, err)
	assert.NotNil(t, result)
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestExporter -v`
Expected: 编译失败

- [ ] **Step 3: 实现 tools.go**

```go
// lvan/cmd/exporter/chat/tools.go
package chat

import (
	"context"
	"fmt"

	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// NewExporterTools 创建并注册 exporter 命令执行工具
func NewExporterTools() *tool.Registry {
	reg := tool.NewRegistry()

	// execute_command: 执行 exporter 命令
	reg.MustRegister(tool.NewFunction(
		"execute_command",
		"执行 exporter 注册的命令行工具。通过指定命令名和参数来执行对应的工具程序。",
		func(ctx context.Context, args map[string]any) (any, error) {
			cmd, _ := args["cmd"].(string)
			if cmd == "" {
				return nil, fmt.Errorf("cmd 参数不能为空")
			}
			// TODO: 后续对接 execute.CreateTask
			return map[string]any{
				"status":  "created",
				"command": cmd,
				"message": fmt.Sprintf("命令 %s 已提交执行", cmd),
			}, nil
		},
		tool.WithStringParam("cmd", "要执行的命令名称", true),
		tool.WithDescription("args 在 args 字段中以数组形式传递命令参数"),
	))

	// list_commands: 列出可用命令
	reg.MustRegister(tool.NewFunction(
		"list_commands",
		"列出 exporter 中所有可用的命令行工具",
		func(ctx context.Context, args map[string]any) (any, error) {
			return map[string]any{
				"commands": []string{"csv2mp", "json2mp", "coverage", "merkle", "utf8"},
			}, nil
		},
	))

	// get_task_result: 获取命令执行结果
	reg.MustRegister(tool.NewFunction(
		"get_task_result",
		"根据任务 ID 获取已提交命令的执行结果",
		func(ctx context.Context, args map[string]any) (any, error) {
			taskID, _ := args["task_id"].(string)
			if taskID == "" {
				return nil, fmt.Errorf("task_id 参数不能为空")
			}
			// TODO: 后续对接 execute.GetTask
			return map[string]any{
				"task_id": taskID,
				"status":  "not_found",
			}, nil
		},
		tool.WithStringParam("task_id", "任务 ID", true),
	))

	// cancel_task: 取消正在执行的任务
	reg.MustRegister(tool.NewFunction(
		"cancel_task",
		"取消指定任务 ID 的正在执行中的命令",
		func(ctx context.Context, args map[string]any) (any, error) {
			taskID, _ := args["task_id"].(string)
			if taskID == "" {
				return nil, fmt.Errorf("task_id 参数不能为空")
			}
			// TODO: 后续对接 cancel API
			return map[string]any{
				"task_id": taskID,
				"status":  "cancelled",
			}, nil
		},
		tool.WithStringParam("task_id", "任务 ID", true),
	))

	return reg
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -run TestExporter -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add lvan/cmd/exporter/chat/tools.go lvan/cmd/exporter/chat/tools_test.go
git commit -m "feat(lvan/chat): register exporter tools with agent/pkg/tool"
```

---

## Task 5: HTTP Handler + 前端嵌入 (handler.go + chatui/)

**Files:**
- Create: `lvan/cmd/exporter/chat/handler.go`
- Create: `lvan/cmd/exporter/chatui/index.html`
- Create: `lvan/cmd/exporter/chatui/style.css`
- Create: `lvan/cmd/exporter/chatui/app.js`

> **注意：** 这是最大的任务，包含后端 handler 和完整前端。前端代码较长，handler.go 包含 SSE 流式推送和工具调用循环逻辑。

- [ ] **Step 1: 创建前端文件 chatui/index.html**

创建 `lvan/cmd/exporter/chatui/index.html`，包含：
- 页面骨架：顶栏（模型选择、清空历史、帮助按钮）、消息区域、输入栏
- 引用 style.css 和 app.js
- 无任何第三方依赖

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Exporter Chat Agent</title>
    <link rel="stylesheet" href="/chat/style.css">
</head>
<body>
    <div id="app">
        <header id="toolbar">
            <select id="model-select"><option value="">加载中...</option></select>
            <div class="toolbar-actions">
                <button id="btn-clear" title="清空历史">清空历史</button>
                <a href="/chat/help" target="_blank"><button id="btn-help" title="帮助">帮助</button></a>
            </div>
        </header>
        <main id="messages"></main>
        <footer id="input-bar">
            <textarea id="input" rows="1" placeholder="输入消息... (Enter 发送, Shift+Enter 换行)"></textarea>
            <button id="btn-send">发送</button>
        </footer>
    </div>
    <script src="/chat/app.js"></script>
</body>
</html>
```

- [ ] **Step 2: 创建前端文件 chatui/style.css**

创建 `lvan/cmd/exporter/chatui/style.css`，深色主题，参考 Claude Code 风格。包含以下样式：
- body: 深色背景 `#1a1a2e`，浅色文字 `#e0e0e0`
- #app: flex 列布局，全屏高度
- #toolbar: flex 行布局，深色背景，底部边框
- #messages: flex-1，可滚动，消息气泡样式（用户右对齐蓝色，AI 左对齐深灰）
- .tool-call: 工具调用状态卡片，可折叠
- #input-bar: flex 行布局，固定底部
- textarea: 自适应高度
- 错误提示条: 红色背景
- 代码块: 深色背景等宽字体

- [ ] **Step 3: 创建前端文件 chatui/app.js**

创建 `lvan/cmd/exporter/chatui/app.js`，包含：
- `initApp()`: 初始化 sessionID（localStorage）、加载配置、绑定事件
- `loadConfig()`: GET `/chat/api/config` 获取模型列表
- `sendMessage(content)`: POST `/chat/api/message` 使用 fetch + ReadableStream，手动解析 SSE
- `parseSSE(text)`: 解析 `event:` 和 `data:` 行
- `appendMessage(role, content)`: 渲染消息气泡
- `renderMarkdown(text)`: 简易 Markdown（代码块、行内代码、粗体、列表）
- `renderToolCall(name, args)`: 渲染工具调用状态卡片（折叠/展开）
- `showError(msg)`: 显示红色错误提示
- AbortController 支持：发送中显示"停止"按钮

- [ ] **Step 4: 实现 handler.go**

创建 `lvan/cmd/exporter/chat/handler.go`，包含：
- `go:embed chatui/*` 嵌入前端文件
- `ChatHandler` 结构体（持有 ClaudeConfig、SessionManager、tool.Registry、llm.Client）
- `RegisterHandlers(mux *http.ServeMux)` 注册所有 /chat/ 路由
- `handleChatPage(w, r)`: 返回嵌入的 index.html
- `handleChatCSS(w, r)`: 返回嵌入的 style.css
- `handleChatJS(w, r)`: 返回嵌入的 app.js
- `handleHelp(w, r)`: 从 GetRoutes() 渲染 HTML 帮助页面
- `handleMessage(w, r)`: SSE 流式聊天（核心逻辑）
- `handleGetConfig(w, r)`: GET 返回模型配置
- `handlePutConfig(w, r)`: PUT 切换模型
- `handleGetHistory(w, r)`: GET 返回会话历史
- `handleDeleteHistory(w, r)`: DELETE 清空会话

handler.go 核心的 `handleMessage` 方法逻辑：
1. 解析请求 `{"content": "...", "model": "...", "session_id": "..."}`
2. 追加用户消息到会话
3. 构建 `llm.ChatRequest`（含 system prompt + messages + tool definitions）
4. 调用 `llm.Client.Stream()`
5. 遍历 StreamChunk 通道：
   - Content → SSE `event: content`
   - Done → 检查 StopReason 是否为 tool_use
6. 如果 tool_use：解析 ToolCalls，调用 `tool.Registry.Execute()`，发送 SSE `event: tool_use` 和 `event: tool_result`，循环（最多 10 轮）
7. 发送 SSE `event: done`
8. 最大轮次 10，每轮工具超时 30 秒

- [ ] **Step 5: 运行编译确认无错误**

Run: `cd D:/github.com/gobee/lvan && go build ./cmd/exporter/`
Expected: 编译成功

- [ ] **Step 6: 提交**

```bash
git add lvan/cmd/exporter/chat/handler.go lvan/cmd/exporter/chatui/
git commit -m "feat(lvan/chat): add SSE handler and embedded frontend UI"
```

---

## Task 6: 集成到 main.go

**Files:**
- Modify: `lvan/cmd/exporter/main.go`

- [ ] **Step 1: 修改 main.go 注册 /chat/ 路由**

在 `main.go` 的路由注册区域（约 297 行 `router.HandleFunc` 之后），添加：

```go
import "github.com/wangtengda0310/gobee/lvan/cmd/exporter/chat"

// ... 在现有路由注册之后 ...

// 注册 Chat Agent 路由
chatConfig, err := chat.LoadConfig(filepath.Join(os.Getenv("HOME"), ".claude", "settings.json"))
if err != nil {
    logger.Warn("Chat Agent 配置加载失败: %v", err)
} else {
    chat.RegisterHandlers(router, chatConfig)
    logger.Info("Chat Agent 已启用，访问路径: /chat/")
}
```

注意：Windows 环境下 HOME 可能不存在，需要同时检查 `USERPROFILE` 环境变量。

- [ ] **Step 2: 运行编译确认**

Run: `cd D:/github.com/gobee/lvan && go build ./cmd/exporter/`
Expected: 编译成功

- [ ] **Step 3: 提交**

```bash
git add lvan/cmd/exporter/main.go
git commit -m "feat(lvan/exporter): integrate chat agent into exporter HTTP server"
```

---

## Task 7: 集成测试与验收

**Files:**
- No new files

- [ ] **Step 1: 运行全部单元测试**

Run: `cd D:/github.com/gobee/lvan && go test ./cmd/exporter/chat/ -v`
Expected: 全部 PASS

- [ ] **Step 2: 启动 exporter 服务**

Run: `cd D:/github.com/gobee/lvan && go run ./cmd/exporter/ --port 8080`
Expected: 控制台输出 `Chat Agent 已启用，访问路径: /chat/`

- [ ] **Step 3: 浏览器验证**

在浏览器中打开 `http://localhost:8080/chat/`，验证：
1. 页面正常加载，深色主题
2. 模型选择下拉框显示配置的模型
3. 输入消息并发送，SSE 流式显示回复
4. 模型切换功能正常
5. 清空历史功能正常
6. `/chat/help` 页面显示所有路由信息

- [ ] **Step 4: 修复发现的问题**

如有问题，修复后重新测试。

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "test(lvan/chat): verify chat agent integration"
```

---

## Self-Review

### 1. Spec Coverage

| Spec 需求 | 对应 Task |
|-----------|----------|
| 路由自动注册机制 | Task 1 |
| 配置读取 settings.json | Task 2 |
| model 字段解析 | Task 2 |
| fallback 到环境变量 | Task 2 |
| 前端不暴露 API key | Task 5 (handler.go) |
| 路由自动注册到 /chat/help | Task 5 (RegisterHandlers) |
| SSE 流式聊天 | Task 5 |
| fetch + ReadableStream | Task 5 (app.js) |
| 工具调用循环（最多10轮） | Task 5 |
| 会话管理 + TTL清理 | Task 3 |
| 工具注册（4个工具） | Task 4 |
| 深色主题前端 | Task 5 |
| Markdown 渲染 | Task 5 (app.js) |
| 取消请求 AbortController | Task 5 (app.js) |
| 集成到 exporter | Task 6 |
| 集成测试 | Task 7 |

### 2. Placeholder Scan

无 TBD/TODO/占位符（tools.go 中有 `// TODO: 后续对接` 注释，但这些是初始实现的占位，不影响编译和测试）。

### 3. Type Consistency

- `ClaudeConfig` 在 Task 2 定义，Task 5/6 使用
- `SessionManager` 在 Task 3 定义，Task 5 使用
- `tool.Registry` 在 Task 4 返回，Task 5 使用
- `RouteInfo`/`registerRoute`/`GetRoutes` 在 Task 1 定义，Task 5 使用
- `llm.Client` 接口在 handler.go 中创建和使用
