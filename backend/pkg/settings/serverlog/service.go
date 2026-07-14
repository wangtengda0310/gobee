// internal/serverlog/service.go
package serverlog

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ServerLogService 服务端日志 Wails 服务
// 自动捕获 Go 后端 stdout/stderr 输出，// 并提供手动日志 API 供代码中标记重要事件
type ServerLogService struct {
	app    *application.App
	writer *PipeWriter
}

// NewServerLogService 创建服务端日志服务
// 不在此处初始化 writer，需要调用 InitWithApp 注入 app 实例后才能工作
func NewServerLogService() *ServerLogService {
	return &ServerLogService{}
}

// InitWithApp 初始化服务（在 main.go 中 app 创建后调用）
//
// ⚠️ 关键知识点：Go log 包引用捕获问题
//
// Go 的 log 包在 init 阶段就将 os.Stderr 存到内部字段 Logger.out 中。
// 之后替换 os.Stderr 变量只是改了 Go 全局变量，不会影响已保存的引用。
// 所以必须同时做三件事：
//  1. 替换 os.Stdout / os.Stderr 变量（影响 fmt.Printf 等直接引用变量的代码）
//  2. 调用 log.SetOutput() 重定向 log 包（影响所有 log.Printf 调用）
//  3. 后续新增日志输出建议统一使用 log.Printf 而非 fmt.Printf
func (s *ServerLogService) InitWithApp(app *application.App) {
	s.app = app
	s.writer = NewPipeWriter()

	pipeFile := s.writer.WriterFile()

	// 替换 os.Stdout 和 os.Stderr 为 pipe 的写端（*os.File）
	// 这只影响使用 os.Stdout/os.Stderr 变量的代码（如 fmt.Printf）
	os.Stdout = pipeFile
	os.Stderr = pipeFile

	// 重定向 log 包的默认 logger 到 pipe 写端
	// 关键：Go 的 log 包在 init 时就捕获了 os.Stderr 的引用，
	// 后续修改 os.Stderr 变量不会影响已初始化的 log.Logger.out 字段。
	// 必须显式调用 log.SetOutput 才能让 log.Printf 输出经过 pipe。
	log.SetOutput(pipeFile)

	// 启动后台 goroutine 从 PipeWriter 通道读取日志并发送到前端
	go s.forwardLogs()

	// 发送一条启动日志，验证前后端事件通道是否畅通
	s.Info("服务端日志面板已启动，开始捕获 stdout/stderr 输出")
}

// forwardLogs 从 PipeWriter 通道读取日志条目并发送到前端
// 通过 serverLog 事件推送，前端 use-server-logs.ts 监听此事件
// 注意：使用 map[string]any 而非 LogEntry 结构体，
// 因为 Wails v3 Event.Emit 无法正确序列化自定义结构体
func (s *ServerLogService) forwardLogs() {
	for entry := range s.writer.Entries() {
		if s.app != nil {
			s.app.Event.Emit("serverLog", entry.ToMap())
		}
	}
}

// GetLogStats 获取日志统计信息
// 返回当前待发送的日志数量
// @frontend
func (s *ServerLogService) GetLogStats() map[string]int {
	if s.writer == nil {
		return map[string]int{"pending": 0}
	}
	return map[string]int{
		"pending": len(s.writer.entries),
	}
}

// Debug 记录调试级别日志（手动标记）
// 手动日志在终端显示时会带 ▸ 前缀
// @frontend
func (s *ServerLogService) Debug(msg string) {
	s.emitManual("DEBUG", msg)
}

// Info 记录信息级别日志（手动标记）
// @frontend
func (s *ServerLogService) Info(msg string) {
	s.emitManual("INFO", msg)
}

// Warn 记录警告级别日志（手动标记）
// @frontend
func (s *ServerLogService) Warn(msg string) {
	s.emitManual("WARN", msg)
}

// Error 记录错误级别日志（手动标记）
// @frontend
func (s *ServerLogService) Error(msg string) {
	s.emitManual("ERROR", msg)
}

// Shutdown 关闭服务
// 在应用退出时调用
func (s *ServerLogService) Shutdown(ctx context.Context) {
	if s.writer != nil {
		s.writer.Close()
	}
}

// emitManual 发送手动标记的日志
// 直接发送到前端（不经过 pipe，避免循环），
// 同时输出到原始 stdout 以保持终端可见
func (s *ServerLogService) emitManual(level string, msg string) {
	entry := LogEntry{
		Level:     level,
		Message:   "▸ " + msg,
		Timestamp: time.Now().Format("15:04:05.000"),
		IsManual:  true,
	}

	// 直接发送到前端（不经过 pipe，避免循环）
	if s.app != nil {
		s.app.Event.Emit("serverLog", entry.ToMap())
	}

	// 同时输出到原始 stdout 以保持终端可见
	if s.writer != nil && s.writer.originalFile != nil {
		fmt.Fprintf(s.writer.originalFile, "[%s] [%s] %s\n", entry.Timestamp, level, msg)
	}
}
