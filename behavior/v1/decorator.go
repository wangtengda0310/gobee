package behavior

// Filter 创建一个过滤装饰器节点，只有条件满足时才执行子节点。
// 条件不满足时直接返回 Failure，不执行子节点。
//
// 参数:
//   - condition: 条件函数，返回 true 表示条件满足
//   - child: 要执行的子节点
//
// 返回值:
//   - Failure: 条件不满足
//   - 其他: 子节点的执行结果
//
// 示例:
//
//	// 只有有弹药时才能射击
//	filter := Filter(func(ctx Context) bool {
//	    return ctx["ammo"].(int) > 0
//	}, shoot)
func Filter(condition func(ctx Context) bool, child Node) Node {
	return func(ctx Context) Result {
		if !condition(ctx) {
			return Failure
		}
		return child(ctx)
	}
}

// Monitor 创建一个监控装饰器节点，持续监控条件并执行子节点。
// 每次执行时先检查条件，条件满足时执行子节点，否则返回 Failure。
// 与 Filter 不同，Monitor 适用于需要持续监控的场景。
//
// 参数:
//   - condition: 监控条件函数
//   - child: 条件满足时要执行的子节点
//
// 返回值:
//   - Failure: 条件不满足
//   - 其他: 子节点的执行结果
//
// 使用场景:
//   - 监控敌人距离，距离足够近时攻击
//   - 监控血量，血量低时逃跑
//
// 示例:
//
//	// 持续监控敌人距离，近则攻击
//	monitor := Monitor(func(ctx Context) bool {
//	    return ctx["enemyDistance"].(float64) < 10.0
//	}, attack)
func Monitor(condition func(ctx Context) bool, child Node) Node {
	return func(ctx Context) Result {
		if !condition(ctx) {
			return Failure
		}
		return child(ctx)
	}
}

// Repeat 创建一个无限重复装饰器节点，永远重复执行子节点。
// 子节点成功或失败后立即重新执行，只有返回 Running 时保持 Running 状态。
//
// 参数:
//   - child: 要无限重复执行的子节点
//
// 返回值:
//   - Running: 子节点正在执行或已完成一次执行（准备下一次）
//   - 注意: 此节点永远不会返回 Success 或 Failure
//
// 与 Repeater 的区别:
//   - Repeater(times, child): 重复指定次数后返回 Success
//   - Repeat(child): 无限重复，永远返回 Running
//
// 警告: 此节点永远不会自然结束，需要外部中断或配合其他节点使用。
//
// 示例:
//
//	// 无限循环巡逻
//	repeat := Repeat(patrol)
func Repeat(child Node) Node {
	return func(ctx Context) Result {
		_ = child(ctx)
		// 无论成功还是失败，都继续执行（返回 Running）
		return Running
	}
}

// Inverter 创建一个反转装饰器节点，反转子节点的执行结果。
// Success 变为 Failure，Failure 变为 Success，Running 保持不变。
//
// 参数:
//   - child: 要反转结果的子节点
//
// 返回值:
//   - Node: 反转装饰器节点函数
//
// 示例:
//
//	// 反转一个失败的条件使其成功
//	inverter := Inverter(failingCondition)
func Inverter(child Node) Node {
	return func(ctx Context) Result {
		result := child(ctx)
		switch result {
		case Success:
			return Failure
		case Failure:
			return Success
		default:
			// 注意: Running 状态不反转，子节点仍在执行中
			return result
		}
	}
}

// Repeater 创建一个重复装饰器节点，重复执行子节点指定次数。
//
// 参数:
//   - times: 重复次数 (-1 表示无限重复，0 表示立即成功)
//   - child: 要重复执行的子节点
//
// 返回值:
//   - Running: 重复执行中（子节点完成但未达到目标次数）
//   - Success: 所有重复次数完成
//
// 注意: 状态通过闭包变量维护。每个 Repeater 实例有独立的 tryCount。
// 如需重置状态，请创建新的 Repeater。
//
// 示例:
//
//	// 重复执行 3 次
//	repeater := Repeater(3, action)
func Repeater(times int, child Node) Node {
	tryCount := 0
	return func(ctx Context) Result {
		// 边界情况: 零次表示立即成功，不执行子节点
		if times == 0 {
			return Success
		}

		result := child(ctx)
		if result == Running {
			return Running
		}

		tryCount++

		if times > 0 && tryCount >= times {
			// 重置以支持复用
			tryCount = 0
			return Success
		}

		return Running
	}
}

// UntilSuccess 创建一个直到成功装饰器节点，重复执行直到子节点成功。
//
// 参数:
//   - child: 要重复执行的子节点
//
// 返回值:
//   - Success: 子节点执行成功
//   - Running: 子节点失败或正在执行，继续重试
//
// 警告: 如果子节点永远不会成功，将无限循环。
//
// 示例:
//
//	// 持续尝试直到攻击成功
//	untilSuccess := UntilSuccess(attackAction)
func UntilSuccess(child Node) Node {
	return func(ctx Context) Result {
		result := child(ctx)
		if result == Success {
			return Success
		}
		return Running
	}
}

// UntilFailure 创建一个直到失败装饰器节点，重复执行直到子节点失败。
//
// 参数:
//   - child: 要重复执行的子节点
//
// 返回值:
//   - Success: 子节点执行失败
//   - Running: 子节点成功或正在执行，继续重试
//
// 警告: 如果子节点永远不会失败，将无限循环。
//
// 示例:
//
//	// 持续巡逻直到检测到敌人
//	untilFailure := UntilFailure(patrolAction)
func UntilFailure(child Node) Node {
	return func(ctx Context) Result {
		result := child(ctx)
		if result == Failure {
			return Success
		}
		return Running
	}
}
