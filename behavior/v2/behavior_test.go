package behavior

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultString(t *testing.T) {
	// 测试Success枚举值的字符串表示
	assert.Equal(t, "Success", Success.String(), "Success枚举值的字符串表示应为'Success'")
	// 测试Failure枚举值的字符串表示
	assert.Equal(t, "Failure", Failure.String(), "Failure枚举值的字符串表示应为'Failure'")
	// 测试Running枚举值的字符串表示
	assert.Equal(t, "Running", Running.String(), "Running枚举值的字符串表示应为'Running'")
	// 测试未知枚举值的字符串表示
	assert.Equal(t, "Unknown", Result(99).String(), "未知枚举值的字符串表示应为'Unknown'")
}

func TestContext(t *testing.T) {
	ctx := make(Context)
	ctx["key"] = "value"
	ctx["number"] = 42

	// 测试字符串类型值的存储和获取
	assert.Equal(t, "value", ctx["key"], "字符串类型值的存储和获取应正确")
	// 测试数值类型值的存储和获取
	assert.Equal(t, 42, ctx["number"], "数值类型值的存储和获取应正确")
}

func TestSequence(t *testing.T) {
	ctx := make(Context)

	// 测试所有子节点都成功的情况
	success1 := NewAction(func(_ Context) Result { return Success })
	success2 := NewAction(func(_ Context) Result { return Success })

	sequence := NewSequence(success1, success2)
	result := sequence.Tick(ctx)
	assert.Equal(t, Success, result, "所有子节点都成功，序列节点应返回成功")

	// 测试第一个子节点失败的情况
	failure := NewAction(func(_ Context) Result { return Failure })
	sequence = NewSequence(failure, success1)
	result = sequence.Tick(ctx)
	assert.Equal(t, Failure, result, "第一个子节点失败，序列节点应返回失败")

	// 测试子节点包含running状态的情况
	running := NewAction(func(_ Context) Result { return Running })
	sequence = NewSequence(success1, running, success2)
	result = sequence.Tick(ctx)
	assert.Equal(t, Running, result, "子节点返回running，序列节点应返回running")

	// 测试动态添加子节点
	sequence = NewSequence()
	sequence.AddChild(success1)
	sequence.AddChild(success2)
	result = sequence.Tick(ctx)
	assert.Equal(t, Success, result, "动态添加的子节点都成功，序列节点应返回成功")

	// 测试空序列节点
	emptySequence := NewSequence()
	result = emptySequence.Tick(ctx)
	assert.Equal(t, Success, result, "空序列节点应返回成功")
}

func TestSelector(t *testing.T) {
	ctx := make(Context)

	// 测试第一个子节点成功的情况
	success1 := NewAction(func(_ Context) Result { return Success })
	success2 := NewAction(func(_ Context) Result { return Success })

	selector := NewSelector(success1, success2)
	result := selector.Tick(ctx)
	assert.Equal(t, Success, result, "第一个子节点成功，选择器节点应返回成功")

	// 测试所有子节点都失败的情况
	failure1 := NewAction(func(_ Context) Result { return Failure })
	failure2 := NewAction(func(_ Context) Result { return Failure })

	selector = NewSelector(failure1, failure2)
	result = selector.Tick(ctx)
	assert.Equal(t, Failure, result, "所有子节点都失败，选择器节点应返回失败")

	// 测试子节点包含running状态的情况
	running := NewAction(func(_ Context) Result { return Running })
	selector = NewSelector(failure1, running, success1)
	result = selector.Tick(ctx)
	assert.Equal(t, Running, result, "子节点返回running，选择器节点应返回running")

	// 测试动态添加子节点
	selector = NewSelector()
	selector.AddChild(failure1)
	selector.AddChild(success1)
	result = selector.Tick(ctx)
	assert.Equal(t, Success, result, "动态添加的子节点中有一个成功，选择器节点应返回成功")

	// 测试空选择器节点
	emptySelector := NewSelector()
	result = emptySelector.Tick(ctx)
	assert.Equal(t, Failure, result, "空选择器节点应返回失败")
}

func TestParallel(t *testing.T) {
	ctx := make(Context)

	// 测试并行节点的成功和失败条件
	success1 := NewAction(func(_ Context) Result { return Success })
	success2 := NewAction(func(_ Context) Result { return Success })
	failure := NewAction(func(_ Context) Result { return Failure })

	// 需要2个成功，但1个失败就返回失败
	parallel := NewParallel(2, 1, success1, success2, failure)
	result := parallel.Tick(ctx)
	assert.Equal(t, Failure, result, "因为有1个失败，并行节点应返回失败")

	// 需要1个成功，3个失败才返回失败
	parallel = NewParallel(1, 3, success1, failure, failure)
	result = parallel.Tick(ctx)
	assert.Equal(t, Success, result, "因为有1个成功，并行节点应返回成功")

	// 测试包含running状态的情况
	running := NewAction(func(_ Context) Result { return Running })
	parallel = NewParallel(2, 2, success1, running, failure)
	result = parallel.Tick(ctx)
	assert.Equal(t, Running, result, "有子节点返回running，并行节点应返回running")

	// 测试动态添加子节点
	parallel = NewParallel(2, 2)
	parallel.AddChild(success1)
	parallel.AddChild(success2)
	result = parallel.Tick(ctx)
	assert.Equal(t, Success, result, "动态添加的子节点中有2个成功，并行节点应返回成功")

	// 测试空并行节点
	emptyParallel := NewParallel(1, 1)
	result = emptyParallel.Tick(ctx)
	assert.Equal(t, Success, result, "空并行节点应返回成功")
}

func TestCondition(t *testing.T) {
	ctx := make(Context)
	ctx["test"] = true

	// 测试条件为true的情况
	trueCond := NewCondition(func(ctx Context) bool {
		return ctx["test"].(bool)
	})
	result := trueCond.Tick(ctx)
	assert.Equal(t, Success, result, "条件为true时，条件节点应返回成功")

	// 测试条件为false的情况
	falseCond := NewCondition(func(ctx Context) bool {
		return false
	})
	result = falseCond.Tick(ctx)
	assert.Equal(t, Failure, result, "条件为false时，条件节点应返回失败")
}

func TestAction(t *testing.T) {
	ctx := make(Context)

	// 测试返回成功的动作节点
	successAction := NewAction(func(_ Context) Result { return Success })
	result := successAction.Tick(ctx)
	assert.Equal(t, Success, result, "动作返回成功，动作节点应返回成功")

	// 测试返回失败的动作节点
	failureAction := NewAction(func(_ Context) Result { return Failure })
	result = failureAction.Tick(ctx)
	assert.Equal(t, Failure, result, "动作返回失败，动作节点应返回失败")

	// 测试返回运行中的动作节点
	runningAction := NewAction(func(_ Context) Result { return Running })
	result = runningAction.Tick(ctx)
	assert.Equal(t, Running, result, "动作返回运行中，动作节点应返回运行中")

	// 测试动作节点对外部变量的修改
	counter := 0
	action := NewAction(func(_ Context) Result {
		counter++
		return Success
	})
	action.Tick(ctx)
	assert.Equal(t, 1, counter, "动作执行后计数器应增加1")
}
