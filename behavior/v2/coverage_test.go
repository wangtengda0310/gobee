package behavior

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// 针对覆盖率不足的部分添加额外的测试

func TestParallelEdgeCases(t *testing.T) {
    ctx := make(Context)
    
    // 测试边界条件：successPolicy大于子节点数量
    success1 := NewAction(func(_ Context) Result { return Success })
    parallel := NewParallel(3, 1, success1) // 需要3个成功，但只有1个子节点
    result := parallel.Tick(ctx)
    assert.Equal(t, Failure, result, "应该失败，因为没有足够的成功")
    
    // 测试边界条件：failurePolicy大于子节点数量
    failure1 := NewAction(func(_ Context) Result { return Failure })
    parallel = NewParallel(1, 3, failure1) // 需要3个失败，但只有1个子节点
    result = parallel.Tick(ctx)
    assert.Equal(t, Failure, result, "应该失败，因为没有足够的成功")
    
    // 测试边界条件：successPolicy和failurePolicy都为0
    parallel = NewParallel(0, 0, success1)
    result = parallel.Tick(ctx)
    assert.Equal(t, Failure, result, "应该失败，因为默认行为")
    
    // 测试所有子节点都返回running的情况
    running1 := NewAction(func(_ Context) Result { return Running })
    running2 := NewAction(func(_ Context) Result { return Running })
    parallel = NewParallel(2, 2, running1, running2)
    result = parallel.Tick(ctx)
    assert.Equal(t, Running, result, "所有子节点都返回running时，并行节点应返回running")
}

func TestRepeaterInfinite(t *testing.T) {
    ctx := make(Context)
    
    // 测试无限重复（times=-1）的情况
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        return Success
    })
    
    repeater := NewRepeater(-1, action)
    
    // 运行几次，确认它一直返回Running
    for i := 0; i < 5; i++ {
        result := repeater.Tick(ctx)
        assert.Equal(t, Running, result, "无限重复器应一直返回running")
    }
    assert.Equal(t, 5, counter, "动作应被执行5次")
}

func TestRepeaterWithZeroTimes(t *testing.T) {
    ctx := make(Context)
    
    // 测试重复0次的情况
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        return Success
    })
    
    repeater := NewRepeater(0, action)
    result := repeater.Tick(ctx)
    assert.Equal(t, Success, result, "重复0次时应立即返回成功")
    assert.Equal(t, 0, counter, "动作不应该被执行")
}

func TestRepeaterWithRunningCompletion(t *testing.T) {
    ctx := make(Context)
    
    // 测试重复器在子节点返回Running后再完成的情况
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        if counter == 1 {
            return Running // 第一次返回Running
        }
        return Success
    })
    
    repeater := NewRepeater(2, action)
    
    // 第一次tick
    result := repeater.Tick(ctx)
    assert.Equal(t, Running, result, "子节点返回running时，重复器应返回running")
    assert.Equal(t, 1, counter, "计数器应增加到1")
    
    // 第二次tick
    result = repeater.Tick(ctx)
    assert.Equal(t, Running, result, "未完成所有重复次数时应返回running")
    assert.Equal(t, 2, counter, "计数器应增加到2")
    
    // 第三次tick
    result = repeater.Tick(ctx)
    assert.Equal(t, Success, result, "完成所有重复次数后应返回成功")
    assert.Equal(t, 3, counter, "计数器应增加到3")
}

func TestNestedCompositeNodes(t *testing.T) {
    ctx := make(Context)
    
    // 测试嵌套的复合节点
    success1 := NewAction(func(_ Context) Result { return Success })
    success2 := NewAction(func(_ Context) Result { return Success })
    failure := NewAction(func(_ Context) Result { return Failure })
    
    // 创建嵌套结构：Selector包含一个Sequence
    sequence := NewSequence(success1, success2)
    selector := NewSelector(failure, sequence)
    
    result := selector.Tick(ctx)
    assert.Equal(t, Success, result, "嵌套的复合节点应正确执行并返回成功")
    
    // 创建更复杂的嵌套结构
    innerSelector := NewSelector(failure, success1)
    outerSequence := NewSequence(innerSelector, success2)
    
    result = outerSequence.Tick(ctx)
    assert.Equal(t, Success, result, "更复杂的嵌套结构应正确执行并返回成功")
}

func TestNestedDecorators(t *testing.T) {
    ctx := make(Context)
    
    // 测试嵌套的装饰器
    success := NewAction(func(_ Context) Result { return Success })
    
    // Inverter嵌套Inverter应该恢复原始结果
    doubleInverter := NewInverter(NewInverter(success))
    result := doubleInverter.Tick(ctx)
    assert.Equal(t, Success, result, "两次反转应恢复原始结果")
    
    // 测试Repeater嵌套的行为
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        return Success
    })
    
    // Repeater嵌套Inverter
    repeater := NewRepeater(2, NewInverter(action))
    
    // 第一次tick
    result = repeater.Tick(ctx)
    assert.Equal(t, Running, result, "未完成所有重复次数时应返回running")
    assert.Equal(t, 1, counter, "计数器应增加到1")
    
    // 第二次tick
    result = repeater.Tick(ctx)
    assert.Equal(t, Success, result, "完成所有重复次数后应返回成功")
    assert.Equal(t, 2, counter, "计数器应增加到2")
}