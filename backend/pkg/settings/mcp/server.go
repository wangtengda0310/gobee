package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer MCP 服务器
type MCPServer struct {
	server       *mcpgo.Server
	config       *settings.MCPConfig
	services     *McpServices
	running      bool
	httpSrv      *http.Server
	connectionMu sync.Mutex
	connections  int
}

// NewMCPServer 创建 MCP 服务器实例
func NewMCPServer(config *settings.MCPConfig, services *McpServices) *MCPServer {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建 MCP 服务器
	s := mcpgo.NewServer(&mcpgo.Implementation{
		Name:    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func",
		Version: "1.0.0",
	}, nil)

	return &MCPServer{
		server:   s,
		config:   config,
		services: services,
		running:  false,
	}
}

// Start 启动 MCP 服务器
func (s *MCPServer) Start(ctx context.Context) error {
	if !s.config.Enabled {
		log.Println("MCP 服务器已禁用，跳过启动")
		return nil
	}

	// 如果已经在运行，先停止
	if s.running {
		log.Println("MCP 服务器已在运行中，先停止")
		_ = s.Stop(ctx)
		time.Sleep(100 * time.Millisecond)
	}

	// 注册所有 tools
	s.registerTools()

	// 启动 MCP 服务器
	go func() {
		s.running = true
		addr := s.config.Address()
		log.Printf("MCP 服务器启动在 http://%s", addr)

		// 使用 StreamableHTTPHandler (2025-03-26 协议版本)
		// 这是 Claude Code 等现代 MCP 客户端期望的传输方式
		// 使用 Stateless 模式避免 session 过期导致的连接问题
		// 注意：Stateless 模式下不统计连接数，因为每个请求都是独立的
		handler := mcpgo.NewStreamableHTTPHandler(func(request *http.Request) *mcpgo.Server {
			return s.server
		}, &mcpgo.StreamableHTTPOptions{
			Stateless: true,
		})

		s.httpSrv = &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("MCP 服务器错误: %v", err)
		}

		s.running = false
		s.connectionMu.Lock()
		s.connections = 0
		s.connectionMu.Unlock()
		log.Println("MCP 服务器已停止")
	}()

	return nil
}

// Stop 停止 MCP 服务器
func (s *MCPServer) Stop(ctx context.Context) error {
	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			return err
		}
	}
	s.running = false
	return nil
}

// IsRunning 检查服务器是否在运行
func (s *MCPServer) IsRunning() bool {
	return s.running
}

// GetConfig 获取当前配置
func (s *MCPServer) GetConfig() *settings.MCPConfig {
	return s.config
}

// GetConnectionCount 获取当前连接数
func (s *MCPServer) GetConnectionCount() int {
	s.connectionMu.Lock()
	defer s.connectionMu.Unlock()
	return s.connections
}

// StartService 启动 MCP 服务（如果已运行则先停止再启动）
func (s *MCPServer) StartService(enabled bool, port int, host string) error {
	s.config.Enabled = enabled
	s.config.Port = port
	s.config.Host = host

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.Start(ctx)
}

// StopService 停止 MCP 服务
func (s *MCPServer) StopService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.config.Enabled = false
	return s.Stop(ctx)
}

// Restart 重启 MCP 服务器
func (s *MCPServer) Restart(ctx context.Context, newConfig *settings.MCPConfig) error {
	// 停止当前服务
	if s.running {
		if err := s.Stop(ctx); err != nil {
			return fmt.Errorf("停止 MCP 服务失败: %v", err)
		}
		// 等待服务完全停止
		time.Sleep(100 * time.Millisecond)
	}

	// 更新配置
	if newConfig != nil {
		s.config = newConfig
	}

	// 重新创建 MCP 服务器实例
	s.server = mcpgo.NewServer(&mcpgo.Implementation{
		Name:    "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func",
		Version: "1.0.0",
	}, nil)

	// 重新启动
	return s.Start(ctx)
}

// UpdateConfigAndRestart 更新配置并重启服务
func (s *MCPServer) UpdateConfigAndRestart(enabled bool, port int, host string) error {
	s.config.Enabled = enabled
	s.config.Port = port
	s.config.Host = host

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.Restart(ctx, s.config)
}

// registerTools 注册所有 MCP tools
func (s *MCPServer) registerTools() {
	registerAllMcpTools(s.server, s.services)
}
