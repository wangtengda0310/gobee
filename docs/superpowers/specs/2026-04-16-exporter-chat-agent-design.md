# Exporter Chat Agent 设计文档

## 概述

为 `lvan/cmd/exporter` 补充一个内置的 AI 聊天页面，集成 exporter 现有的命令执行能力，使用户可以通过自然语言与 AI 对话来执行工具命令。

## 需求总结

- **核心功能**：工具执行助手，AI 可调用 exporter 命令执行工具
- **AI 能力**：使用 `agent/pkg/llm` 包，支持 Anthropic 兼容 API（流式响应 + 工具调用）
- **工具管理**：使用 `agent/pkg/tool` 包的 Registry 和 FunctionTool 管理工具
- **配置来源**：自动从 `~/.claude/settings.json` 读取 API key、base URL、模型名
- **前端技术**：原生 HTML/JS/CSS，通过 `go:embed` 嵌入到 Go 二进制
- **部署方式**：作为 exporter 内置功能，通过 `/chat/` 路径访问

## 架构

### 代码结构

```
lvan/cmd/exporter/
├── chat/
│   ├── config.go       # Claude Code 配置读取与解析
│   ├── handler.go      # HTTP handler + SSE 流式聊天 + 路由自动注册
│   ├── tools.go        # 基于 agent/pkg/tool 注册 exporter 工具
│   ├── chat.go         # 聊天会话管理（消息历史、上下文裁剪）
│   └── route.go        # RouteInfo 结构与自动注册机制
├── chatui/
│   ├── index.html      # 聊天页面骨架
│   ├── style.css       # 深色主题样式
│   └── app.js          # 前端 SSE、消息渲染、UI 交互
└── main.go             # 新增 /chat/ 路由注册
```

### 依赖关系

```
exporter main.go
  └── chat/ 包
        ├── 读取 ~/.claude/settings.json
        ├── agent/pkg/llm/anthropic (LLM 客户端)
        ├── agent/pkg/tool (工具注册与执行)
        └── chatui/ (go:embed 前端文件)
```

## 配置读取

### 配置来源

从 `~/.claude/settings.json` 读取，路径硬编码，不提供自定义环境变量。

### 解析字段

| settings.json 字段 | 用途 |
|--------------------|------|
| `env.ANTHROPIC_AUTH_TOKEN` | API Key |
| `env.ANTHROPIC_BASE_URL` | API 基础 URL |
| `env.ANTHROPIC_DEFAULT_SONNET_MODEL` | Sonnet 模型名 |
| `env.ANTHROPIC_DEFAULT_OPUS_MODEL` | Opus 模型名 |
| `env.ANTHROPIC_DEFAULT_HAIKU_MODEL` | Haiku 模型名 |
| `model` | 当前使用的模型 |

### 配置结构

```go
type ClaudeConfig struct {
    APIKey       string
    BaseURL      string
    Models       map[string]string // {"haiku": "glm-4.5-air", "sonnet": "glm-4.7", "opus": "glm-5.1"}
    DefaultModel string            // 从 model 字段解析出的 tier 名（如 "opus"）
}
```

**model 字段解析**：`settings.json` 中 `model` 字段值为 Claude Code 自有格式（如 `"opus[1m]"`、`"sonnet"`），不能直接作为模型名传给 API。解析规则：
- 提取 `[` 前的部分作为 tier 名（如 `"opus[1m]"` → `"opus"`）
- 通过 tier 名从 `Models` map 中查找实际模型名（如 `"opus"` → `"glm-5.1"`）
- 如果解析失败，默认使用 sonnet tier

### fallback 机制

如果 `settings.json` 不存在或字段缺失，通过 exporter 已有的环境变量机制 fallback：
- `ANTHROPIC_AUTH_TOKEN` 环境变量
- `ANTHROPIC_BASE_URL` 环境变量

### 前端配置接口

前端 `/chat/api/config` 只返回模型信息，不暴露 API key：
```json
{ "models": ["glm-4.5-air", "glm-4.7", "glm-5.1"], "current": "glm-4.7" }
```

## API 设计

### 路由自动注册机制

```go
type RouteInfo struct {
    Path        string
    Method      string   // GET/POST/PUT/DELETE
    Summary     string   // 一句话描述
    Description string   // 详细说明
    Params      []Param
    Response    string
}

type Param struct {
    Name     string
    Type     string
    Required bool
    Desc     string
}

var (
    chatRoutes []RouteInfo
    routeMu    sync.Mutex
)

func registerRoute(info RouteInfo) {
    routeMu.Lock()
    defer routeMu.Unlock()
    chatRoutes = append(chatRoutes, info)
}
```

每个路由注册时调用 `registerRoute()`，`/chat/help` handler 从 `chatRoutes` 动态渲染帮助页面。

### 端点列表

| 端点 | 方法 | 功能 |
|------|------|------|
| `/chat/` | GET | 聊天页面（go:embed 嵌入的 HTML） |
| `/chat/help` | GET | 所有 /chat 路由的汇总帮助页面（自动生成） |
| `/chat/api/message` | POST | 发送消息，返回 SSE 流式响应 |
| `/chat/api/config` | GET | 获取当前模型配置（不含 API key） |
| `/chat/api/config` | PUT | 切换模型 |
| `/chat/api/history` | GET | 获取当前会话消息历史 |
| `/chat/api/history` | DELETE | 清空会话历史 |

### 聊天消息流

```
前端 POST /chat/api/message
  body: { "content": "把这个CSV转成mp格式", "model": "glm-4.7" }
       ↓
handler.go 接收，追加用户消息到会话历史
       ↓
创建 ChatRequest（包含 messages + tool definitions from Registry）
       ↓
调用 llm.Client.Stream()
       ↓
SSE 逐块推送到前端：
  event: content      → AI 文本回复
  event: tool_use     → AI 请求调用工具（前端显示调用状态）
  event: tool_result  → 工具执行结果
  event: done         → 流结束
       ↓
如果 LLM 返回 tool_use（从 DoneChunk.Response 中获取完整结果）：
  → 从 Response.ToolCalls 取出完整的工具调用（anthropic client 已内部处理增量累积）
  → json.Unmarshal 解析 ToolCall.Function.Arguments 字符串为 map[string]any
  → 从 Registry 取出对应 Tool
  → 调用 tool.Execute(ctx, args)
  → 将结果作为 RoleTool 消息追加到历史
  → 再次调用 llm.Client.Stream()
  → 循环直到 LLM 不再请求工具 或 达到最大轮次
```

**工具调用循环保护**：
- 最大轮次限制：10 次（超过后强制中断，返回已执行的摘要）
- 每轮工具执行超时：30 秒
- 单次 LLM 请求超时：由 API_TIMEOUT_MS 配置控制

### SSE 事件格式

SSE 事件名使用自定义名称（与 `agent/pkg/llm` 的 ChunkType 独立，后端负责映射）：

```
event: content
data: {"text": "我来帮你转换这个文件..."}

event: tool_use
data: {"name": "execute_command", "args": {"cmd": "csv2mp", "args": ["input.csv"]}}

event: tool_result
data: {"name": "execute_command", "result": "转换完成", "success": true}

event: error
data: {"message": "API 调用超时"}

event: done
data: {"usage": {"input_tokens": 150, "output_tokens": 80}}
```

后端映射：`llm.ChunkTypeContent` → SSE `content`，`llm.ChunkTypeToolUse` → SSE `tool_use`，`llm.ChunkTypeError` → SSE `error`，`llm.ChunkTypeDone` → SSE `done`。

## 工具注册

基于 `agent/pkg/tool.Registry`，初始注册以下工具：

| 工具名 | 功能 | 参数 |
|--------|------|------|
| `execute_command` | 执行 exporter 命令 | `cmd: string, args: []string, version: string` |
| `list_commands` | 列出可用的 exporter 命令 | 无 |
| `get_task_result` | 获取命令执行结果 | `task_id: string` |
| `cancel_task` | 取消正在执行的任务 | `task_id: string` |

后续可扩展注册更多 exporter 子命令作为独立工具（csv2mp、coverage 等）。

## 会话管理

- 使用内存 map 存储会话（`map[sessionID][]*llm.Message`）
- session ID 通过前端 localStorage 生成和管理
- 最大上下文长度限制（超过时裁剪最早的消息，保留 system prompt）
- system prompt 内置，引导 AI 作为 exporter 工具执行助手
- 并发安全（`sync.RWMutex` 保护会话 map）

### 会话过期与清理

- 每个会话记录最后活跃时间（`map[sessionID]*Session`，Session 包含 Messages 和 LastActive）
- 后台 goroutine 每 10 分钟扫描一次，清理超过 1 小时未活跃的会话
- 最大会话数限制：1000（超过时清理最早未活跃的会话）

## 前端设计

### 页面布局

```
┌──────────────────────────────────────────────────┐
│  [模型选择下拉框]        [清空历史] [帮助]        │  ← 顶栏
├──────────────────────────────────────────────────┤
│                                                  │
│  消息区域（滚动）                                 │
│                                                  │
│  ┌─────────────────────────────────────────┐     │
│  │ AI: 我来帮你执行命令...                   │     │
│  │    🔧 调用 execute_command              │     │
│  │    📋 工具结果: csv2mp 转换完成          │     │
│  │    ✅ CSV 文件已转换为 mp 格式           │     │
│  └─────────────────────────────────────────┘     │
│                                                  │
│  ┌─────────────────────────────────────────┐     │
│  │ 你: 把这个CSV转成mp                      │     │
│  └─────────────────────────────────────────┘     │
│                                                  │
├──────────────────────────────────────────────────┤
│  [输入框                               ] [发送]  │  ← 底栏
└──────────────────────────────────────────────────┘
```

### 交互特性

1. **消息输入**：文本框输入，回车或点击发送
2. **流式显示**：AI 回复通过 SSE 逐字显示
3. **工具调用状态**：显示工具调用过程（调用中 → 执行中 → 完成），可折叠/展开查看详细参数和结果
4. **模型切换**：顶栏下拉框切换 haiku/sonnet/opus
5. **Markdown 渲染**：AI 回复中的代码块、列表等渲染为格式化内容（手写轻量实现）
6. **错误提示**：API 错误、超时等以红色提示条显示
7. **响应式布局**：适配桌面和移动端
8. **取消请求**：发送按钮在请求进行中变为"停止"按钮，点击后通过 AbortController 关闭 fetch 连接
9. **SSE 断线检测**：前端通过 fetch ReadableStream 的 `error` 事件或读取中断检测断线，显示"连接已断开"提示，不自动重连（避免重复发送消息）

### 技术约束

- 不引入任何第三方依赖（无 npm、无 CDN）
- Markdown 渲染用手写轻量实现（支持代码块、行内代码、粗体、列表）
- 深色主题，参考 Claude Code 风格
- 所有前端文件通过 `go:embed` 嵌入到 Go 二进制中

### SSE 连接方式

前端使用 `fetch` + `ReadableStream` 而非 `EventSource`，原因：
- `EventSource` 只支持 GET 请求，无法用于 `POST /chat/api/message`（需要发送请求体）
- `fetch` 支持 `AbortController` 取消请求
- 需手动解析 SSE 格式（`event:` / `data:` 行）

```js
// 前端 SSE 连接示例
const controller = new AbortController();
const response = await fetch('/chat/api/message', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content, model }),
    signal: controller.signal,
});
const reader = response.body.getReader();
const decoder = new TextDecoder();
// 逐块读取并解析 SSE 事件...
```

## System Prompt

内置 system prompt 引导 AI 作为 exporter 工具执行助手：

```
你是 Exporter 工具执行助手。你可以帮助用户通过自然语言执行各种命令行工具。

你可以使用以下工具来完成任务：
- execute_command: 执行 exporter 注册的命令
- list_commands: 列出所有可用命令
- get_task_result: 获取命令执行结果
- cancel_task: 取消正在执行的任务

请用中文回复用户。当用户请求执行操作时，选择合适的工具完成。
```

## 错误处理

| 场景 | 处理方式 |
|------|----------|
| settings.json 不存在 | 日志警告，使用环境变量 fallback |
| API key 缺失 | 前端显示配置错误提示 |
| LLM API 调用失败 | SSE 推送 error 事件，前端显示错误提示条 |
| 工具执行失败 | 将错误信息作为 tool result 返回给 LLM |
| 会话不存在 | 自动创建新会话 |
| 上下文过长 | 裁剪最早的消息，保留 system prompt |
| 工具调用超过 10 轮 | 强制中断，返回已执行的摘要 |
| SSE 连接中断 | 前端显示"连接已断开"提示，不自动重连 |
