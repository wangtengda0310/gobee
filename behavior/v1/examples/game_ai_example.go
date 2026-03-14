package main

import (
	"fmt"
	"time"

	"github.com/wangtengda/gobee/behavior/v1"
)

func main() {
	// 创建一个模拟游戏角色的上下文
	ctx := make(behavior.Context)
	ctx["targetFound"] = false
	ctx["health"] = 100
	ctx["energy"] = 80
	ctx["hasAmmo"] = true
	ctx["enemyDistance"] = 15.0 // 米
	ctx["lastAttackTime"] = time.Time{}

	// 创建条件节点
	checkTarget := behavior.Condition(func(ctx behavior.Context) bool {
		return ctx["targetFound"].(bool)
	})

	checkHealth := behavior.Condition(func(ctx behavior.Context) bool {
		return ctx["health"].(int) > 30
	})

	checkEnergy := behavior.Condition(func(ctx behavior.Context) bool {
		return ctx["energy"].(int) > 20
	})

	checkAmmo := behavior.Condition(func(ctx behavior.Context) bool {
		return ctx["hasAmmo"].(bool)
	})

	checkDistance := behavior.Condition(func(ctx behavior.Context) bool {
		return ctx["enemyDistance"].(float64) < 20.0
	})

	checkCooldown := behavior.Condition(func(ctx behavior.Context) bool {
		lastAttack := ctx["lastAttackTime"].(time.Time)
		return time.Since(lastAttack).Seconds() > 2.0
	})

	// 创建动作节点
	scanForTarget := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Scanning for targets...")
		// 模拟随机找到目标
		if time.Now().UnixNano()%3 == 0 {
			ctx["targetFound"] = true
			fmt.Println("Target found!")
			return behavior.Success
		}
		fmt.Println("No target found.")
		return behavior.Running
	})

	walkToTarget := behavior.Action(func(ctx behavior.Context) behavior.Result {
		distance := ctx["enemyDistance"].(float64)
		if distance > 1.0 {
			distance -= 0.5
			ctx["enemyDistance"] = distance
			fmt.Printf("Walking towards target, distance: %.1f meters\n", distance)
			return behavior.Running
		}
		fmt.Println("Reached target.")
		return behavior.Success
	})

	attackTarget := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Attacking target!")
		energy := ctx["energy"].(int)
		energy -= 10
		ctx["energy"] = energy
		ctx["lastAttackTime"] = time.Now()

		// 50%几率消耗弹药
		if time.Now().UnixNano()%2 == 0 {
			ctx["hasAmmo"] = false
			fmt.Println("Out of ammo!")
		}

		// 30%几率敌人反击
		if time.Now().UnixNano()%10 < 3 {
			health := ctx["health"].(int)
			health -= 20
			ctx["health"] = health
			fmt.Println("Enemy counter-attacked! Health reduced.")
		}

		return behavior.Success
	})

	meleeAttack := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Performing melee attack!")
		energy := ctx["energy"].(int)
		energy -= 5
		ctx["energy"] = energy
		return behavior.Success
	})

	reloadAmmo := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Reloading ammo...")
		// 模拟装弹需要时间
		if time.Now().UnixNano()%3 == 0 {
			ctx["hasAmmo"] = true
			fmt.Println("Reload complete.")
			return behavior.Success
		}
		return behavior.Running
	})

	takeCover := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Taking cover to recover health...")
		health := ctx["health"].(int)
		health += 5
		if health > 100 {
			health = 100
		}
		ctx["health"] = health
		return behavior.Running
	})

	flee := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Fleeing from combat!")
		distance := ctx["enemyDistance"].(float64)
		distance += 2.0
		ctx["enemyDistance"] = distance

		if distance > 50.0 {
			fmt.Println("Successfully fled from combat.")
			ctx["targetFound"] = false
			return behavior.Success
		}
		return behavior.Running
	})

	wait := behavior.Action(func(ctx behavior.Context) behavior.Result {
		fmt.Println("Waiting for energy to recover...")
		energy := ctx["energy"].(int)
		energy += 3
		if energy > 100 {
			energy = 100
		}
		ctx["energy"] = energy
		return behavior.Running
	})

	// 构建行为树
	// 主行为树
	behaviorTree := behavior.Sequence(
		// 首先尝试寻找目标
		scanForTarget,
		checkTarget,

		behavior.Selector(
			// 健康状况良好时的行为
			behavior.Sequence(
				checkHealth,

				behavior.Sequence(
					walkToTarget,
					checkDistance,

					behavior.Selector(
						// 优先使用远程攻击
						behavior.Sequence(
							checkAmmo,
							checkEnergy,
							checkCooldown,
							attackTarget,
						),
						// 没有弹药时使用近战攻击
						behavior.Sequence(
							checkEnergy,
							meleeAttack,
						),
						// 没有能量时重新装弹或等待
						behavior.Selector(
							behavior.Sequence(
								behavior.Inverter(checkAmmo),
								reloadAmmo,
							),
							wait,
						),
					),
				),

				// 健康状况不佳时的行为
				behavior.Selector(
					// 尝试寻找掩护恢复
					takeCover,
					// 掩护无效时逃跑
					flee,
				),
			),
		),
	)

	// 运行行为树
	fmt.Println("===== Behavior Tree Simulation Started =====")
	fmt.Println("Initial state:")
	printState(ctx)

	// 模拟运行100次循环
	for i := 0; i < 100; i++ {
		fmt.Printf("\n--- Cycle %d ---\n", i+1)
		result := behaviorTree(ctx)
		fmt.Printf("Cycle result: %s\n", result)

		// 如果目标丢失，重新开始
		if !ctx["targetFound"].(bool) {
			fmt.Println("Target lost, resetting behavior.")
		}

		// 每5个循环打印一次状态
		if i%5 == 0 {
			printState(ctx)
		}

		// 如果健康为0，结束模拟
		if ctx["health"].(int) <= 0 {
			fmt.Println("Health is zero. Simulation ended.")
			break
		}
	}

	fmt.Println("\n===== Behavior Tree Simulation Ended =====")
	printState(ctx)
}

// 打印当前状态
func printState(ctx behavior.Context) {
	fmt.Println("Current state:")
	fmt.Printf("  Target Found: %v\n", ctx["targetFound"])
	fmt.Printf("  Health: %d\n", ctx["health"])
	fmt.Printf("  Energy: %d\n", ctx["energy"])
	fmt.Printf("  Has Ammo: %v\n", ctx["hasAmmo"])
	fmt.Printf("  Enemy Distance: %.1f meters\n", ctx["enemyDistance"])
}
