# Stream Proxy 协议重放问题修复记录

## 问题列表

### 1. LoginResp 消息序号不匹配

**问题**：`waitLoginResp` 假设 LoginResp 是登录后的第一条消息，但服务端可能先发送其他推送消息（如 SeasonInfoUpateNtf）。

**方案**：改为循环等待，跳过非 LoginResp 消息，添加 10 秒超时。

**修改**：`streamproto/record.go` 的 `waitLoginResp` 函数

**验证日志**：
```
[重放] 跳过消息，等待 LoginResp: SeasonInfoUpateNtf (MsgID=5050)
[重放] 登录成功
```

### 2. 重放消息未追加到表格

**问题**：`main.go` 中调用 `SendMessages` 时 `onMessage` 参数传 `nil`，导致重放消息无法通过 `record:progress` 事件推送到前端。

**方案**：将 `replay_worker.go` 中的 `emitReplayMessage` 改为导出方法 `EmitReplayMessage`，在 `main.go` 的 `SetReplayFunc`/`SetSendFunc` 中传递为回调。

**修改文件**：
- `main.go` 第 152 行（SetReplayFunc）
- `main.go` 第 159 行（SetSendFunc）
- `replay_worker.go` 第 167 行（方法导出）
- `use-record-data.ts` 第 85-113 行（事件处理器更新）

### 3. 重放进度标签不更新

**问题**：重放过程中进度标签显示 `0/0`，不实时更新。

**方案**：在 `ReplayWorker` 中添加 `emitProgress` 方法，使用 `Event.Emit('replay:progress')` 推送进度状态。

**修改文件**：
- `replay_worker.go` 第 131-143 行（emitProgress 方法）
- `use-record-data.ts` 第 74-82 行（事件监听器）

### 7. 连接池与读取生命周期并发竞态修复（2026-06-17）

**问题**：`openid-only` 用例（仅使用账号级变量 `openid`，无 `cityId` 等 Ntf 变量）在连接池复用场景下出现多类并发竞态和帧边界破坏。

**根因分析**：
1. **openid-only 用例缺少帧级排空**：`hasVariable=true` 导致 `skipDrain=true`，但 `watchedIDs` 为空走 `readDrainer` 路径，池连接积压帧未按帧边界排空。
2. **`DrainConn` 裸字节读取破坏帧边界**：原 `DrainConn` 用 `Read(buf)` 读字节，超时退出时可能停在帧中间，导致后续读帧错位。
3. **`readDrainer` 未同步退出导致并发读取竞态**：`cleanup()` 只 `close(stopReader)` + `sleep 50ms` 就 `Return` 连接，残留 `readDrainer` 可能与下一个 borrower 并发读取同一 `net.Conn`。
4. **心跳与 borrower 写竞态**：`heartbeat` 在检查 `state==Idle` 后解锁再 `conn.Write`，可能与 borrower 并发写同一连接。
5. **`IsConnAlive` 丢弃真实数据字节**：`IsConnAlive` 用 `Read(oneByte)` 探测连接存活，会丢弃帧首字节导致错序。
6. **heartbeat 死锁**：heartbeat 失败路径先锁 `pc.mu` 再锁 `p.mu`，与其他路径 `p.mu → pc.mu` 相反。

**修复方案**：
1. `prepareVariableContext` 在 `watchedIDs` 为空且 `borrowedFromPool=true` 时先调用 `DrainConn` 排空积压帧。
2. `DrainConn` 改为按帧读取（`io.ReadFull(header)` → `ParseFrameHeader` → `io.ReadFull(body)`），并返回 `error`。
3. `readDrainer` 增加 `done` channel，`cleanup()` 设置 `SetReadDeadline(time.Now())` 强制中断 `io.ReadFull` 并等待 `done`。
4. `PooledConn` 增加 `writeMu`，`Borrow()` 返回 `writeLockedConn`，`heartbeat` 写时获取 `writeMu`。
5. 从 `GetOrCreate`/`Has`/`Return` 中移除 `IsConnAlive`，删除该函数；改为在 `dialAndLogin`/`AcceptConn` 设置 TCP keepalive。
6. heartbeat 写失败时直接调用 `p.Close(accountID)`，避免锁顺序反转。

**修改文件**：
- `msg/conn_pool.go`: `PooledConn` 增加 `writeMu`，`Borrow()` 返回 `writeLockedConn`，`heartbeat` 获取 `writeMu`，`DrainConn` 按帧读取
- `msg/replay.go`: `replaySession` 增加 `readerDone`，`cleanup()` 同步等待 `readDrainer`，`prepareVariableContext` 在 `watchedIDs` 为空且 `borrowedFromPool=true` 时调用 `DrainConn`
- `msg/transport.go`: `readDrainer` 增加 `done` 参数

**验证**：`go build ./...` 通过；`TestConnPool` 测试通过；多账号并发重放无竞态 panic。

**文档创建时间**：2026-06-17

### 8. 多账号并发重放触发登录限流 code=-600（2026-06-17）

**问题**：`proto-test replay --concurrency 0 --range 1-100` 等场景下，只有部分账号能成功登录，其余账号返回 `HTTP 登录失败: 登录失败: code=-600`，导致前端/表格只看到部分结果。

**根因分析**：
1. `--concurrency 0` 原实现为"不限制并发"，`SendMessages` 会同时启动 `accountCount` 个 goroutine。
2. 每个 goroutine 独立调用 `AuthLogin()` 向登录服发起 HTTP 登录请求。
3. 登录服对同一来源的登录请求做了频率限流（`RATE_LIMIT = -600`），并发度过高时后续请求被拒绝。
4. 登录失败的账号不会发送业务消息，也不会产生服务端返回，因此"结果"只剩成功登录的账号。

**修复方案**：
1. 引入 `DefaultMaxConcurrency = 10` 常量；`MaxConcurrency = 0` 时使用默认值，不再"无限制"。
2. 新增 `RetryConfig`、`ExecuteAccountsWithRetry`、`IsRateLimitError`，对 `code=-600` 限流失败的账号自动重试。
3. 重试时动态降低并发度：失败率 > 50% 时并发度减半，20%~50% 时减 1，最小为 `MinConcurrency`。
4. 将重试与动态并发逻辑提取到独立文件 `msg/retry.go`，供 `SendMessages` 及未来其他 proto 协议路径复用。
5. CLI `proto-test replay` 新增 `--max-retries` 和 `--retry-interval` flag。
6. 将 HTTP/TCP 登录逻辑抽象为 `Authenticator` 接口：`sendMessagesOnce` 接收 `Authenticator`，默认实现为 `HTTPAuthenticator`，便于后续替换登录方式或接入其他 proto 协议路径。
7. 将连接池复用逻辑纳入 Authenticator IoC：新增 `PooledAuthenticator` 包装内部 `Authenticator`，负责 borrow/drain/return；`sendMessagesOnce` 不再直接操作 `AccountConnectionPool`。
8. 引入 `ReplayOptions` 参数对象与 `Replayer` 执行器，将 `SendMessages` / `SendMessagesWithRetry` / `sendMessagesOnce` 的过多参数收敛到配置对象中；旧函数保留为向后兼容包装。
9. CLI（`cobra_replay.go`、`cobra_unity_log.go`、`cobra_send.go`）与 Wails（`cmd/rain-qa-func/wails.go`）的调用点迁移到 `protocol.NewReplayer(...).SendMessages()` / `SendMessagesWithRetry()`。
10. 同步更新 `cobra-help.md`、`msg/CLAUDE.md`。

**修改文件**：
- `msg/retry.go`: 默认并发上限、重试配置、动态并发执行器；`SendMessagesWithRetry` 改为 Replayer 包装
- `msg/replay.go`: `SendMessages` 改为 Replayer 包装；`sendMessagesOnce` 接收 `Authenticator` 与 `accountRunOptions`；删除 `acquireConnection`
- `msg/auth.go`: 新增/更新 `AuthError`/`IsRateLimitError`/`Authenticator`/`HTTPAuthenticator`（含 Return 方法）
- `msg/pooled_authenticator.go`: 新增 `PooledAuthenticator`，将连接池纳入 Authenticator IoC
- `msg/replayer.go`: 新增 `ReplayOptions`/`Replayer` 与核心重放实现；normalize 中自动用 PooledAuthenticator 包装
- `msg/authenticator_test.go`: 新增/更新 Authenticator/PooledAuthenticator/Replayer 单元测试
- `cobra_replay.go`: 新增 `--max-retries`/`--retry-interval` flag；迁移到 `NewReplayer`
- `cobra_unity_log.go`: 迁移到 `NewReplayer`
- `cobra_send.go`: 迁移到 `NewReplayer`
- `cmd/rain-qa-func/wails.go`: Wails 发送函数迁移到 `NewReplayer`，复用全局连接池
- `cobra-help.md`: 更新 flag 说明
- `msg/CLAUDE.md`: 新增默认并发上限、限流重试、Authenticator 接口、ReplayOptions/Replayer 章节，说明 CLI/Wails 已迁移

**验证**：
- 20 账号、`--concurrency 0` 测试：修改前 10 成功 10 失败（code=-600）；方案 A 后 20 全部成功。
- 单元测试 `TestExecuteAccountsWithRetry_*` 覆盖：全部成功、限流重试、不可重试错误、动态并发降低、ctx 取消、空任务。
- 单元测试 `TestPooledAuthenticator_*`、`TestSendMessagesOnce_UsesAuthenticator`、`TestReplayer_SendMessagesWithRetry_*` 验证 Authenticator 注入、连接池 IoC 与 Replayer 默认配置路径。
- `go build ./...` / `go vet ./...` 通过
- 非集成单元测试通过

**文档创建时间**：2026-06-17
