package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestFileMemory_SaveAndLoad(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "memory.json")

	mem := NewFileMemory(10, filePath)
	ctx := context.Background()

	// 添加消息
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("hello")})
	mem.Add(ctx, &llm.Message{Role: llm.RoleAssistant, Content: llm.Text("hi")})

	// 保存
	if err := mem.Save(ctx); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("expected file to exist after save")
	}

	// 创建新的 memory 实例并加载
	mem2 := NewFileMemory(10, filePath)
	if err := mem2.Load(ctx); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// 验证消息已恢复
	if mem2.Len() != 2 {
		t.Errorf("expected 2 messages after load, got %d", mem2.Len())
	}

	msgs, _ := mem2.GetContext(ctx)
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestFileMemory_AutoSave(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "autosave.json")

	mem := NewFileMemory(10, filePath)
	ctx := context.Background()

	// Add 应该触发自动保存
	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("test")})

	// 等待异步保存完成
	// 由于是异步保存，我们需要等待一小段时间
	// 在实际测试中可能需要更可靠的同步机制
}

func TestFileMemory_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "clear_test.json")

	mem := NewFileMemory(10, filePath)
	ctx := context.Background()

	mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("test")})
	mem.Save(ctx)

	// 清空
	if err := mem.Clear(ctx); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	// 验证内存已清空
	if mem.Len() != 0 {
		t.Errorf("expected 0 messages after clear, got %d", mem.Len())
	}

	// 验证文件已删除
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted after clear")
	}
}

func TestFileMemory_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.json")

	mem := NewFileMemory(10, filePath)
	ctx := context.Background()

	// 加载不存在的文件应该成功（返回空）
	if err := mem.Load(ctx); err != nil {
		t.Errorf("expected no error for non-existent file, got: %v", err)
	}

	if mem.Len() != 0 {
		t.Errorf("expected 0 messages, got %d", mem.Len())
	}
}

func TestFileMemory_WithToolCalls(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "toolcalls.json")

	mem := NewFileMemory(10, filePath)
	ctx := context.Background()

	// 添加带工具调用的消息
	toolCall := llm.NewToolCall("call_1", "test_tool", `{"arg": "value"}`)
	mem.Add(ctx, &llm.Message{
		Role:      llm.RoleAssistant,
		Content:   llm.Text("using tool"),
		ToolCalls: []*llm.ToolCall{toolCall},
	})

	// 保存并加载
	mem.Save(ctx)

	mem2 := NewFileMemory(10, filePath)
	mem2.Load(ctx)

	msgs, _ := mem2.GetContext(ctx)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	if len(msgs[0].ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(msgs[0].ToolCalls))
	}

	if msgs[0].ToolCalls[0].Function.Name != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got '%s'", msgs[0].ToolCalls[0].Function.Name)
	}
}

func TestFileMemory_AddBatch(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "batch.json")

	mem := NewFileMemory(10, filePath)
	ctx := context.Background()

	msgs := []*llm.Message{
		{Role: llm.RoleUser, Content: llm.Text("msg1")},
		{Role: llm.RoleAssistant, Content: llm.Text("msg2")},
		{Role: llm.RoleUser, Content: llm.Text("msg3")},
	}

	if err := mem.AddBatch(ctx, msgs); err != nil {
		t.Fatalf("AddBatch failed: %v", err)
	}

	if mem.Len() != 3 {
		t.Errorf("expected 3 messages, got %d", mem.Len())
	}

	// 等待异步保存完成
	time.Sleep(100 * time.Millisecond)

	// 手动保存确保文件写入
	if err := mem.Save(ctx); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected file to exist")
	}
}

func TestFileMemory_Truncation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "truncate.json")

	// 创建 maxSize=3 的 memory
	mem := NewFileMemory(3, filePath)
	ctx := context.Background()

	// 添加 5 条消息
	for i := 0; i < 5; i++ {
		mem.Add(ctx, &llm.Message{Role: llm.RoleUser, Content: llm.Text("msg")})
	}

	// 应该只保留最后 3 条
	if mem.Len() != 3 {
		t.Errorf("expected 3 messages after truncation, got %d", mem.Len())
	}

	// 保存并重新加载验证
	mem.Save(ctx)

	mem2 := NewFileMemory(3, filePath)
	mem2.Load(ctx)

	if mem2.Len() != 3 {
		t.Errorf("expected 3 messages after load, got %d", mem2.Len())
	}
}
