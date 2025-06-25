package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"Time Demo 🚀",
		"1.0.0",
	)
	// Add tool
	tool := mcp.NewTool("current time",
		mcp.WithDescription("Get current time with timezone, Asia/Shanghai is default"),
		mcp.WithString("timezone",
			mcp.Required(),
			mcp.Description("current time timezone"),
		),
	)
	// Add tool handler
	s.AddTool(tool, currentTimeHandler)
	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func currentTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 将Arguments转换为map[string]any再进行索引
	argsMap, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	timezone, ok := argsMap["timezone"].(string)
	if !ok {
		return mcp.NewToolResultError("timezone must be a string"), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse timezone with error: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf(`current time is %s`, time.Now().In(loc))), nil
}
