package mcp

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPServer(t *testing.T) {
	config := DefaultConfig()
	services := &McpServices{} // 空服务用于测试服务器创建

	server := NewMCPServer(config, services)

	require.NotNil(t, server, "NewMCPServer returned nil")
	assert.Equal(t, config, server.config, "Server config not set correctly")
	assert.Equal(t, services, server.services, "Server services not set correctly")
	assert.NotNil(t, server.server, "MCP server not initialized")
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled, "Default config should have Enabled=true")
	assert.Equal(t, 8765, config.Port, "Default port should be 8765")
	assert.Equal(t, "127.0.0.1", config.Host, "Default host should be 127.0.0.1")
}

func TestConfigAddress(t *testing.T) {
	cfg := &settings.MCPConfig{
		Host: "localhost",
		Port: 9000,
	}

	expected := "localhost:9000"
	assert.Equal(t, expected, cfg.Address(), "Address() returned unexpected value")
}
