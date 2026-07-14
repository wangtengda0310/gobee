package settings

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TextResult 创建文本结果（别名）
func TextResult(text string) *mcpgo.CallToolResult {
	return mcp.TextResult(text)
}

// ErrorResult 创建错误结果（别名）
func ErrorResult(err any) *mcpgo.CallToolResult {
	switch v := err.(type) {
	case error:
		return mcp.ErrorResultFromError(v)
	case string:
		return mcp.ErrorResult(v)
	default:
		return mcp.ErrorResult("未知错误")
	}
}
