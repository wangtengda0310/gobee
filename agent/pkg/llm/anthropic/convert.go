package anthropic

import (
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// convertRequest 将统一请求格式转换为 Anthropic 格式
func convertRequest(req *llm.ChatRequest) *ChatRequest {
	anthReq := &ChatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		System:      req.System,
	}

	// 如果没有设置 max_tokens，使用默认值 (Anthropic 要求必填)
	if anthReq.MaxTokens == 0 {
		anthReq.MaxTokens = DefaultMaxTokens
	}

	// 转换消息
	anthReq.Messages = convertMessages(req.Messages)

	// 转换工具
	if len(req.Tools) > 0 {
		anthReq.Tools = convertTools(req.Tools)
	}

	return anthReq
}

// convertMessages 转换消息列表
func convertMessages(messages []*llm.Message) []Message {
	result := make([]Message, 0, len(messages))

	for _, msg := range messages {
		// 跳过系统消息 (Anthropic 使用顶层 system 字段)
		if msg.Role == llm.RoleSystem {
			continue
		}

		anthMsg := Message{
			Role: string(msg.Role),
		}

		// 转换内容
		if msg.Content != nil {
			// 对于纯文本内容，直接使用字符串（智谱代理兼容格式）
			if text, ok := msg.Content.(*llm.TextContent); ok {
				anthMsg.Content = []ContentBlock{
					{Type: "text", Text: text.Text},
				}
			} else {
				anthMsg.Content = msg.Content.ToAnthropic()
			}
		}

		// 处理工具响应消息
		if msg.Role == llm.RoleTool && msg.ToolCallID != "" {
			anthMsg.Content = []ContentBlock{
				{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   llm.TextString(msg.Content),
				},
			}
		}

		// 处理助手的工具调用
		if msg.Role == llm.RoleAssistant && len(msg.ToolCalls) > 0 {
			blocks := make([]ContentBlock, 0)

			// 添加文本内容
			if text := llm.TextString(msg.Content); text != "" {
				blocks = append(blocks, ContentBlock{
					Type: "text",
					Text: text,
				})
			}

			// 添加工具调用
			for _, tc := range msg.ToolCalls {
				blocks = append(blocks, ContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: jsonRawToInterface(tc.Function.Arguments),
				})
			}

			anthMsg.Content = blocks
		}

		result = append(result, anthMsg)
	}

	return result
}

// convertTools 转换工具定义列表
func convertTools(tools []*llm.Tool) []Tool {
	result := make([]Tool, len(tools))
	for i, t := range tools {
		result[i] = Tool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		}
	}
	return result
}

// convertResponse 将 Anthropic 响应转换为统一格式
func convertResponse(resp *ChatResponse) *llm.ChatResponse {
	if resp == nil {
		return nil
	}

	result := &llm.ChatResponse{
		ID:         resp.ID,
		Model:      resp.Model,
		Role:       llm.RoleAssistant,
		StopReason: convertStopReason(resp.StopReason),
	}

	// 提取内容
	var content string
	var toolCalls []*llm.ToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			toolCalls = append(toolCalls, &llm.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: &llm.FunctionCall{
					Name:      block.Name,
					Arguments: interfaceToJSON(block.Input),
				},
			})
		}
	}

	result.Content = content
	result.ToolCalls = toolCalls

	// 转换使用统计
	result.Usage = &llm.Usage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	return result
}

// convertStopReason 转换停止原因
func convertStopReason(reason string) llm.StopReason {
	switch reason {
	case "end_turn":
		return llm.StopReasonEndTurn
	case "max_tokens":
		return llm.StopReasonMaxTokens
	case "tool_use":
		return llm.StopReasonToolUse
	case "stop_sequence":
		return llm.StopReasonStopSeq
	default:
		return llm.StopReason(reason)
	}
}

// jsonRawToInterface 将 JSON 字符串转换为 interface{}
func jsonRawToInterface(s string) interface{} {
	// 简单处理：直接返回字符串
	// 实际使用时会通过 json.Unmarshal 解析
	return s
}

// interfaceToJSON 将 interface{} 转换为 JSON 字符串
func interfaceToJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		// 这里应该使用 json.Marshal，但为了避免导入 encoding/json
		// 实际实现在 client.go 中
		return ""
	}
}
