package behavior

// Sequence 创建一个序列节点，按顺序执行子节点。
// 序列节点是"与"逻辑：所有子节点都成功才返回成功。
//
// 执行规则:
//   - 按顺序执行每个子节点
//   - 如果任何子节点返回 Failure，立即返回 Failure
//   - 如果任何子节点返回 Running，立即返回 Running
//   - 如果所有子节点返回 Success，返回 Success
//
// 参数:
//   - children: 子节点列表
//
// 返回值:
//   - Node: 序列节点函数
//
// 边界情况: 空序列返回 Success
//
// 示例:
//
//	// 先检查血量，再攻击
//	sequence := Sequence(checkHealth, attack)
func Sequence(children ...Node) Node {
	return func(ctx Context) Result {
		if len(children) == 0 {
			return Success
		}
		for _, child := range children {
			result := child(ctx)
			if result != Success {
				return result
			}
		}
		return Success
	}
}

// Selector 创建一个选择器节点，按顺序尝试子节点直到成功。
// 选择器节点是"或"逻辑：任一子节点成功即返回成功。
//
// 执行规则:
//   - 按顺序执行每个子节点
//   - 如果任何子节点返回 Success，立即返回 Success
//   - 如果任何子节点返回 Running，立即返回 Running
//   - 如果所有子节点返回 Failure，返回 Failure
//
// 参数:
//   - children: 子节点列表
//
// 返回值:
//   - Node: 选择器节点函数
//
// 边界情况: 空选择器返回 Failure
//
// 示例:
//
//	// 尝试近战攻击，失败则尝试远程攻击
//	selector := Selector(meleeAttack, rangedAttack)
func Selector(children ...Node) Node {
	return func(ctx Context) Result {
		if len(children) == 0 {
			return Failure
		}
		for _, child := range children {
			result := child(ctx)
			if result != Failure {
				return result
			}
		}
		return Failure
	}
}

// ActiveSelector 创建一个主动选择器节点，每次都重新评估所有子节点。
// 与普通 Selector 不同，ActiveSelector 不会记住上次成功的节点，
// 而是每次都从头开始评估，选择第一个返回非失败的子节点。
//
// 执行规则:
//   - 每次都按顺序评估所有子节点
//   - 返回第一个非 Failure 的子节点结果
//   - 如果所有子节点都返回 Failure，返回 Failure
//
// 参数:
//   - children: 子节点列表（按优先级排序）
//
// 返回值:
//   - Node: 主动选择器节点函数
//
// 边界情况: 空选择器返回 Failure
//
// 使用场景: 适用于需要动态改变优先级的场景，如：
//   - 低优先级行为执行中，高优先级条件满足时立即切换
//   - 巡逻时发现敌人立即追击，追击时失去目标立即返回巡逻
//
// 示例:
//
//	// 每次都重新评估：攻击 > 追击 > 巡逻
//	selector := ActiveSelector(attack, chase, patrol)
func ActiveSelector(children ...Node) Node {
	return func(ctx Context) Result {
		if len(children) == 0 {
			return Failure
		}
		for _, child := range children {
			result := child(ctx)
			if result != Failure {
				return result
			}
		}
		return Failure
	}
}

// Parallel 创建一个并行节点，同时执行所有子节点。
// 并行节点根据成功和失败策略决定最终结果。
//
// 执行规则:
//   - 同时执行所有子节点（同一 tick 内）
//   - 失败数 >= failurePolicy 时，返回 Failure
//   - 成功数 >= successPolicy 时，返回 Success
//   - 有子节点 Running 且未触发策略时，返回 Running
//   - 未满足任何策略时，默认返回 Failure
//
// 参数:
//   - successPolicy: 需要成功的子节点数量
//   - failurePolicy: 需要失败的子节点数量
//   - children: 子节点列表
//
// 返回值:
//   - Node: 并行节点函数
//
// 边界情况: 空并行节点返回 Success
//
// 注意: 失败策略优先于成功策略检查。
//
// 示例:
//
//	// 需要 2 个成功，1 个失败就整体失败
//	parallel := Parallel(2, 1, attack1, attack2, attack3)
func Parallel(successPolicy, failurePolicy int, children ...Node) Node {
	return func(ctx Context) Result {
		if len(children) == 0 {
			return Success
		}

		successes := 0
		failures := 0
		runnings := 0

		for _, child := range children {
			result := child(ctx)
			switch result {
			case Success:
				successes++
			case Failure:
				failures++
			case Running:
				runnings++
			}
		}

		// 注意: 失败策略优先检查，实现快速失败行为
		if failures >= failurePolicy {
			return Failure
		}
		if successes >= successPolicy {
			return Success
		}
		if runnings > 0 {
			return Running
		}

		// 未满足任何策略，默认失败
		return Failure
	}
}
