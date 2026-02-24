# gameactor 架构设计文档

## 概述

gameactor 是一个基于哈希路由的 Go 线程分发系统，专为游戏服务器设计。

### 核心问题

游戏服务器中常见的问题：
1. **玩家数据一致性**：同一玩家的操作必须按顺序执行
2. **并发性能**：不同玩家的操作应该并行执行
3. **简单易用**：API 应该简单直观，减少心智负担

### 解决方案

- **哈希路由**：相同 `playerID` 的任务路由到同一个 Actor
- **Actor 模型**：每个 Actor 是一个独立的 goroutine，串行处理任务
- **CSP 模型**：使用 channel 进行任务分发，避免锁竞争

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         API Layer                           │
│  DispatchBy | DispatchBySync | Dispatch | DispatchWith...  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Dispatcher (全局单例)                      │
│  - actors[]: []*actor                                        │
│  - route(hash) -> actorID                                   │
│  - Submit(hash, task) -> error                              │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┼─────────────┐
                ▼             ▼             ▼
          ┌─────────┐   ┌─────────┐   ┌─────────┐
          │ Actor 0 │   │ Actor 1 │   │ Actor N │
          │ Queue   │   │ Queue   │   │ Queue   │
          │ run()   │   │ run()   │   │ run()   │
          └─────────┘   └─────────┘   └─────────┘
```

## 核心组件

### 1. Dispatcher

**职责**：
- 管理 Actor 池（固定大小）
- 根据哈希值路由任务
- 提供优雅关闭机制

**关键方法**：
```go
type Dispatcher struct {
    actors      []*actor
    numActors   uint64
    stopChan    chan struct{}
    waitGroup   sync.WaitGroup
}

// 路由算法
func (d *Dispatcher) route(hash uint64) uint64 {
    return hash % d.numActors
}

// 提交任务
func (d *Dispatcher) Submit(hash uint64, task Task) error {
    actorID := d.route(hash)
    actor := d.actors[actorID]
    select {
    case actor.queue <- task:
        return nil
    default:
        return ErrQueueFull
    }
}
```

### 2. Actor

**职责**：
- 从队列中取出任务
- 串行执行任务
- 捕获 panic

**关键方法**：
```go
type actor struct {
    id      uint64
    queue   chan Task
    running atomic.Bool
}

func (a *actor) run() {
    for {
        select {
        case task, ok := <-a.queue:
            if !ok { return }
            a.executeTask(task)
        case <-a.dispatcher.stopChan:
            return
        }
    }
}

func (a *actor) executeTask(task Task) {
    defer func() {
        if r := recover(); r != nil {
            // panic 处理
        }
    }()
    task.handler()
}
```

### 3. Task

**职责**：
- 封装任务处理逻辑
- 提供哈希值（用于路由）

**设计**：
```go
type Task struct {
    handler  func() error
    hash     uint64            // 直接指定的哈希（优先级最高）
    hashFunc func(Task) uint64 // 哈希计算函数（次优先级）
}

// 实现 Hashable 接口
func (t Task) Hash() uint64 {
    if t.hash != 0 {
        return t.hash
    }
    if t.hashFunc != nil {
        return t.hashFunc(t)
    }
    return 0
}
```

## API 设计

### 设计原则

1. **简单优先**：推荐使用 `DispatchBy`，最简单
2. **渐进式复杂度**：从简单到复杂：DispatchBy -> Dispatch -> DispatchWithFunc
3. **一致性**：所有 API 都有 Sync 和 Ctx 版本
4. **闭包友好**：支持闭包捕获参数

### API 层次

```
Level 1: DispatchBy(hash, handler)           ⭐ 推荐
Level 2: DispatchBySync(hash, handler)       ⭐ 推荐
Level 3: Dispatch(task Hashable)
Level 4: DispatchWithFunc(fn, task)
Level 5: DispatchWithHash(hash, task)
```

### 命名规范

| 前缀/后缀 | 含义 | 示例 |
|----------|------|------|
| `DispatchBy` | 便利函数，直接传哈希 | `DispatchBy(1001, func(){})` |
| `Sync` | 同步等待结果 | `DispatchBySync(1001, func() error)` |
| `Ctx` | 支持 Context | `DispatchByCtx(ctx, 1001, func(){})` |
| `WithFunc` | 使用哈希函数 | `DispatchWithFunc(fn, task)` |
| `WithHash` | 直接指定哈希 | `DispatchWithHash(1001, task)` |

## 路由算法

### 取模路由

```go
actorID = hash % numActors
```

**特点**：
- 简单高效
- 相同 hash 总是路由到同一个 Actor
- 哈希碰撞概率可接受

**示例**：
```
hash = 1001, numActors = 1000
actorID = 1001 % 1000 = 1

hash = 2001, numActors = 1000
actorID = 2001 % 1000 = 1
```

### 未来优化

- **一致性哈希**：支持动态增减 Actor
- **虚拟节点**：更均匀的负载分布
- **WorkStealing**：负载均衡

## 并发模型

### CSP vs 锁

| 特性 | CSP (Channel) | 锁 (Mutex) |
|------|---------------|-------------|
| 复杂度 | 低 | 高 |
| 调试 | 易 | 难 |
| 性能 | 好 | 更好 |
| 推荐 | ✅ 游戏逻辑 | 高频竞争场景 |

### 为什么选择 CSP？

1. **简化实现**：channel 天然支持队列语义
2. **易于理解**：goroutine + channel 是 Go 的核心哲学
3. **减少竞争**：每个 Actor 独立队列
4. **预留优化空间**：未来可以替换为 WorkStealing

## 错误处理

### Panic 恢复

```go
func (a *actor) executeTask(task Task) {
    defer func() {
        if r := recover(); r != nil {
            // 调用 panicHandler
            if a.dispatcher.panicHandler != nil {
                a.dispatcher.panicHandler(r)
            }
        }
    }()
    task.handler()
}
```

### 错误返回

```go
// 同步版本返回错误
err := DispatchBySync(1001, func() error {
    return errors.New("处理失败")
})

// 异步版本需要自己处理
DispatchBy(1001, func() {
    if err := doSomething(); err != nil {
        log.Printf("错误: %v", err)
    }
})
```

## 关闭流程

### 优雅关闭

```
1. Stop() - 关闭所有队列
2. 执行关闭钩子
3. Wait() - 等待所有任务完成或超时
```

```go
func Shutdown(timeout time.Duration) error {
    // 1. 停止接受新任务
    globalDispatcher.Stop()

    // 2. 执行关闭钩子
    for _, hook := range shutdownHooks {
        hook()
    }

    // 3. 等待所有任务完成或超时
    done := make(chan struct{})
    go func() {
        globalDispatcher.Wait()
        close(done)
    }()

    select {
    case <-done:
        // 正常完成
    case <-time.After(timeout):
        // 超时强制关闭
    }
}
```

## 性能分析

### 基准测试

```
BenchmarkDispatchBy-8     5000000    250 ns/op
BenchmarkDispatchBySync-8 1000000    950 ns/op
```

### 性能瓶颈

1. **Channel 操作**：每次提交都需要 channel 操作
2. **取模运算**：虽然很快，但仍有开销
3. **队列竞争**：多个 goroutine 提交到同一 Actor 时有竞争

### 优化建议

1. **批量提交**：减少 channel 操作次数
2. **预计算哈希**：避免重复计算
3. **调整队列大小**：根据负载调整

## 测试策略

### TDD 驱动

```
1. 编写测试用例
2. 测试失败（预期）
3. 实现功能
4. 测试通过
```

### 隔离测试

使用 `TestDispatcher` 避免测试间干扰：

```go
td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
defer td.Shutdown(5 * time.Second)

td.DispatchBy(1001, func() {
    // 测试逻辑
})
```

### 测试覆盖

- 基本功能：DispatchBy、DispatchBySync
- 并发安全：多 goroutine 并发提交
- 顺序保证：相同 hash 串行执行
- 并行执行：不同 hash 并行执行
- 错误处理：panic、错误返回
- 关闭流程：优雅关闭、超时控制

## 未来规划

### Phase 2: 可观测性

- [ ] Prometheus 指标
- [ ] 任务执行追踪
- [ ] 性能监控

### Phase 3: 高级特性

- [ ] Hybrid Actor Pool（固定+动态）
- [ ] WorkStealing 队列
- [ ] 一致性哈希路由

### Phase 4: 工具链

- [ ] Agent Skill（AI 辅助编程）
- [ ] 性能分析工具
- [ ] 可视化监控

## 参考资料

- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example: Channels](https://gobyexample.com/channels)
- [Actor Model](https://en.wikipedia.org/wiki/Actor_model)
