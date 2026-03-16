package llm

import "errors"

// ChunkType 流式数据块类型
type ChunkType string

const (
	ChunkTypeContent      ChunkType = "content"       // 内容块
	ChunkTypeToolUse      ChunkType = "tool_use"      // 工具调用
	ChunkTypeError        ChunkType = "error"         // 错误
	ChunkTypeDone         ChunkType = "done"          // 完成
	ChunkTypeMessageStart ChunkType = "message_start" // 消息开始 (Anthropic)
	ChunkTypeMessageDelta ChunkType = "message_delta" // 消息增量 (Anthropic)
	ChunkTypeMessageStop  ChunkType = "message_stop"  // 消息结束 (Anthropic)
)

// StreamChunk 流式响应数据块
type StreamChunk struct {
	// Type 数据块类型
	Type ChunkType `json:"type"`

	// Content 文本内容增量
	Content string `json:"content,omitempty"`

	// ToolCalls 工具调用增量
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`

	// Done 是否完成
	Done bool `json:"done,omitempty"`

	// Error 错误信息
	Error error `json:"error,omitempty"`

	// Response 完整响应 (仅在 Done 时有效)
	Response *ChatResponse `json:"response,omitempty"`

	// Usage token 使用统计 (在消息结束时提供)
	Usage *Usage `json:"usage,omitempty"`

	// Delta 增量数据 (用于 Anthropic)
	Delta interface{} `json:"delta,omitempty"`

	// Index 内容块索引 (用于 Anthropic)
	Index int `json:"index,omitempty"`
}

// IsError 检查是否为错误块
func (c *StreamChunk) IsError() bool {
	return c.Type == ChunkTypeError || c.Error != nil
}

// IsDone 检查是否为完成块
func (c *StreamChunk) IsDone() bool {
	return c.Type == ChunkTypeDone || c.Done
}

// NewContentChunk 创建内容块
func NewContentChunk(content string) *StreamChunk {
	return &StreamChunk{
		Type:    ChunkTypeContent,
		Content: content,
	}
}

// NewToolUseChunk 创建工具调用块
func NewToolUseChunk(toolCalls []*ToolCall) *StreamChunk {
	return &StreamChunk{
		Type:      ChunkTypeToolUse,
		ToolCalls: toolCalls,
	}
}

// NewErrorChunk 创建错误块
func NewErrorChunk(err error) *StreamChunk {
	return &StreamChunk{
		Type:  ChunkTypeError,
		Error: err,
	}
}

// NewDoneChunk 创建完成块
func NewDoneChunk(response *ChatResponse) *StreamChunk {
	return &StreamChunk{
		Type:     ChunkTypeDone,
		Done:     true,
		Response: response,
	}
}

// StreamError 流式处理错误
type StreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *StreamError) Error() string {
	return e.Message
}

// NewStreamError 创建流式错误
func NewStreamError(code, message string) error {
	return &StreamError{
		Code:    code,
		Message: message,
	}
}

// IsStreamError 检查是否为流式错误
func IsStreamError(err error) bool {
	var streamErr *StreamError
	return errors.As(err, &streamErr)
}
