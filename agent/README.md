# gobee/agent

Go 语言 AI Agent 开发框架，提供多模型适配、工具链集成、对话管理等通用能力。

## 特性

- **多模型适配** - 统一接口支持 OpenAI、Claude、本地模型
- **MCP 协议** - 完整的 Model Context Protocol 支持
- **工具链集成** - 代码执行、文件操作、搜索等
- **对话管理** - 会话持久化、上下文压缩
- **内网穿透** - OpenClaw Channel 反向连接

## 安装

```bash
go get github.com/wangtengda0310/gobee/agent
```

## 快速开始

### MCP 工具

```bash
# 文件系统工具
go run ./cmd/dir

# 时间查询工具
go run ./cmd/time
```

### OpenClaw Channel

```bash
# 启动 API 服务器（带 Web UI）
go run ./openclaw/channel api-enhanced --port 8080 --token YOUR_TOKEN

# 访问 http://localhost:8080/
```

## 项目结构

```
agent/
├── cmd/            # 可执行程序
├── pkg/            # 公共库
│   ├── llm/        # 多模型适配
│   ├── agent/      # Agent 框架
│   ├── tool/       # 工具链
│   └── memory/     # 对话管理
├── openclaw/       # OpenClaw 模块
└── internal/       # 内部实现
```

## 文档

- [开发者指南](CLAUDE.md) - 架构设计和开发规范
- [OpenClaw Channel](openclaw/channel/README.md) - WebSocket 通道使用说明

## 状态

| 模块 | 状态 |
|------|------|
| MCP 工具 | ✅ 可用 |
| OpenClaw Channel | ✅ 可用 |
| 多模型适配 | 🚧 开发中 |
| Agent 框架 | 📋 规划中 |
| 工具链 | 📋 规划中 |
| 记忆管理 | 📋 规划中 |

## License

MIT
