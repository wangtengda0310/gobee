# GoBee Behavior Tree

GoBee Behavior Tree是一个用Go语言实现的轻量级行为树库，适用于游戏AI、机器人控制、决策系统等需要复杂决策逻辑的应用场景。

## 功能特性

- **完整的行为树节点类型**：
  - 复合节点：Sequence（序列）、Selector（选择器）、Parallel（并行）
  - 条件节点：Condition
  - 动作节点：Action
  - 装饰器节点：Inverter（反转器）、Repeater（重复器）、UntilSuccess（直到成功）、UntilFailure（直到失败）

- **易于扩展**：简单的接口设计，方便扩展自定义节点类型

- **高性能**：轻量级实现，无外部依赖

- **测试覆盖**：完善的单元测试，代码覆盖率达到80.4%

## 快速开始

### 安装

```bash
go get github.com/wangtengda/gobee/behavior/v2
```

### 基本用法

#### 创建一个简单的行为树

```go
package main

import (
    "fmt"
    "github.com/wangtengda/gobee/behavior/v2"
)

func main() {
    // 创建上下文对象，用于在行为树节点间传递数据
    ctx := make(behavior.Context)
    ctx["hungry"] = true
    ctx["hasFood"] = true
    
    // 创建条件节点
    isHungry := behavior.NewCondition(func(ctx behavior.Context) bool {
        return ctx["hungry"].(bool)
    })
    
    hasFood := behavior.NewCondition(func(ctx behavior.Context) bool {
        return ctx["hasFood"].(bool)
    })
    
    // 创建动作节点
    eatFood := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Eating food...")
        ctx["hungry"] = false
        ctx["hasFood"] = false
        return behavior.Success
    })
    
    findFood := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Looking for food...")
        ctx["hasFood"] = true
        return behavior.Success
    })
    
    // 创建复合节点构建行为树
    // 如果饥饿且有食物，则吃食物；否则寻找食物
    behaviorTree := behavior.NewSequence(
        isHungry,
        behavior.NewSelector(
            behavior.NewSequence(hasFood, eatFood),
            findFood,
        ),
    )
    
    // 执行行为树
    result := behaviorTree.Tick(ctx)
    fmt.Printf("Behavior tree result: %s\n", result)
    fmt.Printf("Final state - Hungry: %v, HasFood: %v\n", ctx["hungry"], ctx["hasFood"])
}
```

## 节点类型详解

### 结果类型（Result）

行为树中的每个节点执行后会返回以下三种结果之一：
- `Success`：节点执行成功
- `Failure`：节点执行失败
- `Running`：节点正在执行中

### 复合节点

#### Sequence（序列节点）

按顺序执行子节点，直到所有子节点都返回`Success`，或者某个子节点返回`Failure`。

```go
sequence := behavior.NewSequence(node1, node2, node3)
```

#### Selector（选择器节点）

按顺序执行子节点，直到某个子节点返回`Success`或`Running`，或者所有子节点都返回`Failure`。

```go
selector := behavior.NewSelector(node1, node2, node3)
```

#### Parallel（并行节点）

并行执行所有子节点，根据成功和失败的阈值来决定最终结果。

```go
// 需要至少2个成功，1个失败就返回失败
parallel := behavior.NewParallel(2, 1, node1, node2, node3)
```

### 条件节点

条件节点根据给定的条件函数返回`Success`或`Failure`。

```go
condition := behavior.NewCondition(func(ctx behavior.Context) bool {
    // 条件判断逻辑
    return true
})
```

### 动作节点

动作节点执行具体的行为，并返回执行结果。

```go
action := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
    // 动作执行逻辑
    return behavior.Success
})
```

### 装饰器节点

#### Inverter（反转器）

反转子节点的结果（`Success`变为`Failure`，反之亦然）。

```go
inverter := behavior.NewInverter(childNode)
```

#### Repeater（重复器）

重复执行子节点指定次数或无限重复。

```go
// 重复执行3次
repeater := behavior.NewRepeater(3, childNode)

// 无限重复
infiniteRepeater := behavior.NewRepeater(-1, childNode)
```

#### UntilSuccess（直到成功）

重复执行子节点直到返回`Success`。

```go
untilSuccess := behavior.NewUntilSuccess(childNode)
```

#### UntilFailure（直到失败）

重复执行子节点直到返回`Failure`。

```go
untilFailure := behavior.NewUntilFailure(childNode)
```

## 上下文（Context）

上下文是一个简单的map，用于在行为树的不同节点间传递数据。

```go
ctx := make(behavior.Context)
ctx["key"] = value
```

## 高级示例

查看`example.go`和`cmd/behavior/main.go`文件获取更复杂的使用示例。

## 测试

运行单元测试：

```bash
cd path/to/gobee/behavior
go test -v -cover .
```

## 代码质量检查

使用golangci-lint检查代码质量：

```bash
golangci-lint run
```

## 许可证

本项目采用MIT许可证。

## 贡献指南

欢迎提交Issue和Pull Request来改进这个库。