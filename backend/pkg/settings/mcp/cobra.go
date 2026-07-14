package mcp

// MCP CLI 子命令定义。
// mcp 无子命令时启动 stdio 传输的 MCP 服务器，暴露与 HTTP 模式完全一致的工具集。

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed cobra-help.md
var cobraHelpText string

// NewMCPCmd 创建 mcp 子命令。
// 无子命令时启动 stdio 传输的 MCP 服务器。
func NewMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "启动 stdio MCP 服务器",
		Long:  cobraHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := BuildDefaultMcpServices()
			if err != nil {
				return fmt.Errorf("构建 stdio MCP 服务失败: %w", err)
			}
			server := NewStdioMCPServer(services)
			return server.Start(cmd.Context())
		},
	}
	return cmd
}

// StdioMCPServer stdio 传输的 MCP 服务器
type StdioMCPServer struct {
	server   *mcpgo.Server
	services *McpServices
}

// NewStdioMCPServer 创建 stdio MCP 服务器实例
func NewStdioMCPServer(services *McpServices) *StdioMCPServer {
	s := mcpgo.NewServer(&mcpgo.Implementation{
		Name:    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func",
		Version: "1.0.0",
	}, nil)

	return &StdioMCPServer{
		server:   s,
		services: services,
	}
}

// Start 启动 stdio MCP 服务器，从 stdin 读取请求并写入 stdout
// 使用 SDK 的 StdioTransport 通过 server.Run 启动 stdio 模式
func (s *StdioMCPServer) Start(ctx context.Context) error {
	s.registerStdioTools()
	return s.server.Run(ctx, &mcpgo.StdioTransport{})
}

// registerStdioTools 注册所有 stdio MCP tools
// 工具集与 HTTP 服务器的 registerTools() 完全一致，通过共享函数 registerAllMcpTools 实现
func (s *StdioMCPServer) registerStdioTools() {
	registerAllMcpTools(s.server, s.services)
}
