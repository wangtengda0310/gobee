# msg — 协议解析与运行时

协议测试的底层核心包,负责 TCP 帧编解码、加密/压缩、登录鉴权、重放编排、变量提取、连接池管理。

## 目录结构

| 文件 | 职责 | 关键符号 |
|------|------|----------|
| [frame.go](frame.go) | TCP 帧编解码(4字节头:bodyLen+flags) | `EncodeFrame`, `DecodeFrame`, `DecodedFrame`(params别名), `FrameHeaderSize` |
| [crypto.go](crypto.go) | XOR 加密/解密(body 部分) | `xorEncrypt`, `xorDecrypt`, `FlagEncrypt` |
| [compress.go](compress.go) | Snappy 压缩/解压 | `FlagCompress`, `compressData`, `decompressData` |
| [decoder.go](decoder.go) | 原始字节流→`DecodedFrame` 转换,`RelayAndParse` 代理转发 | `DecodeFrame`, `RelayAndParse`, `DirClientToServer`, `DirServerToClient` |
| [transport.go](transport.go) | 传输层:客户端消息编码、payload→JSON、连接读取 | `EncodeClientMessage`, `payloadToJSON`, `readDrainer`, `sendRawMessage` |
| [auth.go](auth.go) | HTTP 登录鉴权 + TCP LoginReq 发送/LoginResp 等待 | `AuthLogin`, `sendLoginReq`, `waitLoginResp`, `extractAccountFromLoginPayload` |
| [conn_pool.go](conn_pool.go) | 账号连接池(复用已登录 TCP 连接) | `AccountConnectionPool`, `ConnPoolEntry` |
| [frame_mux.go](frame_mux.go) | 帧多路复用器(持续读帧+按 MsgID 缓存+阻塞等待) | `FrameMux`, `NewFrameMux`, `WaitMsg`, `DrainAndStart`, `GetCache` |
| [record.go](record.go) | 录制器运行时(内存驻留,实时回调) | `Recorder`, `NewRecorder`, `RecordFrame`, `ToRecording` |
| [replay.go](replay.go) | 重放编排(账号循环→连接获取→变量上下文→发送循环) | `SendMessages`, `sendMessagesOnce`, `prepareVariableContext`, `resolveVariablePayload`, `replaySession` |
| [variable.go](variable.go) | 向 params 注入 `MessageFactory`，提供 `StripByteStreamPrefix`/`UnmarshalProtoPayload` 别名 | `MessageFactory`, `StripByteStreamPrefix`, `UnmarshalProtoPayload` |
| [variable_runtime.go](variable_runtime.go) | 变量编排(注册、扫描、惰性提取、payload 注入) | `ExtractVariablesForMessage`, `msgNeedsVariable`, `ResolveMessageVariables`, `ScanFieldValuesForVariables` |
| [msg_registry.go](msg_registry.go) | 消息注册表(MsgID→proto.Message 工厂) | `NewMessage`, `GetMsgName` |
| [display.go](display.go) | 展示辅助(消息方向符号、名称美化) | `DirArrow` |
| [fakconn_impl_test.go](fakconn_impl_test.go) | FakeConn 内存连接(测试基建,仅测试构建编译) | `NewFakeConn`, `PushServerFrame`, `WaitClientWrite`, `MakeServerPayload` |

## 核心数据流

### 重放流程(变量提取场景)

```
SendMessages (账号范围循环,并发)
  └─ sendMessagesOnce (单账号)
       ├─ acquireConnection (连接池优先,否则 HTTP登录+TCP拨号)
       ├─ prepareVariableContext
       │    ├─ ScanFieldValuesForVariables → 有变量字段? (含未注册变量,F1 修复)
       │    ├─ 是 → NewFrameMux(conn, watchedIDs)  (未注册变量 watchedIDs 可能为空)
       │    │       └─ DrainAndStart(池连接) 或 readLoop(新连接)
       │    └─ 否 → readDrainer (仅消费,不缓存)
       ├─ 发送循环 (逐条消息 × repeatCount 轮)
       │    ├─ 跳过 Ack/Ntf (direction != "→")
       │    ├─ msgNeedsVariable(msg)?
       │    │    ├─ 是 → ExtractVariablesForMessage (惰性按需)
       │    │    │       └─ WaitMsg(watchedID, 5s) → ExtractFunc → variableStore
       │    │    │       提取失败(超时/未注册/FrameMux 已停) → 跳过发送,计为 failed
       │    │    └─ resolveVariablePayload (注入变量值到 payload)
       │    └─ sendRawMessage (编码+加密+发送)
       └─ cleanup (Stop FrameMux / 归还连接池 / 关闭连接)
```

> **readLoop 异常退出 (F3 修复)**: readLoop 因连接断开(io.EOF)退出时,
> 通过 `signalStopped()` 置 stopped=true + close(done),让进行中的 WaitMsg
> 立即返回错误而非等满 5s 超时,避免多账号并发场景延迟放大。

### 变量提取时序(工会战场景)

```
客户端                          服务器
  │                               │
  │── GuildCityWarDataReq ──────→ │  (前置请求)
  │                               │
  │←── TransportRawNtf ──────────│  (含 PveGuildCityDataNtf → cityId)
  │    [readLoop 缓存到 FrameMux.cache]
  │                               │
  │── TeamSelectGuildCityReq ───→ │  (cityId 已从 cache 提取替换)
  │    [ExtractVariablesForMessage 命中 cache]
  │                               │
```

关键:变量提取是**惰性按需**的——只在发送依赖变量的消息前才触发 `WaitMsg`,确保 Ntf 已被 readLoop 缓存。

## 核心类型

| 类型 | 文件:行 | 说明 |
|------|---------|------|
| `DecodedFrame` | frame.go | 解码后的帧(MsgID, SeqID, Flags, Payload) |
| `RecordMessage` | replay.go:29 | 重放消息(MsgID, PayloadJSON, FieldValues, Direction) |
| `FieldMetaValue` | replay.go:40 | 字段元数据(InputType, VariableName) |
| `FrameMux` | frame_mux.go:30 | 帧多路复用器(缓存 watchedIDs,阻塞 WaitMsg) |
| `replaySession` | replay.go:175 | 重放会话状态(conn, mux, variableStore, cleanup) |
| `Recorder` | record.go:26 | 录制器(内存驻留,RecordFrame → onRecord 回调) |
| `AccountConnectionPool` | conn_pool.go | 账号连接池(GetOrCreate, Return, Close) |

## 连接池并发安全设计（2026-06-17 修复）

连接池 `AccountConnectionPool` 管理多个账号的 TCP 长连接，核心并发安全机制如下：

### 写互斥（writeMu + writeLockedConn）

`PooledConn` 包含 `writeMu sync.Mutex`，`Borrow()` 返回的 `writeLockedConn` 在 `Write` 时自动获取 `writeMu`。
这保证同一时刻只有一个 goroutine 在写底层 `net.Conn`——无论是心跳 goroutine 还是 borrower 的业务发送。

```
heartbeat goroutine          borrower (sendRawMessage)
      │                              │
      ├─ pc.mu.Lock (查 state)       │
      ├─ state==Idle? ──是──┐         │
      ├─ pc.mu.Unlock      │         │
      │                    │         ├─ writeLockedConn.Write
      ├─ writeMu.Lock ◄────┘         │  ├─ pc.writeMu.Lock
      ├─ conn.Write(ping)            │  ├─ conn.Write(frame)
      └─ writeMu.Unlock              │  └─ pc.writeMu.Unlock
                                   │
```

`heartbeat` 在获取 `writeMu` 前检查 `state==Idle`，获取 `writeMu` 后二次确认，防止状态在间隙中改变。

### 读同步退出（readDrainer done channel）

`readDrainer` 是消费服务端推送的 goroutine，在 `cleanup()` 归还连接前必须完全退出，否则下一个 borrower 会与之并发读取同一 `net.Conn`。

```
cleanup()
  ├─ close(stopReader)          // 通知 readDrainer 退出
  ├─ conn.SetReadDeadline(now)  // 强制中断 io.ReadFull
  └─ <-readerDone (500ms 超时) // 同步等待 readDrainer 退出
```

`readDrainer` 函数退出时 `close(done)` 通知 `cleanup()` 同步完成。

### 帧边界排空（DrainConn）

`DrainConn` 按完整帧读取（header → ParseFrameHeader → body），超时返回 `nil`（表示排空完成），而非停在帧中间。
这与裸 `Read(buf)` 不同——后者可能读到任意字节位置，导致后续帧解析错位。

### 连接存活检测（TCP keepalive）

`IsConnAlive` 函数已删除。改为在 `dialAndLogin` 和 `AcceptConn` 中设置 TCP keepalive，由操作系统检测死连接，避免探测读取丢弃帧首字节。

## 默认并发上限与限流重试（2026-06-17）

`SendMessages` 通过 `MaxConcurrency` 控制同时启动的账号 goroutine 数量。为避免大量账号同时 HTTP 登录触发服务端限流（登录服返回 `code=-600`）：

- `MaxConcurrency = 0` 时表示使用 `DefaultMaxConcurrency`（当前为 10），而非"无限制"。
- 显式设置 `MaxConcurrency > 0` 时按指定值执行。
- 相关常量定义在 `[replay.go](replay.go)` 和 `[retry.go](retry.go)` 中。

对于因限流（`code=-600`）失败的账号，`SendMessagesWithRetry` / `ExecuteAccountsWithRetry` 会自动重试：

- 默认最大重试 3 次，间隔 500ms。
- 重试时会根据失败率动态降低并发度：失败率 > 50% 时并发度减半，20%~50% 时减 1。
- 非限流错误不会触发重试。

## Authenticator 接口（2026-06-17）

`sendMessagesOnce` 通过 `Authenticator` 接口解耦"连接获取"与"连接释放"的具体实现：

| 类型/函数 | 文件 | 说明 |
|-----------|------|------|
| `Authenticator` | [auth.go](auth.go) | 接口：`Authenticate(ctx, accountID, skipDrain)` + `Return(accountID, conn, lastSeqID)` |
| `HTTPAuthenticator` | [auth.go](auth.go) | 默认实现：HTTP 登录 + TCP 拨号 + LoginReq/Resp；Return 时关闭连接 |
| `PooledAuthenticator` | [pooled_authenticator.go](pooled_authenticator.go) | 包装内部 Authenticator，优先从连接池借出，失败则回退到 inner；Return 时归还连接池 |
| `AuthError` / `IsRateLimitError` | [auth.go](auth.go) | 结构化登录错误及限流判断 |

连接池复用逻辑已纳入 `Authenticator` IoC：`PooledAuthenticator` 负责 borrow/drain，`HTTPAuthenticator` 负责新建连接，`sendMessagesOnce` 不再直接操作 `AccountConnectionPool`。`ReplayOptions.ConnPool` 仅作为配置传入，`Replayer.normalize` 会自动用 `PooledAuthenticator` 包装当前 `Authenticator`。

## ReplayOptions 与 Replayer（2026-06-17）

为减少 `SendMessages` / `SendMessagesWithRetry` 的过多位置参数，引入参数对象与执行器：

| 类型/函数 | 文件 | 说明 |
|-----------|------|------|
| `ReplayOptions` | [replayer.go](replayer.go) | 一次重放任务的全部配置（连接、账号范围、用例、回调、池、认证器、并发/重试策略） |
| `Replayer` | [replayer.go](replayer.go) | 执行器：`NewReplayer(opts).SendMessages()` / `SendMessagesWithRetry()` |
| `SendMessages`（函数） | [replay.go](replay.go) | 向后兼容包装：构造 `ReplayOptions` 并调用 `Replayer` |
| `SendMessagesWithRetry`（函数） | [retry.go](retry.go) | 向后兼容包装：构造 `ReplayOptions` 并调用 `Replayer` |

CLI（`cobra_replay.go`、`cobra_unity_log.go`、`cobra_send.go`）和 Wails（`cmd/rain-qa-func/wails.go`）均已迁移到 `NewReplayer`。旧函数保留以兼容外部调用和旧测试。

推荐使用方式：

```go
err := protocol.NewReplayer(protocol.ReplayOptions{
    ServerAddr:  "10.254.114.204:18000",
    HTTPAddr:    "10.254.114.204:20144",
    OpenID:      "test",
    RangeStart:  1,
    RangeEnd:    10,
    Messages:    msgs,
    RepeatCount: 1,
    Context:     ctx,
    OnProgress:  onProgress,
    OnMessage:   onMessage,
    ConnPool:    pool,
}).SendMessages()
```

`ReplayOptions` 支持注入自定义 `Authenticator`，便于替换登录逻辑或在测试中 mock 网络。

`sendMessagesOnce` 内部使用 `accountRunOptions` 收敛对单个账号执行一次发送流程所需的参数。`accountRunOptions` 嵌入 `ReplayOptions` 复用批量配置字段，仅保留 `accountID`、`auth`、`grandTotal`、`alreadySent` 等单账号差异字段。

因此不指定 `--concurrency` 时，100 个账号会分 10 个一批启动；遇到限流时自动降低并发并重试，无需手动分批脚本。

## 变量类型

字段 `input_type: "variable"` 支持两类变量，具体实现见 [variables 包](../variables/CLAUDE.md)：

| 变量 | 短名 | 来源 | 说明 |
|------|------|------|------|
| 城池 ID | `cityId` | 服务端 `PveGuildCityDataNtf` / `TransportRawNtf` | 工会城战场景，由 `FrameMux` 缓存并惰性提取 |
| 当前账号 | `openid` | 发送循环预置 `accountID` | 账号级变量，无需 Ntf 提取，适合将当前账号传入 payload |

变量注册表位于 `params` 包，由 `variables` 包在 `init()` 中注册；`msg` 包负责运行时扫描、惰性提取和 payload 注入。`DecodedFrame` 实际定义在 [params/frame.go](../params/frame.go:7)。

## 帧格式

```
┌──────────┬───────┬─────────────────────────────┐
│ bodyLen  │ flags │           body              │
│ 2 bytes  │2 bytes│       bodyLen bytes         │
│ (LE)     │       │ (XOR加密 + 可选Snappy压缩)  │
└──────────┴───────┴─────────────────────────────┘
                   │                             │
                   └─ body = MsgID(2) + SeqID(4) + Payload(n)
```

- `FlagEncrypt` (0x01): body 经 XOR 加密
- `FlagCompress` (0x02): body 经 Snappy 压缩(加密后)
- Payload 格式: 2字节LE长度前缀 + proto 序列化数据(ByteStream)

## 测试索引

| 测试文件 | 覆盖范围 |
|----------|----------|
| [fakconn_test.go](fakconn_test.go) | FakeConn 基础:Push/Wait、字节流语义、deadline 中断、并发安全 |
| [frame_mux_test.go](frame_mux_test.go) | FrameMux:drain 不缓存、WaitMsg cache 命中/阻塞/超时、Stop 中断、并发访问 |
| [lazy_extract_test.go](lazy_extract_test.go) | 惰性变量提取:命中/跳过已有值/超时失败/未注册报错/工会战完整时序 |
| [variable_helper_test.go](variable_helper_test.go) | 测试用 cityId 变量注册与提取 |
| [variable_test.go](variable_test.go) | 变量协议解码:extractGuildCityID 各场景(已迁移到 variables 包) |
| [variable_runtime_test.go](variable_runtime_test.go) | 变量注册/查找/占位符解析 |
| [replay_test.go](replay_test.go) | 账号范围/名称生成 + openid 变量替换/上下文准备（不依赖真实服务器） |
| [conn_pool_test.go](conn_pool_test.go) | 连接池创建/关闭/归还 |
| [auth_test.go](auth_test.go) | LoginReq payload 解析 |
| [transport_test.go](transport_test.go) | 编解码 round-trip |
| [create_role_test.go](create_role_test.go) | **集成测试**(需真实服务器,CI 应跳过) |

### FakeConn 测试基建

详见父目录 [CLAUDE.md](../CLAUDE.md) "测试基建:FakeConn" 章节。

## 相关文档

| 文档 | 路径 |
|------|------|
| 父包文档 | [../CLAUDE.md](../CLAUDE.md) |
| 设计决策 | [../docs/design-decisions.md](../docs/design-decisions.md) |
| 已知问题 | [../docs/known-issues.md](../docs/known-issues.md) |
| 前端页面 | `frontend/src/pages/stream-proxy/CLAUDE.md` |
| 布局文档 | `frontend/docs/layout/pages/stream-proxy/index.md` |
| E2E 测试 | `frontend/e2e/stream-proxy/*.spec.ts` |
| 用例格式 | `cases/proto_cases/CLAUDE.md` |
