# prompt - AI Agent 提示词工具包

[![GoDoc](https://godoc.org/github.com/wangtengda0310/gobee/agent/pkg/prompt?status.svg)](https://godoc.org/github.com/wangtengda0310/gobee/agent/pkg/prompt)

提供 AI Agent 提示词构建和管理工具，帮助快速构建 Agent 并便捷地组织提示词。

## 安装

```bash
go get github.com/wangtengda0310/gobee/agent/pkg/prompt
```

## 快速开始

### 构建系统提示词

```go
package main

import (
    "fmt"

    "github.com/wangtengda0310/gobee/agent/pkg/prompt"
    "github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func main() {
    // 定义工具
    tools := []*llm.Tool{
        llm.NewTool("search", "搜索文档", map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "query": map[string]interface{}{
                    "type":        "string",
                    "description": "搜索关键词",
                },
            },
            "required": []string{"query"},
        }),
    }

    // 构建系统提示词
    system := prompt.NewSystem("代码助手").
        WithDescription("你是一个专业的编程助手，帮助用户解决编程问题").
        WithCapabilities(
            "代码编写和审查",
            "错误诊断和修复",
            "最佳实践建议",
        ).
        WithConstraint("回答必须简洁准确").
        WithConstraint("提供代码示例时添加注释").
        WithTools(tools...).
        Build()

    fmt.Println(system)
}
```

### 管理对话历史

```go
// 创建历史管理器
history := prompt.NewHistory()

// 添加消息（支持链式调用）
history.
    AddUser("如何读取 JSON 文件？").
    AddAssistant("可以使用 encoding/json 包...").
    AddUser("能给个示例吗？")

// 转换为 llm.Message 格式
messages := history.ToMessages()

// 用于 LLM 请求
req := &llm.ChatRequest{
    Messages: messages,
}
```

### 上下文截断

```go
// 设置摘要生成函数
history.SetSummaryGenerator(func(ctx context.Context, content string) (string, error) {
    // 使用 LLM 生成摘要
    resp, err := client.Complete(ctx, &llm.ChatRequest{
        Messages: []*llm.Message{
            {Role: llm.RoleUser, Content: llm.Text("总结以下对话:\n" + content)},
        },
    })
    if err != nil {
        return "", err
    }
    return resp.Content, nil
})

// 使用摘要策略截断
strategy := prompt.NewSummaryTruncateStrategy(10) // 保留最近 10 条
truncated, summary, err := history.Truncate(ctx, strategy)
```

### 结构化输出

```go
// JSON 输出
format := prompt.JSONOutput(map[string]any{
    "type": "object",
    "properties": map[string]any{
        "code":    map[string]any{"type": "string"},
        "explain": map[string]any{"type": "string"},
    },
    "required": []string{"code"},
})

builder := prompt.NewSystem("代码生成器").
    WithOutputFormat(format)
```

## API 参考

### SystemBuilder

| 方法 | 说明 |
|------|------|
| `NewSystem(role)` | 创建构建器 |
| `WithDescription(desc)` | 设置描述 |
| `WithCapabilities(...)` | 添加能力 |
| `WithConstraint(c)` | 添加约束 |
| `WithTools(...)` | 添加工具 |
| `WithOutputFormat(f)` | 设置输出格式 |
| `WithContext(ctx)` | 设置上下文 |
| `Build()` | 构建提示词 |
| `Clone()` | 克隆构建器 |

### History

| 方法 | 说明 |
|------|------|
| `NewHistory()` | 创建历史管理器 |
| `AddUser(content)` | 添加用户消息 |
| `AddAssistant(content)` | 添加助手消息 |
| `AddToolResult(id, name, result)` | 添加工具结果 |
| `ToMessages()` | 转换为 llm.Message |
| `Truncate(ctx, strategy)` | 截断历史 |
| `Clone()` | 克隆历史 |

### 截断策略

| 策略 | 说明 |
|------|------|
| `NewSummaryTruncateStrategy(n)` | LLM 摘要压缩 |
| `NewSlidingWindowStrategy(n)` | 滑动窗口 |
| `NewFixedSystemStrategy(n)` | 固定系统+滑动窗口 |

## License

MIT
