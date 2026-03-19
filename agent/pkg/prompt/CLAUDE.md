# prompt 包 - 开发者指南

AI Agent 提示词构建和管理工具包。

## 架构设计

```
pkg/prompt/
├── doc.go              # 包文档
├── system.go           # System Prompt 构建器
├── history.go          # 对话历史管理
├── truncate.go         # 截断策略（策略模式）
├── tool.go             # 工具描述生成
└── output.go           # 结构化输出格式
```

## 核心类型

### SystemBuilder

链式构建系统提示词，生成 Markdown 格式输出。

**设计要点：**
- 使用 Builder 模式支持链式调用
- `Clone()` 方法支持创建变体配置
- `Reset()` 方法复用构建器实例

### History

管理对话历史，支持与 `llm.ChatCompleter` 集成。

**设计要点：**
- 内部使用 `HistoryItem` 存储，通过 `ToMessages()` 转换为 `llm.Message`
- 摘要生成函数通过 `SetSummaryGenerator()` 注入，避免硬依赖
- 支持 Clone 以便实现分支对话

### TruncateStrategy

策略模式接口，支持多种截断方式。

**已实现策略：**
| 策略 | 说明 | 适用场景 |
|------|------|----------|
| `SummaryTruncateStrategy` | LLM 摘要压缩 | 长对话、需要保留上下文 |
| `SlidingWindowStrategy` | 滑动窗口 | 简单场景、无 LLM 可用 |
| `FixedSystemStrategy` | 固定系统+滑动窗口 | 需要保留系统提示 |

## 与 llm 包的集成

```go
// 复用 llm.Tool 定义
tools := []*llm.Tool{
    llm.NewTool("search", "搜索", params),
}

// 输出转换为 llm.Message
history := prompt.NewHistory()
messages := history.ToMessages()

// 直接用于 ChatRequest
req := &llm.ChatRequest{
    Messages: messages,
    Tools:    tools,
}
```

## 扩展截断策略

实现 `TruncateStrategy` 接口：

```go
type MyStrategy struct{}

func (s *MyStrategy) Truncate(ctx context.Context, messages []HistoryItem) ([]HistoryItem, string, error) {
    // 实现逻辑
}
```

## 测试

```bash
go test ./pkg/prompt/... -v
```
