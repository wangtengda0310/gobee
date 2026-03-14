package behavior

// DecoratorNode 是包装另一个节点并修改其行为的装饰器节点接口。
// 装饰器用于给子节点添加额外功能而不修改子节点本身。
type DecoratorNode interface {
	Node
	// SetChild 设置装饰器的子节点。
	SetChild(child Node)
}

// Inverter 是反转子节点结果的装饰器节点。
// Success 变为 Failure，Failure 变为 Success，Running 保持不变。
//
// 示例:
//
//	// 反转一个失败的条件使其成功
//	inverter := NewInverter(NewCondition(func(ctx Context) bool { return false }))
type Inverter struct {
	child Node
}

// NewInverter 创建一个新的反转装饰器节点。
//
// 参数:
//   - child: 要反转结果的子节点
//
// 返回值:
//   - *Inverter: 反转装饰器节点指针
func NewInverter(child Node) *Inverter {
	return &Inverter{
		child: child,
	}
}

// SetChild 设置反转装饰器的子节点。
func (i *Inverter) SetChild(child Node) {
	i.child = child
}

// Tick 执行反转逻辑。
// 子节点为 nil 时返回 Failure。
func (i *Inverter) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if i.child == nil {
		return Failure
	}

	result := i.child.Tick(ctx)
	switch result {
	case Success:
		return Failure
	case Failure:
		return Success
	default:
		// 注意: Running 不反转 - 子节点仍在执行中
		return result
	}
}

// Repeater 是重复执行子节点指定次数的装饰器节点。
//
// 注意: 状态保存在 tryCount 中。调用 Reset() 重置计数器
// 以支持行为树复用或重新执行。
type Repeater struct {
	child    Node
	times    int // -1 表示无限，0 表示立即成功
	maxTries int // 保存原始 times 值用于可能的 reset
	tryCount int // 当前重复计数
}

// NewRepeater 创建一个新的重复装饰器节点。
//
// 参数:
//   - times: 重复次数 (-1 表示无限，0 表示立即成功)
//   - child: 要重复执行的子节点
//
// 返回值:
//   - Running: 重复执行中（子节点完成但未达到目标次数）
//   - Success: 所有重复次数完成
func NewRepeater(times int, child Node) *Repeater {
	return &Repeater{
		child:    child,
		times:    times,
		maxTries: times,
		tryCount: 0,
	}
}

// SetChild 设置重复装饰器的子节点。
func (r *Repeater) SetChild(child Node) {
	r.child = child
}

// Reset 重置重复装饰器的内部计数器。
// 调用此方法可以在新的执行周期中复用重复器。
func (r *Repeater) Reset() {
	r.tryCount = 0
}

// Tick 执行重复逻辑。
// 子节点为 nil 时返回 Failure。
func (r *Repeater) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if r.child == nil {
		return Failure
	}

	// 边界情况: 零次表示立即成功，不执行子节点
	if r.times == 0 {
		return Success
	}

	result := r.child.Tick(ctx)
	if result == Running {
		return Running
	}

	r.tryCount++

	if r.times > 0 && r.tryCount >= r.times {
		// 注意: 完成时重置 tryCount 以支持行为树复用
		r.tryCount = 0
		return Success
	}

	return Running
}

// UntilSuccess 是重复执行直到成功的装饰器节点。
// 子节点失败或运行时返回 Running，成功时返回 Success。
//
// 警告: 如果子节点永远不会成功，将无限循环。
type UntilSuccess struct {
	child Node
}

// NewUntilSuccess 创建一个新的直到成功装饰器节点。
//
// 参数:
//   - child: 要重复执行的子节点
//
// 返回值:
//   - Success: 子节点成功
//   - Running: 子节点失败或正在执行，继续重试
func NewUntilSuccess(child Node) *UntilSuccess {
	return &UntilSuccess{
		child: child,
	}
}

// SetChild 设置直到成功装饰器的子节点。
func (u *UntilSuccess) SetChild(child Node) {
	u.child = child
}

// Tick 执行直到成功逻辑。
// 子节点为 nil 时返回 Failure。
func (u *UntilSuccess) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if u.child == nil {
		return Failure
	}

	result := u.child.Tick(ctx)
	if result == Success {
		return Success
	}
	// Failure 和 Running 都返回 Running 以继续重试
	return Running
}

// UntilFailure 是重复执行直到失败的装饰器节点。
// 子节点成功或运行时返回 Running，失败时返回 Success。
//
// 警告: 如果子节点永远不会失败，将无限循环。
type UntilFailure struct {
	child Node
}

// NewUntilFailure 创建一个新的直到失败装饰器节点。
//
// 参数:
//   - child: 要重复执行的子节点
//
// 返回值:
//   - Success: 子节点失败
//   - Running: 子节点成功或正在执行，继续重试
func NewUntilFailure(child Node) *UntilFailure {
	return &UntilFailure{
		child: child,
	}
}

// SetChild 设置直到失败装饰器的子节点。
func (u *UntilFailure) SetChild(child Node) {
	u.child = child
}

// Tick 执行直到失败逻辑。
// 子节点为 nil 时返回 Failure。
func (u *UntilFailure) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if u.child == nil {
		return Failure
	}

	result := u.child.Tick(ctx)
	if result == Failure {
		return Success
	}
	// Success 和 Running 都返回 Running 以继续重试
	return Running
}
