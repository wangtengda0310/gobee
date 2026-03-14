# GoBee Behavior Tree

[![Go Reference](https://pkg.go.dev/badge/github.com/wangtengda/gobee/behavior.svg)](https://pkg.go.dev/github.com/wangtengda/gobee/behavior)
[![Go Report Card](https://goreportcard.com/badge/github.com/wangtengda/gobee/behavior)](https://goreportcard.com/report/github.com/wangtengda/gobee/behavior)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

一个用 Go 语言实现的轻量级行为树库，适用于游戏 AI、机器人控制、决策系统等需要复杂决策逻辑的应用场景。

## 特性

- **两种实现风格**
  - **v1**: 纯函数式设计，无状态，易于测试和并行执行
  - **v2**: 面向对象设计，支持状态管理和重置

- **完整的节点类型**
  - 复合节点: Sequence（序列）、Selector（选择器）、Parallel（并行）
  - 条件节点: Condition
  - 动作节点: Action
  - 装饰器节点: Inverter、Repeater、UntilSuccess、UntilFailure

- **高质量代码**
  - 完整的单元测试和模糊测试
  - v1 测试覆盖率 98%+，v2 测试覆盖率 100%
  - 无外部依赖

## 安装

### v1 (函数式)

```bash
go get github.com/wangtengda/gobee/behavior/v1
```

### v2 (面向对象)

```bash
go get github.com/wangtengda/gobee/behavior/v2
```

## 快速开始

### v1 示例

```go
package main

import (
    "fmt"
    "github.com/wangtengda/gobee/behavior/v1"
)

func main() {
    ctx := make(behavior.Context)
    ctx["health"] = 80

    checkHealth := behavior.Condition(func(ctx behavior.Context) bool {
        return ctx["health"].(int) > 50
    })

    attackAction := behavior.Action(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Attacking!")
        return behavior.Success
    })

    tree := behavior.Sequence(checkHealth, attackAction)
    result := tree(ctx)
    fmt.Printf("Result: %s\n", result)
}
```

### v2 示例

```go
package main

import (
    "fmt"
    "github.com/wangtengda/gobee/behavior/v2"
)

func main() {
    ctx := make(behavior.Context)
    ctx["health"] = 80

    checkHealth := behavior.NewCondition(func(ctx behavior.Context) bool {
        return ctx["health"].(int) > 50
    })

    attackAction := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Attacking!")
        return behavior.Success
    })

    tree := behavior.NewSequence(checkHealth, attackAction)
    result := tree.Tick(ctx)
    fmt.Printf("Result: %s\n", result)
}
```

## 文档

- [v1 文档 (函数式)](./v1/README.md)
- [v2 文档 (面向对象)](./v2/README.md)

## 运行测试

```bash
# 运行所有测试
make test

# 运行覆盖率测试
make cover

# 运行 lint
make lint

# 运行模糊测试
make fuzz
```

## 版本历史

查看 [CHANGELOG.md](./CHANGELOG.md) 了解版本变更历史。

## 适用场景

- 游戏 AI 决策系统
- 机器人控制逻辑
- 工作流引擎
- 自动化测试框架
- 任何需要复杂条件决策的场景

## 许可证

本项目采用 [MIT](./LICENSE) 许可证。

## 参考

- [行为树介绍](https://juejin.cn/post/7514549643138531347)
