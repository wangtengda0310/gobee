# AI 使用 gameactor 指导

## AI 触发场景

当用户请求以下操作时，AI 应主动使用 gameactor skill：

### 1. 代码生成

用户请求示例：
- "生成玩家 Actor 代码"
- "创建一个战斗系统的任务分发"
- "用 gameactor 实现聊天功能"

AI 行为：
1. 分析业务场景，确定 hash 字段（通常是 playerID）
2. 生成 `DispatchBy` 调用代码
3. 使用闭包捕获参数
4. 添加错误处理

### 2. 问题诊断

用户请求示例：
- "多个玩家同时登录时数据不一致"
- "任务执行顺序不对"
- "有死锁吗"
- "性能太慢了"

AI 行为：
1. 检查 hash 使用是否正确
2. 查找可能的竞态条件
3. 分析是否有长时间阻塞操作
4. 建议使用 `TestDispatcher` + 断言验证

### 3. 测试生成

用户请求示例：
- "写个单元测试"
- "测试玩家逻辑"
- "验证任务顺序"

AI 行为：
1. 使用 `TestDispatcher` 创建测试
2. 添加 `AssertExecuted` 断言
3. 覆盖正常和异常场景
4. 添加并发测试（多个 goroutine）

### 4. 性能分析

用户请求示例：
- "任务执行太慢"
- "CPU 使用率低"
- "如何提高吞吐量"

AI 行为：
1. 检查 `NumActors` 和 `QueueSize` 配置
2. 分析任务执行时间
3. 建议使用 `GetMetrics()` 查看指标
4. 提供调优建议

### 5. 重构建议

用户请求示例：
- "这段代码如何用 gameactor 改造"
- "如何避免数据竞争"
- "如何保证顺序执行"

AI 行为：
1. 分析现有代码结构
2. 识别需要顺序执行的操作
3. 确定 hash 字段
4. 生成重构后的代码

---

## AI 约束和原则

### 必须遵守

1. **测试隔离**: 必须使用 `TestDispatcher`，不用全局 API
2. **闭包优先**: 推荐闭包捕获参数，不推荐结构体传递
3. **错误处理**: `DispatchBySync` 必须检查返回的 error
4. **资源清理**: 测试中必须调用 `td.Shutdown()`

### 推荐做法

1. **简单优先**: 优先使用 `DispatchBy`，不是 `DispatchWithHash`
2. **引用代码**: 使用 `file:line` 格式引用相关代码
3. **测试先行**: 建议用户先用 `TestDispatcher` 验证
4. **性能意识**: 提醒用户避免在 handler 中长时间阻塞

### 避免陷阱

1. 不要在测试中使用全局 `gameactor.Init()`（sync.Once 问题）
2. 不要期望不同 hash 的任务有顺序保证
3. 不要忘记 `DispatchBy` 是异步的（需要等待或断言）
4. 不要在 handler 中执行 `time.Sleep()` 或网络 IO

---

## 快速决策树

```
用户请求
    │
    ├─ 包含"生成"/"创建"/"实现" + 系统名称
    │   └─> 生成代码（确定 hash 字段）
    │
    ├─ 包含"测试"/"验证"
    │   └─> 生成测试（使用 TestDispatcher）
    │
    ├─ 包含"问题"/"不一致"/"顺序"/"死锁"
    │   └─> 问题诊断（检查 hash 使用）
    │
    ├─ 包含"慢"/"性能"/"吞吐量"
    │   └─> 性能分析（配置、指标）
    │
    ├─ 包含"重构"/"改造"/"优化"
    │   └─> 重构建议（引入 gameactor）
    │
    └─ 其他 gameactor 相关
        └─> 参考基础 SKILL.md
```

---

## 示例对话

### 用户: "生成玩家战斗系统的代码"

**AI 响应流程**:

1. **分析需求**:
   - 确定 hash 字段：`playerID`
   - 确定操作：攻击怪物

2. **生成代码**:
   ```go
   package battle

   import "github.com/wangtengda0310/gobee/gameactor"

   // AttackMonster 玩家攻击怪物
   func AttackMonster(playerID, monsterID uint64, damage int) error {
       return gameactor.DispatchBySync(playerID, func() error {
           // 1. 获取数据
           player := getPlayer(playerID)
           monster := getMonster(monsterID)

           // 2. 执行战斗逻辑
           player.Attack(monster, damage)

           // 3. 保存结果
           return savePlayer(playerID)
       })
   }
   ```

3. **生成测试**:
   ```go
   func TestAttackMonster(t *testing.T) {
       td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
       defer td.Shutdown(5 * time.Second)

       playerID := uint64(1001)
       monsterID := uint64(2001)
       damage := 100

       err := td.DispatchBySync(playerID, func() error {
           return AttackMonster(playerID, monsterID, damage)
       })

       if err != nil {
           t.Errorf("AttackMonster failed: %v", err)
       }

       td.AssertExecuted(playerID, 1)
   }
   ```

4. **提醒注意**:
   - 相同 `playerID` 的攻击按顺序执行
   - 不同 `playerID` 的攻击并行执行
   - 避免在 `AttackMonster` 中执行长时间阻塞操作

---

### 用户: "多个玩家同时登录时数据不一致"

**AI 响应流程**:

1. **分析问题**:
   - 可能是并发访问玩家数据
   - 需要确保同一玩家的操作串行执行

2. **诊断代码**:
   ```go
   // ❌ 问题代码：没有使用 gameactor
   func Login(playerID uint64) error {
       player := getPlayer(playerID)
       player.Online = true
       return savePlayer(playerID)
   }
   ```

3. **提供解决方案**:
   ```go
   // ✅ 正确代码：使用 gameactor 保证顺序
   func Login(playerID uint64) error {
       return gameactor.DispatchBySync(playerID, func() error {
           player := getPlayer(playerID)
           player.Online = true
           return savePlayer(playerID)
       })
   }
   ```

4. **验证方法**:
   ```go
   // 使用 TestDispatcher 验证
   td := gameactor.NewTestDispatcher(t, gameactor.DefaultConfig())
   defer td.Shutdown(5 * time.Second)

   // 并发登录同一玩家
   var wg sync.WaitGroup
   for i := 0; i < 10; i++ {
       wg.Add(1)
       go func() {
           defer wg.Done()
           td.DispatchBySync(1001, func() error {
               return Login(1001)
           })
       }()
   }
   wg.Wait()

   // 验证执行了 10 次
   td.AssertExecuted(1001, 10)
   ```

---

## 代码引用格式

使用 `file:line` 格式引用代码：

- Dispatcher 核心: `gameactor/dispatcher.go:30`
- API 实现: `gameactor/api.go:50`
- 测试工具: `gameactor/testing.go:20`
- 隔离测试: `gameactor/isolation_test.go:100`
- 使用示例: `gameactor/examples/basic/main.go:1`

---

## 常见代码模式

### 模式 1: 基本 CRUD

```go
// Create
func CreatePlayer(playerID uint64, name string) error {
    return gameactor.DispatchBySync(playerID, func() error {
        player := &Player{ID: playerID, Name: name}
        return savePlayer(player)
    })
}

// Update
func UpdatePlayerName(playerID uint64, name string) error {
    return gameactor.DispatchBySync(playerID, func() error {
        player := getPlayer(playerID)
        player.Name = name
        return savePlayer(player)
    })
}

// Delete
func DeletePlayer(playerID uint64) error {
    return gameactor.DispatchBySync(playerID, func() error {
        return deletePlayer(playerID)
    })
}
```

### 模式 2: 事务操作

```go
func TransferItem(fromPlayerID, toPlayerID, itemID int) error {
    // 转移物品涉及两个玩家，需要分别在各自的 Actor 中执行

    // 1. 从发送者移除
    err := gameactor.DispatchBySync(uint64(fromPlayerID), func() error {
        fromInv := getInventory(fromPlayerID)
        return fromInv.Remove(itemID)
    })
    if err != nil {
        return err
    }

    // 2. 添加到接收者
    return gameactor.DispatchBySync(uint64(toPlayerID), func() error {
        toInv := getInventory(toPlayerID)
        return toInv.Add(itemID)
    })
}
```

### 模式 3: 异步日志

```go
func LogPlayerAction(playerID uint64, action string) {
    gameactor.DispatchBy(playerID, func() {
        // 日志记录不需要等待结果
        log.Printf("Player %d: %s", playerID, action)
    })
}
```

---

## 配置调优建议

### Actor 数量

```go
// 默认 1000，适用于大部分场景
// 根据同时在线玩家数调整

NumActors: max(1000, concurrentPlayers / 10)
```

### 队列大小

```go
// 默认 1000
// 根据任务积压情况调整

QueueSize: max(1000, avgTasksPerPlayer * 100)
```

### 监控指标

```go
metrics := gameactor.GetMetrics()
for _, m := range metrics {
    if m.QueueLength > 100 {
        log.Printf("Actor %d queue length: %d", m.ActorID, m.QueueLength)
    }
}
```
