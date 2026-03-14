package behavior

import (
	"testing"
)

// FuzzSimpleBehaviorTree 测试简单的行为树结构
func FuzzSimpleBehaviorTree(f *testing.F) {
	// 定义一些种子值
	testCases := []int{0, 1, 2, 3, 5, 10}
	for _, tc := range testCases {
		f.Add(tc) // 添加模糊测试的种子值
	}

	// 模糊测试函数
	f.Fuzz(func(t *testing.T, n int) {
		// 确保n是正数但不太大
		times := n % 20
		if times < 1 {
			times = 1
		}

		// 创建基本的行为节点
		successAction := Action(func(ctx Context) Result { return Success })
		failureAction := Action(func(ctx Context) Result { return Failure })
		// runningAction := Action(func(ctx Context) Result { return Running })

		// 创建条件节点
		trueCondition := Condition(func(ctx Context) bool { return true })
		falseCondition := Condition(func(ctx Context) bool { return false })

		// 构建各种组合的行为树
		// 1. 简单的动作序列
		sequence := Sequence(
			successAction,
			trueCondition,
			successAction,
		)

		// 2. 选择器
		selector := Selector(
			failureAction,
			falseCondition,
			successAction,
		)

		// 3. 带有装饰器的节点
		repeater := Repeater(times, successAction)
		inverter := Inverter(failureAction)

		// 4. 复杂的组合
		complexTree := Sequence(
			trueCondition,
			Parallel(2, 1, successAction, trueCondition),
			Repeater(times%5+1, selector),
		)

		// 执行所有的树并确保它们不会崩溃
		ctx := make(Context)

		// 执行每个节点多次以测试状态
		for i := 0; i < 5; i++ {
			sequence(ctx)
			selector(ctx)
			repeater(ctx)
			inverter(ctx)
			complexTree(ctx)
		}
	})
}

// FuzzContextOperations 专门测试上下文操作
func FuzzContextOperations(f *testing.F) {
	// 定义一些种子值
	testCases := []int{0, 1, 5, 10}
	for _, tc := range testCases {
		f.Add(tc) // 添加模糊测试的种子值
	}

	// 模糊测试函数
	f.Fuzz(func(t *testing.T, n int) {
		// 使用n作为参数来测试上下文操作
		times := n % 100
		if times < 1 {
			times = 1
		}

		// 创建一个会修改上下文的动作
		counter := 0
		contextAction := Action(func(ctx Context) Result {
			// 读取和修改上下文
			currentCount, exists := ctx["fuzz_count"].(int)
			if !exists {
				currentCount = 0
			}

			currentCount++
			counter++
			ctx["fuzz_count"] = currentCount

			// 根据当前计数返回不同的结果
			switch currentCount % 3 {
			case 0:
				return Running
			case 1:
				return Success
			default:
				return Failure
			}
		})

		// 使用装饰器来重复执行这个动作
		repeater := Repeater(times, contextAction)

		// 执行行为树
		ctx := make(Context)
		for i := 0; i < 3; i++ {
			repeater(ctx)
		}
	})
}
