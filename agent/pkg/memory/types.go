package memory

import (
	"encoding/json"
	"time"
)

// Session 会话
// 表示一个完整的对话会话
type Session struct {
	// ID 会话唯一标识
	ID string `json:"id"`

	// Messages 会话消息列表
	Messages []*MessageWrapper `json:"messages"`

	// CreatedAt 创建时间戳
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt 更新时间戳
	UpdatedAt int64 `json:"updated_at"`

	// Metadata 会话元数据
	Metadata map[string]any `json:"metadata,omitempty"`
}

// MessageWrapper 消息包装器
// 在 llm.Message 基础上添加额外信息
type MessageWrapper struct {
	// ID 消息唯一标识
	ID string `json:"id"`

	// Role 消息角色
	Role string `json:"role"`

	// Content 消息内容（文本形式）
	Content string `json:"content"`

	// Timestamp 消息时间戳
	Timestamp int64 `json:"timestamp"`

	// TokenCount token 数量（可选）
	TokenCount int `json:"token_count,omitempty"`

	// Metadata 消息元数据
	Metadata map[string]any `json:"metadata,omitempty"`
}

// NewSession 创建新会话
func NewSession(id string) *Session {
	now := time.Now().Unix()
	return &Session{
		ID:        id,
		Messages:  make([]*MessageWrapper, 0),
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}
}

// AddMessage 添加消息
func (s *Session) AddMessage(msg *MessageWrapper) {
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now().Unix()
}

// Encoder 编码器接口
// 用于会话的序列化和反序列化
type Encoder interface {
	// Encode 序列化会话
	Encode(session *Session) ([]byte, error)

	// Decode 反序列化会话
	Decode(data []byte, session *Session) error
}

// JSONEncoder JSON 编码器
type JSONEncoder struct{}

// NewJSONEncoder 创建 JSON 编码器
func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{}
}

// Encode 序列化会话为 JSON
func (e *JSONEncoder) Encode(session *Session) ([]byte, error) {
	return json.Marshal(session)
}

// Decode 从 JSON 反序列化会话
func (e *JSONEncoder) Decode(data []byte, session *Session) error {
	return json.Unmarshal(data, session)
}

// Stats 记忆统计信息
type Stats struct {
	// TotalMessages 总消息数
	TotalMessages int

	// TotalTokens 总 token 数
	TotalTokens int

	// UserMessages 用户消息数
	UserMessages int

	// AssistantMessages 助手消息数
	AssistantMessages int

	// ToolMessages 工具消息数
	ToolMessages int

	// SystemMessages 系统消息数
	SystemMessages int
}
