package prompt

import (
	"context"
	"strings"
	"testing"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestSystemBuilder(t *testing.T) {
	t.Run("基础角色", func(t *testing.T) {
		builder := NewSystem("代码助手")
		result := builder.Build()

		if !strings.Contains(result, "# 角色定义") {
			t.Error("Build() 结果应包含角色定义标题")
		}
		if !strings.Contains(result, "你是一个代码助手") {
			t.Error("Build() 结果应包含角色名称")
		}
	})

	t.Run("完整配置", func(t *testing.T) {
		builder := NewSystem("数据分析师").
			WithDescription("专业的数据分析助手").
			WithCapabilities("数据处理", "可视化").
			WithConstraint("回答必须基于数据").
			WithContext("当前项目使用 Python")

		result := builder.Build()

		contains := []string{
			"数据分析师",
			"专业的数据分析助手",
			"数据处理",
			"可视化",
			"回答必须基于数据",
			"当前项目使用 Python",
		}

		for _, s := range contains {
			if !strings.Contains(result, s) {
				t.Errorf("Build() 结果缺少 %q", s)
			}
		}
	})
}

func TestSystemBuilderWithTools(t *testing.T) {
	tools := []*llm.Tool{
		llm.NewTool("search", "搜索文档", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词",
				},
			},
			"required": []string{"query"},
		}),
	}

	builder := NewSystem("助手").WithTools(tools...)
	result := builder.Build()

	if !strings.Contains(result, "search") {
		t.Error("Build() 结果应包含工具名称")
	}
	if !strings.Contains(result, "搜索文档") {
		t.Error("Build() 结果应包含工具描述")
	}
}

func TestSystemBuilderWithOutputFormat(t *testing.T) {
	format := JSONOutput(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"result": map[string]any{"type": "string"},
		},
	})

	builder := NewSystem("助手").WithOutputFormat(format)
	result := builder.Build()

	if !strings.Contains(result, "JSON") {
		t.Error("Build() 结果应包含 JSON 格式说明")
	}
}

func TestSystemBuilderClone(t *testing.T) {
	original := NewSystem("助手").
		WithDescription("原始描述").
		WithConstraint("约束1")

	cloned := original.Clone().
		WithConstraint("约束2")

	originalStr := original.Build()
	clonedStr := cloned.Build()

	if strings.Contains(originalStr, "约束2") {
		t.Error("克隆不应影响原始构建器")
	}
	if !strings.Contains(clonedStr, "约束1") {
		t.Error("克隆应保留原始配置")
	}
	if !strings.Contains(clonedStr, "约束2") {
		t.Error("克隆应包含新添加的约束")
	}
}

func TestHistory(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		h := NewHistory()
		if h.Len() != 0 {
			t.Error("新历史长度应为 0")
		}

		h.AddUser("你好").AddAssistant("你好！有什么可以帮助你的？")
		if h.Len() != 2 {
			t.Errorf("历史长度应为 2, got %d", h.Len())
		}

		messages := h.ToMessages()
		if len(messages) != 2 {
			t.Errorf("消息数量应为 2, got %d", len(messages))
		}
		if messages[0].Role != llm.RoleUser {
			t.Error("第一条消息角色应为 user")
		}
	})

	t.Run("工具调用", func(t *testing.T) {
		h := NewHistory()
		toolCalls := []*llm.ToolCall{
			llm.NewToolCall("id1", "test_tool", `{"arg": "value"}`),
		}

		h.AddAssistantWithTools("调用工具", toolCalls).
			AddToolResult("id1", "test_tool", "工具结果")

		if h.Len() != 2 {
			t.Errorf("历史长度应为 2, got %d", h.Len())
		}

		messages := h.ToMessages()
		if messages[0].Role != llm.RoleAssistant {
			t.Error("第一条消息角色应为 assistant")
		}
		if len(messages[0].ToolCalls) != 1 {
			t.Error("助手消息应包含工具调用")
		}
		if messages[1].Role != llm.RoleTool {
			t.Error("第二条消息角色应为 tool")
		}
	})

	t.Run("克隆", func(t *testing.T) {
		h := NewHistory().
			AddUser("用户1").
			AddAssistant("助手1")

		cloned := h.Clone()
		h.AddUser("用户2")

		if h.Len() != 3 {
			t.Errorf("原始历史长度应为 3, got %d", h.Len())
		}
		if cloned.Len() != 2 {
			t.Errorf("克隆历史长度应为 2, got %d", cloned.Len())
		}
	})
}

func TestHistoryGetLastMessages(t *testing.T) {
	h := NewHistory().
		AddSystem("系统消息").
		AddUser("用户1").
		AddAssistant("助手1").
		AddUser("用户2")

	lastUser := h.GetLastUserMessage()
	if lastUser == nil || lastUser.Content != "用户2" {
		t.Error("GetLastUserMessage() 应返回最后一条用户消息")
	}

	lastAssistant := h.GetLastAssistantMessage()
	if lastAssistant == nil || lastAssistant.Content != "助手1" {
		t.Error("GetLastAssistantMessage() 应返回最后一条助手消息")
	}
}

func TestSlidingWindowTruncate(t *testing.T) {
	strategy := NewSlidingWindowStrategy(3)
	h := NewHistory()

	// 添加 5 条消息
	for i := 1; i <= 5; i++ {
		h.AddUser("消息")
	}

	items, _, err := strategy.Truncate(context.Background(), h.Items())
	if err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Truncate() 返回 %d 条消息, want 3", len(items))
	}
}

func TestSlidingWindowPreserveSystem(t *testing.T) {
	strategy := NewSlidingWindowStrategy(2)
	strategy.PreserveSystem = true

	h := NewHistory().
		AddSystem("系统消息").
		AddUser("用户1").
		AddAssistant("助手1").
		AddUser("用户2")

	items, _, err := strategy.Truncate(context.Background(), h.Items())
	if err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}

	// 应保留系统消息 + 最近 2 条
	if len(items) != 3 {
		t.Errorf("Truncate() 返回 %d 条消息, want 3", len(items))
	}
	if items[0].Role != "system" {
		t.Error("第一条消息应为系统消息")
	}
}

func TestOutputFormat(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		format := JSONOutput(map[string]any{
			"type": "object",
		})
		prompt := format.ToPrompt()

		if !strings.Contains(prompt, "JSON") {
			t.Error("JSON 输出格式应包含 'JSON'")
		}
	})

	t.Run("YAML", func(t *testing.T) {
		format := YAMLOutput(map[string]any{
			"type": "object",
		})
		prompt := format.ToPrompt()

		if !strings.Contains(prompt, "YAML") {
			t.Error("YAML 输出格式应包含 'YAML'")
		}
	})

	t.Run("Markdown", func(t *testing.T) {
		format := MarkdownOutput(map[string]any{
			"description": "生成报告",
			"sections":    []string{"摘要", "详情", "结论"},
		})
		prompt := format.ToPrompt()

		if !strings.Contains(prompt, "Markdown") {
			t.Error("Markdown 输出格式应包含 'Markdown'")
		}
		if !strings.Contains(prompt, "摘要") {
			t.Error("Markdown 输出格式应包含章节")
		}
	})
}

func TestGenerateToolPrompt(t *testing.T) {
	tools := []*llm.Tool{
		llm.NewTool("test", "测试工具", map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"input": map[string]interface{}{
					"type":        "string",
					"description": "输入参数",
				},
			},
			"required": []string{"input"},
		}),
	}

	prompt := GenerateToolPrompt(tools...)
	if !strings.Contains(prompt, "test") {
		t.Error("工具提示应包含工具名称")
	}
	if !strings.Contains(prompt, "测试工具") {
		t.Error("工具提示应包含工具描述")
	}
	if !strings.Contains(prompt, "input") {
		t.Error("工具提示应包含参数名称")
	}
}

func TestGenerateToolSchema(t *testing.T) {
	tools := []*llm.Tool{
		llm.NewTool("test", "测试工具", map[string]interface{}{
			"type": "object",
		}),
	}

	schema := GenerateToolSchema(tools...)
	if !strings.Contains(schema, "test") {
		t.Error("工具 Schema 应包含工具名称")
	}
	if !strings.Contains(schema, "测试工具") {
		t.Error("工具 Schema 应包含工具描述")
	}
}

func TestSummaryTruncateWithoutGenerator(t *testing.T) {
	// 没有设置摘要生成器时，应退化为滑动窗口
	strategy := NewSummaryTruncateStrategy(3)
	h := NewHistory()

	for i := 1; i <= 5; i++ {
		h.AddUser("消息")
	}

	items, summary, err := strategy.Truncate(context.Background(), h.Items())
	if err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Truncate() 返回 %d 条消息, want 3", len(items))
	}
	if summary != "" {
		t.Error("没有摘要生成器时，摘要应为空")
	}
}

func TestFixedSystemTruncate(t *testing.T) {
	strategy := NewFixedSystemStrategy(2)
	h := NewHistory()

	// 添加系统消息和多条对话
	h.AddSystem("系统提示").
		AddUser("用户1").
		AddAssistant("助手1").
		AddUser("用户2").
		AddAssistant("助手2")

	items, _, err := strategy.Truncate(context.Background(), h.Items())
	if err != nil {
		t.Fatalf("Truncate() error = %v", err)
	}

	// 应保留系统消息
	if len(items) < 1 || items[0].Role != "system" {
		t.Error("应保留系统消息")
	}
}
