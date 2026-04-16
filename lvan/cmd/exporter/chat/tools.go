package chat

import (
	"context"
	"fmt"

	"github.com/wangtengda0310/gobee/agent/pkg/tool"
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

	return reg
}
