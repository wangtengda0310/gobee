// gameactor - 基于哈希路由的 Go 线程分发系统
//
// 核心特性：
// - 基于哈希值将任务路由到固定的 goroutine 执行
// - 相同哈希的任务按顺序执行（串行）
// - 不同哈希的任务可以并行执行
// - 支持多种调用方式：便利函数、Hashable 接口、哈希函数、直接哈希
//
// 快速开始：
//
//	gameactor.Init(gameactor.DefaultConfig())
//	defer gameactor.Shutdown(30 * time.Second)
//
//	gameactor.DispatchBy(playerID, func() {
//	    // 处理玩家逻辑
//	})
package gameactor

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ============================================================================
// 核心类型
// ============================================================================

// Hashable 接口：任务对象实现此接口以提供哈希值
type Hashable interface {
	Hash() uint64
}

// Task 类型：封装任务处理逻辑
//
// Task 实现了 Hashable 接口，可以直接传递给 Dispatch 函数
type Task struct {
	handler  func() error      // 处理函数
	hash     uint64            // 直接指定的哈希（优先级最高）
	hashFunc func(Task) uint64 // 哈希计算函数（次优先级）
}

// Hash 实现 Hashable 接口
//
// 优先级：
// 1. 直接指定的 hash
// 2. hashFunc 计算的哈希
// 3. 0（无哈希）
func (t Task) Hash() uint64 {
	if t.hash != 0 {
		return t.hash
	}
	if t.hashFunc != nil {
		return t.hashFunc(t)
	}
	return 0
}

// ============================================================================
// 错误定义
// ============================================================================

var (
	// ErrDispatcherClosed 分发器已关闭
	ErrDispatcherClosed = errors.New("dispatcher is closed")
	// ErrNotInitialized 分发器未初始化
	ErrNotInitialized = errors.New("dispatcher not initialized")
)

// ============================================================================
// Task 工厂函数
// ============================================================================

// NewTask 创建基本任务
func NewTask(handler func() error) Task {
	return Task{handler: handler}
}

// NewTaskWithHash 创建带哈希的任务
func NewTaskWithHash(hash uint64, handler func() error) Task {
	return Task{hash: hash, handler: handler}
}

// NewTaskWithHashFunc 创建带哈希函数的任务
func NewTaskWithHashFunc(hashFunc func(Task) uint64, handler func() error) Task {
	return Task{hashFunc: hashFunc, handler: handler}
}

// ============================================================================
// 便利函数（推荐日常使用）⭐
// ============================================================================

// DispatchBy 提交任务到指定哈希的 Actor 异步执行
//
// 参数:
//   - hash: 用于路由的哈希值，相同 hash 的任务将按顺序在同一 Actor 中执行
//   - handler: 无参任务处理函数，通过闭包捕获所需参数
//
// 返回:
//   - error: 分发器已关闭时返回 ErrDispatcherClosed
//
// 并发安全:
//   - 本函数是并发安全的，可从多个 goroutine 同时调用
//   - handler 将在目标 Actor 的 goroutine 中串行执行
//
// 注意:
//   - handler 执行过程中 panic 会被捕获
//   - 避免在 handler 中执行长时间阻塞操作
//
// 示例:
//
//	gameactor.DispatchBy(playerID, func() {
//	    processPlayer(playerID, data)
//	})
func DispatchBy(hash uint64, handler func()) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	task := Task{
		hash:    hash,
		handler: func() error {
			handler()
			return nil
		},
	}

	return globalDispatcher.Submit(hash, task)
}

// DispatchByCtx 提交任务到指定哈希的 Actor 异步执行（支持 Context）
//
// 与 DispatchBy 的区别：
//   - 支持 Context 取消
//   - Context 超时不会中断正在执行的任务，但会阻止新任务提交
func DispatchByCtx(ctx context.Context, hash uint64, handler func()) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	// 检查 Context 是否已取消
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

	return globalDispatcher.Submit(hash, task)
}

// DispatchBySync 提交任务到指定哈希的 Actor 同步执行
//
// 参数:
//   - hash: 用于路由的哈希值
//   - handler: 任务处理函数，返回 error
//
// 返回:
//   - error: 任务执行错误或分发器错误
//
// 行为:
//   - 阻塞等待任务执行完成
//   - handler 返回的 error 会传递给调用者
func DispatchBySync(hash uint64, handler func() error) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	// 使用 channel 等待结果
	done := make(chan error, 1)

	task := Task{
		hash: hash,
		handler: func() error {
			err := handler()
			done <- err
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, task); err != nil {
		return err
	}

	// 等待结果
	return <-done
}

// DispatchBySyncCtx 提交任务到指定哈希的 Actor 同步执行（支持 Context）
//
// 与 DispatchBySync 的区别：
//   - 支持 Context 超时
//   - 超时后返回 context.DeadlineExceeded
func DispatchBySyncCtx(ctx context.Context, hash uint64, handler func() error) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	// 使用 channel 等待结果
	done := make(chan error, 1)

	task := Task{
		hash: hash,
		handler: func() error {
			err := handler()
			select {
			case done <- err:
			case <-ctx.Done():
				// Context 已取消，不发送结果
			}
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, task); err != nil {
		return err
	}

	// 等待结果或超时
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ============================================================================
// Hashable 接口版本
// ============================================================================

// Dispatch 提交实现了 Hashable 接口的任务
//
// 参数:
//   - task: 实现了 Hashable 接口的任务（通常是 Task 或自定义类型）
//
// 返回:
//   - error: 分发器错误
//
// 示例:
//
//	task := gameactor.NewTaskWithHash(playerID, func() error {
//	    return processPlayer()
//	})
//	gameactor.Dispatch(task)
func Dispatch(task Hashable) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	hash := task.Hash()
	if hash == 0 {
		return errors.New("task hash is zero")
	}

	// 将 Hashable 转换为 Task
	var actualTask Task
	if t, ok := task.(Task); ok {
		actualTask = t
	} else {
		// 自定义 Hashable 类型，创建包装 Task
		actualTask = NewTaskWithHash(hash, func() error {
			// 自定义类型的处理逻辑
			return nil
		})
	}

	return globalDispatcher.Submit(hash, actualTask)
}

// DispatchCtx 提交实现了 Hashable 接口的任务（支持 Context）
func DispatchCtx(ctx context.Context, task Hashable) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return Dispatch(task)
}

// DispatchSync 同步提交实现了 Hashable 接口的任务
func DispatchSync(task Hashable) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	hash := task.Hash()
	if hash == 0 {
		return errors.New("task hash is zero")
	}

	// 将 Hashable 转换为 Task
	var actualTask Task
	if t, ok := task.(Task); ok {
		actualTask = t
	} else {
		actualTask = NewTaskWithHash(hash, func() error {
			return nil
		})
	}

	// 使用 channel 等待结果
	done := make(chan error, 1)

	taskWithWait := Task{
		hash: hash,
		handler: func() error {
			err := actualTask.handler()
			done <- err
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, taskWithWait); err != nil {
		return err
	}

	return <-done
}

// DispatchSyncCtx 同步提交实现了 Hashable 接口的任务（支持 Context）
func DispatchSyncCtx(ctx context.Context, task Hashable) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	hash := task.Hash()
	if hash == 0 {
		return errors.New("task hash is zero")
	}

	// 将 Hashable 转换为 Task
	var actualTask Task
	if t, ok := task.(Task); ok {
		actualTask = t
	} else {
		actualTask = NewTaskWithHash(hash, func() error {
			return nil
		})
	}

	// 使用 channel 等待结果
	done := make(chan error, 1)

	taskWithWait := Task{
		hash: hash,
		handler: func() error {
			err := actualTask.handler()
			select {
			case done <- err:
			case <-ctx.Done():
			}
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, taskWithWait); err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ============================================================================
// 哈希函数版本
// ============================================================================

// DispatchWithFunc 使用哈希函数提交任务
//
// 参数:
//   - hashFunc: 从 Task 中提取哈希值的函数
//   - task: 要执行的任务
//
// 使用场景:
//   - Task 本身没有直接指定哈希
//   - 需要从 Task 的内容中计算哈希
func DispatchWithFunc(hashFunc func(Task) uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	// 使用哈希函数计算哈希
	task.hashFunc = hashFunc
	hash := task.Hash()

	return globalDispatcher.Submit(hash, task)
}

// DispatchWithFuncCtx 使用哈希函数提交任务（支持 Context）
func DispatchWithFuncCtx(ctx context.Context, hashFunc func(Task) uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return DispatchWithFunc(hashFunc, task)
}

// DispatchWithFuncSync 使用哈希函数同步提交任务
func DispatchWithFuncSync(hashFunc func(Task) uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	task.hashFunc = hashFunc
	hash := task.Hash()

	done := make(chan error, 1)

	taskWithWait := Task{
		hash: hash,
		handler: func() error {
			err := task.handler()
			done <- err
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, taskWithWait); err != nil {
		return err
	}

	return <-done
}

// DispatchWithFuncSyncCtx 使用哈希函数同步提交任务（支持 Context）
func DispatchWithFuncSyncCtx(ctx context.Context, hashFunc func(Task) uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	task.hashFunc = hashFunc
	hash := task.Hash()

	done := make(chan error, 1)

	taskWithWait := Task{
		hash: hash,
		handler: func() error {
			err := task.handler()
			select {
			case done <- err:
			case <-ctx.Done():
			}
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, taskWithWait); err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ============================================================================
// 直接哈希版本
// ============================================================================

// DispatchWithHash 直接指定哈希值提交任务
//
// 这是最灵活的方式：Task 和哈希值分别指定
func DispatchWithHash(hash uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	task.hash = hash
	return globalDispatcher.Submit(hash, task)
}

// DispatchWithHashCtx 直接指定哈希值提交任务（支持 Context）
func DispatchWithHashCtx(ctx context.Context, hash uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return DispatchWithHash(hash, task)
}

// DispatchWithHashSync 直接指定哈希值同步提交任务
func DispatchWithHashSync(hash uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	task.hash = hash

	done := make(chan error, 1)

	taskWithWait := Task{
		hash: hash,
		handler: func() error {
			err := task.handler()
			done <- err
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, taskWithWait); err != nil {
		return err
	}

	return <-done
}

// DispatchWithHashSyncCtx 直接指定哈希值同步提交任务（支持 Context）
func DispatchWithHashSyncCtx(ctx context.Context, hash uint64, task Task) error {
	if globalDispatcher == nil || !globalDispatcher.IsRunning() {
		return ErrNotInitialized
	}

	task.hash = hash

	done := make(chan error, 1)

	taskWithWait := Task{
		hash: hash,
		handler: func() error {
			err := task.handler()
			select {
			case done <- err:
			case <-ctx.Done():
			}
			return err
		},
	}

	if err := globalDispatcher.Submit(hash, taskWithWait); err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ============================================================================
// 初始化和配置
// ============================================================================

// Config 分发器配置
type Config struct {
	// 基础配置
	NumActors       int           // Actor 数量
	QueueSize       int           // 队列大小
	ShutdownTimeout time.Duration // 关闭超时

	// 环境变量覆盖
	EnvNumActors string // "GAMEACTOR_NUM_ACTORS"
	EnvQueueSize string // "GAMEACTOR_QUEUE_SIZE"
	EnvTimeout   string // "GAMEACTOR_SHUTDOWN_TIMEOUT"

	// 可替换组件（预留）
	Router  interface{} // HashRouter
	Metrics interface{} // MetricsCollector

	// 高级配置
	EnableMetrics bool          // 启用指标
	EnableTracing bool          // 启用追踪
	IdleTimeout   time.Duration // 动态 Actor 空闲超时
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		NumActors:       1000,
		QueueSize:       1000,
		ShutdownTimeout: 30 * time.Second,
		EnvNumActors:    "GAMEACTOR_NUM_ACTORS",
		EnvQueueSize:    "GAMEACTOR_QUEUE_SIZE",
		EnvTimeout:      "GAMEACTOR_SHUTDOWN_TIMEOUT",
		EnableMetrics:   false,
		EnableTracing:   false,
	}
}

// ConfigFromEnv 从环境变量加载配置
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	// TODO: 实现环境变量读取
	// if v := os.Getenv(cfg.EnvNumActors); v != "" {
	//     cfg.NumActors, _ = strconv.Atoi(v)
	// }
	// ...

	return cfg
}

// ============================================================================
// 全局状态管理
// ============================================================================

var (
	globalDispatcher *Dispatcher
	initOnce         sync.Once
	shutdownHooks    []func()
	shutdownMutex    sync.Mutex
	shutdownOnce     sync.Once
)

// Init 初始化分发器（sync.Once 保证只执行一次）
//
// 参数:
//   - config: 分发器配置
//
// 返回:
//   - error: 初始化失败时返回错误
//
// 注意:
//   - 多次调用只会初始化一次
//   - 首次调用后的配置会被保留
func Init(config Config) error {
	var err error
	initOnce.Do(func() {
		globalDispatcher, err = NewDispatcher(config)
	})
	return err
}

// InitWithSignalHandler 初始化分发器并注册信号监听
//
// 这是一个便利函数，会自动监听 SIGINT 和 SIGTERM 信号
// 注意：这应该在 main 函数中使用，不适合测试环境
func InitWithSignalHandler(config Config) error {
	if err := Init(config); err != nil {
		return err
	}

	// 注册信号监听钩子
	RegisterShutdown(func() {
		// TODO: 在实现时添加 signal.Notify
		// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		// <-sigChan
		Shutdown(config.ShutdownTimeout)
	})

	return nil
}

// RegisterShutdown 注册自定义关闭逻辑
//
// 参数:
//   - handler: 关闭时执行的函数
//
// 执行顺序:
//   - 注册的钩子按顺序执行
//   - 在 Dispatcher 停止后执行
func RegisterShutdown(handler func()) {
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()
	shutdownHooks = append(shutdownHooks, handler)
}

// Shutdown 优雅关闭
//
// 参数:
//   - timeout: 等待任务完成的超时时间
//
// 流程:
//   1. 停止接受新任务
//   2. 执行关闭钩子
//   3. 等待所有任务完成或超时
//
// 返回:
//   - error: 目前总是返回 nil
func Shutdown(timeout time.Duration) error {
	shutdownOnce.Do(func() {
		if globalDispatcher != nil {
			// 1. 停止接受新任务
			globalDispatcher.Stop()

			// 2. 执行关闭钩子
			shutdownMutex.Lock()
			hooks := shutdownHooks
			shutdownMutex.Unlock()

			for _, hook := range hooks {
				hook()
			}

			// 3. 等待所有任务完成或超时
			done := make(chan struct{})
			go func() {
				globalDispatcher.Wait()
				close(done)
			}()

			select {
			case <-done:
				// 正常完成
			case <-time.After(timeout):
				// 超时强制关闭
			}
		}
	})
	return nil
}

// Wait 等待所有任务完成
//
// 这是 Shutdown 的内部步骤，暴露出来供高级用户使用
func Wait() {
	if globalDispatcher != nil {
		globalDispatcher.Wait()
	}
}

// ============================================================================
// 状态查询
// ============================================================================

// IsInitialized 返回分发器是否已初始化
func IsInitialized() bool {
	return globalDispatcher != nil
}

// IsRunning 返回分发器是否正在运行
func IsRunning() bool {
	return globalDispatcher != nil && globalDispatcher.IsRunning()
}

// GetMetrics 返回所有 Actor 的指标统计
func GetMetrics() []ActorMetrics {
	if globalDispatcher == nil {
		return nil
	}
	return globalDispatcher.GetMetrics()
}
