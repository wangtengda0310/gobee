package home

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// EmittedEvent 记录一次 Emit 调用
type EmittedEvent struct {
	Name string
	Data []any
}

// MockEventEmitter 模拟事件发射器，记录所有 Emit 调用供断言
type MockEventEmitter struct {
	mu     sync.RWMutex
	events []EmittedEvent
}

// NewMockEventEmitter 创建模拟事件发射器
func NewMockEventEmitter() *MockEventEmitter {
	return &MockEventEmitter{}
}

// Emit 记录事件调用
func (m *MockEventEmitter) Emit(name string, data ...any) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, EmittedEvent{Name: name, Data: data})
	return false
}

// Events 返回所有记录的事件（线程安全副本）
func (m *MockEventEmitter) Events() []EmittedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]EmittedEvent, len(m.events))
	copy(result, m.events)
	return result
}

// EventsByName 按事件名过滤
func (m *MockEventEmitter) EventsByName(name string) []EmittedEvent {
	var result []EmittedEvent
	for _, e := range m.Events() {
		if e.Name == name {
			result = append(result, e)
		}
	}
	return result
}

// Reset 清空记录
func (m *MockEventEmitter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

// newTestChatService 创建不依赖 Wails App 的测试 ChatService
func newTestChatService(t *testing.T) (*ChatService, *MockEventEmitter) {
	t.Helper()
	tempDir := t.TempDir()
	mockEmitter := NewMockEventEmitter()

	mem := memory.NewFileMemory(100,
		filepath.Join(tempDir, "chat_history.json"),
		memory.WithPreserveSystem(true),
	)

	s := &ChatService{
		emitter:    mockEmitter,
		configFile: filepath.Join(tempDir, "app_chat_config.json"),
		memory:     &memoryAdapter{FileMemory: mem},
		registry:   tool.NewRegistry(),
	}
	return s, mockEmitter
}

// TestNoopEventEmitter 不 panic
func TestNoopEventEmitter_NoPanic(t *testing.T) {
	e := &noopEventEmitter{}
	result := e.Emit("test", "data")
	assert.False(t, result)
}
