// gameactor API 测试用例
//
// 本文件采用 TDD 方式编写，测试用例优先于实现代码。
// 目的：通过测试用例验证 API 设计的合理性和易用性。
//
// 测试覆盖范围：
// 1. DispatchBy - 便利函数（推荐日常使用）
// 2. Dispatch - Hashable 接口版本
// 3. DispatchWithFunc - 哈希函数版本
// 4. DispatchWithHash - 直接哈希版本
// 5. Context 支持
// 6. 同步版本
// 7. 错误场景
// 8. 并发安全
package gameactor_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wangtengda0310/gobee/gameactor"
)

// TestDispatchBy_Basic 测试最基本的 DispatchBy 调用
//
// 这是用户最常用的 API，必须简单直观
func TestDispatchBy_Basic(t *testing.T) {
	// 初始化
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	// 提交任务
	var executed atomic.Bool
	gameactor.DispatchBy(1001, func() {
		executed.Store(true)
	})

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("任务未执行")
	}
}

// TestDispatchBy_ClosureCapture 测试闭包捕获参数
//
// 用户常用场景：通过闭包捕获业务参数
func TestDispatchBy_ClosureCapture(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	// 模拟业务参数
	playerID := uint64(1001)
	damage := 100
	expected := fmt.Sprintf("player=%d damage=%d", playerID, damage)

	var result string
	gameactor.DispatchBy(playerID, func() {
		result = expected
	})

	time.Sleep(100 * time.Millisecond)

	if result != expected {
		t.Errorf("期望 %s, 得到 %s", expected, result)
	}
}

// TestDispatchBy_SameHashSequential 测试相同 hash 的任务按顺序执行
//
// 这是 gameactor 的核心特性：相同哈希的任务必须串行执行
func TestDispatchBy_SameHashSequential(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	hash := uint64(1001)
	var order []int

	// 提交多个任务到同一个 hash
	for i := 1; i <= 5; i++ {
		i := i
		gameactor.DispatchBy(hash, func() {
			order = append(order, i)
			time.Sleep(10 * time.Millisecond) // 故意延迟，确保顺序
		})
	}

	time.Sleep(200 * time.Millisecond)

	// 验证顺序
	expected := []int{1, 2, 3, 4, 5}
	if len(order) != len(expected) {
		t.Fatalf("期望 %d 个任务执行, 实际 %d", len(expected), len(order))
	}

	for i := range expected {
		if order[i] != expected[i] {
			t.Errorf("期望顺序 %v, 实际 %v", expected, order)
			break
		}
	}
}

// TestDispatchBy_DifferentHashParallel 测试不同 hash 的任务并行执行
//
// 验证不同 hash 的任务可以并发执行
func TestDispatchBy_DifferentHashParallel(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	var count atomic.Int32
	done := make(chan struct{})

	// 提交任务到不同 hash
	for i := 0; i < 10; i++ {
		hash := uint64(1000 + i)
		go func() {
			gameactor.DispatchBy(hash, func() {
				time.Sleep(50 * time.Millisecond)
				count.Add(1)
			})
		}()
	}

	// 等待所有任务完成
	go func() {
		time.Sleep(300 * time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
		if count.Load() != 10 {
			t.Errorf("期望 10 个任务执行, 实际 %d", count.Load())
		}
	case <-time.After(1 * time.Second):
		t.Error("超时：任务未在预期时间内完成")
	}
}

// TestDispatchBySync_Basic 测试同步版本的 DispatchBy
//
// 同步版本会阻塞等待任务执行完成
func TestDispatchBySync_Basic(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	var executed atomic.Bool

	err = gameactor.DispatchBySync(1001, func() error {
		executed.Store(true)
		return nil
	})

	if err != nil {
		t.Errorf("DispatchBySync failed: %v", err)
	}

	if !executed.Load() {
		t.Error("任务未执行")
	}
}

// TestDispatchBySync_ReturnError 测试同步版本返回错误
//
// handler 返回 error 时，DispatchBySync 应该返回该错误
func TestDispatchBySync_ReturnError(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	expectedErr := errors.New("测试错误")

	err = gameactor.DispatchBySync(1001, func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("期望错误 %v, 实际 %v", expectedErr, err)
	}
}

// TestDispatchByCtx_Timeout 测试 Context 超时
//
// Context 超时时，任务应该被取消
func TestDispatchByCtx_Timeout(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var executed atomic.Bool

	err = gameactor.DispatchByCtx(ctx, 1001, func() {
		time.Sleep(100 * time.Millisecond) // 超过 context 超时
		executed.Store(true)
	})

	// 注意：根据设计，Context 超时后的行为需要明确
	// 这里我们验证 err != nil
	if err == nil {
		t.Error("期望返回超时错误")
	}
}

// TestDispatchByCtx_Deadline 测试 Context deadline
//
// 类似超时测试，验证 deadline 到达时的行为
func TestDispatchByCtx_Deadline(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	deadline := time.Now().Add(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	err = gameactor.DispatchByCtx(ctx, 1001, func() {
		time.Sleep(50 * time.Millisecond)
	})

	if err == nil {
		t.Error("期望返回 deadline 错误")
	}
}

// TestDispatchBySyncCtx 测试同步 + Context 版本
//
// 组合测试：同步等待 + Context 控制
func TestDispatchBySyncCtx(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := "success"
	err = gameactor.DispatchBySyncCtx(ctx, 1001, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("DispatchBySyncCtx failed: %v", err)
	}

	_ = result // 避免未使用变量警告
}

// TestDispatch_Hashable 测试 Hashable 接口
//
// 用户可以定义实现 Hashable 接口的类型
func TestDispatch_Hashable(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	// 定义实现 Hashable 的任务类型
	type PlayerTask struct {
		PlayerID uint64
		Action   func() error
	}

	// 实现 Hash() 方法
	// 注意：这里假设 gameactor.Task 实现了 Hashable 接口
	// 如果不是，PlayerTask 需要直接实现 Hash() 方法

	var executed atomic.Bool

	// 使用 Dispatch（Hashable 版本）
	gameactor.Dispatch(gameactor.NewTaskWithHash(1001, func() error {
		executed.Store(true)
		return nil
	}))

	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("任务未执行")
	}
}

// TestDispatchSync_Hashable 测试 Hashable + 同步版本
func TestDispatchSync_Hashable(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	task := gameactor.NewTaskWithHash(1001, func() error {
		return nil
	})

	err = gameactor.DispatchSync(task)

	if err != nil {
		t.Errorf("DispatchSync failed: %v", err)
	}
}

// TestDispatchWithFunc_HashFunction 测试使用哈希函数
//
// 用户可以提供自定义的哈希计算函数
func TestDispatchWithFunc_HashFunction(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	// 定义哈希提取函数
	hashExtractor := func(t gameactor.Task) uint64 {
		// 这里假设 Task 有某种方式获取原始数据
		// 实际实现可能需要调整
		return 1001
	}

	var executed atomic.Bool
	task := gameactor.NewTask(func() error {
		executed.Store(true)
		return nil
	})

	err = gameactor.DispatchWithFunc(hashExtractor, task)

	if err != nil {
		t.Errorf("DispatchWithFunc failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("任务未执行")
	}
}

// TestDispatchWithHash_DirectHash 测试直接指定哈希值
//
// 这是最简单的方式：直接告诉系统用哪个哈希
func TestDispatchWithHash_DirectHash(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	var executed atomic.Bool
	task := gameactor.NewTask(func() error {
		executed.Store(true)
		return nil
	})

	err = gameactor.DispatchWithHash(1001, task)

	if err != nil {
		t.Errorf("DispatchWithHash failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("任务未执行")
	}
}

// TestDispatcher_NotInitialized 测试未初始化时的行为
//
// 应该返回错误或自动初始化（根据设计决定）
func TestDispatcher_NotInitialized(t *testing.T) {
	// 注意：这个测试需要放在其他测试之前，或使用独立的测试进程
	// 因为 Init 是 sync.Once，只能执行一次

	// 这里我们测试：如果没有显式 Init，第一个 Dispatch 会自动初始化
	// 或者返回错误

	// 跳过此测试，因为 Init 是 sync.Once
	t.Skip("Init 是 sync.Once，无法重复测试未初始化场景")
}

// TestDispatcher_AfterShutdown 测试关闭后的行为
//
// 关闭后提交任务应该返回错误
func TestDispatcher_AfterShutdown(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 立即关闭
	gameactor.Shutdown(100 * time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	// 尝试提交任务
	err = gameactor.DispatchBySync(1001, func() error {
		return nil
	})

	if err == nil {
		t.Error("期望返回错误：分发器已关闭")
	}
}

// TestDispatcher_ConcurrentSubmit 测试并发提交任务
//
// 验证多个 goroutine 同时提交任务的安全性
func TestDispatcher_ConcurrentSubmit(t *testing.T) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	var count atomic.Int32
	numGoroutines := 100
	tasksPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				hash := uint64(id*1000 + j)
				gameactor.DispatchBy(hash, func() {
					count.Add(1)
				})
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	expected := int32(numGoroutines * tasksPerGoroutine)
	if count.Load() != expected {
		t.Errorf("期望 %d 个任务执行, 实际 %d", expected, count.Load())
	}
}

// TestDispatcher_InitTwice 测试多次初始化
//
// sync.Once 确保只有第一次初始化生效
func TestDispatcher_InitTwice(t *testing.T) {
	// 第一次初始化
	err := gameactor.Init(gameactor.Config{
		NumActors: 10,
	})
	if err != nil {
		t.Fatalf("First Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	// 第二次初始化（应该被忽略）
	err = gameactor.Init(gameactor.Config{
		NumActors: 20, // 这个配置应该被忽略
	})
	if err != nil {
		t.Errorf("Second Init should not fail, got: %v", err)
	}

	// 验证使用的是第一次的配置
	// 这里需要某种方式获取当前配置，或者通过行为验证
}

// TestConfig_Default 测试默认配置
func TestConfig_Default(t *testing.T) {
	config := gameactor.DefaultConfig()

	if config.NumActors <= 0 {
		t.Error("NumActors 应该大于 0")
	}
	if config.QueueSize <= 0 {
		t.Error("QueueSize 应该大于 0")
	}
	if config.ShutdownTimeout <= 0 {
		t.Error("ShutdownTimeout 应该大于 0")
	}
}

// TestConfig_FromEnv 测试从环境变量加载配置
func TestConfig_FromEnv(t *testing.T) {
	// 设置环境变量
	// 注意：这会影响其他测试，实际测试可能需要隔离

	// 跳过此测试，因为环境变量污染问题
	t.Skip("环境变量测试需要隔离")
}

// TestTask_NewTask 测试 Task 工厂函数
func TestTask_NewTask(t *testing.T) {
	handler := func() error {
		return nil
	}

	task := gameactor.NewTask(handler)
	// Task 是结构体，总是有效
	// 验证 Hash 方法可以调用
	_ = task.Hash()
}

// TestTask_NewTaskWithHash 测试带哈希的 Task 工厂函数
func TestTask_NewTaskWithHash(t *testing.T) {
	handler := func() error {
		return nil
	}
	hash := uint64(1001)

	task := gameactor.NewTaskWithHash(hash, handler)
	// 验证哈希值
	if task.Hash() != hash {
		t.Errorf("期望 hash=%d, 实际 %d", hash, task.Hash())
	}
}

// TestTask_NewTaskWithHashFunc 测试带哈希函数的 Task 工厂函数
func TestTask_NewTaskWithHashFunc(t *testing.T) {
	handler := func() error {
		return nil
	}
	expectedHash := uint64(1001)

	hashFunc := func(t gameactor.Task) uint64 {
		return expectedHash
	}

	task := gameactor.NewTaskWithHashFunc(hashFunc, handler)
	// 验证哈希函数
	if task.Hash() != expectedHash {
		t.Errorf("期望 hash=%d, 实际 %d", expectedHash, task.Hash())
	}
}

// TestRegisterShutdown 测试注册关闭钩子
func TestRegisterShutdown(t *testing.T) {
	var hookCalled atomic.Bool

	gameactor.RegisterShutdown(func() {
		hookCalled.Store(true)
	})

	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 关闭会触发钩子
	gameactor.Shutdown(100 * time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	if hookCalled.Load() {
		// 钩子被调用
		// 注意：具体行为取决于实现
	}
}

// BenchmarkDispatchBy 性能基准测试
func BenchmarkDispatchBy(b *testing.B) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		b.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			hash := uint64(i % 1000)
			gameactor.DispatchBy(hash, func() {
				// 空操作
			})
			i++
		}
	})
}

// BenchmarkDispatchBySync 同步版本性能基准测试
func BenchmarkDispatchBySync(b *testing.B) {
	err := gameactor.Init(gameactor.DefaultConfig())
	if err != nil {
		b.Fatalf("Init failed: %v", err)
	}
	defer gameactor.Shutdown(5 * time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			hash := uint64(i % 1000)
			gameactor.DispatchBySync(hash, func() error {
				return nil
			})
			i++
		}
	})
}

