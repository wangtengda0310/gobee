package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher 文件监听器
type Watcher[T any] struct {
	loader     *Loader[T]
	callback   func([]T)
	debounce   time.Duration
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
}

// NewWatcher 创建文件监听器
func NewWatcher[T any](loader *Loader[T]) *Watcher[T] {
	return &Watcher[T]{
		loader:   loader,
		debounce: 200 * time.Millisecond, // 默认防抖 200ms
	}
}

// OnChange 设置变化回调
func (w *Watcher[T]) OnChange(fn func([]T)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callback = fn
}

// SetDebounce 设置防抖时间
func (w *Watcher[T]) SetDebounce(duration time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.debounce = duration
}

// Watch 开始监听文件变化
func (w *Watcher[T]) Watch(ctx context.Context) error {
	// 创建文件监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监听器失败: %w", err)
	}
	defer watcher.Close()

	// 监听文件路径
	filePath := w.loader.basePath
	if err := watcher.Add(filePath); err != nil {
		return fmt.Errorf("添加监听失败: %w", err)
	}

	// 创建上下文
	watchCtx, cancel := context.WithCancel(ctx)
	w.mu.Lock()
	w.cancelFunc = cancel
	w.mu.Unlock()

	// 启动防抖定时器
	var timer *time.Timer
	var debounceMu sync.Mutex

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		for {
			select {
			case <-watchCtx.Done():
				if timer != nil {
					timer.Stop()
				}
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 只处理写入和创建事件
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}

				// 防抖：停止之前的定时器，启动新的
				debounceMu.Lock()
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(w.debounce, func() {
					w.reload()
				})
				debounceMu.Unlock()
			}
		}
	}()

	// 等待上下文取消
	<-watchCtx.Done()
	return nil
}

// reload 重新加载配置
func (w *Watcher[T]) reload() {
	w.mu.Lock()
	callback := w.callback
	w.mu.Unlock()

	if callback == nil {
		return
	}

	data, err := w.loader.Reload()
	if err != nil {
		log.Printf("重新加载配置失败: %v", err)
		return
	}

	callback(data)
}

// Stop 停止监听
func (w *Watcher[T]) Stop() {
	w.mu.Lock()
	if w.cancelFunc != nil {
		w.cancelFunc()
	}
	w.mu.Unlock()

	w.wg.Wait()
}
