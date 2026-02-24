// dispatcher.go - Dispatcher 核心实现
//
// Dispatcher 是 gameactor 的核心组件，负责：
// - 管理多个 Actor（固定数量的 goroutine）
// - 根据哈希值将任务路由到对应的 Actor
// - 确保相同哈希的任务按顺序执行
//
// 设计决策：
// - 使用 FNV-1a 哈希算法：快速且碰撞概率可接受
// - Channel 缓冲队列：减少阻塞等待，提高吞吐量
// - 固定 Actor 池：简化实现，预留未来优化空间（Hybrid/WorkStealing）
package gameactor

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ==============================================================================
// Dispatcher 核心结构
// ==============================================================================

// Dispatcher 任务分发器
//
// 核心职责：
// 1. 管理 Actor 池（固定数量）
// 2. 根据哈希值路由任务到对应 Actor
// 3. 提供优雅关闭机制
type Dispatcher struct {
	// 配置
	config Config

	// Actor 管理
	actors      []*actor      // Actor 池（固定大小）
	numActors   uint64        // Actor 数量

	// 状态管理
	running     atomic.Bool   // 运行状态
	stopped     atomic.Bool   // 是否已停止
	stopping    atomic.Bool   // 是否正在停止

	// 关闭协调
	stopChan    chan struct{} // 停止信号
	stopOnce    sync.Once     // 确保只停止一次
	waitGroup   sync.WaitGroup // 等待所有 Actor 退出

	// 错误处理
	panicHandler func(interface{}) // panic 处理函数
}

// ==============================================================================
// Actor 内部结构
// ==============================================================================

// actor 表示单个任务执行器
//
// 每个 actor 是一个独立的 goroutine，从 channel 中取任务并执行
type actor struct {
	id         uint64            // Actor ID
	dispatcher *Dispatcher       // 所属分发器
	queue      chan Task         // 任务队列（缓冲 channel）
	running    atomic.Bool       // 运行状态
	metrics    actorMetrics      // 指标统计
}

// actorMetrics Actor 指标统计
type actorMetrics struct {
	tasksReceived atomic.Int64 // 接收的任务数
	tasksExecuted atomic.Int64 // 执行的任务数
	tasksFailed   atomic.Int64 // 失败的任务数
	totalDuration atomic.Int64 // 总执行时间（纳秒）
}

// ==============================================================================
// Dispatcher 构造函数
// ==============================================================================

// NewDispatcher 创建新的分发器
//
// 参数:
//   - config: 分发器配置
//
// 返回:
//   - *Dispatcher: 新创建的分发器
//   - error: 配置错误时返回错误
//
// 设计决策:
//   - 预先创建所有 Actor，避免动态创建的开销
//   - 每个 Actor 有独立的缓冲队列，减少竞争
func NewDispatcher(config Config) (*Dispatcher, error) {
	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	d := &Dispatcher{
		config:   config,
		actors:   make([]*actor, config.NumActors),
		numActors: uint64(config.NumActors),
		stopChan: make(chan struct{}),
	}

	// 创建 Actor
	for i := 0; i < config.NumActors; i++ {
		d.actors[i] = &actor{
			id:         uint64(i),
			dispatcher: d,
			queue:      make(chan Task, config.QueueSize),
		}
	}

	// 启动所有 Actor
	d.start()

	return d, nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.NumActors <= 0 {
		return errors.New("NumActors must be positive")
	}
	if config.QueueSize <= 0 {
		return errors.New("QueueSize must be positive")
	}
	if config.ShutdownTimeout <= 0 {
		return errors.New("ShutdownTimeout must be positive")
	}
	return nil
}

// ==============================================================================
// Dispatcher 启动和停止
// ==============================================================================

// start 启动所有 Actor
//
// 设计决策:
//   - 每个 Actor 在独立的 goroutine 中运行
//   - 使用 WaitGroup 跟踪所有 Actor 的生命周期
func (d *Dispatcher) start() {
	d.running.Store(true)
	d.stopped.Store(false)
	d.stopping.Store(false)

	for _, a := range d.actors {
		d.waitGroup.Add(1)
		go a.run()
	}
}

// Stop 停止接受新任务
//
// 行为:
//   1. 设置停止标志
//   2. 关闭所有 Actor 的队列
//   3. 不等待正在执行的任务完成
//
// 设计决策:
//   - 立即返回，不阻塞
//   - 调用者应使用 Wait() 等待所有任务完成
func (d *Dispatcher) Stop() {
	d.stopOnce.Do(func() {
		d.stopping.Store(true)

		// 关闭所有 Actor 的队列
		for _, a := range d.actors {
			close(a.queue)
		}

		// 等待所有 Actor 退出
		d.waitGroup.Wait()

		d.running.Store(false)
		d.stopped.Store(true)
	})
}

// Wait 等待所有任务完成
//
// 行为:
//   - 阻塞直到所有 Actor 退出
//   - 配合 Stop() 使用，实现优雅关闭
//
// 超时处理:
//   - 调用者应使用 context.WithTimeout 实现超时
func (d *Dispatcher) Wait() {
	d.waitGroup.Wait()
}

// ==============================================================================
// 任务提交
// ==============================================================================

// Submit 提交任务到指定的 Actor
//
// 参数:
//   - hash: 用于路由的哈希值
//   - task: 要执行的任务
//
// 返回:
//   - error: 分发器已关闭时返回 ErrDispatcherClosed
//
// 路由算法:
//   actorID = hash % numActors
//
// 并发安全:
//   - 本函数是并发安全的
//   - 使用 select 实现非阻塞提交（队列满时立即返回错误）
func (d *Dispatcher) Submit(hash uint64, task Task) error {
	// 检查是否已停止
	if d.stopping.Load() || d.stopped.Load() {
		return ErrDispatcherClosed
	}

	// 路由到对应的 Actor
	actorID := d.route(hash)
	actor := d.actors[actorID]

	// 非阻塞提交
	select {
	case actor.queue <- task:
		actor.metrics.tasksReceived.Add(1)
		return nil
	default:
		// 队列已满
		return fmt.Errorf("actor %d queue is full", actorID)
	}
}

// SubmitBlocking 阻塞提交任务
//
// 与 Submit 的区别：
//   - 队列满时会阻塞等待，而不是立即返回错误
//   - 适用于对提交可靠性要求高的场景
//
// 参数:
//   - hash: 用于路由的哈希值
//   - task: 要执行的任务
//   - timeout: 阻塞超时时间
//
// 返回:
//   - error: 超时或关闭时返回错误
func (d *Dispatcher) SubmitBlocking(hash uint64, task Task, timeout time.Duration) error {
	// 检查是否已停止
	if d.stopping.Load() || d.stopped.Load() {
		return ErrDispatcherClosed
	}

	// 路由到对应的 Actor
	actorID := d.route(hash)
	actor := d.actors[actorID]

	// 带超时的阻塞提交
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case actor.queue <- task:
			actor.metrics.tasksReceived.Add(1)
			return nil
		case <-timer.C:
			return fmt.Errorf("submit timeout after %v", timeout)
		case <-d.stopChan:
			return ErrDispatcherClosed
		}
	}

	// 无超时限制
	select {
	case actor.queue <- task:
		actor.metrics.tasksReceived.Add(1)
		return nil
	case <-d.stopChan:
		return ErrDispatcherClosed
	}
}

// ==============================================================================
// 路由算法
// ==============================================================================

// route 根据哈希值计算 Actor ID
//
// 算法: actorID = hash % numActors
//
// 设计决策:
//   - 使用取模运算，简单高效
//   - 相同 hash 总是路由到同一个 Actor
//   - 未来可替换为更复杂的路由策略（如一致性哈希）
func (d *Dispatcher) route(hash uint64) uint64 {
	return hash % d.numActors
}

// ==============================================================================
// Actor 执行逻辑
// ==============================================================================

// run Actor 主循环
//
// 行为:
//   1. 从队列中取出任务
//   2. 执行任务（带 panic 恢复）
//   3. 记录指标
//   4. 队列关闭时退出
//
// 设计决策:
//   - 每个 Actor 在独立的 goroutine 中运行
//   - 使用 defer recover() 捕获 panic
//   - panic 会被转换为 error 并调用 panicHandler
func (a *actor) run() {
	defer a.dispatcher.waitGroup.Done()
	a.running.Store(true)
	defer a.running.Store(false)

	for {
		select {
		case task, ok := <-a.queue:
			if !ok {
				// 队列已关闭，退出
				return
			}
			a.executeTask(task)

		case <-a.dispatcher.stopChan:
			// 收到停止信号
			return
		}
	}
}

// executeTask 执行单个任务
//
// 行为:
//   1. 记录开始时间
//   2. 执行 handler（带 panic 恢复）
//   3. 记录执行时间和结果
//   4. 处理 panic 和 error
func (a *actor) executeTask(task Task) {
	start := time.Now()

	// 执行任务（带 panic 恢复）
	defer func() {
		if r := recover(); r != nil {
			a.metrics.tasksFailed.Add(1)

			// 调用 panic handler
			if a.dispatcher.panicHandler != nil {
				a.dispatcher.panicHandler(r)
			}
		}

		// 记录执行时间
		duration := time.Since(start)
		a.metrics.totalDuration.Add(int64(duration))
	}()

	// 执行 handler
	a.metrics.tasksExecuted.Add(1)
	if task.handler != nil {
		if err := task.handler(); err != nil {
			// handler 返回错误，不是 panic
			// 这里可以选择记录日志或调用错误处理器
		}
	}
}

// ==============================================================================
// 状态查询
// ==============================================================================

// IsRunning 返回分发器是否正在运行
func (d *Dispatcher) IsRunning() bool {
	return d.running.Load()
}

// IsStopped 返回分发器是否已停止
func (d *Dispatcher) IsStopped() bool {
	return d.stopped.Load()
}

// ==============================================================================
// 指标查询
// ==============================================================================

// GetMetrics 返回所有 Actor 的指标统计
func (d *Dispatcher) GetMetrics() []ActorMetrics {
	metrics := make([]ActorMetrics, len(d.actors))
	for i, a := range d.actors {
		metrics[i] = ActorMetrics{
			ActorID:       a.id,
			TasksReceived: a.metrics.tasksReceived.Load(),
			TasksExecuted: a.metrics.tasksExecuted.Load(),
			TasksFailed:   a.metrics.tasksFailed.Load(),
			TotalDuration: time.Duration(a.metrics.totalDuration.Load()),
			QueueLength:   len(a.queue),
		}
	}
	return metrics
}

// ActorMetrics Actor 指标
type ActorMetrics struct {
	ActorID       uint64        // Actor ID
	TasksReceived int64         // 接收的任务数
	TasksExecuted int64         // 执行的任务数
	TasksFailed   int64         // 失败的任务数
	TotalDuration time.Duration // 总执行时间
	QueueLength   int           // 当前队列长度
}

// ==============================================================================
// Panic 处理
// ==============================================================================

// SetPanicHandler 设置 panic 处理函数
//
// 参数:
//   - handler: panic 处理函数，接收 panic 值
//
// 使用场景:
//   - 记录 panic 日志
//   - 发送告警通知
//   - 自定义恢复逻辑
func (d *Dispatcher) SetPanicHandler(handler func(interface{})) {
	d.panicHandler = handler
}
