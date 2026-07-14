package home

import (
	"context"
	"time"

	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	prototest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
)

// ChatConfig 聊天配置
type ChatConfig struct {
	Provider        string          `json:"provider"` // "anthropic" | "openai"
	AnthropicConfig AnthropicConfig `json:"anthropicConfig"`
	OpenAIConfig    OpenAIConfig    `json:"openaiConfig"`
	SystemPrompt    string          `json:"systemPrompt"`
}

// AnthropicConfig Anthropic 配置
type AnthropicConfig struct {
	APIKey    string `json:"apiKey"`
	BaseURL   string `json:"baseUrl"`   // 可选，支持代理
	Model     string `json:"model"`     // 默认 "claude-3-5-sonnet-20241022"
	MaxTokens int    `json:"maxTokens"` // 默认 4096
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey    string `json:"apiKey"`
	BaseURL   string `json:"baseUrl"`   // 可选，支持自定义 API
	Model     string `json:"model"`     // 默认 "gpt-4o"
	MaxTokens int    `json:"maxTokens"` // 默认 4096
}

// ChatMessage 对话消息（用于前端交互和 JSON 序列化）
type ChatMessage struct {
	Role      string `json:"role"` // "user" | "assistant"
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"` // Unix 时间戳
}

// ChatSession 对话会话（用于前端交互和 JSON 序列化）
type ChatSession struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt int64         `json:"createdAt"`
	UpdatedAt int64         `json:"updatedAt"`
}

// ChatServices 聊天服务依赖的其他服务容器
type ChatServices struct {
	ExcelCheckService    *exceltest.ExcelCheckService
	RecordControlService *prototest.RecordControlService
	TestCaseService      *prototest.TestCaseService
	RecordFileService    *prototest.RecordFileService
}

// memoryAdapter 官方 FileMemory 的适配器
// 提供与旧 API 兼容的方法
type memoryAdapter struct {
	*memory.FileMemory
}

// GetMessages 获取所有消息（用于前端兼容）
// 将 llm.Message 转换为 ChatMessage 格式
func (m *memoryAdapter) GetMessages() []ChatMessage {
	ctx := context.Background()
	messages, err := m.FileMemory.GetContext(ctx)
	if err != nil {
		return nil
	}

	result := make([]ChatMessage, 0, len(messages))
	for _, msg := range messages {
		cm := ChatMessage{
			Timestamp: time.Now().Unix(),
		}
		switch msg.Role {
		case llm.RoleUser:
			cm.Role = "user"
		case llm.RoleAssistant:
			cm.Role = "assistant"
		case llm.RoleSystem:
			cm.Role = "system"
		}
		cm.Content = llm.TextString(msg.Content)
		result = append(result, cm)
	}

	return result
}
