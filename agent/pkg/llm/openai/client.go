package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

const (
	// DefaultBaseURL OpenAI API 默认地址
	DefaultBaseURL = "https://api.openai.com/v1"
	// DefaultModel 默认模型
	DefaultModel = "gpt-4o"
	// DefaultMaxTokens 默认最大 token
	DefaultMaxTokens = 4096
	// DefaultTimeout 默认超时
	DefaultTimeout = 60 * time.Second
)

// Config OpenAI 客户端配置
type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Option 配置选项函数
type Option func(*Config)

// WithAPIKey 设置 API Key
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithBaseURL 设置 API 地址
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithModel 设置模型
func WithModel(model string) Option {
	return func(c *Config) {
		c.Model = model
	}
}

// WithMaxTokens 设置最大 token
func WithMaxTokens(tokens int) Option {
	return func(c *Config) {
		c.MaxTokens = tokens
	}
}

// WithTimeout 设置超时
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithHTTPClient 设置 HTTP 客户端
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// Client OpenAI 客户端
type Client struct {
	config *Config
	client *http.Client
}

// NewClient 创建 OpenAI 客户端
func NewClient(opts ...Option) (*Client, error) {
	config := &Config{
		BaseURL:   DefaultBaseURL,
		Model:     DefaultModel,
		MaxTokens: DefaultMaxTokens,
		Timeout:   DefaultTimeout,
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API Key 不能为空")
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &Client{
		config: config,
		client: client,
	}, nil
}

// ModelName 返回模型名称
func (c *Client) ModelName() string {
	return c.config.Model
}

// ProviderName 返回提供商名称
func (c *Client) ProviderName() string {
	return "openai"
}

// Complete 发送非流式请求
func (c *Client) Complete(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	// 填充默认值
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
	}
	req.Stream = false

	// 转换请求
	oaiReq := convertRequest(req)

	// 发送请求
	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查错误
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != nil {
			return nil, convertError(errResp.Error, resp.StatusCode)
		}
		return nil, fmt.Errorf("请求失败: %s (状态码: %d)", string(respBody), resp.StatusCode)
	}

	// 解析响应
	var oaiResp ChatResponse
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return convertResponse(&oaiResp), nil
}

// Stream 发送流式请求
func (c *Client) Stream(ctx context.Context, req *llm.ChatRequest) (<-chan *llm.StreamChunk, error) {
	// 填充默认值
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
	}
	req.Stream = true

	// 转换请求
	oaiReq := convertRequest(req)

	// 发送请求
	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != nil {
			return nil, convertError(errResp.Error, resp.StatusCode)
		}
		return nil, fmt.Errorf("请求失败: %s (状态码: %d)", string(respBody), resp.StatusCode)
	}

	// 创建输出通道
	ch := make(chan *llm.StreamChunk, 100)

	// 启动 goroutine 处理 SSE 流
	go c.processStream(resp, ch)

	return ch, nil
}

// processStream 处理 SSE 流
func (c *Client) processStream(resp *http.Response, ch chan<- *llm.StreamChunk) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var accumulatedContent string
	var accumulatedToolCalls []*llm.ToolCall

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否为数据行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// 检查结束标记
		if data == "[DONE]" {
			// 发送最终响应
			finalResp := &llm.ChatResponse{
				ID:         "",
				Model:      c.config.Model,
				Content:    accumulatedContent,
				Role:       llm.RoleAssistant,
				ToolCalls:  accumulatedToolCalls,
				StopReason: llm.StopReasonEndTurn,
			}
			ch <- llm.NewDoneChunk(finalResp)
			return
		}

		// 解析 JSON
		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			ch <- llm.NewErrorChunk(fmt.Errorf("解析流数据失败: %w", err))
			return
		}

		// 处理响应
		if len(streamResp.Choices) > 0 {
			choice := streamResp.Choices[0]

			// 处理内容增量
			if choice.Delta != nil && choice.Delta.Content != "" {
				accumulatedContent += choice.Delta.Content
				ch <- llm.NewContentChunk(choice.Delta.Content)
			}

			// 处理工具调用增量
			if choice.Delta != nil && len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					// 查找或创建工具调用
					found := false
					for i := range accumulatedToolCalls {
						if accumulatedToolCalls[i].ID == tc.ID {
							accumulatedToolCalls[i].Function.Arguments += tc.Function.Arguments
							found = true
							break
						}
					}
					if !found {
						accumulatedToolCalls = append(accumulatedToolCalls, &llm.ToolCall{
							ID:   tc.ID,
							Type: tc.Type,
							Function: &llm.FunctionCall{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						})
					}
				}
			}

			// 处理完成
			if choice.FinishReason != "" {
				finalResp := &llm.ChatResponse{
					ID:         streamResp.ID,
					Model:      streamResp.Model,
					Content:    accumulatedContent,
					Role:       llm.RoleAssistant,
					ToolCalls:  accumulatedToolCalls,
					StopReason: convertFinishReason(choice.FinishReason),
				}
				if streamResp.Usage != nil {
					finalResp.Usage = &llm.Usage{
						InputTokens:  streamResp.Usage.PromptTokens,
						OutputTokens: streamResp.Usage.CompletionTokens,
						TotalTokens:  streamResp.Usage.TotalTokens,
					}
				}
				ch <- llm.NewDoneChunk(finalResp)
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- llm.NewErrorChunk(fmt.Errorf("读取流失败: %w", err))
	}
}

// convertError 转换错误
func convertError(err *ErrorDetail, statusCode int) *llm.LLMError {
	var typ llm.ErrorType
	switch err.Type {
	case "invalid_request_error":
		typ = llm.ErrorTypeInvalidRequest
	case "authentication_error":
		typ = llm.ErrorTypeAuthentication
	case "permission_error":
		typ = llm.ErrorTypePermission
	case "not_found_error":
		typ = llm.ErrorTypeNotFound
	case "rate_limit_error":
		typ = llm.ErrorTypeRateLimit
	case "server_error":
		typ = llm.ErrorTypeServerError
	default:
		typ = llm.ErrorTypeServerError
	}

	return llm.NewLLMError(typ, err.Message).
		WithCode(err.Code).
		WithStatusCode(statusCode).
		WithProvider("openai")
}
