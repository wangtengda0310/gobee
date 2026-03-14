package behavior

// ActionFunc 是动作函数的类型定义。
type ActionFunc func(ctx Context) Result

// Action 表示执行具体行为的动作节点。
// 动作节点是行为树的叶子节点，执行实际的操作。
//
// 示例:
//
//	attack := NewAction(func(ctx Context) Result {
//	    fmt.Println("攻击中...")
//	    return Success
//	})
type Action struct {
	action ActionFunc
}

// NewAction 创建一个新的动作节点。
//
// 参数:
//   - action: 动作函数，接收上下文并返回执行结果
//
// 返回值:
//   - *Action: 动作节点指针
//
// 示例:
//
//	move := NewAction(func(ctx Context) Result {
//	    position := ctx["position"].(int)
//	    ctx["position"] = position + 1
//	    return Success
//	})
func NewAction(action ActionFunc) *Action {
	return &Action{
		action: action,
	}
}

// Tick 执行动作逻辑。
func (a *Action) Tick(ctx Context) Result {
	return a.action(ctx)
}
