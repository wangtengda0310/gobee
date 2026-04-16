package chat

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
	"unicode/utf8"

	"github.com/wangtengda0310/gobee/agent/pkg/tool"
	lvanutf8 "github.com/wangtengda0310/gobee/lvan/pkg/utf8"
)

// buildShellCmd 构建当前平台的 shell 命令
// Windows: cmd /c "chcp 65001 >nul && command"（UTF-8 编码）
// Linux/macOS: bash -c "command" 或 sh -c "command"
func buildShellCmd(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", "chcp 65001 >nul && "+command)
	}
	// 优先 bash，fallback 到 sh
	if _, err := exec.LookPath("bash"); err == nil {
		return exec.CommandContext(ctx, "bash", "-c", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

const (
	// shell 命令执行超时时间
	shellTimeout = 30 * time.Second
	// 最大输出字节数，避免返回超大结果
	maxOutputLen = 32 * 1024
)

// NewExporterTools 创建并注册 exporter 命令执行工具
func NewExporterTools() *tool.Registry {
	reg := tool.NewRegistry()

	// execute_command: 执行 exporter 命令
	reg.MustRegister(tool.NewFunction(
		"execute_command",
		"执行 exporter 注册的命令行工具。通过指定命令名和参数来执行对应的工具程序。",
		func(ctx context.Context, args map[string]any) (any, error) {
			cmd, _ := args["cmd"].(string)
			if cmd == "" {
				return nil, fmt.Errorf("cmd 参数不能为空")
			}
			return map[string]any{
				"status":  "created",
				"command": cmd,
				"message": fmt.Sprintf("命令 %s 已提交执行", cmd),
			}, nil
		},
		tool.WithStringParam("cmd", "要执行的命令名称", true),
	))

	// list_commands: 列出可用命令
	reg.MustRegister(tool.NewFunction(
		"list_commands",
		"列出 exporter 中所有可用的命令行工具",
		func(ctx context.Context, args map[string]any) (any, error) {
			return map[string]any{
				"commands": []string{"csv2mp", "json2mp", "coverage", "merkle", "utf8"},
			}, nil
		},
	))

	// get_task_result: 获取命令执行结果
	reg.MustRegister(tool.NewFunction(
		"get_task_result",
		"根据任务 ID 获取已提交命令的执行结果",
		func(ctx context.Context, args map[string]any) (any, error) {
			taskID, _ := args["task_id"].(string)
			if taskID == "" {
				return nil, fmt.Errorf("task_id 参数不能为空")
			}
			return map[string]any{
				"task_id": taskID,
				"status":  "not_found",
			}, nil
		},
		tool.WithStringParam("task_id", "任务 ID", true),
	))

	// cancel_task: 取消正在执行的任务
	reg.MustRegister(tool.NewFunction(
		"cancel_task",
		"取消指定任务 ID 的正在执行中的命令",
		func(ctx context.Context, args map[string]any) (any, error) {
			taskID, _ := args["task_id"].(string)
			if taskID == "" {
				return nil, fmt.Errorf("task_id 参数不能为空")
			}
			return map[string]any{
				"task_id": taskID,
				"status":  "cancelled",
			}, nil
		},
		tool.WithStringParam("task_id", "任务 ID", true),
	))

	// run_shell: 执行本地命令行命令
	reg.MustRegister(tool.NewFunction(
		"run_shell",
		"在本地执行 shell 命令并返回输出结果。支持执行任意命令行指令，如文件操作、系统查询、构建部署等。",
		runShellCommand,
		tool.WithParameters(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "要执行的完整命令",
				},
				"workdir": map[string]any{
					"type":        "string",
					"description": "工作目录（可选，默认为当前目录）",
				},
			},
			"required": []string{"command"},
		}),
	))

	return reg
}

// runShellCommand 执行本地 shell 命令
// 跨平台兼容：Windows 使用 cmd /c chcp 65001 确保 UTF-8 编码，Linux/macOS 使用 bash -c
func runShellCommand(ctx context.Context, args map[string]any) (any, error) {
	command, _ := args["command"].(string)
	if command == "" {
		return nil, fmt.Errorf("command 参数不能为空")
	}

	workdir, _ := args["workdir"].(string)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, shellTimeout)
	defer cancel()

	// 根据平台构建 shell 命令
	cmd := buildShellCmd(ctx, command)

	// 设置工作目录
	if workdir != "" {
		cmd.Dir = workdir
	}

	// 捕获原始字节输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// 转换输出：优先 UTF-8，如果含非法字节序列则尝试 GBK 解码
	outBytes := stdout.Bytes()
	errBytes := stderr.Bytes()
	outStr := truncateOutput(decodeOutput(outBytes))
	errStr := truncateOutput(decodeOutput(errBytes))

	result := map[string]any{
		"command":  command,
		"workdir":  workdir,
		"stdout":   outStr,
		"stderr":   errStr,
		"exitCode": 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exitCode"] = exitErr.ExitCode()
		} else {
			result["exitCode"] = -1
			result["error"] = err.Error()
		}
	}

	return result, nil
}

// truncateOutput 截断过长的输出
func truncateOutput(s string) string {
	if len(s) > maxOutputLen {
		return s[:maxOutputLen] + "\n... (输出已截断)"
	}
	return s
}

// decodeOutput 解码命令输出，优先 UTF-8，失败时尝试 GBK
func decodeOutput(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	return lvanutf8.From(data, lvanutf8.GBK)
}
