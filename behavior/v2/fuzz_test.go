//go:build gofuzz
// +build gofuzz

package behavior

import (
    "math/rand"
    "time"
)

// FuzzBehaviorTree 模糊测试行为树的各种操作
func FuzzBehaviorTree(f *testing.F) {
    // 初始化随机种子
    rand.Seed(time.Now().UnixNano())
    
    // 添加一些初始测试用例
    f.Add([]byte{0, 1, 2, 3, 4, 5})
    f.Add([]byte{10, 20, 30, 40})
    f.Add([]byte{1, 1, 1, 1})
    
    f.Fuzz(func(t *testing.T, data []byte) {
        if len(data) < 2 { // 确保有足够的数据进行测试
            return
        }
        
        // 使用模糊数据来决定测试的行为树结构和节点类型
        testType := int(data[0]) % 4 // 0: 测试基本节点, 1: 测试复合节点, 2: 测试装饰器节点, 3: 测试混合结构
        ctx := make(Context)
        
        // 为上下文添加一些随机数据
        ctx["value1"] = int(data[1])
        ctx["value2"] = float64(data[1]) / 10.0
        ctx["flag"] = data[1]%2 == 0
        
        switch testType {
        case 0:
            // 测试基本节点（条件和动作）
            testBasicNodes(t, ctx, data)
        case 1:
            // 测试复合节点（Sequence, Selector, Parallel）
            testCompositeNodes(t, ctx, data)
        case 2:
            // 测试装饰器节点
            testDecoratorNodes(t, ctx, data)
        case 3:
            // 测试混合结构
            testMixedStructure(t, ctx, data)
        }
    })
}

// 测试基本节点
func testBasicNodes(t *testing.T, ctx Context, data []byte) {
    // 创建条件节点
    condition := NewCondition(func(ctx Context) bool {
        return ctx["flag"].(bool)
    })
    
    // 执行条件节点
    _ = condition.Tick(ctx)
    
    // 创建动作节点
    action := NewAction(func(ctx Context) Result {
        val, ok := ctx["value1"].(int)
        if !ok {
            return Failure
        }
        
        if val%2 == 0 {
            return Success
        } else if val%3 == 0 {
            return Running
        }
        return Failure
    })
    
    // 执行动作节点
    _ = action.Tick(ctx)
}

// 测试复合节点
func testCompositeNodes(t *testing.T, ctx Context, data []byte) {
    // 创建一些基本节点用于测试复合节点
    successAction := NewAction(func(_ Context) Result { return Success })
    failureAction := NewAction(func(_ Context) Result { return Failure })
    runningAction := NewAction(func(_ Context) Result { return Running })
    
    // 根据模糊数据决定复合节点类型
    nodeType := int(data[0]) % 3
    var compositeNode CompositeNode
    
    switch nodeType {
    case 0:
        // Sequence节点
        compositeNode = NewSequence()
    case 1:
        // Selector节点
        compositeNode = NewSelector()
    case 2:
        // Parallel节点
        successThreshold := 2
        failureThreshold := 1
        if len(data) >= 3 {
            successThreshold = int(data[1])%3 + 1
            failureThreshold = int(data[2])%3 + 1
        }
        compositeNode = NewParallel(successThreshold, failureThreshold)
    }
    
    // 添加随机数量的子节点
    numChildren := int(data[1])%5 + 1
    for i := 0; i < numChildren; i++ {
        // 根据数据选择子节点类型
        childType := int(data[(i+2)%len(data)]) % 3
        var child Node
        
        switch childType {
        case 0:
            child = successAction
        case 1:
            child = failureAction
        case 2:
            child = runningAction
        }
        
        compositeNode.AddChild(child)
    }
    
    // 执行复合节点
    _ = compositeNode.Tick(ctx)
}

// 测试装饰器节点
func testDecoratorNodes(t *testing.T, ctx Context, data []byte) {
    // 创建一些基本节点用于测试装饰器
    action := NewAction(func(ctx Context) Result {
        val := ctx["value1"].(int)
        if val%2 == 0 {
            return Success
        } else if val%3 == 0 {
            return Running
        }
        return Failure
    })
    
    // 根据模糊数据决定装饰器类型
    decoratorType := int(data[0]) % 4
    var decorator Node
    
    switch decoratorType {
    case 0:
        // Inverter
        decorator = NewInverter(action)
    case 1:
        // Repeater
        times := int(data[1]) % 10
        if times == 0 {
            times = -1 // 无限重复
        }
        decorator = NewRepeater(times, action)
    case 2:
        // UntilSuccess
        decorator = NewUntilSuccess(action)
    case 3:
        // UntilFailure
        decorator = NewUntilFailure(action)
    }
    
    // 执行装饰器节点
    _ = decorator.Tick(ctx)
}

// 测试混合结构
func testMixedStructure(t *testing.T, ctx Context, data []byte) {
    // 创建一个混合了各种节点类型的复杂行为树
    selector := NewSelector()
    
    // 添加一个序列节点作为子节点
    sequence := NewSequence()
    
    // 创建一些条件和动作节点
    condition1 := NewCondition(func(ctx Context) bool {
        return ctx["flag"].(bool)
    })
    
    action1 := NewAction(func(_ Context) Result {
        return Success
    })
    
    // 添加一个装饰器节点
    inverter := NewInverter(action1)
    
    // 构建行为树结构
    sequence.AddChild(condition1)
    sequence.AddChild(inverter)
    
    // 添加另一个动作作为选择器的第二个选项
    action2 := NewAction(func(_ Context) Result {
        return Success
    })
    
    selector.AddChild(sequence)
    selector.AddChild(action2)
    
    // 执行混合结构
    _ = selector.Tick(ctx)
}