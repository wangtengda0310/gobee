package mcp

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
)

// mcpSection 统一配置文件中的 MCP section 读写器
var mcpSection = appconfig.New("mcp")

// DefaultConfig 返回默认配置
func DefaultConfig() *settings.MCPConfig {
	return &settings.MCPConfig{
		Enabled: true,
		Port:    8765,
		Host:    "127.0.0.1",
	}
}

// LoadConfig 从统一配置文件加载 MCP section
// section 不存在时返回默认值（不写盘，避免读操作产生副作用）
func LoadConfig() (*settings.MCPConfig, error) {
	var cfg settings.MCPConfig
	if err := mcpSection.Load(&cfg); err != nil {
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	if !mcpSection.Exists() {
		return DefaultConfig(), nil
	}

	return &cfg, nil
}

// SaveConfig 保存 MCP 配置到统一配置文件
func SaveConfig(cfg *settings.MCPConfig) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}
	return mcpSection.Save(cfg)
}
