package chat

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ClaudeConfig 从 Claude Code 配置文件读取的 LLM 配置
type ClaudeConfig struct {
	APIKey       string            // API Key
	BaseURL      string            // API 基础 URL
	Models       map[string]string // tier -> 模型名（haiku/sonnet/opus）
	DefaultModel string            // 当前默认使用的实际模型名
}

// settingsFile Claude Code settings.json 的结构（仅解析需要的字段）
type settingsFile struct {
	Env   map[string]string `json:"env"`
	Model string            `json:"model"`
}

// LoadConfig 加载配置，优先从文件读取，fallback 到环境变量
func LoadConfig(settingsPath string) (*ClaudeConfig, error) {
	_, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Printf("无法读取 Claude Code 配置文件 %s: %v，使用环境变量 fallback\n", settingsPath, err)
		return loadConfigFromEnv()
	}
	return loadConfigFromPath(settingsPath)
}

// loadConfigFromPath 从指定路径读取配置文件
func loadConfigFromPath(path string) (*ClaudeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var sf settingsFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	cfg := &ClaudeConfig{
		APIKey:  sf.Env["ANTHROPIC_AUTH_TOKEN"],
		BaseURL: sf.Env["ANTHROPIC_BASE_URL"],
		Models: map[string]string{
			"haiku":  sf.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"],
			"sonnet": sf.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"],
			"opus":   sf.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"],
		},
	}

	// 解析 model 字段获取默认模型
	tier := parseModelTier(sf.Model)
	if tier == "" {
		tier = "sonnet"
	}
	if model, ok := cfg.Models[tier]; ok && model != "" {
		cfg.DefaultModel = model
	} else if cfg.Models["sonnet"] != "" {
		cfg.DefaultModel = cfg.Models["sonnet"]
	} else {
		// 如果没有配置任何模型，设置一个默认值
		cfg.DefaultModel = "claude-3-sonnet-20240229"
	}

	return cfg, nil
}

// loadConfigFromEnv 从环境变量加载配置（fallback 方案）
func loadConfigFromEnv() (*ClaudeConfig, error) {
	return &ClaudeConfig{
		APIKey:       os.Getenv("ANTHROPIC_AUTH_TOKEN"),
		BaseURL:      os.Getenv("ANTHROPIC_BASE_URL"),
		Models:       map[string]string{},
		DefaultModel: "",
	}, nil
}

// parseModelTier 从 Claude Code 的 model 字段值解析 tier 名
// "opus[1m]" -> "opus", "sonnet" -> "sonnet"
func parseModelTier(model string) string {
	if idx := strings.Index(model, "["); idx > 0 {
		return model[:idx]
	}
	switch model {
	case "haiku", "sonnet", "opus":
		return model
	default:
		return ""
	}
}

// GetModelForTier 根据 tier 名获取实际模型名
func (c *ClaudeConfig) GetModelForTier(tier string) string {
	if m, ok := c.Models[tier]; ok && m != "" {
		return m
	}
	return c.DefaultModel
}