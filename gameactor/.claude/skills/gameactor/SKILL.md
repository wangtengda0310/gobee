---
name: gameactor 线程分发系统
description: 基于哈希路由的 Go 线程分发系统，专为游戏服务器设计
---

# gameactor 线程分发系统

## 项目定位

基于哈希路由的 Go 线程分发系统，专为游戏服务器设计。

**Go Module**: `github.com/wangtengda0310/gobee/gameactor`

**核心价值**: 相同玩家操作串行执行，不同玩家操作并行执行。

---

## 快速开始

### 最简用法

```go
import "github.com/wangtengda0310/gobee/gameactor"

func main() {
    // 初始化
    gameactor.Init(gameactor.DefaultConfig())
    defer gameactor.Shutdown(30 * time.Second)

    // 提交任务
    playerID := uint64(1001)
    gameactor.DispatchBy(playerID, func() {
        // 处理玩家逻辑 - 相同 playerID 的任务按顺序执行
        processPlayer(playerID)
    })
}
```

---

## 核心概念

### 哈希路由

| 概念 | 说明 |
|------|------|
| **Actor** | 固定数量的 goroutine（默认 1000 个） |
| **路由算法** | `actorID = hash % numActors` |
| **串行保证** | 相同 hash 的任务在同一 Actor 中执行 |
| **并行执行** | 不同 hash 的任务在不同 Actor 中执行 |

### API 速查

| 函数 | 类型 | 说明 |
|------|------|------|
| `DispatchBy(hash, handler)` | 异步 | 最简单的 API ⭐ |
| `DispatchBySync(hash, handler)` | 同步 | 等待执行完成 |
| `DispatchByCtx(ctx, hash, handler)` | 异步+Context | 支持超时控制 |
| `Dispatch(task)` | 异步 | Hashable 接口 |
| `DispatchWithHash(hash, task)` | 异步 | 直接指定哈希 |

### 闭包捕获

```go
// 推荐：闭包捕获参数
playerID := uint64(1001)
damage := 100

gameactor.DispatchBy(playerID, func() {
    attackMonster(playerID, damage)  // 闭包自动捕获
})
```

---

## 使用场景

### 场景 1: 玩家战斗系统

```go
// 确保同一玩家的攻击按顺序执行
func attackMonster(playerID, monsterID uint64, damage int) error {
    return gameactor.DispatchBySync(playerID, func() error {
        player := getPlayer(playerID)
        monster := getMonster(monsterID)

        player.Attack(monster, damage)
        return savePlayer(player)
    })
}
```

### 场景 2: 聊天系统

```go
// 玩家聊天消息按顺序处理
func sendChat(playerID uint64, message string) {
    gameactor.DispatchBy(playerID, func() {
        player := getPlayer(playerID)
        player.SendMessage(message)
        saveChatHistory(playerID, message)
    })
}
```

### 场景 3: 库存管理

```go
// 避免并发导致的物品数量错误
func moveItem(playerID, itemID, fromSlot, toSlot int) error {
    return gameactor.DispatchBySync(uint64(playerID), func() error {
        inventory := getInventory(playerID)
        return inventory.Move(itemID, fromSlot, toSlot)
    })
}
```

### 场景 4: 任务系统

```go
// 任务状态变更按顺序执行
func acceptQuest(playerID, questID uint64) error {
    return gameactor.DispatchBySync(playerID, func() error {
        player := getPlayer(playerID)
        quest := getQuest(questID)

        if err := player.AcceptQuest(quest); err != nil {
            return err
        }
        return savePlayer(player)
    })
}
```

### 场景 5: 测试用 TestDispatcher

```go
func TestPlayerLogic(t *testing.T) {
    // 使用 TestDispatcher 避免测试间干扰
    td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
    defer td.Shutdown(5 * time.Second)

    // 提交任务
    td.DispatchBy(1001, func() {
        // 测试逻辑
    })

    // 断言执行
    td.AssertExecuted(1001, 1)
}
```

---

## 项目结构

```
gameactor/
├── api.go              # 对外 API（Dispatch 函数族）
├── dispatcher.go       # Dispatcher 核心实现
├── testing.go          # TestDispatcher 测试辅助
├── isolation_test.go   # 隔离测试（9/9 通过）
├── api_test.go         # API 测试用例
├── examples/basic/     # 使用示例
├── README.md           # 用户文档
├── CLAUDE.md           # AI 开发指南
├── DESIGN.md           # 架构设计
└── TODO.md             # 开发进度
```

---

## 最佳实践

### DO 推荐

- ✅ 使用 `DispatchBy(hash, handler)` 作为主要 API
- ✅ 闭包捕获参数（简单、类型安全）
- ✅ 同一玩家使用相同 hash（保证顺序）
- ✅ 测试使用 `TestDispatcher`（隔离、断言）
- ✅ 同步操作使用 `DispatchBySync`（获取错误）

### DON'T 避免

- ❌ 不要在 handler 中执行长时间阻塞（会阻塞其他任务）
- ❌ 不要使用全局变量存储玩家数据（数据竞争）
- ❌ 不要在测试中使用全局 API（测试间会干扰）
- ❌ 不要忘记调用 `Shutdown()`（资源泄漏）
- ❌ 不要期望不同 hash 的任务有顺序保证

---

## 常见问题

### Q: 如何保证同一玩家的操作按顺序执行？

使用相同的 hash（通常是 playerID）：

```go
gameactor.DispatchBy(playerID, func() {
    // 操作 1
})

gameactor.DispatchBy(playerID, func() {
    // 操作 2 - 一定在操作 1 之后执行
})
```

### Q: 如何获取执行结果？

使用同步版本：

```go
result, err := gameactor.DispatchBySync(playerID, func() error {
    // 返回执行结果
    return doSomething()
})
```

### Q: 如何处理超时？

使用 Context 版本：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := gameactor.DispatchByCtx(ctx, playerID, func() {
    // 任务逻辑
})
```

### Q: 测试时任务没执行？

检查：
1. 是否调用了 `Init()` 初始化
2. 是否等待了足够时间（`time.Sleep`）
3. 推荐使用 `TestDispatcher` + `AssertExecuted`

### Q: 队列满了怎么办？

错误：`actor X queue is full`

解决方案：
1. 增加 `QueueSize` 配置
2. 减少任务提交速率
3. 使用 `DispatchBySync` 阻塞等待

---

## 运行测试

```bash
# 隔离测试
go test -v -run=TestTestDispatcher

# 运行示例
go run examples/basic/main.go

# 并发测试
go test -race ./...

# 基准测试
go test -bench=. -benchmem
```

---

## 配置说明

### 默认配置

```go
config := gameactor.DefaultConfig()
// NumActors: 1000
// QueueSize: 1000
// ShutdownTimeout: 30s
```

### 自定义配置

```go
config := gameactor.Config{
    NumActors:       500,   // Actor 数量
    QueueSize:       2000,  // 队列大小
    ShutdownTimeout: 60 * time.Second,
}
```

---

## 依赖

```go
require (
    // 无外部依赖
    // 仅使用 Go 标准库
)
```

---

## 相关文档

- 用户文档: `gameactor/README.md`
- AI 开发指南: `gameactor/CLAUDE.md`
- 架构设计: `gameactor/DESIGN.md`
- 开发进度: `gameactor/TODO.md`
- **AI 使用指导**: `abilities/AI指导.md` ← AI 阅读此文件以了解如何协助用户

---

## 性能参考

```
BenchmarkDispatchBy-8     5000000    250 ns/op
BenchmarkDispatchBySync-8 1000000    950 ns/op
```

---

## Go Module

```
github.com/wangtengda0310/gobee/gameactor
```
