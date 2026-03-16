// Package mcp 提供 MCP (Model Context Protocol) 扩展。
//
// 基于 github.com/mark3labs/mcp-go 构建，提供:
//   - 工具注册辅助函数
//   - 资源管理
//   - Prompt 模板
//
// 使用示例:
//
//	server := mcp.NewServer("my-tool", "1.0.0")
//	mcp.RegisterTool(server, "echo", "回显输入", handleEcho)
//	server.Start()
package mcp
