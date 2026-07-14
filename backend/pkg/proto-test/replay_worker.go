package prototest

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// ReplayProgress 重放进度（返回给前端）
type ReplayProgress struct {
	Status       string `json:"status"`        // "idle" | "running" | "completed" | "error" | "cancelled"
	Total        int    `json:"total"`         // 总消息数
	Sent         int    `json:"sent"`          // 已发送数
	CurrentMsg   string `json:"current_msg"`   // 当前发送的消息名
	ErrorMessage string `json:"error_message"` // 错误信息（仅 error 状态）
}

// ReplayWorker 异步重放工作器
// 不直接 import streamproto，避免 protoMsg init() 与 rain-robot/xcard_pb 冲突。
// 通过 SetReplayFunc 注入实际的重放函数。
type ReplayWorker struct {
	mu         sync.Mutex
	emitter    EventEmitter
	cancel     context.CancelFunc
	progress   ReplayProgress
	running    bool
	sentCount  atomic.Int64                           // 并发安全的已发送计数
	connPool   func() *protocol.AccountConnectionPool // 延迟获取连接池
	replayFunc func(filePath string, serverAddr string, httpAddr string, openID string, repeatCount int, onProgress func(total, sent int, currentMsg string) bool) error
	sendFunc   func(serverAddr, httpAddr, openID string, messagesJSON string, repeatCount int, rangeStart, rangeEnd int, onProgress func(total, sent int, currentMsg string) bool) error
}

// NewReplayWorker 创建重放工作器
// emitter 可以是 Wails application.App.Event，也可以是任何实现了 EventEmitter 接口的对象。
func NewReplayWorker(emitter EventEmitter) *ReplayWorker {
	return &ReplayWorker{
		emitter: emitter,
		progress: ReplayProgress{
			Status: "idle",
		},
	}
}

// SetReplayFunc 注入实际的重放函数（由 main.go 调用，避免 wails.go 传递依赖 protoMsg）
func (w *ReplayWorker) SetReplayFunc(f func(filePath string, serverAddr string, httpAddr string, openID string, repeatCount int, onProgress func(total, sent int, currentMsg string) bool) error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.replayFunc = f
}

// SetSendFunc 注入消息发送函数
func (w *ReplayWorker) SetSendFunc(f func(serverAddr, httpAddr, openID string, messagesJSON string, repeatCount int, rangeStart, rangeEnd int, onProgress func(total, sent int, currentMsg string) bool) error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sendFunc = f
}

// SetConnPoolFactory 设置连接池工厂（延迟获取，避免循环依赖）
func (w *ReplayWorker) SetConnPoolFactory(f func() *protocol.AccountConnectionPool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.connPool = f
}

// IsRunning 是否正在重放
func (w *ReplayWorker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// GetProgress 获取当前进度
func (w *ReplayWorker) GetProgress() ReplayProgress {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.progress
}

// Start 启动异步重放
func (w *ReplayWorker) Start(filePath string, serverAddr string, httpAddr string, openID string, repeatCount int) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("已有重放任务正在运行")
	}
	if w.replayFunc == nil {
		w.mu.Unlock()
		return fmt.Errorf("重放函数未设置")
	}
	w.running = true
	w.progress = ReplayProgress{
		Status: "running",
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go w.run(ctx, filePath, serverAddr, httpAddr, openID, repeatCount)
	return nil
}

// StartSend 启动异步消息发送（前端"开始重放"/"执行用例"/"重发"共用）
func (w *ReplayWorker) StartSend(serverAddr, httpAddr, openID string, messagesJSON string, repeatCount int, rangeStart, rangeEnd int) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("已有重放任务正在运行")
	}
	if w.sendFunc == nil {
		w.mu.Unlock()
		return fmt.Errorf("发送函数未设置")
	}
	w.running = true
	w.progress = ReplayProgress{
		Status: "running",
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go w.runSend(ctx, serverAddr, httpAddr, openID, messagesJSON, repeatCount, rangeStart, rangeEnd)
	return nil
}

// StartInject 通过录制代理连接注入消息（录制活跃期间同账号重发场景）
func (w *ReplayWorker) StartInject(messagesJSON string, repeatCount int, recordWorker *RecordWorker) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("已有重放任务正在运行")
	}
	w.running = true
	w.progress = ReplayProgress{
		Status: "running",
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	go w.runInject(ctx, messagesJSON, repeatCount, recordWorker)
	return nil
}

// Stop 停止重放
func (w *ReplayWorker) Stop() {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()
	log.Printf("[ReplayWorker] 已请求停止重放")
}

// emitProgress 推送进度到前端（Event.Emit data 必须用 map[string]any）
func (w *ReplayWorker) emitProgress() {
	w.mu.Lock()
	p := w.progress
	w.mu.Unlock()

	w.emitter.Emit("replay:progress", map[string]any{
		"status":        p.Status,
		"total":         p.Total,
		"sent":          p.Sent,
		"latest_msg":    p.CurrentMsg, // 前端期望 "latest_msg" 字段名
		"error_message": p.ErrorMessage,
	})
}

// setError 设置错误状态并推送
func (w *ReplayWorker) setError(msg string) {
	w.mu.Lock()
	w.progress.Status = "error"
	w.progress.ErrorMessage = msg
	w.mu.Unlock()
	w.emitProgress()
	log.Printf("[ReplayWorker] 错误: %s", msg)
}

// finish 清理运行状态
func (w *ReplayWorker) finish() {
	w.mu.Lock()
	w.running = false
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.mu.Unlock()
}

// EmitReplayMessage 推送到前端（用 record:progress 事件以复用前端表格追加逻辑）
// 使用 RecordEntryView 合约类型，确保字段名与前端 bindings 一致
// accountID: 当前账号标识（如 test1、test2）
func (w *ReplayWorker) EmitReplayMessage(name string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) {
	log.Printf("[EmitReplayMessage] name=%s, msgID=%d, seqID=%d, direction=%s, account=%s → 发射 record:progress + replay:result", name, msgID, seqID, direction, accountID)

	// 构建合约类型（唯一转换点：payloadJSON string → RecordEntryView.Payload map）
	entry := entryViewFromJSON(name, msgID, seqID, payloadJSON, offsetMs, direction, accountID)

	// 给发包改包页签的录制表格用
	w.emitter.Emit("record:progress", map[string]any{
		"status":        "running",
		"message_count": 0,
		"server_addr":   "",
		"error_message": "",
		"latest_msg":    entry,
	})

	// 给重放结果页签专用
	w.emitter.Emit("replay:result", entry)
}

// runSend 异步执行消息发送
func (w *ReplayWorker) runSend(ctx context.Context, serverAddr, httpAddr, openID string, messagesJSON string, repeatCount int, rangeStart, rangeEnd int) {
	defer w.finish()

	log.Printf("[ReplayWorker] 开始发送消息: server=%s, repeatCount=%d", serverAddr, repeatCount)

	// 重置原子计数器
	w.sentCount.Store(0)

	done := make(chan error, 1)
	go func() {
		done <- w.sendFunc(serverAddr, httpAddr, openID, messagesJSON, repeatCount, rangeStart, rangeEnd, func(total, sent int, currentMsg string) bool {
			// 并发安全：用原子计数器累加，不用 mutex（避免多 goroutine 锁竞争）
			if currentMsg != "" {
				// 单条消息发送回调：原子 +1
				w.sentCount.Add(1)
			}
			newSent := int(w.sentCount.Load())

			w.mu.Lock()
			w.progress.Total = total
			w.progress.Sent = newSent
			if currentMsg != "" {
				w.progress.CurrentMsg = currentMsg
			}
			w.mu.Unlock()
			w.emitProgress()
			w.mu.Lock()
			cancelled := w.progress.Status == "cancelled"
			w.mu.Unlock()
			return !cancelled
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			w.setError(err.Error())
			return
		}
		w.mu.Lock()
		w.progress.Status = "completed"
		w.mu.Unlock()
		w.emitProgress()
		log.Printf("[ReplayWorker] 消息发送完成")
	case <-ctx.Done():
		w.mu.Lock()
		w.progress.Status = "cancelled"
		w.mu.Unlock()
		w.emitProgress()
		log.Printf("[ReplayWorker] 消息发送已取消")
	}
}

// runInject 通过录制代理连接注入消息
// run 异步执行重放
func (w *ReplayWorker) run(ctx context.Context, filePath string, serverAddr string, httpAddr string, openID string, repeatCount int) {
	defer w.finish()

	log.Printf("[ReplayWorker] 开始重放: file=%s, server=%s, http=%s, openID=%s, repeatCount=%d",
		filePath, serverAddr, httpAddr, openID, repeatCount)

	done := make(chan error, 1)
	go func() {
		done <- w.replayFunc(filePath, serverAddr, httpAddr, openID, repeatCount, func(total, sent int, currentMsg string) bool {
			w.mu.Lock()
			w.progress.Total = total
			w.progress.Sent = sent
			w.progress.CurrentMsg = currentMsg
			w.mu.Unlock()
			w.emitProgress()
			// 检查是否被取消
			w.mu.Lock()
			cancelled := w.progress.Status == "cancelled"
			w.mu.Unlock()
			return !cancelled
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			w.setError(err.Error())
			return
		}
		w.mu.Lock()
		w.progress.Status = "completed"
		w.mu.Unlock()
		w.emitProgress()
		log.Printf("[ReplayWorker] 重放完成")
	case <-ctx.Done():
		w.mu.Lock()
		w.progress.Status = "cancelled"
		w.mu.Unlock()
		w.emitProgress()
		log.Printf("[ReplayWorker] 重放已取消（注意：底层 Replay 可能仍在执行）")
	}
}

// runInject 通过录制代理注入消息
func (w *ReplayWorker) runInject(ctx context.Context, messagesJSON string, repeatCount int, recordWorker *RecordWorker) {
	defer w.finish()

	log.Printf("[ReplayWorker] 开始注入消息: repeatCount=%d", repeatCount)

	w.sentCount.Store(0)

	done := make(chan error, 1)
	go func() {
		_, err := recordWorker.InjectMessages(messagesJSON, repeatCount,
			func(total, sent int, currentMsg string) bool {
				if currentMsg != "" {
					w.sentCount.Add(1)
				}
				w.mu.Lock()
				w.progress.Total = total
				w.progress.Sent = int(w.sentCount.Load())
				if currentMsg != "" {
					w.progress.CurrentMsg = currentMsg
				}
				cancelled := w.progress.Status == "cancelled"
				w.mu.Unlock()
				w.emitProgress()
				return !cancelled
			},
			w.EmitReplayMessage,
		)
		if err != nil {
			done <- err
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			w.setError(err.Error())
			return
		}
		w.mu.Lock()
		w.progress.Status = "completed"
		w.mu.Unlock()
		w.emitProgress()
		log.Printf("[ReplayWorker] 消息注入完成")
	case <-ctx.Done():
		w.mu.Lock()
		w.progress.Status = "cancelled"
		w.mu.Unlock()
		w.emitProgress()
		log.Printf("[ReplayWorker] 消息注入已取消")
	}
}
