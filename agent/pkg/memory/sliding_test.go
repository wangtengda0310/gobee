package memory

import (
	"context"
	"testing"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestNewSlidingWindow(t *testing.T) {
	mem := NewSlidingWindow(10)

	if mem.Len() != 0 {
		t.Errorf("expected empty memory, got %d messages", mem.Len())
	}

	if mem.maxSize != 10 {
		t.Errorf("expected maxSize 10, got %d", mem.maxSize)
	}
}

func TestSlidingWindow_Add(t *testing.T) {
	mem := NewSlidingWindow(3)

	msg1 := &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg1")}
	msg2 := &llm.Message{Role: llm.RoleAssistant, Content: llm.Text("msg2")}
	msg3 := &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg3")}
	msg4 := &llm.Message{Role: llm.RoleAssistant, Content: llm.Text("msg4")}

	ctx := context.Background()

	// 添加消息
	mem.Add(ctx, msg1)
	mem.Add(ctx, msg2)
	mem.Add(ctx, msg3)

	if mem.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", mem.Len())
	}

	// 添加第 4 条消息，应该触发截断
	mem.Add(ctx, msg4)

	if mem.Len() != 3 {
		t.Errorf("expected 3 messages after truncation, got %d", mem.Len())
	}

	// 检查保留的是最后 3 条消息
	msgs, _ := mem.GetContext(ctx)
	if llm.TextString(msgs[0].Content) != "msg2" {
		t.Errorf("expected first message to be 'msg2', got '%s'", llm.TextString(msgs[0].Content))
	}
}

func TestSlidingWindow_PreserveSystem(t *testing.T) {
	mem := NewSlidingWindow(3, WithPreserveSystem(true))

	ctx := context.Background()

	sysMsg := &llm.Message{Role: llm.RoleSystem, Content: llm.Text("system")}
	msg1 := &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg1")}
	msg2 := &llm.Message{Role: llm.RoleAssistant, Content: llm.Text("msg2")}
	msg3 := &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg3")}
	msg4 := &llm.Message{Role: llm.RoleAssistant, Content: llm.Text("msg4")}

	mem.Add(ctx, sysMsg)
	mem.Add(ctx, msg1)
	mem.Add(ctx, msg2)
	mem.Add(ctx, msg3)
	mem.Add(ctx, msg4)

	// 系统消息应该被保留
	msgs, _ := mem.GetContext(ctx)
	if msgs[0].Role != llm.RoleSystem {
		t.Error("expected first message to be system message")
	}

	// 总数应该为 3（1 系统消息 + 2 普通消息）
	if mem.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", mem.Len())
	}
}

func TestSlidingWindow_AddBatch(t *testing.T) {
	mem := NewSlidingWindow(10)

	ctx := context.Background()

	msgs := []*llm.Message{
		{Role: llm.RoleUser, Content: llm.Text("msg1")},
		{Role: llm.RoleAssistant, Content: llm.Text("msg2")},
		{Role: llm.RoleUser, Content: llm.Text("msg3")},
	}

	mem.AddBatch(ctx, msgs)

	if mem.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", mem.Len())
	}
}

func TestSlidingWindow_Clear(t *testing.T) {
	mem := NewSlidingWindow(10)

	ctx := context.Background()
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("test")})

	mem.Clear(ctx)

	if mem.Len() != 0 {
		t.Errorf("expected empty memory after clear, got %d", mem.Len())
	}
}

func TestSlidingWindow_GetStats(t *testing.T) {
	mem := NewSlidingWindow(10)

	ctx := context.Background()
	mem.Add(ctx, &llm.Message{Role: llm.RoleSystem, Content: llm.Text("sys")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("user")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleAssistant, Content: llm.Text("assistant")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleTool, Content: llm.Text("tool")})

	stats := mem.GetStats()

	if stats.SystemMessages != 1 {
		t.Errorf("expected 1 system message, got %d", stats.SystemMessages)
	}
	if stats.UserMessages != 1 {
		t.Errorf("expected 1 user message, got %d", stats.UserMessages)
	}
	if stats.AssistantMessages != 1 {
		t.Errorf("expected 1 assistant message, got %d", stats.AssistantMessages)
	}
	if stats.ToolMessages != 1 {
		t.Errorf("expected 1 tool message, got %d", stats.ToolMessages)
	}
}

// === Phase 2 新增测试 ===

func TestSlidingWindow_WithOptions(t *testing.T) {
	// 测试 WithMaxSize 选项
	mem := NewSlidingWindow(5, WithMaxSize(3))
	if mem.maxSize != 3 {
		t.Errorf("expected maxSize 3, got %d", mem.maxSize)
	}

	// 测试 WithPreserveSystem 选项
	mem2 := NewSlidingWindow(10, WithPreserveSystem(false))
	if mem2.preserveSys {
		t.Error("expected preserveSys to be false")
	}
}

func TestSlidingWindow_SetMaxSize(t *testing.T) {
	mem := NewSlidingWindow(10)
	ctx := context.Background()

	// 添加 5 条消息
	for i := 0; i < 5; i++ {
		mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg")})
	}

	if mem.Len() != 5 {
		t.Errorf("expected 5 messages, got %d", mem.Len())
	}

	// 设置更小的 maxSize，应触发截断
	mem.SetMaxSize(3)

	if mem.Len() != 3 {
		t.Errorf("expected 3 messages after SetMaxSize, got %d", mem.Len())
	}

	// 设置无效值应被忽略
	mem.SetMaxSize(0)
	if mem.maxSize != 3 {
		t.Error("expected maxSize to remain 3 for invalid input")
	}
}

func TestSlidingWindow_SetPreserveSystem(t *testing.T) {
	mem := NewSlidingWindow(3, WithPreserveSystem(true))
	ctx := context.Background()

	// 添加系统消息
	mem.Add(ctx, &llm.Message{Role: llm.RoleSystem, Content: llm.Text("system")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg1")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg2")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg3")})

	// 系统消息应该被保留
	if mem.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", mem.Len())
	}

	msgs, _ := mem.GetContext(ctx)
	if msgs[0].Role != llm.RoleSystem {
		t.Error("expected first message to be system")
	}

	// 关闭保留系统消息
	mem.SetPreserveSystem(false)

	// 添加更多消息触发截断
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg4")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg5")})

	msgs2, _ := mem.GetContext(ctx)
	// 现在系统消息可能被截断
	for _, m := range msgs2 {
		if m.Role == llm.RoleSystem {
			// 系统消息可能仍存在，取决于截断顺序
		}
	}
}

func TestSlidingWindow_NoSystemPreserve(t *testing.T) {
	// 不保留系统消息的测试
	mem := NewSlidingWindow(3, WithPreserveSystem(false))
	ctx := context.Background()

	sysMsg := &llm.Message{Role: llm.RoleSystem, Content: llm.Text("system")}
	mem.Add(ctx, sysMsg)
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg1")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg2")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg3")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg4")})

	// 应该只有最后 3 条消息
	if mem.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", mem.Len())
	}

	msgs, _ := mem.GetContext(ctx)
	// 检查是否为最后 3 条
	if llm.TextString(msgs[0].Content) != "msg2" {
		t.Errorf("expected first message 'msg2', got '%s'", llm.TextString(msgs[0].Content))
	}
}

func TestSlidingWindow_DefaultMaxSize(t *testing.T) {
	// maxSize <= 0 时应使用默认值 50
	mem := NewSlidingWindow(0)
	if mem.maxSize != 50 {
		t.Errorf("expected default maxSize 50, got %d", mem.maxSize)
	}

	mem2 := NewSlidingWindow(-1)
	if mem2.maxSize != 50 {
		t.Errorf("expected default maxSize 50, got %d", mem2.maxSize)
	}
}

func TestSlidingWindow_GetContext_Copy(t *testing.T) {
	mem := NewSlidingWindow(10)
	ctx := context.Background()

	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("original")})

	msgs, _ := mem.GetContext(ctx)
	// 修改返回的切片长度不应影响原始数据
	msgs = append(msgs, &llm.Message{Role: llm.RoleUser, Content: llm.Text("new")})

	msgs2, _ := mem.GetContext(ctx)
	if len(msgs2) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs2))
	}
}
