# MCP stdio 子命令设计

## 背景

`rain-qa-func` 同时提供 Wails GUI 和 Cobra CLI 两种启动方式。GUI 模式下会通过 HTTP 启动一个内置 MCP 服务器（`backend/pkg/settings/mcp/server.go`），暴露项目中的测试/检查能力给外部 MCP 客户端。

当前 `backend/pkg/settings/mcp/cobra.go` 只打印 `--help`，没有实现 CLI 模式下的 MCP 服务器入口。本设计将 `mcp` 子命令实现为 **stdio 传输的 MCP 服务器**，使外部客户端（如 Claude Code）可通过 `stdin/stdout` 调用与 HTTP 版本一致的工具集。

## 目标

1. `rain-qa-func mcp` 无子命令时启动 stdio MCP 服务器。
2. 暴露的工具集与现有 HTTP MCP 服务器完全一致（完整服务）。
3. 复用现有服务初始化逻辑，不重复 inline 组装 `Services`。
4. stdio 模式忽略持久化的 MCP 配置文件，总是直接启动。
5. 服务容器命名与 Wails Service 区分，避免混淆。

## 非目标

- 不修改各业务包的 `RegisterXxxTools` 注册函数。
- 不新增业务工具。
- 不改变 HTTP MCP 服务器的传输协议和默认行为。

## 关键决策

### 1. 传输协议

使用 `modelcontextprotocol/go-sdk` 提供的 stdio 服务器能力。该 SDK 已作为项目依赖存在，无需引入新库。

### 2. 服务命名

为了避免与 Wails 的 `application.Service` 混淆，MCP 专用的服务容器统一采用 `Mcp` 后缀，并按传输方式区分：

- `StdioMcpServices`：stdio 模式下的服务容器，字段为 `StdioMcpJsonCaseService` 等。
- `HttpMcpServices`：HTTP 模式下的服务容器，字段为 `HttpMcpJsonCaseService` 等。

两个容器内部都委托到同一个私有 helper 创建实际的 Service 实例，避免重复代码。

### 3. 配置行为

stdio 模式下不读取 `mcpSection` 配置文件，也不检查 `enabled` 字段。命令被调用即表示用户明确希望启动 stdio MCP 服务器。

### 4. 子命令结构

```
rain-qa-func mcp
```

`mcp` 命令自身就是 stdio 服务入口，不带参数时直接启动。

### 5. 输出约定

stdio 模式下所有正常业务输出必须通过 MCP JSON-RPC 通道返回，避免污染 stdout。运行期日志可输出到 stderr；初始化失败时向 stderr 输出错误并退出。

## 架构

```
外部 MCP 客户端
      │  stdin (JSON-RPC)
      ▼
backend/pkg/settings/mcp/cobra.go
      │
      ▼
NewStdioMCPServer(StdioMcpServices)
      │
      ▼
mcpgo.Server
      │
      ▼
registerTools() ──► 各业务 RegisterXxxTools
      │
      ▼
各业务 Service 的 tool handler
      │
      ▼
stdout (JSON-RPC)
```

## 数据流

1. 外部客户端启动 `rain-qa-func mcp` 子进程。
2. `cobra.go` 调用 `BuildDefaultStdioMcpServices()` 组装服务。
3. 创建 stdio MCP 服务器，将 `StdioMcpServices` 注册到 `mcpgo.Server`。
4. stdio 服务器从 stdin 读取 JSON-RPC 请求，分发到对应 tool handler。
5. handler 调用业务 Service，返回 `CallToolResult`。
6. stdio 服务器将结果序列化为 JSON-RPC 响应并写入 stdout。

## 文件变更

| 文件 | 变更 |
|---|---|
| `backend/pkg/settings/mcp/services.go` | 新增：定义 `StdioMcpServices`、`HttpMcpServices`、构建函数和私有 helper |
| `backend/pkg/settings/mcp/server.go` | 修改：`MCPServer` 使用 `*HttpMcpServices`，构造参数更新 |
| `backend/pkg/settings/mcp/cobra.go` | 修改：实现 `mcp` 子命令为 stdio MCP 服务器入口 |
| `backend/pkg/settings/mcp/cobra-help.md` | 修改：更新帮助文档，说明 stdio 用途和客户端连接方式 |
| `cmd/rain-qa-func/wails.go` | 修改：用 `BuildDefaultHttpMcpServices()` 替换 inline 服务组装 |
| `backend/pkg/settings/mcp/CLAUDE.md` | 更新：补充服务容器、stdio 子命令说明、相关函数索引 |

## 接口设计

### services.go

```go
// StdioMcpServices stdio 模式下暴露的 MCP 服务集合
type StdioMcpServices struct {
    StdioMcpJsonCaseService          *functiontest.JsonCaseService
    StdioMcpFuncCaseConfigService    *functiontest.FuncCaseConfigService
    StdioMcpExcelCheckService        *exceltest.ExcelCheckService
    StdioMcpExcelConfigService       *exceltest.ExcelConfigService
    StdioMcpGameExcelService         *game.GameExcelService
    StdioMcpExcelTestGameService     *exceltest.ExcelTestGameService
    StdioMcpHeroResCheckService      *resourcechecker.HeroResCheckService
    StdioMcpHeroWikiResCheckService  *herowikicheck.HeroWikiResCheckService
    StdioMcpActivityWikiCheckService *activitywikicheck.ActivityWikiCheckService
    StdioMcpConfigService            *settings.MCPConfigService
    StdioMcpFeishuNotifyConfigService *settings.FeishuNotifyConfigService
    StdioMcpRobotExtService          *robotext.RobotExtService
}

// HttpMcpServices HTTP 模式下暴露的 MCP 服务集合
type HttpMcpServices struct {
    HttpMcpJsonCaseService          *functiontest.JsonCaseService
    HttpMcpFuncCaseConfigService    *functiontest.FuncCaseConfigService
    HttpMcpExcelCheckService        *exceltest.ExcelCheckService
    HttpMcpExcelConfigService       *exceltest.ExcelConfigService
    HttpMcpGameExcelService         *game.GameExcelService
    HttpMcpExcelTestGameService     *exceltest.ExcelTestGameService
    HttpMcpHeroResCheckService      *resourcechecker.HeroResCheckService
    HttpMcpHeroWikiResCheckService  *herowikicheck.HeroWikiResCheckService
    HttpMcpActivityWikiCheckService *activitywikicheck.ActivityWikiCheckService
    HttpMcpConfigService            *settings.MCPConfigService
    HttpMcpFeishuNotifyConfigService *settings.FeishuNotifyConfigService
    HttpMcpRobotExtService          *robotext.RobotExtService
}

// BuildDefaultStdioMcpServices 创建 stdio 模式所需的默认 MCP 服务集合
func BuildDefaultStdioMcpServices() (*StdioMcpServices, error)

// BuildDefaultHttpMcpServices 创建 HTTP 模式所需的默认 MCP 服务集合
func BuildDefaultHttpMcpServices() (*HttpMcpServices, error)
```

### cobra.go

```go
func NewMCPCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mcp",
        Short: "启动 stdio MCP 服务器",
        Long:  cobraHelpText,
        RunE: func(cmd *cobra.Command, args []string) error {
            services, err := BuildDefaultStdioMcpServices()
            if err != nil {
                return err
            }
            server := NewStdioMCPServer(services)
            return server.Start(context.Background())
        },
    }
    return cmd
}
```

## 错误处理

- 服务初始化失败：向 stderr 输出错误，进程以非零状态退出。
- 工具调用失败：按 MCP 协议返回 `IsError: true` 的 `CallToolResult`，不中断 stdio 连接。
- stdio 传输错误：记录到 stderr，必要时关闭连接并退出进程。

## 测试策略

1. **单元测试 `services.go`**：`BuildDefaultStdioMcpServices()` 和 `BuildDefaultHttpMcpServices()` 返回非 nil 的服务实例，各字段均被正确赋值。
2. **单元测试 `NewStdioMCPServer`**：创建 stdio 服务器不 panic，调用 `tools/list` 返回非空工具列表。
3. **不直接测试 stdio 端到端交互**：需要外部进程配合，由集成测试或手动测试覆盖。

## 风险与缓解

| 风险 | 缓解 |
|---|---|
| stdio 输出被日志污染 | 禁止向 stdout 打印非 JSON-RPC 内容；日志输出到 stderr |
| 服务初始化与 Wails 强耦合 | 提取 `services.go` 独立组装，不依赖 `application.App` |
| 命名冲突 | 字段统一加 `StdioMcp` / `HttpMcp` 前缀 |

## 参考

- `backend/pkg/settings/mcp/server.go`：现有 HTTP MCP 服务器实现
- `backend/pkg/settings/mcp/cobra.go`：当前占位实现
- `cmd/rain-qa-func/wails.go`：现有服务初始化 inline 代码
- `docs/MCP-USAGE.md`：MCP 工具使用手册
