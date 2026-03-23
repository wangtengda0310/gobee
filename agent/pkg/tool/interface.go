package tool

import (
	"context"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// Tool 工具接口
// 定义了 Agent 可调用工具的标准接口
type Tool interface {
	// Definition 返回工具定义，用于 LLM API 调用
	Definition() *llm.Tool

	// Execute 执行工具
	// ctx: 上下文，用于取消和超时控制
	// args: 工具参数，从 LLM 的 tool_call 中解析
	// 返回: 执行结果或错误
	Execute(ctx context.Context, args map[string]any) (any, error)

	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string
}

// Executor 工具执行器接口
// 管理多个工具的注册和执行
type Executor interface {
	// Register 注册一个或多个工具
	// 如果工具名称已存在，返回错误
	Register(tools ...Tool) error

	// Execute 执行指定名称的工具
	// 如果工具不存在，返回错误
	Execute(ctx context.Context, name string, args map[string]any) (any, error)

	// GetTool 获取指定名称的工具
	// 如果工具不存在，返回 nil, false
	GetTool(name string) (Tool, bool)

	// ListTools 列出所有已注册的工具
	ListTools() []Tool

	// GetDefinitions 获取所有工具的定义
	// 用于构建 LLM 请求的 tools 参数
	GetDefinitions() []*llm.Tool
}
