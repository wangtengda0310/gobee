# 动态变量提取 — 重放时自动从 Ntf 提取参数注入 Req

> 日期: 2026-06-11
> 状态: 设计中
> 关联 commit: `34db192` (cmd/tests/replay_guild_war 中的 TransportRawNtf 解析蓝本)
> 基于 commit: `581dc58` (连接池复用、代理SeqID重写、录制连接移交)

## 1. 背景与动机

### 问题

重放工会城战用例时，`TeamSelectGuildCityReq` 的 `cityId` 字段需要使用服务端动态分配的值（通过 `TransportRawNtf` → `PveGuildCityDataNtf` 推送），而不是录制时的固定值。

当前重放流程（`streamproto/record.go` 中 `sendMessagesOnce`）的 `readDrainer` 只消费服务端推送用于日志，不解析、不缓存、不提取字段值。`cmd/tests/replay_guild_war` 中已实现了完整的解析链路，但它是独立 CLI 工具，QA 需要在 GUI 重放流程中获得相同能力。

### 目标

- QA 在 GUI 重放中可以使用 `${cityId}` 等变量，由程序自动从服务端 Ntf 中提取值并注入到 Req payload 中
- 无变量配置时，重放行为与现有逻辑完全一致（零改动）
- 未来扩展新变量只需在注册表添加一条配置，无需写解析器代码

## 2. 当前架构（commit 581dc58）

### 2.1 三条重放路径

| 路径 | 触发条件 | 连接来源 | 有 readDrainer | 适合变量提取 |
|------|---------|---------|---------------|-------------|
| **标准-连接池** | 正常重放，池中有连接 | `connPool.Borrow()` | ✅ 有 | ✅ 首期 |
| **标准-新建** | 连接池获取失败 | `net.Dial` 新建 | ✅ 有 | ✅ 首期 |
| **注入路径** | 录制中同账号重发 | 录制代理连接 | ❌ 无 | ❌ 后续 |

### 2.2 sendMessagesOnce 当前流程

```
sendMessagesOnce(connPool?)
  │
  ├─ 连接获取
  │   ├── connPool != nil → GetOrCreate → Borrow → DrainConn(100ms)
  │   └── 回退 → HTTP 登录 → TCP 连接 → LoginReq → LoginResp → 等 2 秒
  │
  ├─ 启动 readDrainer（只 log 不缓存）
  │
  ├─ 消息发送循环
  │   for r, i := range messages:
  │     sendRawMessage(conn, msg)
  │     onProgress/onMessage → 前端
  │
  └─ 清理
      ├── borrowedFromPool → connPool.Return()
      └── 新建连接 → conn.Close()
```

### 2.3 DrainConn 的注意点

`DrainConn(100ms)` 在 Borrow 后清空积压推送。这意味着：
- 如果 Ntf 在 Borrow 之前就已到达服务端发送缓冲区，会被 DrainConn 清掉
- 变量提取的 FrameMux 必须在 Borrow 之后立即启动，不能再有 DrainConn 盲清

## 3. 核心设计

### 3.1 生产者-消费者解耦模型

```
FrameMux.readLoop (生产者)         变量表              发送循环 (消费者)
─────────────────────────         ──────              ──────────────────
收到帧                             {                    for msg in msgs:
  → watchedMsgID?                    "cityId": 42       if has ${cityId}:
    → 通用路径解析 → 提取值          }                    注入 cityId=42
    → 存入变量表（覆盖式）                                发送 msg
  → 不是 → 仅 log + 推前端
```

- **无 TriggerMsgID**：Ntf 可在任意时刻到达（登录推送、Req 触发、服务端主动推送），readLoop 持续监听
- **变量取值优先用缓存**：发送前查变量表，有值直接用，没有则阻塞等新 Ntf（超时报错）
- **缓存策略**：只缓存 watchedMsgIDs 的帧，每个 MsgID 只保留最新一帧

### 3.2 变量注册表（短名 + 内部路径映射）

QA 看到的是人可读短名，后端内部映射为完整解析路径。

```go
// VariableDef 变量定义
type VariableDef struct {
    ShortName    string // 短名，如 "cityId"
    DisplayName  string // UI 显示名，如 "🏰 工会城战 - 城池ID"
    OuterMsgName string // 外层监听的消息名，如 "TransportRawNtf"
    InnerMsgName string // 信封内的消息名（空=无信封），如 "PveGuildCityDataNtf"
    FieldPath    string // 字段路径（点分隔），如 "matchedCities"
    PickStrategy string // 数组选择策略，如 "first_not_attacked"
}
```

**数组选择策略 (PickStrategy)**：

| 策略 | 含义 |
|------|------|
| `"index:N"` | 取数组第 N 个元素 |
| `"first_not_attacked"` | 取第一个 `isAttack==false` 的元素 |
| `"last"` | 取最后一个元素 |
| `"any"` | 取第一个非 nil 的元素 |

**工会战变量注册**：

```go
var builtinVariables = []VariableDef{
    {
        ShortName:    "cityId",
        DisplayName:  "🏰 工会城战 - 城池ID",
        OuterMsgName: "TransportRawNtf",
        InnerMsgName: "PveGuildCityDataNtf",
        FieldPath:    "matchedCities",
        PickStrategy: "first_not_attacked",
    },
}
```

### 3.3 首期 MVP：硬编码 cityId 解析（审核修复 M8）

通用反射解析器复杂度被低估（repeated 字段递归反射 + PickStrategy 与 FieldPath 隐含耦合）。首期采用硬编码解析：

```go
// VariableDef 变量定义
type VariableDef struct {
    ShortName    string                                          // "cityId"
    DisplayName  string                                          // "🏰 工会城战 - 城池ID"
    WatchMsgIDs  []uint16                                        // 关注的帧 MsgID 列表
    ExtractFunc  func(frame *DecodedFrame) (any, error)          // 提取函数
}
```

工会战变量注册：
```go
{
    ShortName:   "cityId",
    DisplayName: "🏰 工会城战 - 城池ID",
    WatchMsgIDs: []uint16{
        uint16(pb.EGameMsgID_TransportRawNtf_id),
        uint16(pb.EGameMsgID_PveGuildCityDataNtf_id),
    },
    ExtractFunc: extractGuildCityID,  // 复用 cmd/tests/replay_guild_war 的解析逻辑
}
```

`extractGuildCityID` 直接搬用 `cmd/tests/replay_guild_war/main.go` 中的 `parsePveGuildCityDataFromFrame` + `pickSelectableCityID`，强类型断言无需反射。

**后续扩展**：当有第二个变量需求时，提取通用反射框架，VariableDef 增加 FieldPath/PickStrategy 字段，ExtractFunc 设为 nil 时走通用路径。

### 3.4 TransportRawNtf 信封解析

复用 `cmd/tests/replay_guild_war` 中已验证的解析链路：

```
帧 payload → stripByteStreamPrefix → proto.Unmarshal(TransportRawNtf)
  → raw.MsgId + raw.Data
    → stripByteStreamPrefix(raw.Data) → proto.Unmarshal(内层消息)
      → FieldPath → PickStrategy → 最终值
```

## 4. 代码改动

### 4.1 新增文件

| 文件 | 职责 | 约行数 |
|------|------|--------|
| `streamproto/frame_mux.go` | FrameMux：缓存帧 + 等待特定 MsgID + 并发保护 + 安全退出 | ~150 |
| `params/variable_defs.go` | VariableDef + 注册表 + builtinVariables（ExtractFunc 由 streamproto 注入） | ~30 |
| `params/iteration.go` | IterationConfig + GenerateIterativeMessages + FieldValuesToIterationConfig | ~170 |
| `streamproto/variable_test.go` | FrameMux + 变量解析单元测试（net.Pipe 模拟连接） | ~150 |

### 4.2 改动文件

| 文件 | 改动 |
|------|------|
| `streamproto/record.go` | Recorder 运行时（纯数据类型定义在 `cases/record.go`）；`sendMessagesOnce` 中有变量 → FrameMux 替代 readDrainer；发送前注入变量值 |
| `params/iteration.go` | `IterationConfig` 加 `VarName string` 字段；`GenerateIterativeMessages` 中 `variable` 类型不展开 |
| `wails_replay_control.go` | 新增 `GetAvailableVariables()` API |
| `replay_worker.go` | 无改动（变量逻辑在 streamproto 层） |
| 前端卡片编辑器 | `input_type` 增加 `variable` 选项 + 变量名下拉 |
| 前端重放设置 | 增加变量功能开关 |

### 4.3 不改动的文件

| 文件 | 原因 |
|------|------|
| `conn_pool.go` | 连接池不感知变量，Borrow/Return 逻辑不变 |
| `replay_worker.go` | 变量逻辑在 streamproto 层，Worker 层透传 |
| `record_worker.go` | 注入路径首期不支持变量提取 |
| `frame.go` | 帧编解码不变，只新增 FrameMux 消费 |
| `msg_registry.go` | 消息注册表不变 |

## 5. FrameMux 设计

### 5.1 结构定义

```go
type FrameMux struct {
    conn       net.Conn
    mu         sync.RWMutex              // 保护 cache 和 variableStore（审核修复 C2）
    cache      map[uint16]*DecodedFrame  // 按 MsgID 缓存最新帧（仅 watchedIDs）
    watchedIDs map[uint16]bool           // 需要缓存的 MsgID 集合
    notify     *sync.Cond                // readLoop 写 cache 后 signal 唤醒 waitMsg（审核修复 M1）
    frames     chan *DecodedFrame        // 实时帧流（缓冲 1024）
    done       chan struct{}
    once       sync.Once
    wg         sync.WaitGroup            // 等待 readLoop 退出（审核修复 C3）
}
```

### 5.2 并发安全模型（审核修复 C2）

| 数据 | 写者 | 读者 | 保护机制 |
|------|------|------|---------|
| `cache` | readLoop | waitMsg / resolveVariables | `sync.RWMutex`（readLoop 持写锁，读取方持读锁） |
| `variableStore` | readLoop（提取值时） | 发送循环（resolveVariables） | 与 cache 共用同一把 RWMutex |
| `frames` channel | readLoop | waitMsg | channel 自身线程安全 |
| `done` channel | stop() | readLoop | close(channel) 广播 |

**变量表 (variableStore)** 是 `sendMessagesOnce` 的局部变量（每账号每调用独立），不是全局共享。多账号并发（rangeStart-rangeEnd）自然安全，每个 goroutine 有自己的 FrameMux + variableStore。

### 5.3 退出机制（审核修复 C3）

连接池模式下，readLoop 在 conn 上阻塞读取（`io.ReadFull`）。`stop()` 不能只依赖 done channel——`io.ReadFull` 不会响应 channel 信号。

**退出流程**：
```
stop()
  → close(done)                     // 信号
  → conn.SetReadDeadline(time.Now()) // 强制 io.ReadFull 返回 timeout 错误
  → wg.Wait()                       // 等待 readLoop goroutine 退出
```

**sendMessagesOnce 清理顺序**：
```
defer:
  1. mux.stop()                // 先停 readLoop（SetReadDeadline + Wait）
  2. time.Sleep(50ms)          // 与现有 readDrainer 的退出延迟一致
  3. connPool.Return(account)  // readLoop 已退出，安全归还连接
```

### 5.4 waitMsg 通知机制（审核修复 M1）

```go
// waitMsg 等待特定 MsgID 的帧
// 1. 先查 cache（持读锁）→ 有则立即返回
// 2. 无则进入等待：释放读锁 → notify.Wait()（内部释放写锁）
// 3. readLoop 写入 cache 后 notify.Signal() 唤醒
// 4. 超时由 context.WithTimeout 控制
func (m *FrameMux) waitMsg(msgID uint16, timeout time.Duration) (*DecodedFrame, error)
```

### 5.5 channel 满时行为（审核修复 M2）

frames channel 缓冲区 1024。如果满时 readLoop 丢弃最旧帧（而非阻塞），避免生产者阻塞导致 TCP 缓冲区满断连。对于 watchedIDs 的帧，写入 cache 不受 channel 满影响（cache 是 map 不是 channel）。

### 5.6 DrainAndStart

> **⚠️ 勘误 (2026-06-15 修复)**：本节原描述"drain 阶段检查 watchedIDs 并缓存"已作废。
> 实际实现中 drain 阶段**纯丢弃积压帧，不缓存**。原因：池复用连接的积压帧属于上一次会话残留，
> 缓存会让惰性变量提取（WaitMsg 命中 cache）读到过期数据（如上一个账号的 cityId）。
> 变量提取的数据源必须是 readLoop 启动后、由本会话 Req 触发的新帧。
> 同理 §9 时序场景表中"连接池 Borrow 后积压中有 Ntf → DrainAndStart 检查 watchedIDs → 缓存"一行也作废。

```go
// DrainAndStart 先消费积压帧（替代 DrainConn），然后启动 readLoop
func (m *FrameMux) DrainAndStart(drainTimeout time.Duration, onMessage ReplayMessageCallback, accountID string) {
    // drain 阶段：设置短 deadline 逐帧读取并检查 watchedIDs
    m.conn.SetReadDeadline(time.Now().Add(drainTimeout))
    for { /* 逐帧读取 + DecodeFrame + 检查 watchedIDs 缓存 */ }
    // drain 结束：清除 deadline
    m.conn.SetReadDeadline(time.Time{})
    // 启动 readLoop
    m.wg.Add(1)
    go m.readLoop(onMessage, accountID)
}
```

**关键**：drain 阶段和 readLoop 共享同一个 cache，积压帧中的 watchedIDs 帧也会被缓存。

## 6. sendMessagesOnce 改造

```
sendMessagesOnce(connPool?)
  │
  ├─ 解析 messages 中的变量依赖
  │   → 收集所有 ${varName} 引用
  │   → 查注册表得 watchedMsgIDs
  │
  ├─ 有变量？
  │   ├── 是 → 创建 FrameMux(watchedIDs) + variableStore
  │   └── 否 → 保持现有 readDrainer 逻辑
  │
  ├─ 连接获取
  │   ├── connPool → GetOrCreate → Borrow
  │   │   ├── 有变量 → FrameMux.DrainAndStart(100ms)（积压帧也检查 watchedIDs）
  │   │   └── 无变量 → DrainConn(100ms)（现有逻辑）
  │   └── 回退 → 新建连接 → LoginReq → LoginResp
  │
  ├─ readDrainer 或 FrameMux.readLoop 启动
  │
  ├─ 发送循环:
  │   for msg in messages:
  │     1. resolveVariables(msg, variableStore, mux, timeout)
  │        → 扫描 payloadJSON 中 ${varName} 占位符（字符串匹配）
  │        → 变量表有值 → 替换 payloadJSON 中的值
  │        → 变量表空 → mux.waitMsg() 阻塞等 → 提取后替换
  │        ⚠ 介入点：修改 payloadJSON 字符串（在 EncodeClientMessage 之前）
  │           EncodeClientMessage 内部会 json.Unmarshal → proto.Marshal
  │           占位符必须在 json.Unmarshal 之前被替换为实际值
  │     2. sendRawMessage(conn, msg)
  │     3. onProgress/onMessage → 前端

### 变量注入失败降级策略（审核修复 M5）

变量提取是增强功能，**不应让重放因注入失败而完全中止**：

| 失败场景 | 降级行为 |
|---------|---------|
| payloadJSON 解析失败 | 记录错误日志，使用原始 payload 继续发送 |
| ${varName} 占位符未找到 | 记录警告，使用原始 payload |
| 变量值类型不匹配（期望 int 实际 string） | 尝试 JSON 类型转换，失败则使用原始 payload |
| waitMsg 超时 | 记录警告，使用原始 payload（不含变量的字段保持原值） |

**核心原则**：降级后消息仍然发出（使用录制时的原始值），不会跳过或中止。
  │
  └─ 清理
      ├── borrowedFromPool → connPool.Return()
      └── 新建连接 → conn.Close()
```

### 前端重发路由策略（审核修复 C1）

现有 `collectIterativeStates()` 通过 `input_type !== 'original'` 判断是否有迭代配置，新增 `variable` 类型后会被误判为迭代类型路由到 `sendIterativeMessages`。

**采用方案**：后端 `sendIterativeMessages` 统一处理 variable 类型（不展开，保留 `${varName}` 占位符），前端路由逻辑**无需改动**。理由：
- `sendIterativeMessages` 最终调用 `sendMessagesOnce`，resolveVariables 在此处统一处理
- 避免拆分前端路由为两条路径（降低前端复杂度）
- variable 可以与 range/enum/combo 共存（如 `cityId=variable` + `count=range(1..5)` 做笛卡尔积）

### 注入路径处理

当 `ReplayControlService.SendMessages` 检测到录制代理冲突时，走 `StartInject` 路径。
首期版本：如果 messages 中包含 variable 类型字段且路由到注入路径，返回错误提示
**"变量提取暂不支持录制代理注入模式，请停止录制或使用不同账号"**（复用现有 `message.error()` toast）。

## 7. IterationConfig 扩展

### 7.1 结构体新增字段

```go
type IterationConfig struct {
    Type    string  // "range", "enum", "combo", "original", "variable" ← 新增
    Start   *int
    End     *int
    Step    *int
    Values  []any
    Field   string
    VarName string  // 变量名（如 "cityId"），仅 variable 类型 ← 新增
}
```

### 7.2 fieldValuesToIterationConfig 新增 variable 分支（审核修复 M4）

现有 `params.FieldValuesToIterationConfig(params/iteration.go)` 只处理 range/enum/combo 三种类型，variable 类型走到 default 分支会被跳过，导致 VarName 丢失。需新增：

```go
case "variable":
    configs = append(configs, IterationConfig{
        Type:    "variable",
        Field:   fieldName,
        VarName: fv.VarName,  // 从 FieldValues 中的新字段获取
    })
```

### 7.3 GenerateIterativeMessages 中 variable 类型的处理

```go
case "variable":
    // 不展开消息，保留原始 payload（含 ${varName} 占位符标记）
    // 由后续 resolveVariables 在发送前替换
    messages = append(messages, copyBase(baseMessage))
```

### 7.4 variable 与 combo 混合场景

当 variable 类型的字段出现在 combo 组合中时：
- variable 字段作为单值参与笛卡尔积（不展开）
- 例如 `cityId=variable` + `count=range(1..5)` → 生成 5 条消息，每条的 cityId 都是 `${cityId}` 占位符
- 发送时 resolveVariables 统一替换为实际提取值

### 7.5 变量占位符格式

payload 中 variable 类型字段的值设为特殊 JSON 标记：

```json
{"cityId": {"__var__": "cityId"}, "count": 5}
```

`resolveVariables` 扫描 payload，遇到 `{"__var__": "cityId"}` 时替换为实际值（保持 JSON 类型兼容——cityId 是 number，替换值也是 number）。

## 8. 前端交互

### 8.1 变量功能（已作为正式功能始终开放）

"动态变量提取"已从实验性开关改为正式功能，不再提供 NSwitch 开关，默认始终启用。

- 卡片编辑器的 `inputTypeOptions` 中固定包含 `🔗 变量` 选项
- 调用 `GetAvailableVariables()` 获取可用变量列表

### 8.2 卡片编辑器（审核修复 M7）

**FieldFourState 接口扩展**：新增 `var_name: string` 字段，与后端 `IterationConfig.VarName` 对齐。

**inputTypeOptions**：固定包含 `original`、`range`、`enum`、`combo`、`variable` 五个选项，变量模式作为正式功能始终可用。

**新建 `variable-select.vue`**：与 range-input.vue、enum-select.vue、combo-select.vue 同级，接收 `GetAvailableVariables()` 返回的列表作为 options。

```
┌─────────────────────────────────────┐
│ 字段: cityId     类型: [🔗 变量 ▼]   │
│ ┌─────────────────────────────────┐ │
│ │ 🏰 工会城战 - 城池ID (cityId)  │ │  ← 显示格式: DisplayName (ShortName)
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

前端所有涉及变量的位置统一显示格式为 `{DisplayName} ({ShortName})`，错误提示同样使用 DisplayName。

### 8.3 后端 API

```go
// GetAvailableVariables 返回所有可用变量
func (s *ReplayControlService) GetAvailableVariables() []VariableInfo {
    // 从 builtinVariables 转换为前端友好格式
}
```

## 9. 时序场景覆盖

| 场景 | 处理方式 |
|------|---------|
| Ntf 在登录推送时就到了 | readLoop 缓存 → 变量表有值 → 发送前直接注入 |
| 连接池 Borrow 后积压中有 Ntf | DrainAndStart 检查 watchedIDs → 缓存 → 注入 |
| Ntf 在某条 Req 后才推 | 变量表空 → waitMsg 阻塞 → Ntf 到了继续 |
| 发 consumer Req 时还没 Ntf | waitMsg 阻塞等 → 超时 → 降级用原始 payload |
| Ntf 永远不到 | waitMsg 超时 → 降级用原始 payload |
| 无变量配置 | 走 readDrainer，行为不变 |
| 同名 Ntf 推多次 | 覆盖式缓存，取最新值 |
| 路由到注入路径 | 首期返回错误提示 |
| TransportRawNtf 内层非目标消息 | ExtractFunc 返回 nil，静默跳过，继续等 |
| 空消息列表 | 跳过变量提取，直接返回成功 |
| 全 Ntf 方向消息 | 无 Req 发送，变量提取不被触发 |
| variable + combo 混合 | variable 作为单值参与笛卡尔积 |
| 变量注入失败 | 降级：使用原始 payload 继续发送 |

## 10. 已知限制与后续扩展

1. **注入路径不支持**：首期变量提取仅在标准重放路径（sendMessagesOnce）中生效，注入路径返回明确错误提示
2. **首期硬编码 cityId**：VariableDef.ExtractFunc 模式，无通用反射解析器。第二个变量需求时提取通用框架
3. **多级信封**：当前只支持一层信封（TransportRawNtf），多层嵌套需扩展
4. **变量间依赖**：一个变量依赖另一个变量的值，当前不支持
5. **`${` 自动补全**：后续可在 payload textarea 中实现 `${` 触发变量名补全
6. **连接池变量持久化**：连接 Return 后变量表失效，下次 Borrow 需重新等待。可考虑在连接池维度缓存变量表
7. **Borrow 与心跳竞态**：Borrow() 的 close(stopHeart) 不等待心跳 goroutine 退出，DrainAndStart 需容忍心跳最后一次写入（短延迟或心跳退出 WaitGroup）
8. **GetAvailableVariables 调用时机**：开关从 off→on 时调用一次 + 每次开始重放前调用一次（确保列表最新）
9. **变量提取超时前端反馈**：超时时 replay:progress 的 error_message 包含具体变量名和超时时间；重放前预检查 variable 字段是否有可用变量

## 审核修改记录

| 修复 ID | 优先级 | 问题 | 修改章节 |
|---------|--------|------|---------|
| C1 | Critical | collectIterativeStates() 误路由 variable | §6 注入路径处理 |
| C2 | Critical | FrameMux cache/variableStore 缺并发保护 | §5.2 并发安全模型 |
| C3 | Critical | readLoop 未退出就 Return 连接 | §5.3 退出机制 |
| M1 | Major | waitMsg 缺通知机制 | §5.4 waitMsg 通知机制 |
| M2 | Major | channel 满时行为未定义 | §5.5 channel 满时行为 |
| M3 | Major | resolveVariables 介入点不明确 | §6 发送循环 |
| M4 | Major | fieldValuesToIterationConfig 缺 variable 分支 | §7.2 |
| M5 | Major | 变量注入失败无降级策略 | §6 降级策略表 |
| M6 | Major | 前端开关位置不精确 | §8.1 |
| M7 | Major | FieldFourState 缺 var_name 字段 | §8.2 |
| M8 | Major | 反射通用解析器首期硬编码 | §3.3 |
