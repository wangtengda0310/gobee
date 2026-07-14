package home

import (
	"encoding/json"
	"os"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	"github.com/stretchr/testify/assert"
)

// TestGetDefaultConfig 默认配置结构完整
func TestGetDefaultConfig(t *testing.T) {
	svc, _ := newTestChatService(t)
	cfg := svc.getDefaultConfig()

	assert.Equal(t, "anthropic", cfg.Provider)
	assert.Equal(t, "claude-3-5-sonnet-20241022", cfg.AnthropicConfig.Model)
	assert.Equal(t, 4096, cfg.AnthropicConfig.MaxTokens)
	assert.Equal(t, "gpt-4o", cfg.OpenAIConfig.Model)
	assert.Equal(t, "https://api.openai.com/v1", cfg.OpenAIConfig.BaseURL)
	assert.NotEmpty(t, cfg.SystemPrompt)
}

// TestLoadConfig_FileNotExist 配置文件不存在时返回空配置
func TestLoadConfig_FileNotExist(t *testing.T) {
	svc, _ := newTestChatService(t)
	cfg, err := svc.loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

// TestLoadConfig_ValidFile JSON 文件正确解析
func TestLoadConfig_ValidFile(t *testing.T) {
	svc, _ := newTestChatService(t)

	// 写入配置文件
	testCfg := &ChatConfig{
		Provider: "openai",
		OpenAIConfig: OpenAIConfig{
			APIKey: "test-key",
			Model:  "gpt-4o-mini",
		},
	}
	data, _ := json.MarshalIndent(testCfg, "", "  ")
	err := os.WriteFile(svc.configFile, data, 0644)
	assert.NoError(t, err)

	cfg, err := svc.loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "openai", cfg.Provider)
	assert.Equal(t, "test-key", cfg.OpenAIConfig.APIKey)
	assert.Equal(t, "gpt-4o-mini", cfg.OpenAIConfig.Model)
}

// TestSaveConfig_WritesFile 保存后文件可读
func TestSaveConfig_WritesFile(t *testing.T) {
	svc, _ := newTestChatService(t)

	cfg := &ChatConfig{
		Provider: "anthropic",
		AnthropicConfig: AnthropicConfig{
			APIKey: "sk-test",
			Model:  "glm-5.1",
		},
	}
	err := svc.saveConfig(cfg)
	assert.NoError(t, err)

	// 验证文件内容
	data, err := os.ReadFile(svc.configFile)
	assert.NoError(t, err)

	var loaded ChatConfig
	err = json.Unmarshal(data, &loaded)
	assert.NoError(t, err)
	assert.Equal(t, "glm-5.1", loaded.AnthropicConfig.Model)
}

// TestApplyClaudeCodeConfig_AuthToken 映射 API Key
func TestApplyClaudeCodeConfig_AuthToken(t *testing.T) {
	cfg := &ChatConfig{}
	cs := &settings.ClaudeCodeSettings{
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "test-token",
		},
	}
	applyClaudeCodeConfig(cs, cfg)
	assert.Equal(t, "test-token", cfg.AnthropicConfig.APIKey)
}

// TestApplyClaudeCodeConfig_BaseURL 映射 Base URL
func TestApplyClaudeCodeConfig_BaseURL(t *testing.T) {
	cfg := &ChatConfig{}
	cs := &settings.ClaudeCodeSettings{
		Env: map[string]string{
			"ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
		},
	}
	applyClaudeCodeConfig(cs, cfg)
	assert.Equal(t, "https://open.bigmodel.cn/api/anthropic", cfg.AnthropicConfig.BaseURL)
}

// TestApplyClaudeCodeConfig_ModelPriority Opus > Sonnet > Haiku 优先级
func TestApplyClaudeCodeConfig_ModelPriority(t *testing.T) {
	// 仅 Opus 配置时使用 Opus
	t.Run("OnlyOpus", func(t *testing.T) {
		cfg := &ChatConfig{}
		cs := &settings.ClaudeCodeSettings{
			Env: map[string]string{
				"ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1",
			},
		}
		applyClaudeCodeConfig(cs, cfg)
		assert.Equal(t, "glm-5.1", cfg.AnthropicConfig.Model)
	})

	// Opus 优先于 Sonnet
	t.Run("OpusOverSonnet", func(t *testing.T) {
		cfg := &ChatConfig{}
		cs := &settings.ClaudeCodeSettings{
			Env: map[string]string{
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5.1",
			},
		}
		applyClaudeCodeConfig(cs, cfg)
		assert.Equal(t, "glm-5.1", cfg.AnthropicConfig.Model)
	})

	// 全部配置时 Opus 优先
	t.Run("AllModels", func(t *testing.T) {
		cfg := &ChatConfig{}
		cs := &settings.ClaudeCodeSettings{
			Env: map[string]string{
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5.1",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4-flash",
			},
		}
		applyClaudeCodeConfig(cs, cfg)
		assert.Equal(t, "glm-5.1", cfg.AnthropicConfig.Model)
	})
}

// TestApplyClaudeCodeConfig_NilSettings nil 时不 panic
func TestApplyClaudeCodeConfig_NilSettings(t *testing.T) {
	cfg := &ChatConfig{Provider: "openai"}
	applyClaudeCodeConfig(nil, cfg)
	assert.Equal(t, "openai", cfg.Provider) // 不被修改
}

// TestApplyClaudeCodeConfig_NilEnv Env 为 nil 时不 panic
func TestApplyClaudeCodeConfig_NilEnv(t *testing.T) {
	cfg := &ChatConfig{Provider: "openai"}
	cs := &settings.ClaudeCodeSettings{Env: nil}
	applyClaudeCodeConfig(cs, cfg)
	assert.Equal(t, "openai", cfg.Provider)
}

// TestApplyClaudeCodeConfig_SetsDefaultProvider Provider 为空时设为 anthropic
func TestApplyClaudeCodeConfig_SetsDefaultProvider(t *testing.T) {
	cfg := &ChatConfig{}
	cs := &settings.ClaudeCodeSettings{Env: map[string]string{}}
	applyClaudeCodeConfig(cs, cfg)
	assert.Equal(t, "anthropic", cfg.Provider)
}
