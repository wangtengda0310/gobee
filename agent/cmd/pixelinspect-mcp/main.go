package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Find pixelinspect executable in PATH or relative path
func findPixelInspect() string {
	// Try relative path first (for development)
	relativePath := "../../../lvan/cmd/pixelinspect/pixelinspect.exe"
	if absPath, err := filepath.Abs(relativePath); err == nil {
		if _, err := exec.LookPath(absPath); err == nil {
			return absPath
		}
	}
	// Try to find in PATH
	if path, err := exec.LookPath("pixelinspect.exe"); err == nil {
		return path
	}
	// Fallback to default
	return "pixelinspect.exe"
}

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"PixelInspect 🎨",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool: inspect_pixel
	inspectTool := mcp.NewTool("inspect_pixel",
		mcp.WithDescription("Inspect pixel values in an image. Supports single point, rectangle, multiple points, or all pixels. Returns color information including RGB, HSL, and hex values."),
		mcp.WithString("image_path",
			mcp.Required(),
			mcp.Description("Path to the input image file"),
		),
		mcp.WithString("region",
			mcp.Description("Region to inspect as JSON string. Examples: '{\"x\":100,\"y\":200}' for single point, '{\"x\":0,\"y\":0,\"width\":100,\"height\":100}' for rectangle, '[{\"x\":0,\"y\":0},{\"x\":100,\"y\":200}]' for multiple points, or 'all' for entire image. Default: 'all'"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: 'list' (readable text), 'stats' (region statistics), 'json' (JSON data). Default: 'list'"),
			mcp.Enum("list", "stats", "json"),
		),
		mcp.WithString("output_file",
			mcp.Description("Optional output file path. If not specified, output is returned as text"),
		),
	)

	// Add tool handler
	s.AddTool(inspectTool, inspectPixelHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func inspectPixelHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	// Parse required parameters
	imagePath, ok := argsMap["image_path"].(string)
	if !ok || imagePath == "" {
		return mcp.NewToolResultError("image_path is required and must be a string"), nil
	}

	// Build command arguments
	args := []string{imagePath}

	// Region (optional, defaults to "all")
	if region, ok := argsMap["region"].(string); ok && region != "" {
		args = append(args, "--region", region)
	} else {
		args = append(args, "--region", "all")
	}

	// Format (optional, defaults to "list")
	if format, ok := argsMap["format"].(string); ok && format != "" {
		args = append(args, "--format", format)
	}

	// Output file (optional)
	if outputFile, ok := argsMap["output_file"].(string); ok && outputFile != "" {
		args = append(args, "--output", outputFile)
	}

	// Execute command
	pixelinspectPath := findPixelInspect()
	cmd := exec.CommandContext(ctx, pixelinspectPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to execute pixelinspect: %v\nOutput: %s", err, string(output))), nil
	}

	resultText := string(output)

	// If output file was specified, include file info
	if outputFile, ok := argsMap["output_file"].(string); ok && outputFile != "" {
		if _, err := filepath.Abs(outputFile); err == nil {
			resultText += fmt.Sprintf("\n\nOutput saved to: %s\n", outputFile)
		}
	}

	return mcp.NewToolResultText(resultText), nil
}
