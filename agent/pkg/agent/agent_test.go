package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// MockLLM 模拟 LLM 客户端
type MockLLM struct {
	responses []*llm.ChatResponse
	callCount int
}

func (m *MockLLM) Complete(_ context.Context, _ *llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.callCount >= len(m.responses) {
		return &llm.ChatResponse{
			Content:    "default response",
			StopReason: llm.StopReasonEndTurn,
		}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *MockLLM) Stream(_ context.Context, _ *llm.ChatRequest) (<-chan *llm.StreamChunk, error) {
	ch := make(chan *llm.StreamChunk, 1)
	ch <- llm.NewDoneChunk(&llm.ChatResponse{Content: "stream response"})
	close(ch)
	return ch, nil
}

func TestNew(t *testing.T) {
	ag := New()
	if ag == nil {
		t.Error("expected agent to be created")
	}
	if ag.config.MaxLoops != 10 {
		t.Errorf("expected default max loops 10, got %d", ag.config.MaxLoops)
	}
}

func TestWithLLM(t *testing.T) {
	mockLLM := &MockLLM{}
	ag := New(WithLLM(mockLLM))

	if ag.config.LLM == nil {
		t.Error("expected LLM to be set")
	}
}

func TestWithSystemPrompt(t *testing.T) {
	ag := New(WithSystemPrompt("你是一个助手"))

	if ag.config.SystemPrompt != "你是一个助手" {
		t.Errorf("expected system prompt, got '%s'", ag.config.SystemPrompt)
	}
}

func TestWithTools(t *testing.T) {
	tool1 := tool.NewFunction("test", "测试工具", nil)

	ag := New(WithTools(tool1))

	if len(ag.config.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(ag.config.Tools))
	}
}

func TestWithHooks(t *testing.T) {
	startCalled := false
	hooks := &Hooks{
		OnStart: func(input string) {
			startCalled = true
		},
	}

	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{Content: "Hello!", StopReason: llm.StopReasonEndTurn},
		},
	}

	ag := New(WithLLM(mockLLM), WithHooks(hooks))

	_, err := ag.Run(context.Background(), "Hi")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !startCalled {
		t.Error("expected OnStart hook to be called")
	}
}

func TestAgent_Run_SimpleResponse(t *testing.T) {
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{
				Content:    "Hello! How can I help you?",
				StopReason: llm.StopReasonEndTurn,
			},
		},
	}

	ag := New(WithLLM(mockLLM))

	result, err := ag.Run(context.Background(), "Hi")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.Content != "Hello! How can I help you?" {
		t.Errorf("expected 'Hello! How can I help you?', got '%s'", result.Content)
	}

	if result.LoopCount != 1 {
		t.Errorf("expected 1 loop, got %d", result.LoopCount)
	}
}

func TestAgent_Run_WithToolCall(t *testing.T) {
	// 创建测试工具
	timeTool := tool.NewFunction("get_time", "获取当前时间",
		func(ctx context.Context, args map[string]any) (any, error) {
			return "2024-01-01T00:00:00Z", nil
		},
	)

	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{
				Content:    "",
				StopReason: llm.StopReasonToolUse,
				ToolCalls: []*llm.ToolCall{
					llm.NewToolCall("call_1", "get_time", "{}"),
				},
			},
			{
				Content:    "当前时间是 2024-01-01T00:00:00Z",
				StopReason: llm.StopReasonEndTurn,
			},
		},
	}

	toolCallHookCalled := false
	ag := New(
		WithLLM(mockLLM),
		WithTools(timeTool),
		WithHooks(&Hooks{
			OnToolCall: func(name string, args map[string]any) {
				toolCallHookCalled = true
				if name != "get_time" {
					t.Errorf("expected tool name 'get_time', got '%s'", name)
				}
			},
		}),
	)

	result, err := ag.Run(context.Background(), "现在几点了？")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !toolCallHookCalled {
		t.Error("expected OnToolCall hook to be called")
	}

	if result.LoopCount != 2 {
		t.Errorf("expected 2 loops, got %d", result.LoopCount)
	}
}

func TestAgent_AddTool(t *testing.T) {
	ag := New()

	newTool := tool.NewFunction("new_tool", "新工具", nil)
	if err := ag.AddTool(newTool); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 重复添加应该失败
	if err := ag.AddTool(newTool); err == nil {
		t.Error("expected error for duplicate tool")
	}
}

func TestLoopController(t *testing.T) {
	controller := NewLoopController(5)

	if controller.MaxLoops != 5 {
		t.Errorf("expected max loops 5, got %d", controller.MaxLoops)
	}

	if !controller.ShouldContinue() {
		t.Error("expected to continue")
	}

	for range 5 {
		controller.Increment()
	}

	controller.CheckMaxLoops()
	if controller.State != LoopStateMaxLoops {
		t.Errorf("expected state MaxLoops, got %v", controller.State)
	}
}

func TestResult_Duration(t *testing.T) {
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{Content: "test", StopReason: llm.StopReasonEndTurn},
		},
	}

	ag := New(WithLLM(mockLLM))

	start := time.Now()
	result, err := ag.Run(context.Background(), "test")
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.Duration > elapsed {
		t.Error("result duration should not exceed actual elapsed time")
	}
}

// === Phase 3 新增测试 ===

func TestWithMaxTokens(t *testing.T) {
	mockLLM := &MockLLM{}
	ag := New(WithLLM(mockLLM), WithMaxTokens(2048))

	if ag.config.MaxTokens != 2048 {
		t.Errorf("expected MaxTokens 2048, got %d", ag.config.MaxTokens)
	}
}

func TestWithTemperature(t *testing.T) {
	mockLLM := &MockLLM{}
	ag := New(WithLLM(mockLLM), WithTemperature(0.5))

	if ag.config.Temperature != 0.5 {
		t.Errorf("expected Temperature 0.5, got %f", ag.config.Temperature)
	}
}

func TestWithTimeout(t *testing.T) {
	ag := New(WithTimeout(30 * time.Second))

	if ag.config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", ag.config.Timeout)
	}
}

func TestAgent_SetMemory(t *testing.T) {
	ag := New()
	mem := memory.NewSlidingWindow(10)

	ag.SetMemory(mem)

	// 验证 memory 被设置（通过内部字段）
	if ag.memory == nil {
		t.Error("expected memory to be set")
	}
}

func TestAgent_ClearHistory(t *testing.T) {
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{Content: "response", StopReason: llm.StopReasonEndTurn},
		},
	}

	ag := New(WithLLM(mockLLM))

	// 运行一次产生历史
	_, _ = ag.Run(context.Background(), "test")

	// 清空历史
	ag.ClearHistory()

	// 历史应该被重置
	if ag.history == nil {
		t.Error("expected history to be non-nil after ClearHistory")
	}
}

func TestAgent_MaxLoops(t *testing.T) {
	// 创建一个总是返回工具调用的 MockLLM
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{StopReason: llm.StopReasonToolUse, ToolCalls: []*llm.ToolCall{llm.NewToolCall("call_1", "test_tool", "{}")}},
			{StopReason: llm.StopReasonToolUse, ToolCalls: []*llm.ToolCall{llm.NewToolCall("call_2", "test_tool", "{}")}},
			{StopReason: llm.StopReasonToolUse, ToolCalls: []*llm.ToolCall{llm.NewToolCall("call_3", "test_tool", "{}")}},
		},
	}

	testTool := tool.NewFunction("test_tool", "测试工具", func(ctx context.Context, args map[string]any) (any, error) {
		return "result", nil
	})

	// 设置最大循环次数为 2
	ag := New(WithLLM(mockLLM), WithTools(testTool), WithMaxLoops(2))

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 应该在达到最大循环次数后停止
	if result.LoopCount > 2 {
		t.Errorf("expected max 2 loops, got %d", result.LoopCount)
	}
}

func TestAgent_MultipleToolCalls(t *testing.T) {
	// 创建返回多个工具调用的 MockLLM
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{
				StopReason: llm.StopReasonToolUse,
				ToolCalls: []*llm.ToolCall{
					llm.NewToolCall("call_1", "tool_a", "{}"),
					llm.NewToolCall("call_2", "tool_b", "{}"),
				},
			},
			{Content: "done", StopReason: llm.StopReasonEndTurn},
		},
	}

	toolA := tool.NewFunction("tool_a", "工具A", func(ctx context.Context, args map[string]any) (any, error) {
		return "result_a", nil
	})
	toolB := tool.NewFunction("tool_b", "工具B", func(ctx context.Context, args map[string]any) (any, error) {
		return "result_b", nil
	})

	ag := New(WithLLM(mockLLM), WithTools(toolA, toolB))

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 应该记录所有工具调用
	if len(result.ToolCalls) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(result.ToolCalls))
	}
}

func TestAgent_ToolError(t *testing.T) {
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{
				StopReason: llm.StopReasonToolUse,
				ToolCalls: []*llm.ToolCall{
					llm.NewToolCall("call_1", "error_tool", "{}"),
				},
			},
			{Content: "handled error", StopReason: llm.StopReasonEndTurn},
		},
	}

	errorTool := tool.NewFunction("error_tool", "错误工具", func(ctx context.Context, args map[string]any) (any, error) {
		return nil, errors.New("tool execution failed")
	})

	ag := New(WithLLM(mockLLM), WithTools(errorTool))

	result, err := ag.Run(context.Background(), "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Agent 应该处理工具错误并继续
	if result.LoopCount != 2 {
		t.Errorf("expected 2 loops, got %d", result.LoopCount)
	}
}

func TestAgent_LLMError(t *testing.T) {
	// 创建返回错误的 MockLLM
	errorLLM := &ErrorMockLLM{err: errors.New("LLM error")}

	ag := New(WithLLM(errorLLM))

	_, err := ag.Run(context.Background(), "test")
	if err == nil {
		t.Error("expected error from LLM")
	}
}

// ErrorMockLLM 模拟 LLM 错误
type ErrorMockLLM struct {
	err error
}

func (m *ErrorMockLLM) Complete(_ context.Context, _ *llm.ChatRequest) (*llm.ChatResponse, error) {
	return nil, m.err
}

func (m *ErrorMockLLM) Stream(_ context.Context, _ *llm.ChatRequest) (<-chan *llm.StreamChunk, error) {
	return nil, m.err
}

func TestHooks_OnLLMCall(t *testing.T) {
	llmCallCalled := false
	var capturedMessages []*llm.Message

	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{Content: "response", StopReason: llm.StopReasonEndTurn},
		},
	}

	hooks := &Hooks{
		OnLLMCall: func(messages []*llm.Message) {
			llmCallCalled = true
			capturedMessages = messages
		},
	}

	ag := New(WithLLM(mockLLM), WithHooks(hooks))

	_, err := ag.Run(context.Background(), "test input")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !llmCallCalled {
		t.Error("expected OnLLMCall hook to be called")
	}

	if len(capturedMessages) < 1 {
		t.Error("expected messages to be captured")
	}
}

func TestHooks_OnError(t *testing.T) {
	errorCalled := false

	errorLLM := &ErrorMockLLM{err: errors.New("test error")}

	hooks := &Hooks{
		OnError: func(err error, state *State) {
			errorCalled = true
		},
	}

	ag := New(WithLLM(errorLLM), WithHooks(hooks))

	_, _ = ag.Run(context.Background(), "test")

	if !errorCalled {
		t.Error("expected OnError hook to be called")
	}
}

func TestLoopController_MarkMethods(t *testing.T) {
	controller := NewLoopController(5)

	// 测试 MarkDone
	controller.MarkDone()
	if controller.State != LoopStateDone {
		t.Errorf("expected state Done, got %v", controller.State)
	}

	// 重置并测试 MarkError
	controller.State = LoopStateContinue
	controller.MarkError()
	if controller.State != LoopStateError {
		t.Errorf("expected state Error, got %v", controller.State)
	}

	// 重置并测试 MarkTimeout
	controller.State = LoopStateContinue
	controller.MarkTimeout()
	if controller.State != LoopStateTimeout {
		t.Errorf("expected state Timeout, got %v", controller.State)
	}
}

func TestLoopState_String(t *testing.T) {
	tests := []struct {
		state    LoopState
		expected string
	}{
		{LoopStateContinue, "continue"},
		{LoopStateDone, "done"},
		{LoopStateError, "error"},
		{LoopStateTimeout, "timeout"},
		{LoopStateMaxLoops, "max_loops"},
		{LoopState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("LoopState(%d).String() = %s, want %s", tt.state, got, tt.expected)
		}
	}
}

func TestNewLoopController_DefaultValue(t *testing.T) {
	// maxLoops <= 0 时应使用默认值 10
	controller := NewLoopController(0)
	if controller.MaxLoops != 10 {
		t.Errorf("expected default MaxLoops 10, got %d", controller.MaxLoops)
	}

	controller2 := NewLoopController(-5)
	if controller2.MaxLoops != 10 {
		t.Errorf("expected default MaxLoops 10, got %d", controller2.MaxLoops)
	}
}

// TestAgent_MemoryIntegration 测试 memory 集成
func TestAgent_MemoryIntegration(t *testing.T) {
	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{Content: "第一次回复", StopReason: llm.StopReasonEndTurn},
			{Content: "第二次回复", StopReason: llm.StopReasonEndTurn},
		},
	}

	// 创建滑动窗口 memory
	mem := memory.NewSlidingWindow(10)

	ag := New(WithLLM(mockLLM), WithSystemPrompt("你是助手"))
	ag.SetMemory(mem)

	ctx := context.Background()

	// 第一次对话
	_, err := ag.Run(ctx, "你好")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证 memory 中有消息
	if mem.Len() < 2 {
		t.Errorf("expected at least 2 messages in memory, got %d", mem.Len())
	}

	// 第二次对话
	_, err = ag.Run(ctx, "再见")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证 memory 持续累积
	if mem.Len() < 4 {
		t.Errorf("expected at least 4 messages in memory after second run, got %d", mem.Len())
	}
}

// TestAgent_MemoryWithContext 测试 memory 提供历史上下文
func TestAgent_MemoryWithContext(t *testing.T) {
	var callCount int
	var lastMessages []*llm.Message

	mockLLM := &MockLLM{
		responses: []*llm.ChatResponse{
			{Content: "回复1", StopReason: llm.StopReasonEndTurn},
			{Content: "回复2", StopReason: llm.StopReasonEndTurn},
		},
	}

	mem := memory.NewSlidingWindow(10)

	ag := New(
		WithLLM(mockLLM),
		WithHooks(&Hooks{
			OnLLMCall: func(messages []*llm.Message) {
				callCount++
				lastMessages = messages
			},
		}),
	)
	ag.SetMemory(mem)

	ctx := context.Background()

	// 第一次对话
	_, _ = ag.Run(ctx, "问题1")

	_ = callCount // 记录调用次数
	firstMsgCount := len(lastMessages)

	// 第二次对话 - 应该包含历史
	_, _ = ag.Run(ctx, "问题2")

	// 验证 LLM 调用时包含了历史消息
	if len(lastMessages) <= firstMsgCount {
		t.Errorf("expected more messages in second call due to history, got %d vs %d", len(lastMessages), firstMsgCount)
	}
}
