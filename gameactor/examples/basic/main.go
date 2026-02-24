// basic 示例 - gameactor 基本使用
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wangtengda0310/gobee/gameactor"
)

func main() {
	fmt.Println("=== gameactor 基本使用示例 ===\n")

	// 示例 1: 基本任务提交
	fmt.Println("示例 1: 基本任务提交")
	basicExample()

	fmt.Println("\n示例 2: 串行执行")
	sequentialExample()

	fmt.Println("\n示例 3: 并行执行")
	parallelExample()

	fmt.Println("\n示例 4: 同步执行")
	syncExample()

	fmt.Println("\n示例 5: 闭包捕获参数")
	closureExample()

	fmt.Println("\n示例 6: 错误处理")
	errorHandlingExample()

	fmt.Println("\n所有示例完成！")
}

// basicExample 基本任务提交
func basicExample() {
	// 使用 TestDispatcher 避免与其他示例冲突
	td := gameactor.NewTestDispatcher(nil, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	var executed atomic.Bool

	// 提交任务
	td.DispatchBy(1001, func() {
		executed.Store(true)
		fmt.Println("  任务已执行")
	})

	// 等待执行
	time.Sleep(100 * time.Millisecond)

	if executed.Load() {
		fmt.Println("  ✓ 任务执行成功")
	} else {
		fmt.Println("  ✗ 任务未执行")
	}
}

// sequentialExample 串行执行示例
func sequentialExample() {
	td := gameactor.NewTestDispatcher(nil, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	hash := uint64(1001)
	var order []int
	var mu sync.Mutex

	// 提交多个任务到同一个 hash
	for i := 1; i <= 5; i++ {
		i := i
		td.DispatchBy(hash, func() {
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
		})
	}

	// 等待执行
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if len(order) == 5 {
		fmt.Printf("  ✓ 5 个任务按顺序执行: %v\n", order)
	} else {
		fmt.Printf("  ✗ 只执行了 %d 个任务\n", len(order))
	}
	mu.Unlock()
}

// parallelExample 并行执行示例
func parallelExample() {
	td := gameactor.NewTestDispatcher(nil, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	var count atomic.Int32
	start := time.Now()

	// 提交任务到不同 hash（会分配到不同 Actor）
	for i := 0; i < 10; i++ {
		hash := uint64(1000 + i)
		td.DispatchBy(hash, func() {
			time.Sleep(50 * time.Millisecond)
			count.Add(1)
		})
	}

	// 等待执行
	time.Sleep(300 * time.Millisecond)

	duration := time.Since(start)
	if count.Load() == 10 {
		fmt.Printf("  ✓ 10 个任务并行执行，耗时: %v\n", duration)
	} else {
		fmt.Printf("  ✗ 只执行了 %d 个任务\n", count.Load())
	}
}

// syncExample 同步执行示例
func syncExample() {
	td := gameactor.NewTestDispatcher(nil, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	err := td.DispatchBySync(1001, func() error {
		fmt.Println("  同步任务执行中...")
		return nil
	})

	if err == nil {
		fmt.Println("  ✓ 同步任务执行成功")
	} else {
		fmt.Printf("  ✗ 同步任务失败: %v\n", err)
	}
}

// closureExample 闭包捕获参数示例
func closureExample() {
	td := gameactor.NewTestDispatcher(nil, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	playerID := uint64(1001)
	damage := 100

	var result string
	td.DispatchBy(playerID, func() {
		result = fmt.Sprintf("玩家 %d 造成 %d 点伤害", playerID, damage)
	})

	time.Sleep(100 * time.Millisecond)

	if result != "" {
		fmt.Printf("  ✓ %s\n", result)
	} else {
		fmt.Println("  ✗ 闭包捕获失败")
	}
}

// errorHandlingExample 错误处理示例
func errorHandlingExample() {
	td := gameactor.NewTestDispatcher(nil, gameactor.DefaultConfig())
	defer td.Shutdown(5 * time.Second)

	err := td.DispatchBySync(1001, func() error {
		return fmt.Errorf("处理失败")
	})

	if err != nil {
		fmt.Printf("  ✓ 捕获到错误: %v\n", err)
	} else {
		fmt.Println("  ✗ 应该返回错误")
	}
}
