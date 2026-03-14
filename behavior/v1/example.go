package behavior

import (
	"fmt"
	"time"
)

// Example 展示了行为树的基本用法
type Example struct {}

// Run 运行一个简单的行为树示例
func (e *Example) Run() {
	// 创建一个简单的上下文
	ctx := make(Context)
	ctx["count"] = 0

	// 创建一些基本节点
	successAction := Action(func(ctx Context) Result {
		fmt.Println("执行成功动作")
		return Success
	})

	condition := Condition(func(ctx Context) bool {
		count := ctx["count"].(int)
		return count < 3
	})

	counterAction := Action(func(ctx Context) Result {
		count := ctx["count"].(int)
		count++
		ctx["count"] = count
		fmt.Printf("计数器增加到: %d\n", count)
		return Success
	})

	waitAction := Action(func(ctx Context) Result {
		fmt.Println("等待中...")
		return Running
	})

	// 构建行为树
	tree := Sequence(
		// 检查条件
		condition,
		// 执行计数器动作
		counterAction,
		// 选择器: 尝试执行成功动作或等待
		Selector(
			successAction,
			waitAction,
		),
	)

	// 运行行为树多次
	fmt.Println("=== 行为树示例开始 ===")
	for i := 0; i < 5; i++ {
		fmt.Printf("\n--- 迭代 %d ---", i+1)
		result := tree(ctx)
		fmt.Printf("\n结果: %s\n", result)
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("\n=== 行为树示例结束 ===")
}

// DemoBehaviorTree 是一个简单的导出函数，用于演示行为树功能
func DemoBehaviorTree() {
	example := &Example{}
	example.Run()
}