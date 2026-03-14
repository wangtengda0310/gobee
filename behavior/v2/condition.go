package behavior

// ConditionFunc 是条件函数的类型定义。
type ConditionFunc func(ctx Context) bool

// Condition 表示评估布尔表达式的条件节点。
// 条件节点是行为树的叶子节点，用于决策判断。
//
// 示例:
//
//	checkHealth := NewCondition(func(ctx Context) bool {
//	    return ctx["health"].(int) > 50
//	})
type Condition struct {
	condition ConditionFunc
}

// NewCondition 创建一个新的条件节点。
//
// 参数:
//   - condition: 条件函数，接收上下文并返回布尔值
//
// 返回值:
//   - *Condition: 条件节点指针
//
// 示例:
//
//	hasTarget := NewCondition(func(ctx Context) bool {
//	    _, exists := ctx["target"]
//	    return exists
//	})
func NewCondition(condition ConditionFunc) *Condition {
	return &Condition{
		condition: condition,
	}
}

// Tick 执行条件逻辑。
// 条件为真返回 Success，条件为假返回 Failure。
func (c *Condition) Tick(ctx Context) Result {
	if c.condition(ctx) {
		return Success
	}
	return Failure
}
