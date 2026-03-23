package tool

import (
	"context"
	"testing"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestNewFunction(t *testing.T) {
	handler := func(ctx context.Context, args map[string]any) (any, error) {
		return "test result", nil
	}

	tool := NewFunction("test_tool", "测试工具", handler,
		WithStringParam("query", "查询内容", true),
	)

	if tool.Name() != "test_tool" {
		t.Errorf("expected name 'test_tool', got '%s'", tool.Name())
	}

	if tool.Description() != "测试工具" {
		t.Errorf("expected description '测试工具', got '%s'", tool.Description())
	}

	def := tool.Definition()
	if def.Function.Name != "test_tool" {
		t.Errorf("expected definition name 'test_tool', got '%s'", def.Function.Name)
	}
}

func TestFunctionTool_Execute(t *testing.T) {
	handler := func(ctx context.Context, args map[string]any) (any, error) {
		name, _ := args["name"].(string)
		return "Hello, " + name, nil
	}

	tool := NewFunction("greet", "问候工具", handler,
		WithStringParam("name", "名称", true),
	)

	result, err := tool.Execute(context.Background(), map[string]any{"name": "World"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("expected 'Hello, World', got '%v'", result)
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	tool1 := NewFunction("tool1", "工具1", nil)
	tool2 := NewFunction("tool2", "工具2", nil)

	// 测试注册
	if err := registry.Register(tool1, tool2); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 测试重复注册
	if err := registry.Register(tool1); err == nil {
		t.Error("expected error for duplicate registration")
	}

	// 测试获取工具
	got, exists := registry.GetTool("tool1")
	if !exists {
		t.Error("expected tool1 to exist")
	}
	if got.Name() != "tool1" {
		t.Errorf("expected 'tool1', got '%s'", got.Name())
	}

	// 测试列表
	tools := registry.ListTools()
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}

	// 测试定义
	defs := registry.GetDefinitions()
	if len(defs) != 2 {
		t.Errorf("expected 2 definitions, got %d", len(defs))
	}

	// 测试注销
	registry.Unregister("tool1")
	if registry.Count() != 1 {
		t.Errorf("expected 1 tool after unregister, got %d", registry.Count())
	}
}

func TestToolResult(t *testing.T) {
	// 测试成功结果
	result := &ToolResult{
		ToolCallID: "call_123",
		Name:       "test_tool",
		Result:     "success",
	}

	if result.IsError() {
		t.Error("expected no error")
	}

	m := result.ToMap()
	if m["success"] != true {
		t.Error("expected success to be true")
	}

	// 测试错误结果
	errResult := &ToolResult{
		ToolCallID: "call_456",
		Name:       "error_tool",
		Error:      ErrExecutionFailed,
	}

	if !errResult.IsError() {
		t.Error("expected error")
	}

	errMap := errResult.ToMap()
	if errMap["success"] != false {
		t.Error("expected success to be false")
	}
}

// === Phase 1 新增测试 ===

func TestWithNumberParam(t *testing.T) {
	tool := NewFunction("calc", "计算工具", nil,
		WithNumberParam("value", "数值参数", true),
	)

	def := tool.Definition()
	props, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	valueProp, ok := props["value"].(map[string]any)
	if !ok {
		t.Fatal("expected value property to be a map")
	}

	if valueProp["type"] != "number" {
		t.Errorf("expected type 'number', got '%v'", valueProp["type"])
	}

	// 检查 required
	required, ok := def.Function.Parameters["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "value" {
		t.Error("expected 'value' to be required")
	}
}

func TestWithBooleanParam(t *testing.T) {
	tool := NewFunction("toggle", "开关工具", nil,
		WithBooleanParam("enabled", "是否启用", false),
	)

	def := tool.Definition()
	props, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	enabledProp, ok := props["enabled"].(map[string]any)
	if !ok {
		t.Fatal("expected enabled property to be a map")
	}

	if enabledProp["type"] != "boolean" {
		t.Errorf("expected type 'boolean', got '%v'", enabledProp["type"])
	}

	// 检查非 required
	required, ok := def.Function.Parameters["required"].([]string)
	if ok && len(required) > 0 {
		t.Error("expected 'enabled' to not be required")
	}
}

func TestWithParameters(t *testing.T) {
	customParams := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "搜索查询",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "返回数量限制",
			},
		},
		"required": []string{"query"},
	}

	tool := NewFunction("search", "搜索工具", nil,
		WithParameters(customParams),
	)

	def := tool.Definition()
	props, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	if _, ok := props["query"]; !ok {
		t.Error("expected query property")
	}
	if _, ok := props["limit"]; !ok {
		t.Error("expected limit property")
	}
}

func TestFunctionTool_SetDescription(t *testing.T) {
	tool := NewFunction("test", "原始描述", nil)

	// 链式设置新描述
	tool.SetDescription("新描述")

	if tool.Description() != "新描述" {
		t.Errorf("expected '新描述', got '%s'", tool.Description())
	}
}

func TestFunctionTool_SetParameters(t *testing.T) {
	tool := NewFunction("test", "测试工具", nil)

	newParams := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}

	// 链式设置参数
	tool.SetParameters(newParams)

	def := tool.Definition()
	if def.Function.Parameters["type"] != "object" {
		t.Error("expected parameters type to be object")
	}
}

func TestFunctionTool_NoHandler(t *testing.T) {
	tool := NewFunction("no_handler", "无处理函数的工具", nil)

	_, err := tool.Execute(context.Background(), map[string]any{})
	if err != ErrNoHandler {
		t.Errorf("expected ErrNoHandler, got %v", err)
	}
}

func TestToolError(t *testing.T) {
	// 测试带 Cause 的错误
	cause := ErrToolNotFound
	toolErr := NewToolError("test_tool", "工具未找到", cause)

	// 测试 Error() 方法
	errStr := toolErr.Error()
	if errStr != "test_tool: 工具未找到: 工具未找到" {
		t.Errorf("unexpected error string: %s", errStr)
	}

	// 测试 Unwrap() 方法
	unwrapped := toolErr.Unwrap()
	if unwrapped != cause {
		t.Error("expected Unwrap to return cause")
	}

	// 测试不带 Cause 的错误
	toolErrNoCause := NewToolError("another_tool", "执行失败", nil)
	errStrNoCause := toolErrNoCause.Error()
	if errStrNoCause != "another_tool: 执行失败" {
		t.Errorf("unexpected error string: %s", errStrNoCause)
	}
}

func TestBatchResult(t *testing.T) {
	results := []*ToolResult{
		{ToolCallID: "call_1", Name: "tool_a", Result: "ok"},
		{ToolCallID: "call_2", Name: "tool_b", Error: ErrExecutionFailed},
		{ToolCallID: "call_3", Name: "tool_a", Result: "ok2"},
	}

	batch := &BatchResult{
		Results:      results,
		SuccessCount: 2,
		ErrorCount:   1,
	}

	// 测试 HasErrors
	if !batch.HasErrors() {
		t.Error("expected HasErrors to be true")
	}

	// 测试 GetByToolCallID
	res := batch.GetByToolCallID("call_2")
	if res == nil || res.Name != "tool_b" {
		t.Error("expected to find tool_b result")
	}

	notFound := batch.GetByToolCallID("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent id")
	}

	// 测试 GetByName
	toolAResults := batch.GetByName("tool_a")
	if len(toolAResults) != 2 {
		t.Errorf("expected 2 results for tool_a, got %d", len(toolAResults))
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	registry.Register(NewFunction("tool1", "工具1", nil))
	registry.Register(NewFunction("tool2", "工具2", nil))

	if registry.Count() != 2 {
		t.Errorf("expected 2 tools, got %d", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("expected 0 tools after clear, got %d", registry.Count())
	}
}

func TestRegistry_MustRegister(t *testing.T) {
	registry := NewRegistry()
	tool1 := NewFunction("must_tool", "必须注册的工具", nil)

	// 正常注册不应 panic
	registry.MustRegister(tool1)

	if registry.Count() != 1 {
		t.Errorf("expected 1 tool, got %d", registry.Count())
	}

	// 重复注册应 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate registration")
		}
	}()

	registry.MustRegister(tool1)
}

func TestRegistry_Execute_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}

	// 检查错误类型
	toolErr, ok := err.(*ToolError)
	if !ok {
		t.Errorf("expected ToolError, got %T", err)
	}
	if toolErr.Name != "nonexistent" {
		t.Errorf("expected tool name 'nonexistent', got '%s'", toolErr.Name)
	}
}

func TestBatchExecutor_ExecuteBatch(t *testing.T) {
	registry := NewRegistry()

	// 注册多个工具
	registry.Register(NewFunction("tool1", "工具1", func(ctx context.Context, args map[string]any) (any, error) {
		return "result1", nil
	}))
	registry.Register(NewFunction("tool2", "工具2", func(ctx context.Context, args map[string]any) (any, error) {
		return "result2", nil
	}))

	executor := NewBatchExecutor(registry, 2)

	calls := []*llm.ToolCall{
		llm.NewToolCall("call_1", "tool1", "{}"),
		llm.NewToolCall("call_2", "tool2", "{}"),
	}

	batch := executor.ExecuteBatch(context.Background(), calls)

	if batch.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", batch.SuccessCount)
	}

	if batch.ErrorCount != 0 {
		t.Errorf("expected 0 errors, got %d", batch.ErrorCount)
	}
}

func TestBatchExecutor_ExecuteSequential(t *testing.T) {
	registry := NewRegistry()

	var executionOrder []string

	registry.Register(NewFunction("first", "第一个工具", func(ctx context.Context, args map[string]any) (any, error) {
		executionOrder = append(executionOrder, "first")
		return "r1", nil
	}))
	registry.Register(NewFunction("second", "第二个工具", func(ctx context.Context, args map[string]any) (any, error) {
		executionOrder = append(executionOrder, "second")
		return "r2", nil
	}))

	executor := NewBatchExecutor(registry, 4)

	calls := []*llm.ToolCall{
		llm.NewToolCall("call_1", "first", "{}"),
		llm.NewToolCall("call_2", "second", "{}"),
	}

	batch := executor.ExecuteSequential(context.Background(), calls)

	if batch.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", batch.SuccessCount)
	}

	// 验证顺序执行
	if len(executionOrder) != 2 || executionOrder[0] != "first" || executionOrder[1] != "second" {
		t.Errorf("expected sequential execution order, got %v", executionOrder)
	}
}

func TestBatchExecutor_Registry(t *testing.T) {
	registry := NewRegistry()
	executor := NewBatchExecutor(registry, 4)

	if executor.Registry() != registry {
		t.Error("expected Registry() to return the underlying registry")
	}
}

func TestBatchExecutor_Empty(t *testing.T) {
	registry := NewRegistry()
	executor := NewBatchExecutor(registry, 4)

	// 空调用
	batch := executor.ExecuteBatch(context.Background(), nil)
	if batch == nil || len(batch.Results) != 0 {
		t.Error("expected empty batch result")
	}

	batchSeq := executor.ExecuteSequential(context.Background(), nil)
	if batchSeq == nil || len(batchSeq.Results) != 0 {
		t.Error("expected empty batch result")
	}
}
