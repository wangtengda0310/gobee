package behavior

// Sequence 是按顺序执行子节点的复合节点。
// 序列节点是"与"逻辑：所有子节点都成功才返回成功。
//
// 执行规则:
//   - 按顺序执行每个子节点
//   - 如果任何子节点返回 Failure，立即返回 Failure
//   - 如果任何子节点返回 Running，立即返回 Running
//   - 如果所有子节点返回 Success，返回 Success
//
// 示例:
//
//	// 先检查血量，再瞄准，最后攻击
//	sequence := NewSequence(checkHealth, aim, attack)
type Sequence struct {
	Children []Node
}

// NewSequence 创建一个新的序列节点。
//
// 参数:
//   - children: 子节点列表
//
// 返回值:
//   - *Sequence: 序列节点指针
//
// 边界情况: 空序列返回 Success
func NewSequence(children ...Node) *Sequence {
	return &Sequence{
		Children: children,
	}
}

// AddChild 添加一个子节点到序列。
func (s *Sequence) AddChild(child Node) {
	s.Children = append(s.Children, child)
}

// Tick 执行序列逻辑。
// 边界情况: 空序列返回 Success。
//
// 注意: 当前实现不支持从 Running 状态恢复。
// 每次 tick 从第一个子节点开始。如需有状态序列，
// 请在外部跟踪当前子节点索引或使用其他实现。
func (s *Sequence) Tick(ctx Context) Result {
	if len(s.Children) == 0 {
		// 边界情况: 空序列成功（没有失败就是成功）
		return Success
	}
	for _, child := range s.Children {
		result := child.Tick(ctx)
		if result != Success {
			return result
		}
	}
	return Success
}

// Selector 是按顺序尝试子节点的复合节点。
// 选择器节点是"或"逻辑：任一子节点成功即返回成功。
//
// 执行规则:
//   - 按顺序执行每个子节点
//   - 如果任何子节点返回 Success，立即返回 Success
//   - 如果任何子节点返回 Running，立即返回 Running
//   - 如果所有子节点返回 Failure，返回 Failure
//
// 示例:
//
//	// 尝试近战攻击，失败则尝试远程攻击
//	selector := NewSelector(meleeAttack, rangedAttack)
type Selector struct {
	Children []Node
}

// NewSelector 创建一个新的选择器节点。
//
// 参数:
//   - children: 子节点列表
//
// 返回值:
//   - *Selector: 选择器节点指针
//
// 边界情况: 空选择器返回 Failure
func NewSelector(children ...Node) *Selector {
	return &Selector{
		Children: children,
	}
}

// AddChild 添加一个子节点到选择器。
func (s *Selector) AddChild(child Node) {
	s.Children = append(s.Children, child)
}

// Tick 执行选择器逻辑。
// 边界情况: 空选择器返回 Failure。
//
// 注意: 当前实现不支持从 Running 状态恢复。
// 每次 tick 从第一个子节点开始。
func (s *Selector) Tick(ctx Context) Result {
	if len(s.Children) == 0 {
		// 边界情况: 空选择器失败（没有成功就是失败）
		return Failure
	}
	for _, child := range s.Children {
		result := child.Tick(ctx)
		if result != Failure {
			return result
		}
	}
	return Failure
}

// Parallel 是同时执行所有子节点的复合节点。
// 使用成功策略和失败策略决定最终结果。
//
// 策略:
//   - SuccessPolicy: 需要成功的子节点数量
//   - FailurePolicy: 需要失败的子节点数量
//
// 评估顺序（所有子节点 tick 后检查）:
//  1. 失败数 >= FailurePolicy，返回 Failure
//  2. 成功数 >= SuccessPolicy，返回 Success
//  3. 有子节点 Running，返回 Running
//  4. 未满足任何策略，返回 Failure
//
// 示例:
//
//	// 需要 2 个成功，1 个失败就整体失败
//	parallel := NewParallel(2, 1, attack1, attack2, attack3)
type Parallel struct {
	Children      []Node
	SuccessPolicy int // 需要成功的子节点数量
	FailurePolicy int // 需要失败的子节点数量
}

// NewParallel 创建一个新的并行节点。
//
// 参数:
//   - successPolicy: 需要成功的子节点数量
//   - failurePolicy: 需要失败的子节点数量
//   - children: 子节点列表
//
// 返回值:
//   - *Parallel: 并行节点指针
//
// 边界情况: 空并行节点返回 Success
//
// 注意: 失败策略在成功策略之前检查。
// 如果同时满足两个策略，返回 Failure。
func NewParallel(successPolicy, failurePolicy int, children ...Node) *Parallel {
	return &Parallel{
		Children:      children,
		SuccessPolicy: successPolicy,
		FailurePolicy: failurePolicy,
	}
}

// AddChild 添加一个子节点到并行节点。
func (p *Parallel) AddChild(child Node) {
	p.Children = append(p.Children, child)
}

// Tick 执行并行逻辑。
// 边界情况: 空并行节点返回 Success。
func (p *Parallel) Tick(ctx Context) Result {
	if len(p.Children) == 0 {
		// 边界情况: 空并行成功（没有失败就是成功）
		return Success
	}

	successes := 0
	failures := 0
	runnings := 0

	// 执行所有子节点并统计结果
	for _, child := range p.Children {
		result := child.Tick(ctx)
		switch result {
		case Success:
			successes++
		case Failure:
			failures++
		case Running:
			runnings++
		}
	}

	// 注意: 失败优先检查 - 这是故意的设计
	// 实现并行执行中的快速失败行为
	if failures >= p.FailurePolicy {
		return Failure
	}
	if successes >= p.SuccessPolicy {
		return Success
	}
	if runnings > 0 {
		return Running
	}

	// 未满足任何策略 - 默认失败
	// 处理情况如: 需要 2 个成功但只得到 1 个且没有失败
	return Failure
}
