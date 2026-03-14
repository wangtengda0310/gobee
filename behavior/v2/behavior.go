// Package behavior 提供行为树实现，用于 AI 决策逻辑。
package behavior

// Result 表示行为树节点执行后的返回结果。
type Result int

const (
	// Success 表示节点执行成功。
	Success Result = iota
	// Failure 表示节点执行失败。
	Failure
	// Running 表示节点正在执行中，需要再次 tick。
	Running
)

// String 返回 Result 的字符串表示。
func (r Result) String() string {
	switch r {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	case Running:
		return "Running"
	default:
		return "Unknown"
	}
}
