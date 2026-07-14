// internal/serverlog/writer.go
package serverlog

import (
	"bufio"
	"os"
	"strings"
	"time"
)

// PipeWriter 使用 os.Pipe 拦截 stdout 输出
// 将写端赋值给 os.Stdout（类型 *os.File），从读端读取日志行
type PipeWriter struct {
	reader       *os.File      // pipe 读端
	writer       *os.File      // pipe 写端（赋值给 os.Stdout）
	originalFile *os.File      // 原始 os.Stdout（供手动日志输出到终端）
	entries      chan LogEntry // 日志条目通道
	done         chan struct{} // 关闭信号
}

// NewPipeWriter 创建 PipeWriter
// 保存原始 stdout 引用，创建 pipe，启动消费 goroutine
func NewPipeWriter() *PipeWriter {
	r, w, err := os.Pipe()
	if err != nil {
		panic("serverlog: failed to create pipe: " + err.Error())
	}

	pw := &PipeWriter{
		reader:       r,
		writer:       w,
		originalFile: os.Stdout, // 保存原始 stdout 供后续终端输出使用
		entries:      make(chan LogEntry, 200),
		done:         make(chan struct{}),
	}

	// 启动消费 goroutine
	go pw.consume()

	return pw
}

// WriterFile 返回写端 *os.File，用于赋值给 os.Stdout
func (pw *PipeWriter) WriterFile() *os.File {
	return pw.writer
}

// Entries 返回日志条目通道，供 service 层读取
func (pw *PipeWriter) Entries() <-chan LogEntry {
	return pw.entries
}

// Write 实现 io.Writer 接口（供直接写入使用）
func (pw *PipeWriter) Write(p []byte) (n int, err error) {
	return pw.writer.Write(p)
}

// Close 关闭 pipe，释放资源
func (pw *PipeWriter) Close() {
	pw.writer.Close()
	close(pw.done)
}

// consume 从 pipe 读端读取日志行，推断级别后发送到通道
func (pw *PipeWriter) consume() {
	scanner := bufio.NewScanner(pw.reader)
	// 允许较长的单行
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		level := inferLogLevel(line)
		pw.entries <- LogEntry{
			Level:     level,
			Message:   line,
			Timestamp: time.Now().Format("15:04:05.000"),
			IsManual:  false,
		}
	}

	// scanner 结束后关闭通道
	close(pw.entries)
}

// inferLogLevel 从日志内容推断日志级别
func inferLogLevel(content string) string {
	upper := strings.ToUpper(content)
	switch {
	case strings.Contains(upper, "[DEBUG]"):
		return "DEBUG"
	case strings.Contains(upper, "[WARN]") || strings.Contains(upper, "[WARNING]"):
		return "WARN"
	case strings.Contains(upper, "[ERROR]") || strings.Contains(upper, "[ERR]"):
		return "ERROR"
	case strings.Contains(upper, "[INFO]"):
		return "INFO"
	default:
		return "INFO"
	}
}
