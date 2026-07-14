package feishu

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// InterceptedMessage 被劫持的消息
type InterceptedMessage struct {
	ID        string    `json:"id"`        // 唯一ID
	RobotGUID string    `json:"robotGuid"` // 机器人GUID
	MsgType   string    `json:"msgType"`   // 消息类型：text / interactive
	Content   string    `json:"content"`   // 消息内容（文本或JSON）
	Timestamp time.Time `json:"timestamp"` // 发送时间
}

// InterceptService 劫持服务
// 用于在测试阶段劫持飞书消息发送，在本地弹窗显示而非真正发送到飞书服务器
type InterceptService struct {
	messages       []InterceptedMessage
	mu             sync.RWMutex
	maxMessages    int                          // 最大保留消息数
	enabled        bool                         // 劫持开关
	onMessageAdded func(msg InterceptedMessage) // 新消息回调（用于推送事件到前端）
}

// globalInterceptService 全局劫持服务实例
var globalInterceptService *InterceptService
var globalOnce sync.Once

// NewInterceptService 创建劫持服务实例
func NewInterceptService() *InterceptService {
	globalOnce.Do(func() {
		globalInterceptService = &InterceptService{
			messages:    make([]InterceptedMessage, 0),
			maxMessages: 50,
			enabled:     false,
		}
	})
	return globalInterceptService
}

// GetInterceptService 获取全局劫持服务实例
func GetInterceptService() *InterceptService {
	if globalInterceptService == nil {
		return NewInterceptService()
	}
	return globalInterceptService
}

// SetEnabled 设置劫持开关
// @frontend
func (s *InterceptService) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// IsEnabled 检查劫持开关状态
// @frontend
func (s *InterceptService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// GetMessages 获取所有消息
// @frontend
func (s *InterceptService) GetMessages() []InterceptedMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]InterceptedMessage, len(s.messages))
	copy(result, s.messages)
	return result
}

// ClearMessages 清空消息
// @frontend
func (s *InterceptService) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = make([]InterceptedMessage, 0)
}

// AddMessage 添加消息（内部调用）
// 返回添加的消息，供调用方通过事件推送到前端
func (s *InterceptService) AddMessage(robotGUID, msgType, content string) InterceptedMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := InterceptedMessage{
		ID:        uuid.New().String(),
		RobotGUID: robotGUID,
		MsgType:   msgType,
		Content:   content,
		Timestamp: time.Now(),
	}

	// 超过最大数量时移除最旧的消息
	if len(s.messages) >= s.maxMessages {
		s.messages = s.messages[1:]
	}
	s.messages = append(s.messages, msg)
	return msg
}

// SetMaxMessages 设置最大保留消息数
func (s *InterceptService) SetMaxMessages(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if max > 0 {
		s.maxMessages = max
		// 如果当前消息数超过新上限，移除最旧的消息
		if len(s.messages) > s.maxMessages {
			s.messages = s.messages[len(s.messages)-s.maxMessages:]
		}
	}
}
