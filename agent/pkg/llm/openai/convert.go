package openai

import (
	"agent/pkg/llm"
)

// convertRequest 将统一请求格式转换为 OpenAI 格式
func convertRequest(req *llm.ChatRequest) *ChatRequest {
	oaiReq := &ChatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Stop:        req.Stop,
	}

	// 转换消息
	oaiReq.Messages = convertMessages(req.Messages)

	// 转换工具
	if len(req.Tools) > 0 {
		oaiReq.Tools = convertTools(req.Tools)
	}

	return oaiReq
}

// convertMessages 转换消息列表
func convertMessages(messages []*llm.Message) []Message {
	result := make([]Message, len(messages))
	for i, msg := range messages {
		result[i] = convertMessage(msg)
	}
	return result
}

// convertMessage 转换单条消息
func convertMessage(msg *llm.Message) Message {
	oaiMsg := Message{
		Role:       string(msg.Role),
		ToolCallID: msg.ToolCallID,
		Name:       msg.Name,
	}

	// 转换内容
	if msg.Content != nil {
		oaiMsg.Content = msg.Content.ToOpenAI()
	}

	// 转换工具调用
	if len(msg.ToolCalls) > 0 {
		oaiMsg.ToolCalls = convertToolCalls(msg.ToolCalls)
	}

	return oaiMsg
}

// convertToolCalls 转换工具调用列表
func convertToolCalls(toolCalls []*llm.ToolCall) []ToolCall {
	result := make([]ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		result[i] = ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}

// convertTools 转换工具定义列表
func convertTools(tools []*llm.Tool) []Tool {
	result := make([]Tool, len(tools))
	for i, t := range tools {
		result[i] = Tool{
			Type: t.Type,
			Function: FunctionDef{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		}
	}
	return result
}

// convertResponse 将 OpenAI 响应转换为统一格式
func convertResponse(resp *ChatResponse) *llm.ChatResponse {
	if resp == nil || len(resp.Choices) == 0 {
		return nil
	}

	choice := resp.Choices[0]
	result := &llm.ChatResponse{
		ID:         resp.ID,
		Model:      resp.Model,
		Role:       llm.RoleAssistant,
		StopReason: convertFinishReason(choice.FinishReason),
	}

	// 提取文本内容
	if content, ok := choice.Message.Content.(string); ok {
		result.Content = content
	}

	// 转换工具调用
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]*llm.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = &llm.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: &llm.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	// 转换使用统计
	result.Usage = &llm.Usage{
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		TotalTokens:  resp.Usage.TotalTokens,
	}

	return result
}

// convertFinishReason 转换完成原因
func convertFinishReason(reason string) llm.StopReason {
	switch reason {
	case "stop":
		return llm.StopReasonEndTurn
	case "length":
		return llm.StopReasonMaxTokens
	case "tool_calls", "function_call":
		return llm.StopReasonToolUse
	default:
		return llm.StopReason(reason)
	}
}
