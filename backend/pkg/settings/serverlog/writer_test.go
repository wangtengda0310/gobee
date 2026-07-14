// internal/serverlog/writer_test.go
package serverlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"包含 [DEBUG]", "2026/04/02 [DEBUG] test message", "DEBUG"},
		{"包含 [WARN]", "2026/04/02 [WARN] warning msg", "WARN"},
		{"包含 [WARNING]", "2026/04/02 [WARNING] warning msg", "WARN"},
		{"包含 [ERROR]", "2026/04/02 [ERROR] error msg", "ERROR"},
		{"包含 [ERR]", "2026/04/02 [ERR] error msg", "ERROR"},
		{"包含 [INFO]", "2026/04/02 [INFO] info msg", "INFO"},
		{"默认为 INFO", "2026/04/02 some random output", "INFO"},
		{"空字符串", "", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPipeWriter(t *testing.T) {
	pw := NewPipeWriter()
	defer pw.Close()

	// 写入一行日志
	n, err := pw.Write([]byte("test log line\n"))
	assert.NoError(t, err)
	assert.Equal(t, 14, n)

	// 从通道读取
	entry := <-pw.Entries()
	assert.Equal(t, "test log line", entry.Message)
	assert.Equal(t, "INFO", entry.Level)
	assert.False(t, entry.IsManual)
}

func TestPipeWriterMultipleLines(t *testing.T) {
	pw := NewPipeWriter()
	defer pw.Close()

	// 写入多行
	pw.Write([]byte("line1\nline2\nline3\n"))

	entries := make([]LogEntry, 0, 3)
	for i := 0; i < 3; i++ {
		entry := <-pw.Entries()
		entries = append(entries, entry)
	}

	assert.Len(t, entries, 3)
	assert.Equal(t, "line1", entries[0].Message)
	assert.Equal(t, "line2", entries[1].Message)
	assert.Equal(t, "line3", entries[2].Message)
}

func TestPipeWriterLevelInference(t *testing.T) {
	pw := NewPipeWriter()
	defer pw.Close()

	pw.Write([]byte("[ERROR] something failed\n"))

	entry := <-pw.Entries()
	assert.Equal(t, "ERROR", entry.Level)
	assert.Contains(t, entry.Message, "[ERROR] something failed")
}

func TestPipeWriterIncompleteLine(t *testing.T) {
	pw := NewPipeWriter()
	defer pw.Close()

	// 写入不完整的行
	pw.Write([]byte("partial"))

	// 再写入补全的行
	pw.Write([]byte(" continued\n"))

	entry := <-pw.Entries()
	assert.Equal(t, "partial continued", entry.Message)
}
