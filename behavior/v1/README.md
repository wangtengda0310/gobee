# GoBee Behavior Tree v1

一个用Go语言编写的纯函数式行为树库，专为游戏AI和决策系统设计。

## 特性

- 纯函数式设计，无状态，易于测试和并行执行
- 支持常见的行为树节点类型：动作、条件、序列、选择器、并行和装饰器
- 简单直观的API设计
- 完整的单元测试和模糊测试
- 高性能，无内存泄漏

## 安装

```bash
go get github.com/wangtengda/gobee/behavior/v1
```

## 核心概念

### 结果类型 (Result)

行为树中的每个节点执行后都会返回以下结果之一：

- `Success`: 节点执行成功
- `Failure`: 节点执行失败
- `Running`: 节点正在执行中

### 上下文 (Context)

上下文是一个键值对映射 (`map[string]interface{}`)，用于在节点之间共享数据。

### 节点 (Node)

节点是行为树的基本构建块，在v1版本中，节点被定义为一个函数类型：

```go
type Node func(ctx Context) Result
```

## 节点类型

### 动作节点 (Action)

动作节点执行具体的行为逻辑。

```go
// 创建一个成功的动作节点
successAction := Action(func(ctx Context) Result {
    // 执行具体操作
    return Success
})

// 创建一个会修改上下文的动作节点
counterAction := Action(func(ctx Context) Result {
    count, exists := ctx["count"].(int)
    if !exists {
        count = 0
    }
    ctx["count"] = count + 1
    return Success
})
```

### 条件节点 (Condition)

条件节点评估一个布尔表达式并返回成功或失败。

```go
// 创建一个始终为真的条件节点
trueCondition := Condition(func(ctx Context) bool {
    return true
})

// 创建一个基于上下文的条件节点
countCheck := Condition(func(ctx Context) bool {
    count, exists := ctx["count"].(int)
    return exists && count > 10
})
```

### 复合节点 (Composite)

#### 序列节点 (Sequence)

序列节点按顺序执行子节点，只有当所有子节点都成功时才返回成功。如果任何子节点失败或运行中，则序列节点返回相应的结果。

```go
// 创建一个序列节点
sequence := Sequence(
    checkCondition,
    performAction1,
    performAction2,
)
```

#### 选择器节点 (Selector)

选择器节点按顺序执行子节点，只要有一个子节点成功就返回成功。如果所有子节点都失败，则选择器节点返回失败。

```go
// 创建一个选择器节点
selector := Selector(
    tryAction1,
    tryAction2,
    fallbackAction,
)
```

#### 并行节点 (Parallel)

并行节点同时执行所有子节点，并根据成功和失败的策略返回结果。

```go
// 创建一个并行节点，需要2个成功才能成功，1个失败就失败
parallel := Parallel(2, 1, 
    action1,
    action2,
    action3,
)
```

### 装饰器节点 (Decorator)

#### 反转节点 (Inverter)

反转子节点的结果（成功变失败，失败变成功，但运行中保持不变）。

```go
// 创建一个反转节点
inverter := Inverter(action)
```

#### 重复节点 (Repeater)

重复执行子节点指定的次数。

```go
// 创建一个重复执行3次的节点
repeater := Repeater(3, action)
```

#### 直到成功节点 (UntilSuccess)

重复执行子节点直到子节点返回成功。

```go
// 创建一个直到成功节点
untilSuccess := UntilSuccess(action)
```

#### 直到失败节点 (UntilFailure)

重复执行子节点直到子节点返回失败。

```go
// 创建一个直到失败节点
untilFailure := UntilFailure(action)
```

## 使用示例

### 基本用法

以下是使用行为树库的基本示例：

1. 首先创建一个新的Go项目：

```bash
mkdir my_behavior_tree_app
cd my_behavior_tree_app
go mod init my_behavior_tree_app
```

2. 安装行为树库：

```bash
go get github.com/wangtengda/gobee/behavior/v1
```

3. 创建main.go文件：

```go
package main

import (
    "fmt"
    "github.com/wangtengda/gobee/behavior/v1"
)

func main() {
    // 创建上下文
    ctx := make(behavior.Context)
    ctx["targetFound"] = true
    ctx["health"] = 80
    
    // 创建节点
    checkTarget := behavior.Condition(func(ctx behavior.Context) bool {
        return ctx["targetFound"].(bool)
    })
    
    checkHealth := behavior.Condition(func(ctx behavior.Context) bool {
        return ctx["health"].(int) > 50
    })
    
    attackAction := behavior.Action(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Attacking target!")
        return behavior.Success
    })
    
    fleeAction := behavior.Action(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Fleeing from danger!")
        return behavior.Success
    })
    
    // 构建行为树
    behaviorTree := behavior.Sequence(
        checkTarget,
        behavior.Selector(
            behavior.Sequence(
                checkHealth,
                attackAction,
            ),
            fleeAction,
        ),
    )
    
    // 执行行为树
    result := behaviorTree(ctx)
    fmt.Printf("Behavior tree result: %s\n", result)
}
```

4. 运行程序：

```bash
go run main.go
```

### 运行演示示例

库中包含了一个简单的演示函数，可以通过以下方式使用：

```go
package main

import (
    "github.com/wangtengda/gobee/behavior/v1"
)

func main() {
    // 运行内置的行为树演示
    behavior.DemoBehaviorTree()
}
```

## 运行测试

### 单元测试

```bash
go test -v
```

### 覆盖率测试

```bash
go test -coverprofile=coverage.out
```

### 模糊测试

```bash
go test -fuzz=FuzzSimpleBehaviorTree -fuzztime=10s
go test -fuzz=FuzzContextOperations -fuzztime=10s
```

## 代码质量检查

```bash
golangci-lint run
```

## 版本历史

- **v1**: 纯函数式设计，使用函数闭包实现无状态的行为树
- **v2**: 基于对象的设计，提供更多高级功能

## 适用场景

- 游戏AI决策系统
- 机器人控制逻辑
- 工作流引擎
- 自动化测试框架
- 任何需要复杂条件决策的场景

## 性能优化建议

1. 对于频繁执行的行为树，考虑缓存不变的子树结构
2. 避免在节点函数中执行耗时操作，可以考虑使用异步模式
3. 对于复杂的行为树，考虑使用子树来组织和重用逻辑
```

## 运行测试

### 单元测试

```bash
go test -v
```

### 覆盖率测试

```bash
go test -coverprofile=coverage.out
```

### 模糊测试

```bash
go test -fuzz=FuzzSimpleBehaviorTree -fuzztime=10s
go test -fuzz=FuzzContextOperations -fuzztime=10s
```

## 代码质量检查

```bash
golangci-lint run
```

## 版本历史

- **v1**: 纯函数式设计，使用函数闭包实现无状态的行为树
- **v2**: 基于对象的设计，提供更多高级功能

## 适用场景

- 游戏AI决策系统
- 机器人控制逻辑
- 工作流引擎
- 自动化测试框架
- 任何需要复杂条件决策的场景

## 性能优化建议

1. 对于频繁执行的行为树，考虑缓存不变的子树结构
2. 避免在节点函数中执行耗时操作，可以考虑使用异步模式
3. 对于复杂的行为树，考虑使用子树来组织和重用逻辑