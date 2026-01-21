package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const gridsplitPath = "lvan/cmd/gridsplit/gridsplit.exe"

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"GridSplit 🖼️",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Tool 1: split_image
	splitTool := mcp.NewTool("split_image",
		mcp.WithDescription("Split an image into a grid of smaller images. Supports multiple splitting modes: equal spacing, color detection, and edge detection."),
		mcp.WithString("image_path",
			mcp.Required(),
			mcp.Description("Path to the input image file"),
		),
		mcp.WithString("output_dir",
			mcp.Required(),
			mcp.Description("Directory to save the split images"),
		),
		mcp.WithString("mode",
			mcp.Description("Splitting mode: 'equal' (equal spacing), 'color' (color detection), 'edge' (edge detection). Default: 'equal'"),
			mcp.Enum("equal", "color", "edge"),
		),
		mcp.WithNumber("rows",
			mcp.Description("Number of rows (0 = auto detect). Default: 3"),
		),
		mcp.WithNumber("cols",
			mcp.Description("Number of columns (0 = auto detect). Default: 3"),
		),
		mcp.WithString("line_color",
			mcp.Description("Line color for color detection mode, format: #RRGGBB. Default: #FFFFFF"),
		),
		mcp.WithNumber("color_tolerance",
			mcp.Description("Color tolerance for color detection mode (0-255). Default: 30"),
		),
		mcp.WithNumber("edge_threshold",
			mcp.Description("Edge threshold for edge detection mode (0-255). Default: 50"),
		),
		mcp.WithString("output_format",
			mcp.Description("Output format: 'png', 'jpg', 'jpeg'. Default: 'png'"),
			mcp.Enum("png", "jpg", "jpeg"),
		),
		mcp.WithNumber("padding",
			mcp.Description("Padding around each split image in pixels. Default: 0"),
		),
		mcp.WithBoolean("auto_detect",
			mcp.Description("Auto-detect grid dimensions. Default: false"),
		),
	)

	// Tool 2: detect_grid_size
	detectTool := mcp.NewTool("detect_grid_size",
		mcp.WithDescription("Detect the grid structure (rows and columns) in an image automatically."),
		mcp.WithString("image_path",
			mcp.Required(),
			mcp.Description("Path to the input image file"),
		),
	)

	// Add tool handlers
	s.AddTool(splitTool, splitImageHandler)
	s.AddTool(detectTool, detectGridSizeHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func splitImageHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	// Parse required parameters
	imagePath, ok := argsMap["image_path"].(string)
	if !ok || imagePath == "" {
		return mcp.NewToolResultError("image_path is required and must be a string"), nil
	}

	outputDir, ok := argsMap["output_dir"].(string)
	if !ok || outputDir == "" {
		return mcp.NewToolResultError("output_dir is required and must be a string"), nil
	}

	// Build command arguments
	args := []string{imagePath, outputDir}

	if mode, ok := argsMap["mode"].(string); ok && mode != "" {
		args = append(args, "-m", mode)
	}

	if rows, ok := argsMap["rows"].(float64); ok {
		args = append(args, "-r", fmt.Sprintf("%d", int(rows)))
	}

	if cols, ok := argsMap["cols"].(float64); ok {
		args = append(args, "-c", fmt.Sprintf("%d", int(cols)))
	}

	if lineColor, ok := argsMap["line_color"].(string); ok && lineColor != "" {
		args = append(args, "--line-color", lineColor)
	}

	if colorTolerance, ok := argsMap["color_tolerance"].(float64); ok {
		args = append(args, "--color-tolerance", fmt.Sprintf("%d", int(colorTolerance)))
	}

	if edgeThreshold, ok := argsMap["edge_threshold"].(float64); ok {
		args = append(args, "--edge-threshold", fmt.Sprintf("%d", int(edgeThreshold)))
	}

	if outputFormat, ok := argsMap["output_format"].(string); ok && outputFormat != "" {
		args = append(args, "-f", outputFormat)
	}

	if padding, ok := argsMap["padding"].(float64); ok {
		args = append(args, "--padding", fmt.Sprintf("%d", int(padding)))
	}

	if autoDetect, ok := argsMap["auto_detect"].(bool); ok && autoDetect {
		args = append(args, "--auto")
	}

	// Execute command
	cmd := exec.CommandContext(ctx, gridsplitPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to execute gridsplit: %v\nOutput: %s", err, string(output))), nil
	}

	// List output files
	files, _ := filepath.Glob(filepath.Join(outputDir, "*.png"))
	allFiles := files
	jpgFiles, _ := filepath.Glob(filepath.Join(outputDir, "*.jpg"))
	allFiles = append(allFiles, jpgFiles...)
	jpegFiles, _ := filepath.Glob(filepath.Join(outputDir, "*.jpeg"))
	allFiles = append(allFiles, jpegFiles...)

	resultText := string(output)
	if len(allFiles) > 0 {
		resultText += "\n\nGenerated files:\n"
		for _, f := range allFiles {
			resultText += fmt.Sprintf("  - %s\n", f)
		}
	}

	return mcp.NewToolResultText(resultText), nil
}

func detectGridSizeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	// Parse required parameters
	imagePath, ok := argsMap["image_path"].(string)
	if !ok || imagePath == "" {
		return mcp.NewToolResultError("image_path is required and must be a string"), nil
	}

	// Build command arguments for dry-run with auto-detect
	args := []string{imagePath, "--dry-run", "--auto", "-v"}

	// Execute command
	cmd := exec.CommandContext(ctx, gridsplitPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to execute gridsplit: %v\nOutput: %s", err, string(output))), nil
	}

	// Parse output to extract grid information
	outputStr := string(output)
	resultText := fmt.Sprintf("Grid detection results for: %s\n\n", imagePath)
	resultText += outputStr

	return mcp.NewToolResultText(resultText), nil
}
