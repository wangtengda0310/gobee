package chat

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExporterTools(t *testing.T) {
	reg := NewExporterTools()
	require.NotNil(t, reg)
	assert.Equal(t, 5, reg.Count(), "应注册 5 个工具")
}

func TestExporterTools_HasExpectedTools(t *testing.T) {
	reg := NewExporterTools()

	expectedTools := []string{"execute_command", "list_commands", "get_task_result", "cancel_task", "run_shell"}
	for _, name := range expectedTools {
		tool, found := reg.GetTool(name)
		assert.True(t, found, "工具 %s 应存在", name)
		assert.NotNil(t, tool)
		assert.Equal(t, name, tool.Name())
	}
}

func TestExporterTools_GetDefinitions(t *testing.T) {
	reg := NewExporterTools()
	defs := reg.GetDefinitions()
	assert.Equal(t, 5, len(defs))

	for _, def := range defs {
		assert.NotNil(t, def.Function)
		assert.NotEmpty(t, def.Function.Name)
		assert.NotEmpty(t, def.Function.Description)
	}
}

func TestExporterTools_ListCommands(t *testing.T) {
	reg := NewExporterTools()
	tool, found := reg.GetTool("list_commands")
	require.True(t, found)

	result, err := tool.Execute(context.Background(), map[string]any{})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestExporterTools_ExecuteCommand(t *testing.T) {
	reg := NewExporterTools()
	tool, found := reg.GetTool("execute_command")
	require.True(t, found)

	result, err := tool.Execute(context.Background(), map[string]any{
		"cmd": "test_command",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "created", result.(map[string]any)["status"])
	assert.Equal(t, "test_command", result.(map[string]any)["command"])
}

func TestExporterTools_GetTaskResult(t *testing.T) {
	reg := NewExporterTools()
	tool, found := reg.GetTool("get_task_result")
	require.True(t, found)

	result, err := tool.Execute(context.Background(), map[string]any{
		"task_id": "task-123",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "task-123", result.(map[string]any)["task_id"])
	assert.Equal(t, "not_found", result.(map[string]any)["status"])
}

func TestExporterTools_CancelTask(t *testing.T) {
	reg := NewExporterTools()
	tool, found := reg.GetTool("cancel_task")
	require.True(t, found)

	result, err := tool.Execute(context.Background(), map[string]any{
		"task_id": "task-456",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "task-456", result.(map[string]any)["task_id"])
	assert.Equal(t, "cancelled", result.(map[string]any)["status"])
}
func TestExporterTools_RunShell(t *testing.T) {
	reg := NewExporterTools()
	shellTool, found := reg.GetTool("run_shell")
	require.True(t, found)

	t.Run("基本命令执行", func(t *testing.T) {
		result, err := shellTool.Execute(context.Background(), map[string]any{
			"command": "echo hello",
		})
		require.NoError(t, err)
		m := result.(map[string]any)
		assert.Equal(t, 0, m["exitCode"])
		assert.Contains(t, m["stdout"], "hello")
	})

	t.Run("命令执行失败", func(t *testing.T) {
		result, err := shellTool.Execute(context.Background(), map[string]any{
			"command": "nonexistent_cmd_xyz",
		})
		require.NoError(t, err)
		m := result.(map[string]any)
		assert.NotEqual(t, 0, m["exitCode"])
	})

	t.Run("缺少 command 参数", func(t *testing.T) {
		_, err := shellTool.Execute(context.Background(), map[string]any{})
		assert.Error(t, err)
	})
}
