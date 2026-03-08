package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RunMCPServer starts the MCP server
func RunMCPServer(client *Client, transport string) error {
	// Create MCP server
	s := server.NewMCPServer(
		"openclaw-channel",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Add chat tool
	chatTool := mcp.NewTool("openclaw_chat",
		mcp.WithDescription("Send a message to OpenClaw and get a response. Use this to interact with the AI assistant."),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("The message to send to OpenClaw"),
		),
		mcp.WithString("session",
			mcp.Description("Optional session key (default: agent:main:main)"),
		),
	)

	s.AddTool(chatTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		message, ok := args["message"].(string)
		if !ok || message == "" {
			return mcp.NewToolResultError("message is required"), nil
		}

		session := "agent:main:main"
		if s, ok := args["session"].(string); ok && s != "" {
			session = s
		}

		// Ensure connection
		if !client.IsConnected() {
			if err := client.Connect(); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to connect: %v", err)), nil
			}
		}

		// Send message
		response, err := client.Chat(session, message)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get response: %v", err)), nil
		}

		return mcp.NewToolResultText(response), nil
	})

	// Add status tool
	statusTool := mcp.NewTool("openclaw_status",
		mcp.WithDescription("Get the current status of the OpenClaw gateway. Use this to check connectivity and session information."),
	)

	s.AddTool(statusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Ensure connection
		if !client.IsConnected() {
			if err := client.Connect(); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to connect: %v", err)), nil
			}
		}

		status, err := client.GetStatus()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get status: %v", err)), nil
		}

		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to marshal status"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})

	// Add history tool
	historyTool := mcp.NewTool("openclaw_history",
		mcp.WithDescription("Get chat history from an OpenClaw session. Use this to see recent messages."),
		mcp.WithString("session",
			mcp.Description("Optional session key (default: agent:main:main)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of messages to retrieve (default: 10)"),
		),
	)

	s.AddTool(historyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			args = make(map[string]interface{})
		}

		session := "agent:main:main"
		if s, ok := args["session"].(string); ok && s != "" {
			session = s
		}

		limit := 10
		if l, ok := args["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}

		// Ensure connection
		if !client.IsConnected() {
			if err := client.Connect(); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to connect: %v", err)), nil
			}
		}

		history, err := client.GetHistory(session, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get history: %v", err)), nil
		}

		data, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			return mcp.NewToolResultError("failed to marshal history"), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	})

	// Start server based on transport
	switch transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			return fmt.Errorf("MCP server error: %w", err)
		}
	case "sse":
		// SSE transport would need additional setup
		return fmt.Errorf("SSE transport not yet implemented, use stdio")
	default:
		return fmt.Errorf("unknown transport: %s (use stdio or sse)", transport)
	}

	return nil
}
