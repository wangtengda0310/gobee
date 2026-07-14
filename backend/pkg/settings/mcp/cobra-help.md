# mcp — MCP 服务器管理工具

## 简介

`mcp` 子命令用于启动一个 **stdio 传输** 的内置 MCP（Model Context Protocol）服务器。

当外部 MCP 客户端（如 Claude Code）通过 stdio 启动本程序时，客户端可以通过 `stdin/stdout` 调用项目中的所有 MCP 工具。

## 用法

```bash
rain-qa-func mcp
```

无子命令时直接启动 stdio MCP 服务器。

## 启动方式

外部客户端通常会使用如下方式启动：

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

## 暴露的工具

stdio 模式下暴露的工具与 GUI 模式下启动的 HTTP MCP 服务器完全一致，包括：

- 功能测试用例管理
- 功能测试配置 / Excel 配置 / MCP 配置
- Excel 检查
- 武将 Wiki 检查
- 活动 Wiki 检查
- 游戏数据查询
- Robot 工会扩展
- 战斗测试

完整工具列表及参数说明见项目文档 `docs/MCP-USAGE.md`。

## 配置行为

stdio 模式**忽略**持久化的 MCP 配置文件，被调用即直接启动。绑定地址、端口等 HTTP 相关配置在 stdio 模式下不适用。

## 输出约定

stdio 模式下所有业务输出均通过 MCP JSON-RPC 通道返回，避免污染 stdout。运行期日志输出到 stderr。

## 示例

```bash
# 启动 stdio MCP 服务器
rain-qa-func mcp

# 查看帮助
rain-qa-func mcp --help
```

## 更多信息

详见 `backend/pkg/settings/mcp/` 目录及 `CLAUDE.md`。
