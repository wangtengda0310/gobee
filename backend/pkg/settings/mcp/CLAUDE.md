# MCP 服务器

`backend/pkg/settings/mcp/` 是内置 MCP 服务器的实现包，负责接收 MCP 客户端请求并暴露项目中的测试/检查能力为 Tools。

## 目录结构

```
mcp/
├── cobra.go          # stdio MCP 服务器入口（CLI 子命令）
├── cobra-help.md     # --help 文档
├── config.go          # MCP 配置加载与持久化
├── server.go          # HTTP MCP 服务器启动、停止
├── services.go        # 统一 MCP 服务容器（stdio 与 HTTP 共享）
├── tools.go           # MCP 工具注册（HTTP 与 stdio 共享）
├── server_test.go     # 单元测试（HTTP 服务器）
├── services_test.go   # 单元测试（服务容器）
├── cobra_test.go      # 单元测试（stdio 服务器）
└── tests/             # 手动/集成测试脚本
    └── test_mcp.bat
```

## 核心类型

| 类型 | 文件位置 | 说明 |
|------|----------|------|
| `MCPServer` | [server.go](server.go) | HTTP MCP 服务器实例，包含 http 服务与工具注册 |
| `McpServices` | [services.go](services.go) | 统一服务容器，聚合所有需要暴露给 MCP 的业务服务，stdio 与 HTTP 模式共享 |
| `StdioMCPServer` | [cobra.go](cobra.go) | stdio 传输的 MCP 服务器，用于 CLI 模式 |
| `settings.MCPConfig` | [wails.go:257](../wails.go:257) | MCP 配置结构（启用状态、端口、绑定地址） |

## 核心函数

| 函数/方法 | 文件位置 | 职责 |
|-----------|----------|------|
| `NewMCPServer` | [server.go](server.go) | 创建 HTTP MCP 服务器实例 |
| `MCPServer.Start` | [server.go](server.go) | 注册 tools 并启动 StreamableHTTP 服务 |
| `MCPServer.Stop` | [server.go](server.go) | 优雅关闭 HTTP 服务 |
| `MCPServer.registerTools` | [server.go](server.go) | 注册所有 MCP tools（内部调用 `registerAllMcpTools`） |
| `BuildDefaultMcpServices` | [services.go](services.go) | 创建默认 MCP 服务集合，stdio 与 HTTP 模式共用 |
| `registerAllMcpTools` | [tools.go](tools.go) | 注册所有 MCP tools（HTTP 与 stdio 共享） |
| `NewStdioMCPServer` | [cobra.go](cobra.go) | 创建 stdio MCP 服务器实例 |
| `StdioMCPServer.Start` | [cobra.go](cobra.go) | 启动 stdio 服务器，从 stdin 读取请求并写入 stdout |
| `DefaultConfig` | [config.go](config.go) | 返回默认 MCP 配置（127.0.0.1:8765，默认启用） |
| `LoadConfig` | [config.go](config.go) | 从文件加载 MCP 配置，不存在则创建默认配置 |
| `SaveConfig` | [config.go](config.go) | 保存 MCP 配置到文件 |

## 工具注册清单

`registerAllMcpTools` 统一注册所有业务 tools，被 HTTP 和 stdio 两种传输模式复用：

| 领域 | 注册函数 | 来源包 |
|------|----------|--------|
| 功能测试用例管理 | `functiontest.RegisterJsonCaseTools` | `pkg/function-test` |
| 功能测试配置 / Excel 配置 / MCP 配置 | `settings.RegisterConfigTools` | `pkg/settings` |
| Excel 检查 | `exceltest.RegisterExcelCheckTools` | `pkg/excel-test` |
| Excel 配置 | `exceltest.RegisterExcelConfigTools` | `pkg/excel-test` |
| 武将 Wiki 检查 | `herowikicheck.RegisterWikiCheckTools` | `pkg/hero-wiki-check` |
| 活动 Wiki 检查 | `activitywikicheck.RegisterActivityWikiCheckTools` | `pkg/activity-wiki-check` |
| 游戏数据（Excel 测试页专用） | `exceltest.RegisterGameExcelTools` | `pkg/excel-test` |
| 游戏数据（完整版） | `mcputil.RegisterGameExcelTools` | `pkg/common/mcp` |
| Robot 工会扩展 | `mcputil.RegisterRobotGuildTools` | `pkg/common/mcp` |
| 战斗测试 | `functiontest.RegisterFightCaseTools` | `pkg/function-test` |

完整可用工具列表及参数说明见项目文档 [MCP 接口使用手册](../../../../docs/MCP-USAGE.md)。

## 传输协议

- 使用 `modelcontextprotocol/go-sdk` 的 `StreamableHTTPHandler`（2025-03-26 协议版本）。
- 采用 `Stateless` 模式，避免 session 过期导致的连接问题。
- 默认监听地址：`http://127.0.0.1:8765`。

## CLI stdio 模式

- `rain-qa-func mcp` 无子命令时启动 stdio MCP 服务器。
- stdio 模式忽略 `mcpSection` 配置文件，直接启动并暴露完整工具集。
- 所有业务输出通过 JSON-RPC 返回 stdout；日志输出到 stderr。
- 外部客户端可通过以下配置连接：
  ```json
  {
    "mcpServers": {
      "rain-qa-func": {
        "command": "rain-qa-func",
        "args": ["mcp"]
      }
    }
  }
  ```

## 测试

| 测试文件 | 类型 | 覆盖范围 |
|----------|------|----------|
| [server_test.go](server_test.go) | 单元测试 | 默认配置、配置地址、服务器创建 |
| [services_test.go](services_test.go) | 单元测试 | `BuildDefaultMcpServices` 创建的服务非空验证 |
| [cobra_test.go](cobra_test.go) | 单元测试 | `NewStdioMCPServer` 创建、`NewMCPCmd` 子命令创建 |
| [tests/test_mcp.bat](tests/test_mcp.bat) | 手动集成测试 | 通过 curl 调用各 tool 验证响应 |

## 开发注意事项

- 新增业务 tool 应在对应业务包内实现注册函数，再由 [tools.go](tools.go) 的 `registerAllMcpTools` 调用。
- 修改 MCP 配置保存后会自动调用 `UpdateConfigAndRestart` 重启服务，见 [wails.go:321](../wails.go:321)。
- Stateless 模式下每个请求独立，因此 [GetConnectionCount](server.go) 当前保持为 0。
- stdio 与 HTTP 两种传输模式共享同一套 `McpServices` 服务实例，传输层差异保留在 `MCPServer` / `StdioMCPServer` 类型上。
