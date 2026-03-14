package behavior

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExportDOTString(t *testing.T) {
	// Create a simple behavior tree
	action1 := NewAction(func(_ Context) Result { return Success })
	action2 := NewAction(func(_ Context) Result { return Success })
	condition := NewCondition(func(_ Context) bool { return true })

	sequence := NewSequence(condition, action1)
	selector := NewSelector(sequence, action2)

	// Export to DOT
	dot, err := ExportDOTString(selector)
	assert.NoError(t, err, "ExportDOTString should not error")
	assert.Contains(t, dot, "digraph behavior_tree", "DOT should contain digraph declaration")
	assert.Contains(t, dot, "Selector", "DOT should contain Selector node")
	assert.Contains(t, dot, "Sequence", "DOT should contain Sequence node")
	assert.Contains(t, dot, "Condition", "DOT should contain Condition node")
	assert.Contains(t, dot, "Action", "DOT should contain Action node")
}

func TestExportDOTWithDecorators(t *testing.T) {
	action := NewAction(func(_ Context) Result { return Success })
	inverter := NewInverter(action)
	repeater := NewRepeater(3, inverter)

	dot, err := ExportDOTString(repeater)
	assert.NoError(t, err, "ExportDOTString should not error")
	assert.Contains(t, dot, "Repeater", "DOT should contain Repeater node")
	assert.Contains(t, dot, "Inverter", "DOT should contain Inverter node")
	assert.Contains(t, dot, "Action", "DOT should contain Action node")
}

func TestExportDOTWithParallel(t *testing.T) {
	action1 := NewAction(func(_ Context) Result { return Success })
	action2 := NewAction(func(_ Context) Result { return Success })
	parallel := NewParallel(1, 1, action1, action2)

	dot, err := ExportDOTString(parallel)
	assert.NoError(t, err, "ExportDOTString should not error")
	assert.Contains(t, dot, "Parallel", "DOT should contain Parallel node")
}

func TestExportDOTWithAdvancedNodes(t *testing.T) {
	action := NewAction(func(_ Context) Result { return Success })

	// Test Retry
	retry := NewRetry(3, action)
	dot, err := ExportDOTString(retry)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Retry", "DOT should contain Retry node")

	// Test Delay
	delay := NewDelay(5, action)
	dot, err = ExportDOTString(delay)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Delay", "DOT should contain Delay node")

	// Test Limiter
	limiter := NewLimiter(10, action)
	dot, err = ExportDOTString(limiter)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Limiter", "DOT should contain Limiter node")

	// Test Timeout
	timeout := NewTimeout(5*time.Second, action)
	dot, err = ExportDOTString(timeout)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Timeout", "DOT should contain Timeout node")

	// Test RandomSelector
	randomSelector := NewRandomSelector(action, action)
	dot, err = ExportDOTString(randomSelector)
	assert.NoError(t, err)
	assert.Contains(t, dot, "RandomSelector", "DOT should contain RandomSelector node")
}

func TestExportDOTWithUntilDecorators(t *testing.T) {
	action := NewAction(func(_ Context) Result { return Success })

	// Test UntilSuccess
	untilSuccess := NewUntilSuccess(action)
	dot, err := ExportDOTString(untilSuccess)
	assert.NoError(t, err)
	assert.Contains(t, dot, "UntilSuccess", "DOT should contain UntilSuccess node")

	// Test UntilFailure
	untilFailure := NewUntilFailure(action)
	dot, err = ExportDOTString(untilFailure)
	assert.NoError(t, err)
	assert.Contains(t, dot, "UntilFailure", "DOT should contain UntilFailure node")
}

func TestExportDOTWithInfiniteRepeater(t *testing.T) {
	action := NewAction(func(_ Context) Result { return Success })
	repeater := NewRepeater(-1, action)

	dot, err := ExportDOTString(repeater)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Repeater", "DOT should contain Repeater node")
	assert.Contains(t, dot, "∞", "Infinite repeater should show infinity symbol")
}

func TestExportDOTStructure(t *testing.T) {
	// Create a tree structure and verify the DOT output structure
	action := NewAction(func(_ Context) Result { return Success })
	sequence := NewSequence(action)

	dot, err := ExportDOTString(sequence)
	assert.NoError(t, err)

	// Check that the DOT contains expected structural elements
	lines := strings.Split(dot, "\n")
	hasDigraph := false
	hasNodeDef := false
	hasEdgeDef := false

	for _, line := range lines {
		if strings.Contains(line, "digraph") {
			hasDigraph = true
		}
		if strings.Contains(line, "node [") {
			hasNodeDef = true
		}
		if strings.Contains(line, "edge [") {
			hasEdgeDef = true
		}
	}

	assert.True(t, hasDigraph, "DOT should have digraph declaration")
	assert.True(t, hasNodeDef, "DOT should have node definition")
	assert.True(t, hasEdgeDef, "DOT should have edge definition")
}

func TestExportDOTWithNestedTree(t *testing.T) {
	// Create a complex nested tree
	action1 := NewAction(func(_ Context) Result { return Success })
	action2 := NewAction(func(_ Context) Result { return Failure })
	action3 := NewAction(func(_ Context) Result { return Running })

	condition := NewCondition(func(_ Context) bool { return true })

	// Nested structure: Selector -> [Sequence, Inverter -> Action]
	sequence := NewSequence(condition, action1)
	inverter := NewInverter(action2)
	selector := NewSelector(sequence, inverter, action3)

	dot, err := ExportDOTString(selector)
	assert.NoError(t, err)

	// Verify all nodes are present
	assert.Contains(t, dot, "Selector")
	assert.Contains(t, dot, "Sequence")
	assert.Contains(t, dot, "Condition")
	assert.Contains(t, dot, "Action")
	assert.Contains(t, dot, "Inverter")

	// Verify edges exist (arrow notation)
	assert.Contains(t, dot, "->", "DOT should contain edge arrows")
}

func TestExportDOTNodeColors(t *testing.T) {
	// Test that different node types have different colors
	tests := []struct {
		name     string
		node     Node
		contains string
	}{
		{"Sequence", NewSequence(), "#90EE90"},
		{"Selector", NewSelector(), "#87CEEB"},
		{"Parallel", NewParallel(1, 1), "#DDA0DD"},
		{"Condition", NewCondition(func(_ Context) bool { return true }), "#FFD700"},
		{"Action", NewAction(func(_ Context) Result { return Success }), "#FFA07A"},
		{"Inverter", NewInverter(nil), "#F0E68C"},
		{"Repeater", NewRepeater(3, nil), "#F0E68C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dot, err := ExportDOTString(tt.node)
			assert.NoError(t, err)
			assert.Contains(t, dot, tt.contains, "Node should have correct color")
		})
	}
}

func TestExportDOTWithEmptyNodes(t *testing.T) {
	// Test empty composite nodes
	emptySequence := NewSequence()
	dot, err := ExportDOTString(emptySequence)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Sequence", "Empty Sequence should be exported")

	emptySelector := NewSelector()
	dot, err = ExportDOTString(emptySelector)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Selector", "Empty Selector should be exported")

	emptyParallel := NewParallel(1, 1)
	dot, err = ExportDOTString(emptyParallel)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Parallel", "Empty Parallel should be exported")
}

func TestExportDOTWithAddChild(t *testing.T) {
	// Test nodes created with AddChild
	action := NewAction(func(_ Context) Result { return Success })

	sequence := NewSequence()
	sequence.AddChild(action)

	dot, err := ExportDOTString(sequence)
	assert.NoError(t, err)
	assert.Contains(t, dot, "Sequence")
	assert.Contains(t, dot, "Action")

	selector := NewSelector()
	selector.AddChild(action)
	selector.AddChild(action)

	dot, err = ExportDOTString(selector)
	assert.NoError(t, err)
	// Should have multiple Action nodes
	actionCount := strings.Count(dot, "Action")
	assert.GreaterOrEqual(t, actionCount, 1, "Should have at least one Action node")
}

// VisualizableNode 测试
type customVisualizableNode struct {
	name string
}

func (n *customVisualizableNode) Tick(ctx Context) Result {
	return Success
}

func (n *customVisualizableNode) VisualName() string {
	return n.name
}

func TestExportDOTWithVisualizableNode(t *testing.T) {
	customNode := &customVisualizableNode{name: "CustomNode"}
	dot, err := ExportDOTString(customNode)
	assert.NoError(t, err)
	assert.Contains(t, dot, "CustomNode", "Custom VisualizableNode should use VisualName")
}
