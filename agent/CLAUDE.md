# gobee/agent - AI Agent 开发框架

Go 语言实现的 AI Agent 开发框架，提供通用能力为其他项目使用。

## 项目结构

```
agent/
├── cmd/                    # 可执行程序入口
│   ├── dir/               # 文件系统 MCP 工具
│   ├── time/              # 时间查询 MCP 工具
│   └── llmtest/           # LLM 接口测试工具
├── pkg/                    # 公共库
│   ├── llm/               # 多模型适配层 (已完成)
│   ├── agent/             # Agent 框架核心
│   ├── tool/              # 工具链集成
│   ├── memory/            # 对话/记忆管理
│   └── mcp/               # MCP 协议扩展
├── openclaw/              # OpenClaw 相关模块
│   └── channel/           # WebSocket 通道 + 内网穿透
└── internal/              # 内部实现（不对外暴露）
```

## 技术栈

- **Go 1.23.4**
- **MCP**: `github.com/mark3labs/mcp-go`
- **CLI**: `github.com/spf13/cobra`
- **WebSocket**: `github.com/gorilla/websocket`

## 开发规范

### 模块设计原则

1. **接口优先**: 公共 API 必须先定义接口
2. **依赖注入**: 通过构造函数传入依赖
3. **配置外置**: 支持环境变量和配置文件
4. **错误包装**: 使用 `fmt.Errorf("操作失败: %w", err)` 保留上下文

### 代码风格

```bash
# 格式化
go fmt ./...

# 静态检查
go vet ./...

# 运行测试
go test ./...
```

### 命名约定

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写单词 | `agent`, `llm`, `tool` |
| 接口 | 动词+er | `ChatCompleter`, `ToolExecutor` |
| 结构体 | 名词 | `Session`, `Message`, `Tool` |
| 常量 | 全大写 | `MaxRetries`, `DefaultTimeout` |

## 子模块文档

- [LLM 包](pkg/llm/doc.go) - 多模型适配接口
- [OpenClaw Channel](openclaw/channel/CLAUDE.md) - WebSocket 通道、内网穿透

## 开发路线

### Phase 1: 多模型适配 (pkg/llm) ✅
- [x] 统一的 Chat Completion 接口
- [x] OpenAI 适配器
- [x] Claude/Anthropic 适配器
- [ ] 本地模型适配器 (Ollama)

### Phase 2: Agent 框架 (pkg/agent)
- [ ] Agent 核心循环
- [ ] 工具调用机制
- [ ] 多 Agent 协作

### Phase 3: 工具链 (pkg/tool)
- [ ] 代码执行
- [ ] 文件操作
- [ ] 搜索集成
- [ ] MCP 工具扩展

### Phase 4: 记忆管理 (pkg/memory)
- [ ] 会话持久化
- [ ] 上下文压缩
- [ ] 向量存储集成

## 快速开始

```bash
# 运行 MCP 工具
go run ./cmd/dir
go run ./cmd/time

# 运行 OpenClaw Channel
go run ./openclaw/channel api-enhanced --port 8080

# 测试 LLM 接口 (使用智谱代理)
ANTHROPIC_API_KEY="xxx" go run ./cmd/llmtest \
  -provider anthropic \
  -model "glm-5" \
  -base-url "https://open.bigmodel.cn/api/anthropic/v1" \
  -prompt "Hello"

# 测试流式响应
ANTHROPIC_API_KEY="xxx" go run ./cmd/llmtest \
  -provider anthropic \
  -model "glm-5" \
  -base-url "https://open.bigmodel.cn/api/anthropic/v1" \
  -prompt "Hello" \
  -stream
```

## LLM 包使用示例

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/wangtengda0310/gobee/agent/pkg/llm"
    "github.com/wangtengda0310/gobee/agent/pkg/llm/anthropic"
)

func main() {
    // 创建客户端
    client, err := anthropic.NewClient(
        anthropic.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
        anthropic.WithModel("claude-sonnet-4-20250514"),
    )
    if err != nil {
        panic(err)
    }

    // 发送请求
    req := &llm.ChatRequest{
        Messages: []*llm.Message{
            {Role: llm.RoleUser, Content: llm.Text("Hello!")},
        },
    }
    resp, err := client.Complete(context.Background(), req)
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Content)
}
```

## 依赖管理

```bash
# 添加依赖
go get github.com/example/package

# 整理依赖
go mod tidy

# 更新依赖
go get -u ./...
```
