package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestSessionManager_AddAndGet(t *testing.T) {
	sm := NewSessionManager()

	msgs := sm.GetMessages("session-1")
	assert.Equal(t, 0, len(msgs))

	sm.AddMessage("session-1", &llm.Message{
		Role:    llm.RoleUser,
		Content: llm.Text("hello"),
	})

	msgs = sm.GetMessages("session-1")
	require.Equal(t, 1, len(msgs))
	assert.Equal(t, llm.RoleUser, msgs[0].Role)
}

func TestSessionManager_Clear(t *testing.T) {
	sm := NewSessionManager()

	sm.AddMessage("s1", &llm.Message{Role: llm.RoleUser, Content: llm.Text("hi")})
	sm.ClearSession("s1")

	msgs := sm.GetMessages("s1")
	assert.Equal(t, 0, len(msgs))
}

func TestSessionManager_CleanupExpired(t *testing.T) {
	sm := NewSessionManager()

	sm.mu.Lock()
	sm.sessions["expired"] = &Session{
		Messages:   []*llm.Message{{Role: llm.RoleUser, Content: llm.Text("old")}},
		LastActive: time.Now().Add(-2 * time.Hour),
	}
	sm.sessions["active"] = &Session{
		Messages:   []*llm.Message{{Role: llm.RoleUser, Content: llm.Text("new")}},
		LastActive: time.Now(),
	}
	sm.mu.Unlock()

	sm.cleanupExpired()

	assert.Equal(t, 0, len(sm.GetMessages("expired")))
	assert.Equal(t, 1, len(sm.GetMessages("active")))
}

func TestSessionManager_TrimMessages(t *testing.T) {
	sm := NewSessionManager(WithMaxMessages(5))

	for i := range 8 {
		sm.AddMessage("s1", &llm.Message{
			Role:    llm.RoleUser,
			Content: llm.Text("msg"),
		})
		_ = i
	}

	msgs := sm.GetMessages("s1")
	assert.Equal(t, 5, len(msgs), "应裁剪到最大消息数")
}
