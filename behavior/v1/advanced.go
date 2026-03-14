package behavior

import (
	"math/rand"
	"time"
)

// RandomSelector 创建一个随机选择器节点，以随机顺序执行子节点。
// 与 Selector 类似，但子节点执行顺序随机。
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
//   - Node: 随机选择器节点函数
//
// 边界情况: 空选择器返回 Failure
//
// 注意: 随机种子在节点创建时初始化，使用当前时间戳。
//
// 示例:
//
//	// 随机尝试不同的攻击策略
//	selector := RandomSelector(meleeAttack, rangedAttack, magicAttack)
func RandomSelector(children ...Node) Node {
	// 注意: RNG 在节点创建时初始化一次，保证同一执行周期内随机序列一致
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(ctx Context) Result {
		if len(children) == 0 {
			return Failure
		}

		// Fisher-Yates 洗牌算法 - O(n) 原地随机排列
		indices := make([]int, len(children))
		for i := range indices {
			indices[i] = i
		}
		for i := len(indices) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			indices[i], indices[j] = indices[j], indices[i]
		}

		for _, idx := range indices {
			result := children[idx](ctx)
			if result != Failure {
				return result
			}
		}
		return Failure
	}
}

// Retry 创建一个重试装饰器节点，在失败时重试。
// 与 Repeater 不同，Retry 只在子节点失败时重试。
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
//
// 示例:
//
//	// 最多重试 3 次
//	retry := Retry(3, unreliableAction)
func Retry(maxTries int, child Node) Node {
	tryCount := 0
	return func(ctx Context) Result {
		result := child(ctx)

		switch result {
		case Success:
			// 重置以支持复用
			tryCount = 0
			return Success
		case Failure:
			tryCount++
			if maxTries > 0 && tryCount >= maxTries {
				// 重置以支持复用
				tryCount = 0
				return Failure
			}
			// 继续重试，返回 Running 表示正在进行
			return Running
		case Running:
			return Running
		}

		return Failure
	}
}

// Timeout 创建一个超时装饰器节点，在指定时间后强制失败。
// 计时器在首次 tick 时启动，子节点完成或超时时重置。
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
//
// 示例:
//
//	// 搜索操作最多执行 5 秒
//	timeout := Timeout(5*time.Second, searchAction)
func Timeout(duration time.Duration, child Node) Node {
	started := false
	var startTime time.Time
	return func(ctx Context) Result {
		// 首次 tick 初始化计时器
		if !started {
			started = true
			startTime = time.Now()
		}

		// 执行子节点前检查超时
		if time.Since(startTime) > duration {
			// 重置状态以支持复用
			started = false
			return Failure
		}

		result := child(ctx)
		if result != Running {
			// 子节点完成时重置状态
			started = false
		}

		return result
	}
}

// Delay 创建一个延迟装饰器节点，延迟指定 tick 数后执行子节点。
// 延迟期间不执行子节点，只返回 Running。
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
//
// 示例:
//
//	// 延迟 3 个 tick 后攻击
//	delay := Delay(3, attackAction)
func Delay(delayTicks int, child Node) Node {
	tickCount := 0
	return func(ctx Context) Result {
		// 延迟阶段 - 返回 Running，不执行子节点
		if tickCount < delayTicks {
			tickCount++
			return Running
		}

		result := child(ctx)
		if result != Running {
			// 子节点完成时重置延迟
			tickCount = 0
		}

		return result
	}
}

// Limiter 创建一个限制装饰器节点，限制成功执行的次数。
// 达到限制后，后续调用立即返回 Failure。
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
//
// 示例:
//
//	// 最多成功治疗 3 次
//	limiter := Limiter(3, healAction)
func Limiter(maxCalls int, child Node) Node {
	callCount := 0
	return func(ctx Context) Result {
		// 执行子节点前检查限制
		if maxCalls > 0 && callCount >= maxCalls {
			return Failure
		}

		result := child(ctx)
		// 只有成功执行才计数
		if result == Success {
			callCount++
		}

		return result
	}
}
