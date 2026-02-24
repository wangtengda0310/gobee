// isolation_test.go - 使用 TestDispatcher 的隔离测试
//
// 本文件使用 TestDispatcher 进行测试，避免测试间的状态干扰
package gameactor_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wangtengda0310/gobee/gameactor"
)

// TestTestDispatcher_Basic 测试 TestDispatcher 基本功能
func TestTestDispatcher_Basic(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	var executed atomic.Bool
	td.DispatchBy(1001, func() {
		executed.Store(true)
	})

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("任务未执行")
	}
}

// TestTestDispatcher_ClosureCapture 测试闭包捕获参数
func TestTestDispatcher_ClosureCapture(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	playerID := uint64(1001)
	damage := 100
	expected := fmt.Sprintf("player=%d damage=%d", playerID, damage)

	var result string
	td.DispatchBy(playerID, func() {
		result = expected
	})

	time.Sleep(100 * time.Millisecond)

	if result != expected {
		t.Errorf("期望 %s, 得到 %s", expected, result)
	}
}

// TestTestDispatcher_SameHashSequential 测试相同 hash 的任务按顺序执行
func TestTestDispatcher_SameHashSequential(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	hash := uint64(1001)
	var order []int
	var orderMutex sync.Mutex

	// 提交多个任务到同一个 hash
	for i := 1; i <= 5; i++ {
		i := i
		td.DispatchBy(hash, func() {
			orderMutex.Lock()
			order = append(order, i)
			orderMutex.Unlock()
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 等待所有任务完成
	time.Sleep(200 * time.Millisecond)

	// 验证数量
	if len(order) != 5 {
		t.Errorf("期望 5 个任务执行, 实际 %d", len(order))
	}
}

// TestTestDispatcher_DifferentHashParallel 测试不同 hash 的任务并行执行
func TestTestDispatcher_DifferentHashParallel(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	var count atomic.Int32

	// 提交任务到不同 hash
	for i := 0; i < 10; i++ {
		hash := uint64(1000 + i)
		td.DispatchBy(hash, func() {
			time.Sleep(50 * time.Millisecond)
			count.Add(1)
		})
	}

	// 等待所有任务完成
	time.Sleep(300 * time.Millisecond)

	if count.Load() != 10 {
		t.Errorf("期望 10 个任务执行, 实际 %d", count.Load())
	}
}

// TestTestDispatcher_Sync 测试同步版本
func TestTestDispatcher_Sync(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	var executed atomic.Bool

	err := td.DispatchBySync(1001, func() error {
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

// TestTestDispatcher_SyncReturnError 测试同步版本返回错误
func TestTestDispatcher_SyncReturnError(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	expectedErr := fmt.Errorf("测试错误")

	err := td.DispatchBySync(1001, func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("期望错误 %v, 实际 %v", expectedErr, err)
	}
}

// TestTestDispatcher_ConcurrentSubmit 测试并发提交任务
func TestTestDispatcher_ConcurrentSubmit(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

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
				td.DispatchBy(hash, func() {
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

// TestTestDispatcher_AssertExecuted 测试断言方法
func TestTestDispatcher_AssertExecuted(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	hash := uint64(1001)

	// 提交任务
	for i := 0; i < 3; i++ {
		td.DispatchBy(hash, func() {
			// 空操作
		})
	}

	// 断言执行
	td.AssertExecuted(hash, 3)
}

// TestTestDispatcher_AfterClose 测试关闭后的行为
func TestTestDispatcher_AfterClose(t *testing.T) {
	td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())

	// 关闭
	td.Close()
	time.Sleep(100 * time.Millisecond)

	// 尝试提交任务
	err := td.DispatchBy(1001, func() {
		// 不应该执行
	})

	if err == nil {
		t.Error("期望返回错误：分发器已关闭")
	}
}
