package llm

// ChatRequest 表示聊天补全请求
type ChatRequest struct {
	// Model 指定使用的模型
	Model string `json:"model"`

	// Messages 对话消息列表
	Messages []*Message `json:"messages"`

	// MaxTokens 最大生成 token 数
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature 采样温度 (0-2)
	Temperature float64 `json:"temperature,omitempty"`

	// TopP 核采样参数
	TopP float64 `json:"top_p,omitempty"`

	// Tools 可用工具列表
	Tools []*Tool `json:"tools,omitempty"`

	// Stream 是否启用流式响应
	Stream bool `json:"stream,omitempty"`

	// System 系统提示词 (可选，某些提供商在顶层使用)
	System string `json:"system,omitempty"`

	// Stop 停止生成的序列
	Stop []string `json:"stop,omitempty"`
}

// ChatResponse 表示聊天补全响应
type ChatResponse struct {
	// ID 响应唯一标识
	ID string `json:"id"`

	// Model 使用的模型
	Model string `json:"model"`

	// Content 生成的文本内容
	Content string `json:"content"`

	// Role 响应角色 (通常为 "assistant")
	Role Role `json:"role"`

	// ToolCalls 工具调用列表
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`

	// Usage token 使用统计
	Usage *Usage `json:"usage,omitempty"`

	// StopReason 停止原因
	StopReason StopReason `json:"stop_reason,omitempty"`
}

// Usage token 使用统计
type Usage struct {
	// InputTokens 输入 token 数
	InputTokens int `json:"input_tokens"`

	// OutputTokens 输出 token 数
	OutputTokens int `json:"output_tokens"`

	// TotalTokens 总 token 数
	TotalTokens int `json:"total_tokens"`
}

// StopReason 表示响应停止的原因
type StopReason string

const (
	StopReasonEndTurn   StopReason = "end_turn"      // 正常结束
	StopReasonMaxTokens StopReason = "max_tokens"    // 达到最大 token
	StopReasonToolUse   StopReason = "tool_use"      // 工具调用
	StopReasonStopSeq   StopReason = "stop_sequence" // 遇到停止序列
)
