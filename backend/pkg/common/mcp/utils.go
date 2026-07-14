package mcp

import (
	"fmt"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TextResult 创建文本结果
func TextResult(text string) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: text},
		},
	}
}

// ErrorResultFromError 创建错误结果（从 error 类型）
func ErrorResultFromError(err error) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		IsError: true,
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: fmt.Sprintf("错误: %v", err)},
		},
	}
}

// ErrorResult 创建错误结果（从字符串）
func ErrorResult(errMsg string) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		IsError: true,
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: errMsg},
		},
	}
}
