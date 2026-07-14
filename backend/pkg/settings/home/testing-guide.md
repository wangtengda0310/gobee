# ChatService 单元测试编写指南

## 概述

`home` 包的 ChatService 是 AI 聊天服务核心，基于 gobee Agent 框架实现流式对话和工具调用。本文档指导如何为该服务编写单元测试。

## 文件结构

| 文件 | 职责 |
|------|------|
| `wails.go` | ChatService 结构体、构造函数、所有导出方法（@frontend） |
| `types.go` | 类型定义（ChatConfig, ChatMessage, ChatSession 等） |
| `config.go` | 配置加载/保存/Claude Code 合并 |
| `client.go` | LLM 客户端创建和模型名解析 |
| `agent.go` | EventEmitter 接口、Agent 创建、流式事件处理 |

## 核心接口

### EventEmitter

解耦 Wails `*application.App` 的事件系统，使 ChatService 可脱离 Wails 测试。

```go
// 定义在 agent.go
type EventEmitter interface {
    Emit(name string, data ...any) bool
}
```

`*application.EventManager` 隐式实现此接口（签名完全一致），构造时直接赋值 `app.Event`。

### llm.ChatCompleter

gobee 框架提供的 LLM 客户端接口，mock 实现在 `gobee/agent/pkg/llm/mock`。

## Mock 策略

### 1. MockEventEmitter — 自建

记录所有 Emit 调用供断言，放在 `mock_test.go` 中：

```go
type EmittedEvent struct {
    Name string
    Data []any
}

type MockEventEmitter struct {
    mu     sync.RWMutex
    events []EmittedEvent
}

func (m *MockEventEmitter) Emit(name string, data ...any) bool {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.events = append(m.events, EmittedEvent{Name: name, Data: data})
    return false
}

func (m *MockEventEmitter) Events() []EmittedEvent {
    m.mu.RLock()
    defer m.mu.RUnlock()
    result := make([]EmittedEvent, len(m.events))
    copy(result, m.events)
    return result
}

func (m *MockEventEmitter) EventsByName(name string) []EmittedEvent {
    var result []EmittedEvent
    for _, e := range m.Events() {
        if e.Name == name {
            result = append(result, e)
        }
    }
    return result
}
```

### 2. mock.Client — gobee 框架提供

```go
import mockllm "github.com/wangtengda0310/gobee/agent/pkg/llm/mock"
```

**关键注意事项：**

- `AddToolCallResponse(name, argsJSON)` 和 `AddTextResponse(text)` 只对 `Complete` 方法生效
- Agent 的 `RunStream` 调用的是 `Stream` 方法，需要手动构造 `StreamChunks` 字段
- 工具调用场景需要用 `llm.NewToolUseChunk` 构造流式响应：

```go
mockClient := mockllm.NewClient()
mockClient.StreamChunks = [][]*llm.StreamChunk{
    {
        // 模拟 LLM 返回工具调用
        llm.NewToolUseChunk([]*llm.ToolCall{
            llm.NewToolCall("call_1", "get_all_excels", `{"dirPath": "/test"}`),
        }),
        llm.NewDoneChunk(&llm.ChatResponse{
            StopReason: llm.StopReasonToolUse,
            ToolCalls: []*llm.ToolCall{
                llm.NewToolCall("call_1", "get_all_excels", `{"dirPath": "/test"}`),
            },
        }),
    },
}
```

### 3. 测试辅助构造函数

```go
// newTestChatService 创建不依赖 Wails App 的测试实例
func newTestChatService(t *testing.T) (*ChatService, *MockEventEmitter) {
    t.Helper()
    tempDir := t.TempDir()
    mockEmitter := NewMockEventEmitter()

    mem := memory.NewFileMemory(100,
        filepath.Join(tempDir, "chat_history.json"),
        memory.WithPreserveSystem(true),
    )

    s := &ChatService{
        emitter:    mockEmitter,
        configFile: filepath.Join(tempDir, "app_chat_config.json"),
        memory:     &memoryAdapter{FileMemory: mem},
        registry:   tool.NewRegistry(),
    }
    return s, mockEmitter
}
```

## 按职责的测试模式

### 配置管理（config_test.go）

纯逻辑测试，不依赖 Wails 或 LLM：

```go
func TestApplyClaudeCodeConfig_ModelPriority(t *testing.T) {
    // Sonnet > Opus > Haiku 优先级
    cfg := &ChatConfig{}
    env := map[string]string{
        "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1",
    }
    applyClaudeCodeConfig(&settings.ClaudeCodeSettings{Env: env}, cfg)
    assert.Equal(t, "glm-5.1", cfg.AnthropicConfig.Model)
}
```

配置文件 IO 测试使用 `t.TempDir()` 隔离：

```go
func TestSaveAndGetConfig(t *testing.T) {
    svc, _ := newTestChatService(t)
    cfg := svc.getDefaultConfig()
    cfg.AnthropicConfig.Model = "test-model"

    err := svc.SaveConfig(cfg)
    assert.NoError(t, err)

    // 重新加载
    svc.config = nil
    loaded, err := svc.GetConfig()
    assert.NoError(t, err)
    assert.Equal(t, "test-model", loaded.AnthropicConfig.Model)
}
```

### 客户端创建（client_test.go）

验证各种配置组合：

```go
func TestCreateClient_AnthropicNoKey(t *testing.T) {
    svc, _ := newTestChatService(t)
    svc.config = &ChatConfig{Provider: "anthropic"}

    _, err := svc.createClient()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "API Key")
}

func TestCreateClient_AnthropicSuccess(t *testing.T) {
    svc, _ := newTestChatService(t)
    svc.config = &ChatConfig{
        Provider: "anthropic",
        AnthropicConfig: AnthropicConfig{
            APIKey: "test-key",
            Model:  "test-model",
        },
    }

    client, err := svc.createClient()
    assert.NoError(t, err)
    assert.NotNil(t, client)
}
```

### 事件处理（agent_test.go）

测试 `processAgentEvents` 的事件转发逻辑。通过构造 `agent.StreamEvent` 发送到 buffered channel，验证 MockEventEmitter 收到的事件：

```go
func TestProcessAgentEvents_ContentChunks(t *testing.T) {
    svc, emitter := newTestChatService(t)
    ch := make(chan *agent.StreamEvent, 3)

    // 构造事件序列
    ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "你好"}
    ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "世界"}
    ch <- &agent.StreamEvent{Type: agent.EventTypeDone}
    close(ch)

    ctx := context.Background()
    svc.processAgentEvents(ctx, ch)

    chunks := emitter.EventsByName("chatStreamChunk")
    assert.Len(t, chunks, 2)

    // 验证 seq 递增
    assert.Equal(t, int64(1), chunks[0].Data[0].(map[string]any)["seq"])
    assert.Equal(t, int64(2), chunks[1].Data[0].(map[string]any)["seq"])
}

func TestProcessAgentEvents_ContextCancelled(t *testing.T) {
    svc, emitter := newTestChatService(t)
    ch := make(chan *agent.StreamEvent, 1)

    ctx, cancel := context.WithCancel(context.Background())
    cancel() // 立即取消

    ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "test"}
    close(ch)

    svc.processAgentEvents(ctx, ch)

    errors := emitter.EventsByName("chatStreamError")
    assert.Len(t, errors, 1)
}
```

### 注意：两层事件通道

`processAgentEvents` 中 `EventTypeToolCall` 和 `EventTypeToolResult` 分支只做日志记录（log.Printf），不调用 Emit。工具调用相关的事件由 Agent hooks 层的 `OnToolCall`/`OnToolResult` 触发。测试时注意区分：

| 事件类型 | processAgentEvents 行为 | Hooks 行为 |
|---------|------------------------|-----------|
| EventTypeContent | Emit("chatStreamChunk") | 无 |
| EventTypeToolCall | 仅日志 | Emit("chatToolCall") |
| EventTypeToolResult | 仅日志 | Emit("chatToolResult") |
| EventTypeDone | return | Emit("chatStreamDone") |
| EventTypeError | return | Emit("chatStreamError") |

## 事件名称常量

测试中断言事件名时使用以下常量值：

| 事件名 | 触发场景 | Data 格式 |
|--------|---------|-----------|
| `chatStreamStart` | 开始处理输入 | `string` (输入内容) |
| `chatStreamChunk` | 内容增量 | `map[string]any{"content": string, "seq": int}` |
| `chatStreamDone` | 处理完成 | `nil` |
| `chatToolCall` | 工具调用 | `map[string]any{"name": string, "args": map}` |
| `chatToolResult` | 工具结果 | `map[string]any{"name": string, "result": any, "error": error}` |
| `chatStreamError` | 错误 | `string` (错误信息) |

## 测试风格约定

- 使用 `testify/assert` 断言库
- 测试函数命名：`Test方法名_场景描述`
- 中文注释描述测试目的
- 使用 `t.TempDir()` 隔离文件 IO
- 使用 `t.Helper()` 标记辅助函数
- 使用 `go test -race` 检测竞态条件

## 常见陷阱

1. **mock.Client 的 AddToolCallResponse 不适用于 Stream**：Agent.RunStream 调用 Stream 方法，需要手动构造 `StreamChunks` 字段
2. **processAgentEvents 会阻塞**：必须关闭 channel 或发送 Done/Error 事件才能返回
3. **事件 seq 从 1 开始**：不是 0
4. **SaveConfig 会重置 agent 和 currentClient**：测试连续操作时注意副作用
5. **ChatService 有读写锁**：并发测试使用 `-race` 标志
