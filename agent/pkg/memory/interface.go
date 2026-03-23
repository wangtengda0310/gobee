package memory

import (
	"context"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// Memory 记忆接口
// 定义了对话历史管理的基本操作
type Memory interface {
	// Add 添加单条消息
	Add(ctx context.Context, msg *llm.Message) error

	// AddBatch 批量添加消息
	AddBatch(ctx context.Context, msgs []*llm.Message) error

	// GetContext 获取当前上下文消息
	// 返回的消息列表可直接用于 LLM 请求
	GetContext(ctx context.Context) ([]*llm.Message, error)

	// Clear 清空记忆
	Clear(ctx context.Context) error

	// Len 返回消息数量
	Len() int
}

// PersistentMemory 持久化记忆接口
// 扩展 Memory 接口，支持持久化存储
type PersistentMemory interface {
	Memory

	// Save 保存记忆到存储
	Save(ctx context.Context) error

	// Load 从存储加载记忆
	Load(ctx context.Context) error
}

// Compressor 记忆压缩器接口
// 用于压缩长对话历史
type Compressor interface {
	// Compress 压缩消息列表
	// 返回压缩后的消息列表
	Compress(ctx context.Context, messages []*llm.Message) ([]*llm.Message, error)
}
