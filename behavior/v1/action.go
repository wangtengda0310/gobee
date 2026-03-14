package behavior

// Action 创建一个动作节点函数。
// 动作节点执行具体的行为逻辑，如移动、攻击、交互等。
//
// 参数:
//   - action: 动作函数，接收上下文并返回执行结果
//
// 返回值:
//   - Node: 可直接在行为树中使用的节点函数
//
// 示例:
//
//	attack := Action(func(ctx Context) Result {
//	    fmt.Println("攻击中...")
//	    return Success
//	})
func Action(action func(ctx Context) Result) Node {
	return action
}
