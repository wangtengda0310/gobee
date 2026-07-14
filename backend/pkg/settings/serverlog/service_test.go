// internal/serverlog/service_test.go
package serverlog

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLogEntry_ToMap 验证 LogEntry.ToMap() 转换为 map[string]any 的正确性
func TestLogEntry_ToMap(t *testing.T) {
	t.Run("标准条目所有字段映射正确", func(t *testing.T) {
		entry := LogEntry{
			Level:     "ERROR",
			Message:   "连接数据库失败",
			Timestamp: "10:30:00.123",
			IsManual:  true,
		}
		m := entry.ToMap()

		assert.Equal(t, "ERROR", m["level"])
		assert.Equal(t, "连接数据库失败", m["message"])
		assert.Equal(t, "10:30:00.123", m["timestamp"])
		assert.Equal(t, true, m["isManual"])
	})

	t.Run("isManual 为 false", func(t *testing.T) {
		entry := LogEntry{
			Level:     "INFO",
			Message:   "常规日志",
			Timestamp: "10:30:00.000",
			IsManual:  false,
		}
		m := entry.ToMap()
		assert.Equal(t, false, m["isManual"])
	})

	t.Run("包含正确的 4 个 key", func(t *testing.T) {
		entry := LogEntry{Level: "DEBUG", Message: "test", Timestamp: "00:00:00.000", IsManual: false}
		m := entry.ToMap()

		expectedKeys := map[string]bool{
			"level": true, "message": true, "timestamp": true, "isManual": true,
		}
		for key := range m {
			assert.True(t, expectedKeys[key], "意外的 key: %s", key)
		}
		assert.Len(t, m, 4)
	})
}

// TestNewServerLogService 验证服务创建的初始状态
func TestNewServerLogService(t *testing.T) {
	svc := NewServerLogService()

	assert.NotNil(t, svc, "服务不应为 nil")
	assert.Nil(t, svc.app, "创建时 app 应为 nil")
	assert.Nil(t, svc.writer, "创建时 writer 应为 nil")
}

// TestServerLogService_GetLogStats_NilWriter 验证 writer 为 nil 时 GetLogStats 的安全性
func TestServerLogService_GetLogStats_NilWriter(t *testing.T) {
	svc := NewServerLogService()

	// 不调用 InitWithApp，writer 为 nil
	stats := svc.GetLogStats()

	assert.Equal(t, 0, stats["pending"], "writer 为 nil 时 pending 应为 0")
}

// TestServerLogService_emitManual_NoPanic 验证 emitManual 在 app 和 writer 为 nil 时不 panic
func TestServerLogService_emitManual_NoPanic(t *testing.T) {
	svc := NewServerLogService()

	// app 和 writer 均为 nil，调用不应 panic
	assert.NotPanics(t, func() {
		svc.emitManual("INFO", "测试消息")
	}, "emitManual 在 app 和 writer 为 nil 时不应 panic")
}

// TestPipeWriter_LogSetOutput_Integration 验证 PipeWriter 与 log.SetOutput 的集成
// 确认 log.Printf 输出经过 pipe 后能被正确读取
func TestPipeWriter_LogSetOutput_Integration(t *testing.T) {
	pw := NewPipeWriter()
	defer pw.Close()

	// 保存原始 log 输出并在测试结束后恢复
	originalOutput := log.Writer()
	defer log.SetOutput(originalOutput)

	// 将 log 输出重定向到 PipeWriter
	log.SetOutput(pw.WriterFile())

	// 通过 log.Printf 写入测试消息（使用 [DEBUG] 标签以便推断级别）
	log.Printf("[DEBUG] integration check")

	// 从 entries 通道读取，带超时保护
	select {
	case entry := <-pw.Entries():
		assert.Contains(t, entry.Message, "[DEBUG] integration check")
		assert.Equal(t, "DEBUG", entry.Level)
		assert.False(t, entry.IsManual)
		assert.NotEmpty(t, entry.Timestamp)
	case <-time.After(2 * time.Second):
		t.Fatal("等待 log.Printf 输出超时，pipe 可能未正确连接")
	}
}
