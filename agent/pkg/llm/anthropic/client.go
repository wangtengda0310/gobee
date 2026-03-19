package anthropic

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

// Config Anthropic 客户端配置
type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
	Version    string
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

// WithVersion 设置 API 版本
func WithVersion(version string) Option {
	return func(c *Config) {
		c.Version = version
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

// Client Anthropic 客户端
type Client struct {
	config *Config
	client *http.Client
}

// NewClient 创建 Anthropic 客户端
func NewClient(opts ...Option) (*Client, error) {
	config := &Config{
		BaseURL:   DefaultBaseURL,
		Model:     DefaultModel,
		MaxTokens: DefaultMaxTokens,
		Version:   DefaultVersion,
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
	return "anthropic"
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
	anthReq := convertRequest(req)

	// 发送请求
	body, err := json.Marshal(anthReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", c.config.Version)

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
	var anthResp ChatResponse
	if err := json.Unmarshal(respBody, &anthResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return convertResponse(&anthResp), nil
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
	anthReq := convertRequest(req)

	// 发送请求
	body, err := json.Marshal(anthReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", c.config.Version)
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

	var (
		messageID      string
		model          string
		accumulated    string
		toolCalls      []*llm.ToolCall
		currentToolIdx = -1
		inputUsage     int
		outputUsage    int
	)

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否为事件类型行
		if strings.HasPrefix(line, "event: ") {
			// 事件类型在下一行的 data 中
			continue
		}

		// 检查是否为数据行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// 解析事件
		var event StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			ch <- llm.NewErrorChunk(fmt.Errorf("解析流数据失败: %w", err))
			return
		}

		// 处理不同类型的事件
		switch event.Type {
		case EventTypeMessageStart:
			if event.Message != nil {
				messageID = event.Message.ID
				model = event.Message.Model
				inputUsage = event.Message.Usage.InputTokens
			}

		case EventTypeContentBlockStart:
			// 内容块开始
			if event.ContentBlock != nil {
				if event.ContentBlock.Type == "tool_use" {
					// 开始新的工具调用
					toolCalls = append(toolCalls, &llm.ToolCall{
						ID:   event.ContentBlock.ID,
						Type: "function",
						Function: &llm.FunctionCall{
							Name: event.ContentBlock.Name,
						},
					})
					currentToolIdx = len(toolCalls) - 1
				}
			}

		case EventTypeContentBlockDelta:
			if event.Delta != nil {
				switch event.Delta.Type {
				case "text_delta":
					// 文本增量
					accumulated += event.Delta.Text
					ch <- llm.NewContentChunk(event.Delta.Text)

				case "input_json_delta":
					// 工具输入增量
					if currentToolIdx >= 0 && currentToolIdx < len(toolCalls) {
						toolCalls[currentToolIdx].Function.Arguments += event.Delta.PartialJSON
					}
				}
			}

		case EventTypeContentBlockStop:
			// 内容块结束
			currentToolIdx = -1

		case EventTypeMessageDelta:
			// 消息增量，包含 stop_reason 和输出 token
			if event.Delta != nil {
				if event.Delta.StopReason != "" {
					// 构建最终响应
					finalResp := &llm.ChatResponse{
						ID:         messageID,
						Model:      model,
						Content:    accumulated,
						Role:       llm.RoleAssistant,
						ToolCalls:  toolCalls,
						StopReason: convertStopReason(event.Delta.StopReason),
						Usage: &llm.Usage{
							InputTokens:  inputUsage,
							OutputTokens: outputUsage,
							TotalTokens:  inputUsage + outputUsage,
						},
					}
					if event.Delta.Usage != nil {
						outputUsage = event.Delta.Usage.OutputTokens
						finalResp.Usage.OutputTokens = outputUsage
						finalResp.Usage.TotalTokens = inputUsage + outputUsage
					}
					ch <- llm.NewDoneChunk(finalResp)
				}
			}

		case EventTypeMessageStop:
			// 消息结束
			return

		case EventTypePing:
			// 心跳，忽略

		case EventTypeError:
			// 错误
			if event.Error != nil {
				ch <- llm.NewErrorChunk(convertError(event.Error, 0))
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
	case "api_error":
		typ = llm.ErrorTypeServerError
	case "overloaded_error":
		typ = llm.ErrorTypeOverloaded
	default:
		typ = llm.ErrorTypeServerError
	}

	return llm.NewLLMError(typ, err.Message).
		WithStatusCode(statusCode).
		WithProvider("anthropic")
}
