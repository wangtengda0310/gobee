package tool

import (
	"context"
	"sync"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// Registry 工具注册表
// 线程安全的工具管理器
type Registry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewRegistry 创建新的工具注册表
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册一个或多个工具
// 实现 Executor 接口
func (r *Registry) Register(tools ...Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, t := range tools {
		name := t.Name()
		if _, exists := r.tools[name]; exists {
			return NewToolError(name, "工具已存在", ErrToolAlreadyExists)
		}
		r.tools[name] = t
	}
	return nil
}

// Execute 执行指定名称的工具
// 实现 Executor 接口
func (r *Registry) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
	r.mu.RLock()
	t, exists := r.tools[name]
	r.mu.RUnlock()

	if !exists {
		return nil, NewToolError(name, "工具未找到", ErrToolNotFound)
	}

	return t.Execute(ctx, args)
}

// GetTool 获取指定名称的工具
// 实现 Executor 接口
func (r *Registry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	return t, exists
}

// ListTools 列出所有已注册的工具
// 实现 Executor 接口
func (r *Registry) ListTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetDefinitions 获取所有工具的定义
// 实现 Executor 接口
func (r *Registry) GetDefinitions() []*llm.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]*llm.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.Definition())
	}
	return defs
}

// Unregister 注销指定名称的工具
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

// Clear 清空所有工具
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]Tool)
}

// Count 返回已注册工具数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// MustRegister 注册工具，如果失败则 panic
// 用于初始化时注册工具
func (r *Registry) MustRegister(tools ...Tool) {
	if err := r.Register(tools...); err != nil {
		panic(err)
	}
}
