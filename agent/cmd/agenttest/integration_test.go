//go:build integration
// +build integration

package main

import (
	"context"
	"testing"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/agent"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/llm/mock"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// TestIntegration_AgentWithTools 测试 Agent 与工具的集成
func TestIntegration_AgentWithTools(t *testing.T) {
	// 创建 Mock LLM
	mockLLM := mock.NewClient()
	mockLLM.AddToolCallResponse("get_time", `{}`)
	mockLLM.AddTextResponse("当前时间是 2024-01-01T00:00:00Z")

	// 创建工具
	timeTool := tool.NewFunction("get_time", "获取当前时间",
		func(ctx context.Context, args map[string]any) (any, error) {
			return map[string]any{
				"time":     "2024-01-01T00:00:00Z",
				"timezone": "UTC",
			}, nil
		},
	)

	// 创建 Agent
	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithSystemPrompt("你是一个助手"),
		agent.WithTools(timeTool),
		agent.WithMaxLoops(5),
	)

	// 执行任务
	result, err := ag.Run(context.Background(), "现在几点了？")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证结果
	if result.LoopCount != 2 {
		t.Errorf("expected 2 loops, got %d", result.LoopCount)
	}

	if len(result.ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(result.ToolCalls))
	}
}

// TestIntegration_AgentWithMemory 测试 Agent 与记忆的集成
func TestIntegration_AgentWithMemory(t *testing.T) {
	mockLLM := mock.NewClient()
	mockLLM.AddTextResponse("你好！我是 AI 助手")
	mockLLM.AddTextResponse("你刚才说：你好！我是 AI 助手")

	// 创建记忆
	mem := memory.NewSlidingWindow(10)

	// 创建 Agent
	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithSystemPrompt("你是一个助手"),
	)
	ag.SetMemory(mem)

	// 第一次对话
	_, err := ag.Run(context.Background(), "你好")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 第二次对话
	result, err := ag.Run(context.Background(), "我刚才说了什么？")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content == "" {
		t.Error("expected non-empty response")
	}
}

// TestIntegration_MultipleToolCalls 测试多工具调用
func TestIntegration_MultipleToolCalls(t *testing.T) {
	mockLLM := mock.NewClient()
	mockLLM.AddResponse(&llm.ChatResponse{
		StopReason: llm.StopReasonToolUse,
		ToolCalls: []*llm.ToolCall{
			llm.NewToolCall("call_1", "tool_a", `{}`),
			llm.NewToolCall("call_2", "tool_b", `{}`),
		},
	})
	mockLLM.AddTextResponse("两个工具都执行完成")

	// 创建工具
	toolA := tool.NewFunction("tool_a", "工具A",
		func(ctx context.Context, args map[string]any) (any, error) {
			return "result_a", nil
		},
	)
	toolB := tool.NewFunction("tool_b", "工具B",
		func(ctx context.Context, args map[string]any) (any, error) {
			return "result_b", nil
		},
	)

	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithTools(toolA, toolB),
	)

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ToolCalls) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(result.ToolCalls))
	}
}

// TestIntegration_ErrorHandling 测试错误处理
func TestIntegration_ErrorHandling(t *testing.T) {
	mockLLM := mock.NewClient()
	mockLLM.AddToolCallResponse("error_tool", `{}`)
	mockLLM.AddTextResponse("工具执行失败，但我继续处理")

	// 创建会出错的工具
	errorTool := tool.NewFunction("error_tool", "错误工具",
		func(ctx context.Context, args map[string]any) (any, error) {
			return nil, tool.ErrExecutionFailed
		},
	)

	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithTools(errorTool),
	)

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Agent 应该能够处理工具错误并继续
	if result.LoopCount != 2 {
		t.Errorf("expected 2 loops, got %d", result.LoopCount)
	}
}

// TestIntegration_MaxLoops 测试最大循环次数
func TestIntegration_MaxLoops(t *testing.T) {
	mockLLM := mock.NewClient()
	// 总是返回工具调用
	for i := 0; i < 10; i++ {
		mockLLM.AddToolCallResponse("loop_tool", `{}`)
	}

	loopTool := tool.NewFunction("loop_tool", "循环工具",
		func(ctx context.Context, args map[string]any) (any, error) {
			return "result", nil
		},
	)

	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithTools(loopTool),
		agent.WithMaxLoops(3),
	)

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.LoopCount > 3 {
		t.Errorf("expected max 3 loops, got %d", result.LoopCount)
	}
}

// TestIntegration_Hooks 测试钩子函数
func TestIntegration_Hooks(t *testing.T) {
	mockLLM := mock.NewClient()
	mockLLM.AddTextResponse("response")

	var startCalled, loopCalled, doneCalled bool

	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithHooks(&agent.Hooks{
			OnStart: func(input string) {
				startCalled = true
			},
			OnLoop: func(loopCount int, state *agent.State) {
				loopCalled = true
			},
			OnDone: func(result *agent.Result) {
				doneCalled = true
			},
		}),
	)

	_, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !startCalled {
		t.Error("expected OnStart to be called")
	}
	if !loopCalled {
		t.Error("expected OnLoop to be called")
	}
	if !doneCalled {
		t.Error("expected OnDone to be called")
	}
}

// TestIntegration_Timeout 测试超时控制
func TestIntegration_Timeout(t *testing.T) {
	mockLLM := mock.NewClient()
	mockLLM.AddTextResponse("response")

	ag := agent.New(
		agent.WithLLM(mockLLM),
		agent.WithTimeout(5*time.Second),
	)

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Duration > 5*time.Second {
		t.Errorf("execution took longer than timeout: %v", result.Duration)
	}
}
