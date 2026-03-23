package memory

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// FileMemory 文件持久化记忆
// 组合 SlidingWindow 并添加文件持久化能力
type FileMemory struct {
	*SlidingWindow
	filePath string
	mu       sync.RWMutex
}

// NewFileMemory 创建文件持久化记忆
// maxSize: 最大消息数量
// filePath: 持久化文件路径
func NewFileMemory(maxSize int, filePath string, opts ...Option) *FileMemory {
	fm := &FileMemory{
		SlidingWindow: NewSlidingWindow(maxSize, opts...),
		filePath:      filePath,
	}

	// 尝试加载已有数据
	ctx := context.Background()
	_ = fm.Load(ctx) // 忽略错误，文件可能不存在

	return fm
}

// Save 保存记忆到文件
// 实现 PersistentMemory 接口
func (m *FileMemory) Save(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	msgs, err := m.GetContext(ctx)
	if err != nil {
		return err
	}

	// 转换为可序列化格式
	data := &fileData{
		Messages: make([]*messageData, len(msgs)),
	}

	for i, msg := range msgs {
		data.Messages[i] = &messageData{
			Role:       msg.Role,
			Content:    llm.TextString(msg.Content),
			ToolCallID: msg.ToolCallID,
			Name:       msg.Name,
		}
		if len(msg.ToolCalls) > 0 {
			data.Messages[i].ToolCalls = msg.ToolCalls
		}
	}

	// 确保目录存在
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 序列化并写入文件
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, jsonData, 0644)
}

// Load 从文件加载记忆
// 实现 PersistentMemory 接口
func (m *FileMemory) Load(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 读取文件
	jsonData, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在不是错误
		}
		return err
	}

	var data fileData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	// 清空现有消息
	m.SlidingWindow.Clear(ctx)

	// 恢复消息
	for _, md := range data.Messages {
		msg := &llm.Message{
			Role:       md.Role,
			Content:    llm.Text(md.Content),
			ToolCallID: md.ToolCallID,
			Name:       md.Name,
			ToolCalls:  md.ToolCalls,
		}
		m.SlidingWindow.Add(ctx, msg)
	}

	return nil
}

// Add 添加单条消息并自动保存
func (m *FileMemory) Add(ctx context.Context, msg *llm.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.SlidingWindow.Add(ctx, msg); err != nil {
		return err
	}

	// 异步保存，避免阻塞
	go func() {
		_ = m.Save(context.Background())
	}()

	return nil
}

// AddBatch 批量添加消息并自动保存
func (m *FileMemory) AddBatch(ctx context.Context, msgs []*llm.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.SlidingWindow.AddBatch(ctx, msgs); err != nil {
		return err
	}

	// 异步保存
	go func() {
		_ = m.Save(context.Background())
	}()

	return nil
}

// Clear 清空记忆并删除文件
func (m *FileMemory) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.SlidingWindow.Clear(ctx); err != nil {
		return err
	}

	// 删除持久化文件
	if err := os.Remove(m.filePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// fileData 文件存储格式
type fileData struct {
	Messages []*messageData `json:"messages"`
}

// messageData 消息存储格式
type messageData struct {
	Role       llm.Role         `json:"role"`
	Content    string           `json:"content"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []*llm.ToolCall  `json:"tool_calls,omitempty"`
}
