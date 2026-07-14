package notification

import (
	"errors"
	"fmt"
	"sync"
)

// ErrStopProcessing 表示停止后续处理的特殊错误
var ErrStopProcessing = errors.New("stop processing")

// CheckResultDispatcher 检查结果分发器
// 使用观察者模式，将检查结果事件分发到所有注册的输出通道
type CheckResultDispatcher struct {
	handlers []CheckResultHandler
	mu       sync.RWMutex
}

// NewDispatcher 创建新的分发器
func NewDispatcher() *CheckResultDispatcher {
	return &CheckResultDispatcher{
		handlers: make([]CheckResultHandler, 0),
	}
}

// Register 注册输出通道
// 注册后，该通道将收到所有检查结果事件
func (d *CheckResultDispatcher) Register(handler CheckResultHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, handler)
}

// Dispatch 分发事件到所有注册的通道
// 返回所有处理过程中的错误，不会因为某个通道失败而中断其他通道
// 如果某个处理器返回 ErrStopProcessing，则停止后续处理
func (d *CheckResultDispatcher) Dispatch(event *CheckResultEvent) []error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var errors []error
	for _, handler := range d.handlers {
		if err := handler.Handle(event); err != nil {
			if err == ErrStopProcessing {
				// 停止后续处理
				break
			}
			errors = append(errors, fmt.Errorf("[%s] %w", handler.Name(), err))
		}
	}
	return errors
}

// HandlerCount 返回已注册的通道数量
func (d *CheckResultDispatcher) HandlerCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.handlers)
}
