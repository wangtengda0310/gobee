package memory

import (
	"context"
	"sync"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// SlidingWindow 滑动窗口记忆
// 保留最近的 N 条消息，可选保留系统消息
type SlidingWindow struct {
	messages    []*llm.Message
	maxSize     int
	preserveSys bool
	compressor  Compressor
	mu          sync.RWMutex
}

// NewSlidingWindow 创建滑动窗口记忆
// maxSize: 最大消息数量
// opts: 可选配置
func NewSlidingWindow(maxSize int, opts ...Option) *SlidingWindow {
	if maxSize <= 0 {
		maxSize = 50
	}

	s := &SlidingWindow{
		messages:    make([]*llm.Message, 0),
		maxSize:     maxSize,
		preserveSys: true,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Add 添加单条消息
// 实现 Memory 接口
func (s *SlidingWindow) Add(_ context.Context, msg *llm.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)

	// 检查是否需要截断
	if len(s.messages) > s.maxSize {
		s.truncate()
	}

	return nil
}

// AddBatch 批量添加消息
// 实现 Memory 接口
func (s *SlidingWindow) AddBatch(_ context.Context, msgs []*llm.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msgs...)

	// 检查是否需要截断
	if len(s.messages) > s.maxSize {
		s.truncate()
	}

	return nil
}

// GetContext 获取当前上下文消息
// 实现 Memory 接口
func (s *SlidingWindow) GetContext(_ context.Context) ([]*llm.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回消息的副本，避免外部修改
	result := make([]*llm.Message, len(s.messages))
	copy(result, s.messages)
	return result, nil
}

// Clear 清空记忆
// 实现 Memory 接口
func (s *SlidingWindow) Clear(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = make([]*llm.Message, 0)
	return nil
}

// Len 返回消息数量
// 实现 Memory 接口
func (s *SlidingWindow) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}

// truncate 内部截断方法
// 调用前必须持有锁
func (s *SlidingWindow) truncate() {
	if len(s.messages) <= s.maxSize {
		return
	}

	// 如果需要保留系统消息
	if s.preserveSys {
		var sysMsgs []*llm.Message
		var otherMsgs []*llm.Message

		for _, msg := range s.messages {
			if msg.Role == llm.RoleSystem {
				sysMsgs = append(sysMsgs, msg)
			} else {
				otherMsgs = append(otherMsgs, msg)
			}
		}

		// 保留系统消息 + 最近的非系统消息
		keepCount := max(0, s.maxSize-len(sysMsgs))

		if len(otherMsgs) > keepCount {
			otherMsgs = otherMsgs[len(otherMsgs)-keepCount:]
		}

		s.messages = append(sysMsgs, otherMsgs...)
	} else {
		// 简单滑动窗口
		s.messages = s.messages[len(s.messages)-s.maxSize:]
	}
}

// GetStats 获取统计信息
func (s *SlidingWindow) GetStats() *Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &Stats{TotalMessages: len(s.messages)}

	for _, msg := range s.messages {
		switch msg.Role {
		case llm.RoleUser:
			stats.UserMessages++
		case llm.RoleAssistant:
			stats.AssistantMessages++
		case llm.RoleTool:
			stats.ToolMessages++
		case llm.RoleSystem:
			stats.SystemMessages++
		}
	}

	return stats
}

// SetMaxSize 设置最大消息数量
func (s *SlidingWindow) SetMaxSize(size int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if size > 0 {
		s.maxSize = size
		// 立即截断
		if len(s.messages) > s.maxSize {
			s.truncate()
		}
	}
}

// SetPreserveSystem 设置是否保留系统消息
func (s *SlidingWindow) SetPreserveSystem(preserve bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preserveSys = preserve
}
