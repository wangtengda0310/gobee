package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"Cursor MCP Host",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("echo",
			mcp.WithDescription("返回输入内容"),
			mcp.WithString("message", mcp.Required(), mcp.Description("需要回显的内容")),
		),
		handleEcho,
	)

	s.AddTool(
		mcp.NewTool("listDir",
			mcp.WithDescription("列出指定目录的文件结构"),
			mcp.WithString("path", mcp.Required(), mcp.Description("要列出文件结构的目录路径")),
			mcp.WithString("depth", mcp.Description("遍历深度，默认为1")),
		),
		handleListDir,
	)

	s.AddTool(
		mcp.NewTool("readFile",
			mcp.WithDescription("读取文件内容"),
			mcp.WithString("path", mcp.Required(), mcp.Description("文件路径")),
		),
		handleReadFile,
	)

	s.AddTool(
		mcp.NewTool("writeFile",
			mcp.WithDescription("写入文件内容"),
			mcp.WithString("path", mcp.Required(), mcp.Description("文件路径")),
			mcp.WithString("content", mcp.Required(), mcp.Description("要写入的内容")),
		),
		handleWriteFile,
	)

	s.AddTool(
		mcp.NewTool("searchInFiles",
			mcp.WithDescription("在文件中搜索内容"),
			mcp.WithString("pattern", mcp.Required(), mcp.Description("搜索模式")),
			mcp.WithString("path", mcp.Required(), mcp.Description("搜索路径")),
			mcp.WithString("filePattern", mcp.Description("文件匹配模式，例如 *.go")),
		),
		handleSearchInFiles,
	)

	fmt.Println("Cursor MCP 主机已启动...")
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

func handleEcho(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	msg, err := req.RequireString("message")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("ECHO: %s", msg)), nil
}

func handleListDir(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dirPath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	depth := 1
	depthStr := req.GetString("depth", "1")
	if depthStr != "" {
		if d, err := parseInt(depthStr); err == nil && d > 0 {
			depth = d
		}
	}

	result, err := listDirStructure(dirPath, depth)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("列出目录结构失败: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

func handleReadFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取文件失败: %v", err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func handleWriteFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("写入文件失败: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("成功写入文件: %s", filePath)), nil
}

func handleSearchInFiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, err := req.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	searchPath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	filePattern := req.GetString("filePattern", "*")

	results, err := searchFiles(searchPath, pattern, filePattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索失败: %v", err)), nil
	}

	return mcp.NewToolResultText(results), nil
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func listDirStructure(rootPath string, maxDepth int) (string, error) {
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return "", fmt.Errorf("目录不存在: %s", rootPath)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("目录结构 (%s):\n", rootPath))

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算当前路径相对于根路径的深度
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// 跳过根目录自身
		if relPath == "." {
			return nil
		}

		// 计算深度
		depth := len(strings.Split(relPath, string(os.PathSeparator)))
		if depth > maxDepth {
			// 如果超过最大深度且是目录，跳过该目录
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 添加适当的缩进
		indent := strings.Repeat("  ", depth-1)

		// 添加文件或目录名
		if info.IsDir() {
			result.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, info.Name()))
		} else {
			result.WriteString(fmt.Sprintf("%s📄 %s\n", indent, info.Name()))
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func searchFiles(rootPath, searchPattern, filePattern string) (string, error) {
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return "", fmt.Errorf("路径不存在: %s", rootPath)
	}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("在 %s 中搜索 '%s' (文件模式: %s):\n\n", rootPath, searchPattern, filePattern))

	count := 0
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件是否匹配模式
		matched, err := filepath.Match(filePattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		// 读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		// 搜索内容
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, searchPattern) {
				relPath, _ := filepath.Rel(rootPath, path)
				results.WriteString(fmt.Sprintf("%s:%d: %s\n", relPath, i+1, line))
				count++
				// 限制结果数量
				if count >= 100 {
					results.WriteString("\n... 结果太多，仅显示前100条 ...\n")
					return filepath.SkipDir
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if count == 0 {
		results.WriteString("未找到匹配结果\n")
	} else {
		results.WriteString(fmt.Sprintf("\n共找到 %d 个匹配项\n", count))
	}

	return results.String(), nil
}
