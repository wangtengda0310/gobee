package tests

import (
	"fmt"
	"testing"

	"github.com/wangtengda0310/gobee/agent/pkg/agent"
)

// TestAgentEventTypes 打印所有 Agent 事件类型
func TestAgentEventTypes(t *testing.T) {
	fmt.Printf("=== Agent 事件类型 ===\n")
	fmt.Printf("EventTypeContent: %s\n", agent.EventTypeContent)

	// 通过反射获取所有 EventType 常量
	// 这里手动列出已知的类型
	fmt.Printf("\n已知的事件类型:\n")
	fmt.Printf("  - content: 内容增量\n")
	fmt.Printf("  - tool_call: 工具调用\n")
	fmt.Printf("  - tool_result: 工具结果\n")
	fmt.Printf("  - done: 完成\n")
	fmt.Printf("  - error: 错误\n")
}
