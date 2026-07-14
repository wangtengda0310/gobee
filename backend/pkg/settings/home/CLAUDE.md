# home 包 — ChatService 聊天服务

## 概述

提供 AI 聊天服务，集成 gobee Agent 框架，支持流式对话、工具调用和配置管理。

## 文件结构

| 文件 | 职责 |
|------|------|
| `wails.go` | ChatService 结构体、构造函数、所有 @frontend 导出方法 |
| `types.go` | 类型定义（ChatConfig, ChatMessage, ChatSession, memoryAdapter 等） |
| `config.go` | 配置加载/保存/Claude Code 合并 |
| `client.go` | LLM 客户端创建、请求构建、模型名解析 |
| `agent.go` | EventEmitter 接口、Agent 创建、流式事件处理 |
| `testing-guide.md` | 单元测试编写指南（Mock 策略、测试模式、常见陷阱） |

## 核心接口

### EventEmitter

解耦 Wails `*application.App` 的事件系统，使 ChatService 可脱离 Wails 测试。

```go
type EventEmitter interface {
    Emit(name string, data ...any) bool
}
```

`*application.EventManager` 隐式实现此接口，构造时直接赋值 `app.Event`。

### llm.ChatCompleter

gobee 框架提供的 LLM 客户端接口，mock 实现在 `gobee/agent/pkg/llm/mock`。

## Wails 绑定

`@frontend` 是项目内部文档约定，不是 Wails v3 指令。Wails 绑定基于导出方法自动发现（`application.NewService[*ChatService]`），跨文件安全。

## 事件名称

| 事件名 | 触发场景 |
|--------|---------|
| `chatStreamStart` | 开始处理输入 |
| `chatStreamChunk` | 内容增量（附带 seq 序列号） |
| `chatStreamDone` | 处理完成 |
| `chatToolCall` | 工具调用 |
| `chatToolResult` | 工具结果 |
| `chatStreamError` | 错误 |

## 测试

运行：`go test ./backend/pkg/settings/home/ -v`

详细测试编写指南见 [testing-guide.md](testing-guide.md)。

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/home/home.spec.ts`](../../../../frontend/e2e/home/home.spec.ts) | 页面加载、聊天容器、消息输入、配置面板布局、流式生成 |

