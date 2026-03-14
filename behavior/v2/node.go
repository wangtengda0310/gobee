package behavior

// Context 提供在节点执行期间传递数据的上下文。
// 使用 map 存储键值对，可在节点间共享状态。
type Context map[string]interface{}

// Node 是行为树节点接口。
// 所有行为树节点都必须实现此接口。
type Node interface {
	// Tick 执行节点逻辑并返回结果。
	//
	// 参数:
	//   - ctx: 上下文对象，用于在节点间传递数据
	//
	// 返回值:
	//   - Success: 节点执行成功
	//   - Failure: 节点执行失败
	//   - Running: 节点正在执行中
	Tick(ctx Context) Result
}

// CompositeNode 是可添加子节点的复合节点接口。
// Sequence、Selector、Parallel 等复合节点实现此接口。
type CompositeNode interface {
	Node
	// AddChild 添加一个子节点。
	AddChild(child Node)
}

// ResettableNode 是可重置状态的节点接口。
// 有状态的节点（如 Repeater、Retry）实现此接口以支持重用。
type ResettableNode interface {
	Node
	// Reset 重置节点的内部状态。
	Reset()
}
