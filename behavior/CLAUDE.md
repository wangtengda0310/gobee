# GoBee Behavior Tree - Claude Code 项目指南

本文档为 Claude Code 提供项目上下文，帮助 AI 助手更好地理解和维护此项目。

## 项目概述

GoBee Behavior Tree 是一个用 Go 语言实现的行为树库，提供两种实现风格：
- **v1**: 纯函数式设计，节点定义为 `func(Context) Result`
- **v2**: 面向对象设计，节点定义为 `Node` 接口

## 目录结构

```
behavior/
├── README.md              # 项目主文档
├── CLAUDE.md              # 本文件 - Claude Code 指南
├── LICENSE                # MIT 许可证
├── CHANGELOG.md           # 版本变更记录
├── Makefile               # 构建/测试命令
├── .gitignore             # Git 忽略配置
├── .golangci.yml          # 代码质量检查配置
├── .github/
│   └── workflows/
│       └── ci.yml         # GitHub Actions CI 配置
├── v1/                    # 函数式实现
│   ├── behavior.go        # 核心类型定义 (Result, Context, Node)
│   ├── action.go          # Action 节点
│   ├── condition.go       # Condition 节点
│   ├── composite.go       # 复合节点 (Sequence, Selector, Parallel)
│   ├── decorator.go       # 装饰器节点 (Inverter, Repeater, etc.)
│   ├── advanced.go        # 高级节点 (Retry, Timeout, Delay, Limiter)
│   ├── example.go         # 示例代码
│   └── *_test.go          # 测试文件
└── v2/                    # 面向对象实现
    ├── behavior.go        # Result 类型定义
    ├── node.go            # Node 接口定义
    ├── action.go          # Action 节点
    ├── condition.go       # Condition 节点
    ├── composite.go       # 复合节点
    ├── decorator.go       # 装饰器节点
    ├── advanced.go        # 高级节点
    ├── visualizer.go      # DOT 格式可视化导出
    └── *_test.go          # 测试文件
```

## 核心概念

### Result 类型

所有节点执行后返回三种结果之一：
```go
type Result int

const (
    Success  Result = iota  // 执行成功
    Failure                 // 执行失败
    Running                 // 正在执行中
)
```

### Context 上下文

用于在节点间传递数据：
```go
type Context map[string]interface{}
```

### Node 定义

**v1 (函数式):**
```go
type Node func(ctx Context) Result
```

**v2 (面向对象):**
```go
type Node interface {
    Tick(ctx Context) Result
}
```

## 节点类型

### 复合节点 (Composite)

| 节点 | 描述 | 行为 |
|------|------|------|
| Sequence | 序列节点 | 顺序执行，全部成功才成功 |
| Selector | 选择器节点 | 顺序执行，首个成功即成功 |
| Parallel | 并行节点 | 同时执行所有子节点 |
| RandomSelector | 随机选择器 | 随机顺序执行子节点 |

### 装饰器节点 (Decorator)

| 节点 | 描述 |
|------|------|
| Inverter | 反转子节点结果 |
| Repeater | 重复执行子节点 N 次 |
| UntilSuccess | 重复直到成功 |
| UntilFailure | 重复直到失败 |
| Retry | 失败时重试 |
| Timeout | 超时返回失败 |
| Delay | 延迟 N 次 tick 后执行 |
| Limiter | 限制成功执行次数 |

## 编码规范

### 命名约定

- **v1**: 构造函数使用大写开头，如 `Sequence()`, `Selector()`
- **v2**: 构造函数使用 `New` 前缀，如 `NewSequence()`, `NewSelector()`

### 注释规范

**重要：所有注释必须使用中文**

- 所有导出类型和函数必须有注释
- 注释以类型/函数名开头
- 使用完整句子

#### 公开方法注释 (针对使用者)

公开方法必须包含以下内容：
- **功能描述**: 简要说明方法的作用
- **参数说明**: 描述每个参数的含义和约束
- **返回值说明**: 描述返回值的含义
- **使用示例**: 复杂方法应提供简短的使用示例

示例：
```go
// NewRetry 创建一个重试装饰器节点。
// 当子节点失败时重试，最多重试 maxTries 次。
//
// 参数:
//   - maxTries: 最大重试次数 (-1 表示无限重试)
//   - child: 要执行的子节点
//
// 返回值:
//   - Success: 子节点执行成功
//   - Failure: 重试次数耗尽仍未成功
//   - Running: 正在重试中
func NewRetry(maxTries int, child Node) *Retry {
    // ...
}
```

#### 实现注释 (针对 Code Reviewer)

复杂逻辑、边界条件、设计决策需要添加注释说明：
- 解释"为什么"而不仅仅是"做什么"
- 说明边界条件的处理方式
- 标注潜在的性能考量
- 使用 `// 注意:` 或 `// NOTE:` 标注重要的设计决策

示例：
```go
func (r *Repeater) Tick(ctx Context) Result {
    // 注意: 完成时重置 tryCount 以支持行为树复用。
    // 这意味着同一个 Repeater 实例可以在不同的执行周期中多次 tick。

    if r.times == 0 {
        // 边界情况: 零次表示立即成功，不执行子节点。
        return Success
    }

    // ...
}
```

### 测试规范

- 测试文件与源文件同目录
- 使用 `github.com/stretchr/testify/assert` 断言
- 每个节点类型有对应的单元测试
- 包含模糊测试 (Fuzz Test)

## 常用命令

```bash
# 运行所有测试
make test

# 运行覆盖率测试
make cover

# 运行代码检查
make lint

# 运行模糊测试
make fuzz

# 清理构建产物
make clean
```

## API 示例

### v1 函数式用法

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

    attack := behavior.Action(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Attacking!")
        return behavior.Success
    })

    tree := behavior.Sequence(checkHealth, attack)
    result := tree(ctx)
    fmt.Println(result) // Success
}
```

### v2 面向对象用法

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

    attack := behavior.NewAction(func(ctx behavior.Context) behavior.Result {
        fmt.Println("Attacking!")
        return behavior.Success
    })

    tree := behavior.NewSequence(checkHealth, attack)
    result := tree.Tick(ctx)
    fmt.Println(result) // Success
}
```

## 可视化

v2 支持 DOT 格式导出，可用 Graphviz 渲染：

```go
dot, _ := behavior.ExportDOTString(tree)
fmt.Println(dot)
// 渲染: dot -Tpng tree.dot -o tree.png
```

## 版本兼容性

- Go 版本: 1.20+
- 依赖: 仅 `github.com/stretchr/testify` (测试)

## 修改代码时的注意事项

1. **保持两个版本的一致性**: 新功能应同时添加到 v1 和 v2
2. **添加测试**: 新功能必须有对应的单元测试
3. **更新文档**: 更新 README.md 和 CHANGELOG.md
4. **运行测试**: 提交前确保 `make test` 通过
5. **代码检查**: 提交前确保 `make lint` 通过

## 常见问题

### Q: v1 和 v2 应该用哪个？

- **v1**: 适合简单场景，函数式风格，无状态，易于测试
- **v2**: 适合复杂场景，面向对象，支持状态管理和 Reset

### Q: 如何处理空子节点？

- 空 Sequence 返回 Success
- 空 Selector 返回 Failure
- 空 Parallel 返回 Success

### Q: 如何重置节点状态？

v2 中支持 Reset 的节点：
- Repeater.Reset()
- Retry.Reset()
- Timeout.Reset()
- Delay.Reset()
- Limiter.Reset()
