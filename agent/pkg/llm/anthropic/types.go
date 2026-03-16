package anthropic

import "time"

// Anthropic API 请求和响应类型定义

const (
	// DefaultBaseURL Anthropic API 默认地址
	DefaultBaseURL = "https://api.anthropic.com/v1"
	// DefaultModel 默认模型
	DefaultModel = "claude-sonnet-4-20250514"
	// DefaultMaxTokens 默认最大 token (Anthropic 要求必填)
	DefaultMaxTokens = 4096
	// DefaultVersion API 版本
	DefaultVersion = "2023-06-01"
	// DefaultTimeout 默认超时
	DefaultTimeout = 60 * time.Second
)

// SSE 事件类型常量
const (
	EventTypeMessageStart      = "message_start"
	EventTypeContentBlockStart = "content_block_start"
	EventTypeContentBlockDelta = "content_block_delta"
	EventTypeContentBlockStop  = "content_block_stop"
	EventTypeMessageDelta      = "message_delta"
	EventTypeMessageStop       = "message_stop"
	EventTypePing              = "ping"
	EventTypeError             = "error"
)

// ChatRequest Anthropic Messages API 请求
type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`

	// 可选参数
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"top_p,omitempty"`
	TopK          int      `json:"top_k,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
	Stream        bool     `json:"stream,omitempty"`
	Tools         []Tool   `json:"tools,omitempty"`
}

// Message Anthropic 消息格式
type Message struct {
	Role    string      `json:"role"`    // "user" 或 "assistant"
	Content interface{} `json:"content"` // ContentBlock 数组
}

// ContentBlock 内容块
type ContentBlock struct {
	Type string `json:"type"` // "text", "image", "tool_use", "tool_result"

	// 文本内容
	Text string `json:"text,omitempty"`

	// 图像内容
	Source *ImageSource `json:"source,omitempty"`

	// 工具调用
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`

	// 工具结果
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   interface{} `json:"content,omitempty"`
	IsError   bool        `json:"is_error,omitempty"`
}

// ImageSource 图像源
type ImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png", "image/gif", "image/webp"
	Data      string `json:"data"`
}

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ChatResponse Anthropic Messages API 响应
type ChatResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// Usage token 使用统计
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type         string        `json:"type"`
	Index        int           `json:"index,omitempty"`
	Delta        *StreamDelta  `json:"delta,omitempty"`
	Message      *ChatResponse `json:"message,omitempty"`
	ContentBlock *ContentBlock `json:"content_block,omitempty"`
	Error        *ErrorDetail  `json:"error,omitempty"`
}

// StreamDelta 流式增量
type StreamDelta struct {
	Type        string `json:"type,omitempty"` // "text_delta", "input_json_delta"
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
	Usage       *Usage `json:"usage,omitempty"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
}
