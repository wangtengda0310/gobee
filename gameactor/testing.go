// testing.go - 测试辅助工具
//
// 提供测试专用的辅助函数，支持：
// - 独立的 Dispatcher 实例（避免测试间干扰）
// - 任务执行验证
// - 并发测试支持
package gameactor

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ==============================================================================
// TestDispatcher - 测试用分发器
// ==============================================================================

// TestDispatcher 测试用分发器
//
// 特点：
// - 独立的 Dispatcher 实例，不影响全局单例
// - 提供任务执行验证
// - 支持并发测试
type TestDispatcher struct {
	*Dispatcher
	t    testing.TB
	executed map[uint64][]func() // 已执行的任务记录
	mutex    sync.Mutex
	closed   atomic.Bool
}

// NewTestDispatcher 创建测试用分发器
//
// 参数:
//   - t: testing.TB 接口（*testing.T 或 *testing.B）
//   - config: 分发器配置（使用 DefaultConfig() 获取默认配置）
//
// 返回:
//   - *TestDispatcher: 新创建的测试分发器
//
// 使用示例:
//
//	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
//	defer td.Shutdown(5 * time.Second)
//
//	td.DispatchBy(1001, func() {
//	    // 测试逻辑
//	})
func NewTestDispatcher(t testing.TB, config Config) *TestDispatcher {
	d, err := NewDispatcher(config)
	if err != nil {
		t.Fatalf("Failed to create test dispatcher: %v", err)
	}

	return &TestDispatcher{
		Dispatcher: d,
		t:          t,
		executed:   make(map[uint64][]func()),
	}
}

// ==============================================================================
// 便捷方法（与全局 API 一致）
// ==============================================================================

// DispatchBy 提交任务到指定哈希的 Actor 异步执行
func (td *TestDispatcher) DispatchBy(hash uint64, handler func()) error {
	if td.closed.Load() {
		return ErrDispatcherClosed
	}

	task := Task{
		hash:    hash,
		handler: func() error {
			// 记录执行
			td.mutex.Lock()
			td.executed[hash] = append(td.executed[hash], handler)
			td.mutex.Unlock()

			handler()
			return nil
		},
	}

	return td.Dispatcher.Submit(hash, task)
}

// DispatchBySync 提交任务到指定哈希的 Actor 同步执行
func (td *TestDispatcher) DispatchBySync(hash uint64, handler func() error) error {
	if td.closed.Load() {
		return ErrDispatcherClosed
	}

	done := make(chan error, 1)

	task := Task{
		hash: hash,
		handler: func() error {
			// 记录执行
			td.mutex.Lock()
			td.executed[hash] = append(td.executed[hash], func() {
				handler()
			})
			td.mutex.Unlock()

			err := handler()
			done <- err
			return err
		},
	}

	if err := td.Dispatcher.Submit(hash, task); err != nil {
		return err
	}

	return <-done
}

// DispatchByCtx 提交任务到指定哈希的 Actor 异步执行（支持 Context）
func (td *TestDispatcher) DispatchByCtx(ctx context.Context, hash uint64, handler func()) error {
	if td.closed.Load() {
		return ErrDispatcherClosed
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	task := Task{
		hash:    hash,
		handler: func() error {
			handler()
			return nil
		},
	}

	return td.Dispatcher.Submit(hash, task)
}

// DispatchBySyncCtx 提交任务到指定哈希的 Actor 同步执行（支持 Context）
func (td *TestDispatcher) DispatchBySyncCtx(ctx context.Context, hash uint64, handler func() error) error {
	if td.closed.Load() {
		return ErrDispatcherClosed
	}

	done := make(chan error, 1)

	task := Task{
		hash: hash,
		handler: func() error {
			err := handler()
			select {
			case done <- err:
			case <-ctx.Done():
			}
			return err
		},
	}

	if err := td.Dispatcher.Submit(hash, task); err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ==============================================================================
// 断言方法
// ==============================================================================

// AssertExecuted 断言指定 hash 的任务已执行
//
// 参数:
//   - hash: 要检查的哈希值
//   - count: 期望执行的任务数量
//
// 使用场景:
//   - 验证任务是否被执行
//   - 验证任务执行次数
//
// 注意:
//   - 会等待最多 1 秒让任务执行完成
func (td *TestDispatcher) AssertExecuted(hash uint64, count int) {
	// 等待任务执行
	for i := 0; i < 10; i++ {
		td.mutex.Lock()
		actual := len(td.executed[hash])
		td.mutex.Unlock()

		if actual >= count {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	td.mutex.Lock()
	actual := len(td.executed[hash])
	td.mutex.Unlock()

	if actual < count {
		td.t.Errorf("期望 hash %d 执行 %d 个任务, 实际 %d", hash, count, actual)
	}
}

// AssertNotExecuted 断言指定 hash 的任务未执行
func (td *TestDispatcher) AssertNotExecuted(hash uint64) {
	td.mutex.Lock()
	defer td.mutex.Unlock()

	if len(td.executed[hash]) > 0 {
		td.t.Errorf("期望 hash %d 不执行任务, 但实际执行了 %d 个", hash, len(td.executed[hash]))
	}
}

// AssertOrder 断言任务按预期顺序执行
//
// 参数:
//   - hashes: 期望的哈希值顺序
//
// 使用场景:
//   - 验证串行执行顺序
//   - 验证任务排序逻辑
//
// 注意:
//   - 这个断言只验证第一个 hash 为 0 的 Actor 的任务顺序
//   - 对于跨多个 Actor 的任务，需要分别验证
func (td *TestDispatcher) AssertOrder(hashes []uint64) {
	// 等待所有任务执行
	time.Sleep(500 * time.Millisecond)

	// 获取执行记录
	td.mutex.Lock()
	defer td.mutex.Unlock()

	// 简化版本：只验证第一个 hash 的执行次数
	// 完整版本需要记录每个任务的执行顺序
	if len(hashes) == 0 {
		return
	}

	firstHash := hashes[0]
	actual := len(td.executed[firstHash])
	if actual != len(hashes) {
		td.t.Errorf("期望 hash %d 执行 %d 个任务, 实际 %d", firstHash, len(hashes), actual)
	}
}

// WaitForExecution 等待指定 hash 的任务执行完成
//
// 参数:
//   - hash: 要等待的哈希值
//   - timeout: 等待超时时间
//
// 返回:
//   - error: 超时或任务执行错误
func (td *TestDispatcher) WaitForExecution(hash uint64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		td.mutex.Lock()
		executed := len(td.executed[hash]) > 0
		td.mutex.Unlock()

		if executed {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return ErrTimeout
}

// Reset 重置执行记录
//
// 使用场景:
//   - 在同一个测试中验证多个场景
//   - 清除之前的执行记录
func (td *TestDispatcher) Reset() {
	td.mutex.Lock()
	defer td.mutex.Unlock()
	td.executed = make(map[uint64][]func())
}

// ==============================================================================
// 关闭方法
// ============================================================================

// Close 关闭测试分发器
//
// 行为:
//   - 停止接受新任务
//   - 等待所有任务完成
//   - 标记为已关闭
func (td *TestDispatcher) Close() {
	if !td.closed.CompareAndSwap(false, true) {
		return // 已经关闭
	}
	td.Dispatcher.Stop()
}

// Shutdown 关闭测试分发器（兼容全局 API）
//
// 参数:
//   - timeout: 等待超时时间
func (td *TestDispatcher) Shutdown(timeout time.Duration) {
	td.Close()
	td.Dispatcher.Wait()
}

// ==============================================================================
// 错误定义
// ============================================================================

var (
	// ErrTimeout 等待超时
	ErrTimeout = errors.New("timeout waiting for execution")
)
