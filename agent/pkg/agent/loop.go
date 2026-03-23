package agent

// LoopState 循环状态
// 用于跟踪执行循环的状态
type LoopState int

const (
	// LoopStateContinue 继续循环
	LoopStateContinue LoopState = iota
	// LoopStateDone 完成
	LoopStateDone
	// LoopStateError 错误
	LoopStateError
	// LoopStateTimeout 超时
	LoopStateTimeout
	// LoopStateMaxLoops 达到最大循环次数
	LoopStateMaxLoops
)

// String 返回状态字符串
func (s LoopState) String() string {
	switch s {
	case LoopStateContinue:
		return "continue"
	case LoopStateDone:
		return "done"
	case LoopStateError:
		return "error"
	case LoopStateTimeout:
		return "timeout"
	case LoopStateMaxLoops:
		return "max_loops"
	default:
		return "unknown"
	}
}

// LoopController 循环控制器
// 控制执行循环的行为
type LoopController struct {
	// MaxLoops 最大循环次数
	MaxLoops int

	// CurrentLoop 当前循环次数
	CurrentLoop int

	// State 当前状态
	State LoopState
}

// NewLoopController 创建循环控制器
func NewLoopController(maxLoops int) *LoopController {
	if maxLoops <= 0 {
		maxLoops = 10
	}
	return &LoopController{
		MaxLoops:    maxLoops,
		CurrentLoop: 0,
		State:       LoopStateContinue,
	}
}

// ShouldContinue 检查是否应该继续循环
func (c *LoopController) ShouldContinue() bool {
	return c.State == LoopStateContinue && c.CurrentLoop < c.MaxLoops
}

// Increment 增加循环计数
func (c *LoopController) Increment() {
	c.CurrentLoop++
}

// MarkDone 标记完成
func (c *LoopController) MarkDone() {
	c.State = LoopStateDone
}

// MarkError 标记错误
func (c *LoopController) MarkError() {
	c.State = LoopStateError
}

// MarkTimeout 标记超时
func (c *LoopController) MarkTimeout() {
	c.State = LoopStateTimeout
}

// CheckMaxLoops 检查是否达到最大循环次数
func (c *LoopController) CheckMaxLoops() {
	if c.CurrentLoop >= c.MaxLoops {
		c.State = LoopStateMaxLoops
	}
}
