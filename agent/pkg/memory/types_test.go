package memory

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	session := NewSession("test-session-1")

	if session.ID != "test-session-1" {
		t.Errorf("expected ID 'test-session-1', got '%s'", session.ID)
	}

	if session.Messages == nil {
		t.Error("expected Messages to be initialized")
	}

	if len(session.Messages) != 0 {
		t.Errorf("expected empty messages, got %d", len(session.Messages))
	}

	if session.Metadata == nil {
		t.Error("expected Metadata to be initialized")
	}

	// 验证时间戳
	now := time.Now().Unix()
	if session.CreatedAt > now || session.CreatedAt < now-5 {
		t.Error("expected CreatedAt to be close to current time")
	}
}

func TestSession_AddMessage(t *testing.T) {
	session := NewSession("test-session")

	msg := &MessageWrapper{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now().Unix(),
	}

	session.AddMessage(msg)

	if len(session.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(session.Messages))
	}

	if session.Messages[0].ID != "msg-1" {
		t.Errorf("expected message ID 'msg-1', got '%s'", session.Messages[0].ID)
	}

	// 验证 UpdatedAt 被设置
	if session.UpdatedAt == 0 {
		t.Error("expected UpdatedAt to be set")
	}

	// 添加更多消息
	session.AddMessage(&MessageWrapper{ID: "msg-2", Role: "assistant", Content: "Hi"})
	session.AddMessage(&MessageWrapper{ID: "msg-3", Role: "user", Content: "How are you?"})

	if len(session.Messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(session.Messages))
	}
}

func TestJSONEncoder(t *testing.T) {
	encoder := NewJSONEncoder()

	session := NewSession("encode-test")
	session.AddMessage(&MessageWrapper{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Test message",
		Timestamp: 1234567890,
	})
	session.Metadata["key"] = "value"

	// 测试 Encode
	data, err := encoder.Encode(session)
	if err != nil {
		t.Errorf("unexpected encode error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty encoded data")
	}

	// 测试 Decode
	decoded := &Session{}
	err = encoder.Decode(data, decoded)
	if err != nil {
		t.Errorf("unexpected decode error: %v", err)
	}

	if decoded.ID != "encode-test" {
		t.Errorf("expected ID 'encode-test', got '%s'", decoded.ID)
	}

	if len(decoded.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(decoded.Messages))
	}

	if decoded.Messages[0].Content != "Test message" {
		t.Errorf("expected content 'Test message', got '%s'", decoded.Messages[0].Content)
	}

	if decoded.Metadata["key"] != "value" {
		t.Errorf("expected metadata key 'value', got '%v'", decoded.Metadata["key"])
	}
}

func TestJSONEncoder_InvalidDecode(t *testing.T) {
	encoder := NewJSONEncoder()

	// 测试无效 JSON
	decoded := &Session{}
	err := encoder.Decode([]byte("invalid json"), decoded)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMessageWrapper(t *testing.T) {
	msg := &MessageWrapper{
		ID:         "msg-123",
		Role:       "user",
		Content:    "Hello world",
		Timestamp:  1234567890,
		TokenCount: 10,
		Metadata: map[string]any{
			"source": "test",
		},
	}

	if msg.ID != "msg-123" {
		t.Errorf("expected ID 'msg-123', got '%s'", msg.ID)
	}

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", msg.Role)
	}

	if msg.Content != "Hello world" {
		t.Errorf("expected content 'Hello world', got '%s'", msg.Content)
	}

	if msg.TokenCount != 10 {
		t.Errorf("expected TokenCount 10, got %d", msg.TokenCount)
	}

	if msg.Metadata["source"] != "test" {
		t.Errorf("expected metadata source 'test', got '%v'", msg.Metadata["source"])
	}
}

func TestStats(t *testing.T) {
	stats := &Stats{
		TotalMessages:      10,
		TotalTokens:        500,
		UserMessages:       4,
		AssistantMessages:  4,
		ToolMessages:       1,
		SystemMessages:     1,
	}

	if stats.TotalMessages != 10 {
		t.Errorf("expected TotalMessages 10, got %d", stats.TotalMessages)
	}

	// 验证消息计数总和
	total := stats.UserMessages + stats.AssistantMessages + stats.ToolMessages + stats.SystemMessages
	if total != stats.TotalMessages {
		t.Errorf("message count mismatch: %d vs %d", total, stats.TotalMessages)
	}
}
