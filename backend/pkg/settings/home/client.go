package home

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/llm/anthropic"
	"github.com/wangtengda0310/gobee/agent/pkg/llm/openai"
)

// createClient 根据当前配置创建 LLM 客户端
func (s *ChatService) createClient() (llm.ChatCompleter, error) {
	if s.config == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	log.Printf("[ChatService] createClient, provider: %s", s.config.Provider)

	switch s.config.Provider {
	case "anthropic":
		apiKey := s.config.AnthropicConfig.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API Key 未配置")
		}

		baseURL := s.config.AnthropicConfig.BaseURL
		if baseURL == "" {
			baseURL = "https://api.anthropic.com/v1"
		}

		model := s.config.AnthropicConfig.Model
		if model == "" {
			model = "claude-3-5-sonnet-20241022"
		}

		maxTokens := s.config.AnthropicConfig.MaxTokens
		if maxTokens == 0 {
			maxTokens = 4096
		}

		log.Printf("[ChatService] 创建 Anthropic 客户端, baseURL: %s, model: %s", baseURL, model)

		return anthropic.NewClient(
			anthropic.WithAPIKey(apiKey),
			anthropic.WithBaseURL(baseURL),
			anthropic.WithModel(model),
			anthropic.WithMaxTokens(maxTokens),
		)

	case "openai":
		apiKey := s.config.OpenAIConfig.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API Key 未配置")
		}

		baseURL := s.config.OpenAIConfig.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}

		model := s.config.OpenAIConfig.Model
		if model == "" {
			model = "gpt-4o"
		}

		maxTokens := s.config.OpenAIConfig.MaxTokens
		if maxTokens == 0 {
			maxTokens = 4096
		}

		return openai.NewClient(
			openai.WithAPIKey(apiKey),
			openai.WithBaseURL(baseURL),
			openai.WithModel(model),
			openai.WithMaxTokens(maxTokens),
		)

	default:
		return nil, fmt.Errorf("不支持的提供商: %s", s.config.Provider)
	}
}

// buildRequest 构建 LLM 请求（包含历史消息、系统提示和工具定义）
func (s *ChatService) buildRequest() *llm.ChatRequest {
	ctx := context.Background()

	// 从 memory 获取历史消息
	messages, err := s.memory.GetContext(ctx)
	if err != nil {
		messages = []*llm.Message{}
	}

	// 添加系统提示
	systemPrompt := s.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "你是一个有帮助的AI助手。"
	}

	req := &llm.ChatRequest{
		Model:    s.getModelName(),
		Messages: messages,
		System:   systemPrompt,
		Stream:   true,
	}

	// 添加工具定义（如果注册表中有工具）
	if tools := s.registry.GetDefinitions(); len(tools) > 0 {
		req.Tools = tools
	}

	return req
}

// getModelName 返回当前配置的模型名称
func (s *ChatService) getModelName() string {
	if s.config == nil {
		return ""
	}

	switch s.config.Provider {
	case "anthropic":
		if s.config.AnthropicConfig.Model != "" {
			return s.config.AnthropicConfig.Model
		}
		return "claude-3-5-sonnet-20241022"
	case "openai":
		if s.config.OpenAIConfig.Model != "" {
			return s.config.OpenAIConfig.Model
		}
		return "gpt-4o"
	default:
		return ""
	}
}
