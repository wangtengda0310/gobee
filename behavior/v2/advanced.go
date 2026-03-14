package behavior

import (
	"math/rand"
	"time"
)

// RandomSelector 是以随机顺序执行子节点的选择器节点。
// 与 Selector 类似，但子节点执行顺序随机。
//
// 示例:
//
//	// 随机尝试不同的攻击策略
//	selector := NewRandomSelector(meleeAttack, rangedAttack, magicAttack)
type RandomSelector struct {
	children []Node
	rng      *rand.Rand
}

// NewRandomSelector 创建一个新的随机选择器节点。
//
// 执行规则:
//   - 使用 Fisher-Yates 洗牌算法随机排序
//   - 按随机顺序执行子节点
//   - 首个非 Failure 的结果即返回
//   - 所有子节点都失败才返回 Failure
//
// 参数:
//   - children: 子节点列表
//
// 返回值:
//   - *RandomSelector: 随机选择器节点指针
//
// 边界情况: 无子节点时返回 Failure
//
// 注意: 随机种子在节点创建时使用当前时间戳初始化。
func NewRandomSelector(children ...Node) *RandomSelector {
	return &RandomSelector{
		children: children,
		// 注意: RNG 在节点创建时初始化一次，保证同一执行周期内随机序列一致
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddChild 添加一个子节点到随机选择器。
func (r *RandomSelector) AddChild(child Node) {
	r.children = append(r.children, child)
}

// Tick 执行随机选择器逻辑。
// 无子节点时返回 Failure。
func (r *RandomSelector) Tick(ctx Context) Result {
	if len(r.children) == 0 {
		return Failure
	}

	// Fisher-Yates 洗牌算法 - O(n) 原地随机排列
	// 注意: 洗牌索引而不是节点，保留原始顺序
	indices := make([]int, len(r.children))
	for i := range indices {
		indices[i] = i
	}
	for i := len(indices) - 1; i > 0; i-- {
		j := r.rng.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	for _, idx := range indices {
		result := r.children[idx].Tick(ctx)
		if result != Failure {
			return result
		}
	}
	return Failure
}

// Retry 是在失败时重试的装饰器节点。
// 与 Repeater 不同，Retry 只在子节点失败时重试。
//
// 示例:
//
//	// 最多重试 3 次连接
//	retry := NewRetry(3, connectAction)
type Retry struct {
	child    Node
	maxTries int // -1 表示无限重试
	tryCount int // 当前重试计数
}

// NewRetry 创建一个新的重试装饰器节点。
//
// 执行规则:
//   - 子节点成功时返回 Success，重置计数器
//   - 子节点失败时增加重试计数
//   - 达到最大重试次数后返回 Failure
//   - 未达到最大次数时返回 Running（表示正在重试）
//
// 参数:
//   - maxTries: 最大重试次数 (-1 表示无限重试)
//   - child: 要执行的子节点
//
// 返回值:
//   - Success: 子节点执行成功
//   - Failure: 达到最大重试次数仍未成功
//   - Running: 正在重试中
//
// 警告: maxTries=-1 时，如果子节点永远不成功将无限循环。
func NewRetry(maxTries int, child Node) *Retry {
	return &Retry{
		child:    child,
		maxTries: maxTries,
		tryCount: 0,
	}
}

// SetChild 设置重试装饰器的子节点。
func (r *Retry) SetChild(child Node) {
	r.child = child
}

// Reset 重置重试装饰器的计数器。
// 调用此方法可以在新的执行周期中复用重试器。
func (r *Retry) Reset() {
	r.tryCount = 0
}

// Tick 执行重试逻辑。
// 子节点为 nil 时返回 Failure。
func (r *Retry) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if r.child == nil {
		return Failure
	}

	result := r.child.Tick(ctx)

	switch result {
	case Success:
		// 重置以支持复用
		r.tryCount = 0
		return Success
	case Failure:
		r.tryCount++
		if r.maxTries > 0 && r.tryCount >= r.maxTries {
			// 重置以支持复用
			r.tryCount = 0
			return Failure
		}
		// 继续重试，返回 Running 表示正在进行
		return Running
	case Running:
		return Running
	}

	return Failure
}

// Timeout 是在指定时间后强制失败的装饰器节点。
// 计时器在首次 tick 时启动，子节点完成或超时时重置。
//
// 示例:
//
//	// 搜索操作最多执行 5 秒
//	timeout := NewTimeout(5*time.Second, searchAction)
type Timeout struct {
	child     Node
	duration  time.Duration
	startTime time.Time
	started   bool
}

// NewTimeout 创建一个新的超时装饰器节点。
//
// 执行规则:
//   - 首次 tick 启动计时器
//   - 超时后立即返回 Failure
//   - 子节点完成（成功或失败）时重置计时器
//
// 参数:
//   - duration: 最大允许执行时间
//   - child: 要执行的子节点
//
// 返回值:
//   - Success: 子节点在超时前成功
//   - Failure: 超时或子节点失败
//   - Running: 子节点仍在执行且未超时
func NewTimeout(duration time.Duration, child Node) *Timeout {
	return &Timeout{
		child:    child,
		duration: duration,
		started:  false,
	}
}

// SetChild 设置超时装饰器的子节点。
func (t *Timeout) SetChild(child Node) {
	t.child = child
}

// Reset 重置超时装饰器的状态。
// 调用此方法可以在新的执行周期中复用超时器。
func (t *Timeout) Reset() {
	t.started = false
	t.startTime = time.Time{}
}

// Tick 执行超时逻辑。
// 子节点为 nil 时返回 Failure。
func (t *Timeout) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if t.child == nil {
		return Failure
	}

	// 首次 tick 初始化计时器
	if !t.started {
		t.started = true
		t.startTime = time.Now()
	}

	// 执行子节点前检查超时
	if time.Since(t.startTime) > t.duration {
		// 重置状态以支持复用
		t.started = false
		return Failure
	}

	result := t.child.Tick(ctx)
	if result != Running {
		// 子节点完成时重置状态
		t.started = false
	}

	return result
}

// Delay 是延迟指定 tick 数后执行子节点的装饰器节点。
// 延迟期间不执行子节点，只返回 Running。
//
// 示例:
//
//	// 延迟 3 个 tick 后攻击
//	delay := NewDelay(3, attackAction)
type Delay struct {
	child      Node
	delayTicks int // 延迟的 tick 数
	tickCount  int // 当前 tick 计数
}

// NewDelay 创建一个新的延迟装饰器节点。
//
// 执行规则:
//   - 延迟期间返回 Running，不执行子节点
//   - 延迟结束后执行子节点
//   - 子节点完成时重置延迟计数
//
// 参数:
//   - delayTicks: 延迟的 tick 数
//   - child: 延迟后要执行的子节点
//
// 返回值:
//   - Running: 延迟期间或子节点正在执行
//   - Success/Failure: 延迟结束后子节点的结果
//
// 用途: 用于引入延迟，如冷却时间、动画延迟等。
func NewDelay(delayTicks int, child Node) *Delay {
	return &Delay{
		child:      child,
		delayTicks: delayTicks,
		tickCount:  0,
	}
}

// SetChild 设置延迟装饰器的子节点。
func (d *Delay) SetChild(child Node) {
	d.child = child
}

// Reset 重置延迟装饰器的计数器。
// 调用此方法可以在新的执行周期中复用延迟器。
func (d *Delay) Reset() {
	d.tickCount = 0
}

// Tick 执行延迟逻辑。
// 子节点为 nil 时返回 Failure。
func (d *Delay) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if d.child == nil {
		return Failure
	}

	// 延迟阶段 - 返回 Running，不执行子节点
	if d.tickCount < d.delayTicks {
		d.tickCount++
		return Running
	}

	result := d.child.Tick(ctx)
	if result != Running {
		// 子节点完成时重置延迟
		d.tickCount = 0
	}

	return result
}

// Limiter 是限制成功执行次数的装饰器节点。
// 达到限制后，后续调用立即返回 Failure。
//
// 示例:
//
//	// 最多成功治疗 3 次
//	limiter := NewLimiter(3, healAction)
type Limiter struct {
	child     Node
	maxCalls  int // -1 表示无限制
	callCount int // 当前成功调用计数
}

// NewLimiter 创建一个新的限制装饰器节点。
//
// 执行规则:
//   - 执行子节点
//   - 子节点成功时增加计数
//   - 达到限制后立即返回 Failure，不执行子节点
//
// 参数:
//   - maxCalls: 最大成功执行次数 (-1 表示无限制)
//   - child: 要执行的子节点
//
// 返回值:
//   - Success: 子节点成功且未达到限制
//   - Failure: 达到限制或子节点失败
//   - Running: 子节点正在执行
//
// 注意: 只有成功执行才计数，Failure 和 Running 不计数。
// 用途: 限制资源使用、限流等场景。
func NewLimiter(maxCalls int, child Node) *Limiter {
	return &Limiter{
		child:     child,
		maxCalls:  maxCalls,
		callCount: 0,
	}
}

// SetChild 设置限制装饰器的子节点。
func (l *Limiter) SetChild(child Node) {
	l.child = child
}

// Reset 重置限制装饰器的计数器。
// 调用此方法可以在新的执行周期中复用限制器。
func (l *Limiter) Reset() {
	l.callCount = 0
}

// Tick 执行限制逻辑。
// 子节点为 nil 时返回 Failure。
func (l *Limiter) Tick(ctx Context) Result {
	// 边界情况: nil 子节点返回 Failure
	if l.child == nil {
		return Failure
	}

	// 执行子节点前检查限制
	if l.maxCalls > 0 && l.callCount >= l.maxCalls {
		return Failure
	}

	result := l.child.Tick(ctx)
	// 只有成功执行才计数
	if result == Success {
		l.callCount++
	}

	return result
}
