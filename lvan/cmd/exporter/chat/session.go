package chat

import (
	"sync"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

const (
	defaultMaxMessages = 100               // 默认最大消息数
	defaultSessionTTL  = 1 * time.Hour     // 默认会话过期时间
	defaultMaxSessions = 1000              // 默认最大会话数
	cleanupInterval    = 10 * time.Minute  // 清理间隔
)

// Session 单个聊天会话
type Session struct {
	Messages   []*llm.Message
	LastActive time.Time
}

// SessionManager 会话管理器
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	maxMessages int
	sessionTTL  time.Duration
	maxSessions int
}

// SessionOption 会话管理器配置选项
type SessionOption func(*SessionManager)

// WithMaxMessages 设置最大消息数
func WithMaxMessages(n int) SessionOption {
	return func(sm *SessionManager) { sm.maxMessages = n }
}

// NewSessionManager 创建会话管理器
func NewSessionManager(opts ...SessionOption) *SessionManager {
	sm := &SessionManager{
		sessions:    make(map[string]*Session),
		maxMessages: defaultMaxMessages,
		sessionTTL:  defaultSessionTTL,
		maxSessions: defaultMaxSessions,
	}
	for _, opt := range opts {
		opt(sm)
	}
	go sm.cleanupLoop()
	return sm
}

// AddMessage 向指定会话追加消息
func (sm *SessionManager) AddMessage(sessionID string, msg *llm.Message) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, ok := sm.sessions[sessionID]
	if !ok {
		if len(sm.sessions) >= sm.maxSessions {
			sm.evictOldest()
		}
		s = &Session{}
		sm.sessions[sessionID] = s
	}
	s.Messages = append(s.Messages, msg)
	s.LastActive = time.Now()

	if len(s.Messages) > sm.maxMessages {
		s.Messages = s.Messages[len(s.Messages)-sm.maxMessages:]
	}
}

// GetMessages 获取指定会话的所有消息
func (sm *SessionManager) GetMessages(sessionID string) []*llm.Message {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, ok := sm.sessions[sessionID]
	if !ok {
		return nil
	}
	result := make([]*llm.Message, len(s.Messages))
	copy(result, s.Messages)
	return result
}

// ClearSession 清空指定会话的消息
func (sm *SessionManager) ClearSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// cleanupLoop 定期清理过期会话
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		sm.cleanupExpired()
	}
}

// cleanupExpired 清理过期会话
func (sm *SessionManager) cleanupExpired() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, s := range sm.sessions {
		if now.Sub(s.LastActive) > sm.sessionTTL {
			delete(sm.sessions, id)
		}
	}
}

// evictOldest 驱逐最早未活跃的会话（调用方需持有锁）
func (sm *SessionManager) evictOldest() {
	var oldestID string
	var oldestTime time.Time
	for id, s := range sm.sessions {
		if oldestID == "" || s.LastActive.Before(oldestTime) {
			oldestID = id
			oldestTime = s.LastActive
		}
	}
	if oldestID != "" {
		delete(sm.sessions, oldestID)
	}
}
