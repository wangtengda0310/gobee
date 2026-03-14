package behavior

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestInverter(t *testing.T) {
    ctx := make(Context)
    
    // 测试反转成功结果
    successNode := NewAction(func(_ Context) Result { return Success })
    inverter := NewInverter(successNode)
    result := inverter.Tick(ctx)
    assert.Equal(t, Failure, result, "成功结果应被反转为失败")
    
    // 测试反转失败结果
    failureNode := NewAction(func(_ Context) Result { return Failure })
    inverter = NewInverter(failureNode)
    result = inverter.Tick(ctx)
    assert.Equal(t, Success, result, "失败结果应被反转为成功")
    
    // 测试反转运行中结果
    runningNode := NewAction(func(_ Context) Result { return Running })
    inverter = NewInverter(runningNode)
    result = inverter.Tick(ctx)
    assert.Equal(t, Running, result, "运行中结果不应被反转")
    
    // 测试动态设置子节点
    inverter = NewInverter(nil)
    inverter.SetChild(successNode)
    result = inverter.Tick(ctx)
    assert.Equal(t, Failure, result, "动态设置子节点后应正确反转结果")
    
    // 测试空子节点
    inverter = NewInverter(nil)
    result = inverter.Tick(ctx)
    assert.Equal(t, Failure, result, "空子节点应返回失败")
}

func TestRepeater(t *testing.T) {
    ctx := make(Context)
    
    // 测试有限次数重复器
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        return Success
    })
    
    // 重复3次
    repeater := NewRepeater(3, action)
    
    // 第一次执行
    result := repeater.Tick(ctx)
    assert.Equal(t, Running, result, "未达到重复次数应返回运行中")
    assert.Equal(t, 1, counter, "计数器应增加1")
    
    // 第二次执行
    result = repeater.Tick(ctx)
    assert.Equal(t, Running, result, "未达到重复次数应返回运行中")
    assert.Equal(t, 2, counter, "计数器应增加到2")
    
    // 第三次执行（应完成）
    result = repeater.Tick(ctx)
    assert.Equal(t, Success, result, "达到重复次数应返回成功")
    assert.Equal(t, 3, counter, "计数器应增加到3")
    
    // 测试包含运行中子节点的重复器
    runningAction := NewAction(func(_ Context) Result { return Running })
    repeater = NewRepeater(2, runningAction)
    result = repeater.Tick(ctx)
    assert.Equal(t, Running, result, "子节点返回运行中时，重复器应返回运行中")
    
    // 测试动态设置子节点
    repeater = NewRepeater(1, nil)
    repeater.SetChild(action)
    result = repeater.Tick(ctx)
    assert.Equal(t, Success, result, "动态设置子节点后应正确执行")
    assert.Equal(t, 4, counter, "计数器应增加到4")
    
    // 测试空子节点
    repeater = NewRepeater(1, nil)
    result = repeater.Tick(ctx)
    assert.Equal(t, Failure, result, "空子节点应返回失败")

    // 测试 Reset 方法
    counter = 0
    repeater = NewRepeater(3, action)

    // 执行2次
    repeater.Tick(ctx)
    repeater.Tick(ctx)
    assert.Equal(t, 2, counter, "执行2次后计数器应为2")

    // 重置后重新执行
    repeater.Reset()
    counter = 0

    // 重置后应该从头开始计数
    repeater.Tick(ctx)
    assert.Equal(t, 1, counter, "重置后执行1次计数器应为1")
}

func TestUntilSuccess(t *testing.T) {
    ctx := make(Context)
    
    // 测试直到成功装饰器（子节点最终会成功）
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        if counter >= 3 {
            return Success
        }
        return Failure
    })
    
    untilSuccess := NewUntilSuccess(action)
    
    // 第一次执行（失败）
    result := untilSuccess.Tick(ctx)
    assert.Equal(t, Running, result, "子节点失败时应返回运行中")
    assert.Equal(t, 1, counter, "计数器应增加1")
    
    // 第二次执行（失败）
    result = untilSuccess.Tick(ctx)
    assert.Equal(t, Running, result, "子节点失败时应返回运行中")
    assert.Equal(t, 2, counter, "计数器应增加到2")
    
    // 第三次执行（成功）
    result = untilSuccess.Tick(ctx)
    assert.Equal(t, Success, result, "子节点成功时应返回成功")
    assert.Equal(t, 3, counter, "计数器应增加到3")
    
    // 测试子节点立即成功的情况
    successAction := NewAction(func(_ Context) Result { return Success })
    untilSuccess = NewUntilSuccess(successAction)
    result = untilSuccess.Tick(ctx)
    assert.Equal(t, Success, result, "子节点立即成功时应返回成功")
    
    // 测试子节点返回运行中的情况
    runningAction := NewAction(func(_ Context) Result { return Running })
    untilSuccess = NewUntilSuccess(runningAction)
    result = untilSuccess.Tick(ctx)
    assert.Equal(t, Running, result, "子节点返回运行中时应返回运行中")
    
    // 测试动态设置子节点
    untilSuccess = NewUntilSuccess(nil)
    untilSuccess.SetChild(successAction)
    result = untilSuccess.Tick(ctx)
    assert.Equal(t, Success, result, "动态设置子节点后应正确执行")
    
    // 测试空子节点
    untilSuccess = NewUntilSuccess(nil)
    result = untilSuccess.Tick(ctx)
    assert.Equal(t, Failure, result, "空子节点应返回失败")
}

func TestUntilFailure(t *testing.T) {
    ctx := make(Context)
    
    // 测试直到失败装饰器（子节点最终会失败）
    counter := 0
    action := NewAction(func(_ Context) Result {
        counter++
        if counter >= 3 {
            return Failure
        }
        return Success
    })
    
    untilFailure := NewUntilFailure(action)
    
    // 第一次执行（成功）
    result := untilFailure.Tick(ctx)
    assert.Equal(t, Running, result, "子节点成功时应返回运行中")
    assert.Equal(t, 1, counter, "计数器应增加1")
    
    // 第二次执行（成功）
    result = untilFailure.Tick(ctx)
    assert.Equal(t, Running, result, "子节点成功时应返回运行中")
    assert.Equal(t, 2, counter, "计数器应增加到2")
    
    // 第三次执行（失败）
    result = untilFailure.Tick(ctx)
    assert.Equal(t, Success, result, "子节点失败时应返回成功")
    assert.Equal(t, 3, counter, "计数器应增加到3")
    
    // 测试子节点立即失败的情况
    failureAction := NewAction(func(_ Context) Result { return Failure })
    untilFailure = NewUntilFailure(failureAction)
    result = untilFailure.Tick(ctx)
    assert.Equal(t, Success, result, "子节点立即失败时应返回成功")
    
    // 测试子节点返回运行中的情况
    runningAction := NewAction(func(_ Context) Result { return Running })
    untilFailure = NewUntilFailure(runningAction)
    result = untilFailure.Tick(ctx)
    assert.Equal(t, Running, result, "子节点返回运行中时应返回运行中")
    
    // 测试动态设置子节点
    untilFailure = NewUntilFailure(nil)
    untilFailure.SetChild(failureAction)
    result = untilFailure.Tick(ctx)
    assert.Equal(t, Success, result, "动态设置子节点后应正确执行")
    
    // 测试空子节点
    untilFailure = NewUntilFailure(nil)
    result = untilFailure.Tick(ctx)
    assert.Equal(t, Failure, result, "空子节点应返回失败")
}