package behavior

// Condition 创建一个条件节点函数。
// 条件节点评估布尔表达式，返回成功或失败。
//
// 参数:
//   - condition: 条件函数，接收上下文并返回布尔值
//
// 返回值:
//   - Node: 条件节点函数，条件为真返回 Success，否则返回 Failure
//
// 示例:
//
//	checkHealth := Condition(func(ctx Context) bool {
//	    return ctx["health"].(int) > 50
//	})
func Condition(condition func(ctx Context) bool) Node {
	return func(ctx Context) Result {
		if condition(ctx) {
			return Success
		}
		return Failure
	}
}
