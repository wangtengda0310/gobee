package home

import (
	"encoding/json"
	"fmt"
	"os"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
)

// loadConfig 从文件加载配置，并合并 Claude Code 配置和环境变量
func (s *ChatService) loadConfig() (*ChatConfig, error) {
	var cfg ChatConfig

	// 1. 尝试从项目配置文件加载
	data, err := os.ReadFile(s.configFile)
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("解析项目配置失败: %w", err)
		}
	}

	// 2. 尝试从 Claude Code 配置补充（优先级：项目配置 > Claude Code 配置）
	claudeSettings, err := settings.LoadClaudeCodeConfig()
	if err == nil {
		applyClaudeCodeConfig(claudeSettings, &cfg)
	}

	// 3. 环境变量后备
	if cfg.AnthropicConfig.APIKey == "" {
		cfg.AnthropicConfig.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if cfg.OpenAIConfig.APIKey == "" {
		cfg.OpenAIConfig.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	return &cfg, nil
}

// applyClaudeCodeConfig 将 Claude Code 配置应用到项目配置
// 纯函数，不依赖 ChatService 实例状态
func applyClaudeCodeConfig(claudeSettings *settings.ClaudeCodeSettings, cfg *ChatConfig) {
	if claudeSettings == nil || claudeSettings.Env == nil {
		return
	}

	// 映射 Anthropic 配置
	if token, ok := claudeSettings.Env["ANTHROPIC_AUTH_TOKEN"]; ok && token != "" {
		cfg.AnthropicConfig.APIKey = token
	}

	if baseURL, ok := claudeSettings.Env["ANTHROPIC_BASE_URL"]; ok && baseURL != "" {
		cfg.AnthropicConfig.BaseURL = baseURL
	}

	// 模型优先级：Opus > Sonnet > Haiku
	// 聊天机器人需要工具调用能力，优先使用最强模型
	if model, ok := claudeSettings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"]; ok && model != "" {
		cfg.AnthropicConfig.Model = model
	} else if model, ok := claudeSettings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"]; ok && model != "" {
		cfg.AnthropicConfig.Model = model
	} else if model, ok := claudeSettings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"]; ok && model != "" {
		cfg.AnthropicConfig.Model = model
	}

	// 如果没有配置提供商，默认使用 anthropic
	if cfg.Provider == "" {
		cfg.Provider = "anthropic"
	}
}

// saveConfig 将配置序列化为 JSON 写入文件
func (s *ChatService) saveConfig(config *ChatConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configFile, data, 0644)
}

// getDefaultConfig 返回默认配置
func (s *ChatService) getDefaultConfig() *ChatConfig {
	return &ChatConfig{
		Provider: "anthropic",
		AnthropicConfig: AnthropicConfig{
			APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
			Model:     "claude-3-5-sonnet-20241022",
			MaxTokens: 4096,
		},
		OpenAIConfig: OpenAIConfig{
			APIKey:    os.Getenv("OPENAI_API_KEY"),
			BaseURL:   "https://api.openai.com/v1",
			Model:     "gpt-4o",
			MaxTokens: 4096,
		},
		SystemPrompt: "你是一个QA测试助手。你可以使用提供的工具来帮助用户检查Excel配表、查询数据、执行测试等。当用户提出与配表相关的问题时，请主动使用工具获取数据后再回答，不要凭猜测回答。",
	}
}
