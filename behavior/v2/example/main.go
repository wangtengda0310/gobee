package main

import (
	"fmt"

	"github.com/wangtengda/gobee/behavior/v2"
)

func main() {
	// 调用示例函数来展示行为树的使用
	fmt.Println("Running behavior tree example...")
	Example()

	// 创建一个更复杂的行为树示例
	fmt.Println("\nRunning complex behavior tree example...")
	runComplexExample()
}

// runComplexExample demonstrates a more complex behavior tree
func runComplexExample() {
	// 创建上下文
	ctx := make(behavior.Context)
	ctx["hungry"] = true
	ctx["hasFood"] = true
	ctx["energy"] = 50

	// 创建条件节点
	isHungry := behavior.NewCondition(func(ctx behavior.Context) bool {
		return ctx["hungry"].(bool)
	})

	hasFood := behavior.NewCondition(func(ctx behavior.Context) bool {
		return ctx["hasFood"].(bool)
	})

	hasEnergy := behavior.NewCondition(func(ctx behavior.Context) bool {
		return ctx["energy"].(int) > 20
	})

	// 创建动作节点
	eatFood := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Eating food...")
		ctx["hungry"] = false
		ctx["hasFood"] = false
		energy := ctx["energy"].(int)
		ctx["energy"] = energy + 30
		fmt.Printf("Hunger: %v, Energy: %d\n", ctx["hungry"], ctx["energy"])
		return behavior.Success
	})

	findFood := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Looking for food...")
		energy := ctx["energy"].(int)
		ctx["energy"] = energy - 10
		fmt.Printf("Energy after searching: %d\n", ctx["energy"])

		// 模拟找到食物
		if ctx["energy"].(int) > 10 {
			ctx["hasFood"] = true
			fmt.Println("Found food!")
			return behavior.Success
		}
		fmt.Println("Too tired to find food.")
		return behavior.Failure
	})

	sleep := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Sleeping to regain energy...")
		ctx["energy"] = 100
		fmt.Printf("Energy restored: %d\n", ctx["energy"])
		return behavior.Success
	})

	// 创建行为树结构
	// 1. 如果饥饿且有食物，就吃食物
	// 2. 如果饥饿但没有食物且有能量，就寻找食物
	// 3. 如果没有能量，就睡觉

	eatSequence := behavior.NewSequence(
		isHungry,
		hasFood,
		eatFood,
	)

	findFoodSequence := behavior.NewSequence(
		isHungry,
		behavior.NewInverter(hasFood), // 没有食物
		hasEnergy,
		findFood,
	)

	// 使用选择器来尝试不同的行为
	behaviorTree := behavior.NewSelector(
		eatSequence,
		findFoodSequence,
		behavior.NewInverter(hasEnergy), // 如果没有能量
		sleep,
	)

	// 运行行为树直到所有行为都完成
	fmt.Println("\nStarting behavior tree execution...")
	for i := 0; i < 5; i++ { // 运行5个循环
		fmt.Printf("\n--- Cycle %d ---", i+1)
		result := behaviorTree.Tick(ctx)
		fmt.Printf("\nCycle result: %s\n", result)

		// 如果角色不饿且有能量，我们可以结束循环
		if !ctx["hungry"].(bool) && ctx["energy"].(int) > 20 {
			fmt.Println("Character is satisfied. Ending behavior.")
			break
		}

		// 每两个循环后，角色可能再次感到饥饿
		if i > 0 && i%2 == 0 {
			ctx["hungry"] = true
			fmt.Println("Character is getting hungry again.")
		}
	}

	fmt.Println("\nComplex behavior tree example completed.")
}
