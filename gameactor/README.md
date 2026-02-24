# gameactor

基于哈希路由的 Go 线程分发系统，专为游戏服务器设计。

## 核心特性

- **哈希路由**：相同哈希的任务按顺序在同一 goroutine 中执行
- **并行处理**：不同哈希的任务可以并行执行
- **简单易用**：推荐使用 `DispatchBy`，零心智负担
- **多种 API**：支持便利函数、Hashable 接口、哈希函数、直接哈希
- **优雅关闭**：支持超时控制和关闭钩子

## 快速开始

### 安装

#### 方式 1: 安装 Go 包

```bash
go get github.com/wangtengda0310/gobee/gameactor
```

#### 方式 2: 安装 AI Skill（推荐）

安装 AI Skill 后，可以在任何项目中获得智能辅助：

```bash
# 从远程安装
go install github.com/wangtengda0310/gobee/gameactor/cmd/install-skill@latest
gameactor-install-skill

# 或在 gameactor 仓库中
cd gameactor
go run cmd/install-skill/main.go
```

安装后，AI 会自动识别 gameactor 并提供以下能力：
- **代码生成**: "生成玩家战斗系统的代码"
- **问题诊断**: "诊断并发问题"
- **测试生成**: "写个单元测试"
- **性能分析**: "任务执行太慢"
- **重构建议**: "如何改造这段代码"

详细说明见 [AI 辅助开发](#ai-辅助开发) 章节。

### 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/wangtengda0310/gobee/gameactor"
)

func main() {
    // 初始化
    gameactor.Init(gameactor.DefaultConfig())
    defer gameactor.Shutdown(30 * time.Second)

    // 提交任务
    gameactor.DispatchBy(1001, func() {
        fmt.Println("处理玩家 1001 的任务")
    })

    // 等待任务完成
    time.Sleep(100 * time.Millisecond)
}
```

## API 文档

### 便利函数（推荐）⭐

#### DispatchBy - 异步执行

```go
gameactor.DispatchBy(hash uint64, handler func())
```

最简单的调用方式，handler 通过闭包捕获参数。

```go
playerID := uint64(1001)
damage := 100

gameactor.DispatchBy(playerID, func() {
    attackMonster(playerID, damage)
})
```

#### DispatchBySync - 同步执行

```go
err := gameactor.DispatchBySync(hash uint64, handler func() error)
```

阻塞等待任务完成，返回执行错误。

```go
err := gameactor.DispatchBySync(1001, func() error {
    return savePlayerData(1001)
})
if err != nil {
    log.Printf("保存失败: %v", err)
}
```

#### Context 支持

```go
// 异步 + Context
gameactor.DispatchByCtx(ctx, hash, func() {
    // 任务逻辑
})

// 同步 + Context
err := gameactor.DispatchBySyncCtx(ctx, hash, func() error {
    return doSomething()
})
```

### Hashable 接口

适用于复杂任务对象：

```go
type PlayerTask struct {
    PlayerID uint64
    Action   func() error
}

func (t PlayerTask) Hash() uint64 {
    return t.PlayerID
}

// 使用
gameactor.Dispatch(PlayerTask{
    PlayerID: 1001,
    Action:   func() error { return doSomething() },
})
```

### 直接哈希

使用 Task 对象：

```go
task := gameactor.NewTask(func() error {
    return doSomething()
})

gameactor.DispatchWithHash(1001, task)
```

### 哈希函数

使用自定义哈希计算函数：

```go
hashExtractor := func(t gameactor.Task) uint64 {
    // 自定义哈希逻辑
    return calculateHash(t)
}

task := gameactor.NewTask(func() error {
    return doSomething()
})

gameactor.DispatchWithFunc(hashExtractor, task)
```

## 配置

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
    NumActors:       500,
    QueueSize:       2000,
    ShutdownTimeout: 60 * time.Second,
}

gameactor.Init(config)
```

### 环境变量

```go
// 支持通过环境变量覆盖配置
// GAMEACTOR_NUM_ACTORS=500
// GAMEACTOR_QUEUE_SIZE=2000
// GAMEACTOR_SHUTDOWN_TIMEOUT=60s

config := gameactor.ConfigFromEnv()
gameactor.Init(config)
```

## 优雅关闭

### 基本关闭

```go
gameactor.Shutdown(30 * time.Second)
```

### 关闭钩子

```go
gameactor.RegisterShutdown(func() {
    fmt.Println("保存数据...")
    saveData()
})

gameactor.Shutdown(30 * time.Second)
```

### 信号监听

```go
// 自动监听 SIGINT 和 SIGTERM
gameactor.InitWithSignalHandler(gameactor.DefaultConfig())

// 程序会阻塞等待信号
```

## 测试

### 使用 TestDispatcher

```go
func TestPlayerLogic(t *testing.T) {
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

## 最佳实践

### 1. 使用闭包捕获参数

```go
// ✅ 推荐
playerID := 1001
damage := 100
gameactor.DispatchBy(playerID, func() {
    attackMonster(playerID, damage)
})

// ❌ 不推荐（需要额外结构体）
params := struct {
    PlayerID uint64
    Damage   int
}{PlayerID: 1001, Damage: 100}
gameactor.DispatchBy(params.PlayerID, func() {
    attackMonster(params.PlayerID, params.Damage)
})
```

### 2. 相同玩家使用相同哈希

```go
// 确保同一玩家的操作按顺序执行
gameactor.DispatchBy(playerID, func() {
    attack(playerID)
})

gameactor.DispatchBy(playerID, func() {
    defense(playerID)
})
```

### 3. 避免长时间阻塞

```go
// ❌ 不推荐：长时间阻塞
gameactor.DispatchBy(playerID, func() {
    time.Sleep(10 * time.Second) // 阻塞其他任务
})

// ✅ 推荐：使用 goroutine 或异步操作
gameactor.DispatchBy(playerID, func() {
    go func() {
        // 异步操作
    }()
})
```

### 4. 错误处理

```go
// 同步版本可以获取错误
err := gameactor.DispatchBySync(playerID, func() error {
    return savePlayer(playerID)
})
if err != nil {
    log.Printf("保存失败: %v", err)
}

// 异步版本需要自己处理错误
gameactor.DispatchBy(playerID, func() {
    if err := savePlayer(playerID); err != nil {
        log.Printf("保存失败: %v", err)
    }
})
```

## 性能

### 基准测试

```
BenchmarkDispatchBy-8     5000000    250 ns/op
BenchmarkDispatchBySync-8 1000000    950 ns/op
```

### 调优建议

1. **Actor 数量**：根据并发度调整，默认 1000
2. **队列大小**：根据任务积压情况调整，默认 1000
3. **监控指标**：使用 `GetMetrics()` 查看队列长度和执行时间

## 示例

### 玩家战斗系统

```go
func attackMonster(playerID, monsterID uint64, damage int) error {
    return gameactor.DispatchBySync(playerID, func() error {
        player := getPlayer(playerID)
        monster := getMonster(monsterID)

        player.Attack(monster, damage)
        return savePlayer(player)
    })
}
```

### 聊天系统

```go
func sendChat(playerID uint64, message string) {
    gameactor.DispatchBy(playerID, func() {
        player := getPlayer(playerID)
        player.SendMessage(message)
    })
}
```

## 故障排查

### 任务未执行

检查 Dispatcher 是否初始化：

```go
if !gameactor.IsRunning() {
    gameactor.Init(gameactor.DefaultConfig())
}
```

### 队列满

错误：`actor X queue is full`

解决方案：
1. 增加 `QueueSize`
2. 减少任务提交速率
3. 使用 `DispatchBySync` 阻塞等待

### 性能问题

使用指标查看：

```go
metrics := gameactor.GetMetrics()
for _, m := range metrics {
    fmt.Printf("Actor %d: 队列=%d, 已执行=%d\n",
        m.ActorID, m.QueueLength, m.TasksExecuted)
}
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

---

## AI 辅助开发

gameactor 提供了 AI Skill 支持，使用 Claude Code 时可以获得智能辅助。

### 安装 AI Skill

#### 方式 1: 从远程安装

```bash
go install github.com/wangtengda0310/gobee/gameactor/cmd/install-skill@latest
gameactor-install-skill
```

#### 方式 2: 从仓库安装

```bash
cd gameactor
go run cmd/install-skill/main.go
```

#### 方式 3: 自定义安装位置

```bash
gameactor-install-skill -target /path/to/skills
```

#### 验证安装

```bash
# 检查安装目录
ls ~/.claude/skills/gameactor/

# 应该看到以下文件：
# SKILL.md
# README.md
# abilities/
#   └── AI指导.md
```

安装后，重启 Claude Code 即可使用 AI Skill。

### 如何使用 AI Skill

#### 方式 1: 直接提问（推荐）

在 Claude Code 对话中直接描述需求：

```
"生成玩家战斗系统的代码"
"诊断并发问题：多个玩家同时登录数据不一致"
"写个单元测试验证任务顺序"
```

#### 方式 2: 引用 Skill 文件

```
"请阅读 @gameactor/.claude/skills/gameactor/SKILL.md 了解详细用法"
```

### AI 支持的操作

| 操作 | 示例问题 | AI 会做什么 |
|------|----------|-------------|
| **代码生成** | "生成聊天系统代码" | 分析场景 → 生成 DispatchBy 代码 |
| **问题诊断** | "数据不一致" | 检查 hash 使用 → 提供解决方案 |
| **测试生成** | "写个单元测试" | 使用 TestDispatcher → 添加断言 |
| **性能分析** | "执行太慢" | 检查配置 → 查看指标 → 调优建议 |
| **重构建议** | "如何改造" | 分析代码 → 引入 gameactor |

### Skill 文件位置

```
gameactor/.claude/skills/gameactor/
├── README.md       # Skill 目录说明
├── SKILL.md        # 完整使用指南
└── abilities/
    └── AI指导.md   # AI 能力指导
```

### 快速示例

#### 用户: "生成玩家战斗系统"

**AI 响应**:
```go
func AttackMonster(playerID, monsterID uint64, damage int) error {
    return gameactor.DispatchBySync(playerID, func() error {
        player := getPlayer(playerID)
        monster := getMonster(monsterID)
        player.Attack(monster, damage)
        return savePlayer(player)
    })
}
```

#### 用户: "测试用 TestDispatcher"

**AI 响应**:
```go
func TestAttackMonster(t *testing.T) {
    td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
    defer td.Shutdown(5 * time.Second)

    err := td.DispatchBySync(1001, func() error {
        return AttackMonster(1001, 2001, 100)
    })

    if err != nil {
        t.Errorf("AttackMonster failed: %v", err)
    }

    td.AssertExecuted(1001, 1)
}
```

### 详细文档

- **使用指南**: @gameactor/.claude/skills/gameactor/SKILL.md
- **AI 能力**: @gameactor/.claude/skills/gameactor/abilities/AI指导.md
- **架构设计**: @gameactor/DESIGN.md
- **开发指南**: @gameactor/CLAUDE.md
