package home

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetModelName_Anthropic 返回配置的 Anthropic 模型名
func TestGetModelName_Anthropic(t *testing.T) {
	svc, _ := newTestChatService(t)
	svc.config = &ChatConfig{
		Provider:        "anthropic",
		AnthropicConfig: AnthropicConfig{Model: "glm-5.1"},
	}
	assert.Equal(t, "glm-5.1", svc.getModelName())
}

// TestGetModelName_AnthropicDefault 模型名为空时返回默认值
func TestGetModelName_AnthropicDefault(t *testing.T) {
	svc, _ := newTestChatService(t)
	svc.config = &ChatConfig{Provider: "anthropic"}
	assert.Equal(t, "claude-3-5-sonnet-20241022", svc.getModelName())
}

// TestGetModelName_OpenAI 返回配置的 OpenAI 模型名
func TestGetModelName_OpenAI(t *testing.T) {
	svc, _ := newTestChatService(t)
	svc.config = &ChatConfig{
		Provider:     "openai",
		OpenAIConfig: OpenAIConfig{Model: "gpt-4o-mini"},
	}
	assert.Equal(t, "gpt-4o-mini", svc.getModelName())
}

// TestGetModelName_UnknownProvider 未知 provider 返回空字符串
func TestGetModelName_UnknownProvider(t *testing.T) {
	svc, _ := newTestChatService(t)
	svc.config = &ChatConfig{Provider: "unknown"}
	assert.Equal(t, "", svc.getModelName())
}

// TestGetModelName_NilConfig config 为 nil 返回空字符串
func TestGetModelName_NilConfig(t *testing.T) {
	svc, _ := newTestChatService(t)
	svc.config = nil
	assert.Equal(t, "", svc.getModelName())
}

// TestCreateClient_NilConfig config 为 nil 返回 error
func TestCreateClient_NilConfig(t *testing.T) {
	svc, _ := newTestChatService(t)
	svc.config = nil
	_, err := svc.createClient()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "配置未初始化")
}
