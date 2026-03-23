package tool

import (
	"context"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// HandlerFunc 工具处理函数类型
// ctx: 上下文，用于取消和超时控制
// args: 从 LLM tool_call.arguments 解析的参数
// 返回: 任意可 JSON 序列化的结果或错误
type HandlerFunc func(ctx context.Context, args map[string]any) (any, error)

// FunctionTool 函数工具实现
// 最常用的 Tool 实现，包装一个 Go 函数
type FunctionTool struct {
	name        string
	description string
	parameters  map[string]any
	handler     HandlerFunc
}

// NewFunction 创建函数工具
// name: 工具名称，应使用 snake_case 格式
// description: 工具描述，清晰说明工具的用途
// handler: 工具处理函数
// opts: 可选配置，如 WithParameters, WithStringParam 等
//
// 使用示例:
//
//	tool := NewFunction("get_weather", "获取天气信息",
//	    func(ctx context.Context, args map[string]any) (any, error) {
//	        city := args["city"].(string)
//	        return map[string]any{"city": city, "temp": 25}, nil
//	    },
//	    WithStringParam("city", "城市名称", true),
//	)
func NewFunction(name, description string, handler HandlerFunc, opts ...FunctionOption) *FunctionTool {
	f := &FunctionTool{
		name:        name,
		description: description,
		handler:     handler,
		parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Definition 返回工具定义
// 实现 Tool 接口
func (f *FunctionTool) Definition() *llm.Tool {
	return llm.NewTool(f.name, f.description, f.parameters)
}

// Execute 执行工具
// 实现 Tool 接口
func (f *FunctionTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	if f.handler == nil {
		return nil, ErrNoHandler
	}
	return f.handler(ctx, args)
}

// Name 返回工具名称
// 实现 Tool 接口
func (f *FunctionTool) Name() string {
	return f.name
}

// Description 返回工具描述
// 实现 Tool 接口
func (f *FunctionTool) Description() string {
	return f.description
}

// SetDescription 设置描述
// 用于链式构建
func (f *FunctionTool) SetDescription(desc string) *FunctionTool {
	f.description = desc
	return f
}

// SetParameters 设置参数定义
// 用于链式构建
func (f *FunctionTool) SetParameters(params map[string]any) *FunctionTool {
	f.parameters = params
	return f
}
