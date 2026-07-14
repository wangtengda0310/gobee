package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStdioMCPServer 验证 stdio MCP 服务器创建
func TestNewStdioMCPServer(t *testing.T) {
	services, err := BuildDefaultMcpServices()
	require.NoError(t, err)

	server := NewStdioMCPServer(services)
	require.NotNil(t, server)
	assert.NotNil(t, server.server)
	assert.NotNil(t, server.services)
}

// TestNewMCPCmd 验证 mcp 子命令创建
func TestNewMCPCmd(t *testing.T) {
	cmd := NewMCPCmd()
	require.NotNil(t, cmd)
	assert.Equal(t, "mcp", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}
