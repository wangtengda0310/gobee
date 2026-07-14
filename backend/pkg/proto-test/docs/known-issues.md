# proto-test 已知问题

## 已解决的问题

详见 [ISSUES-RESOLVED.md](../ISSUES-RESOLVED.md)：

1. **LoginResp 消息序号不匹配** — `waitLoginResp` 改为循环等待
2. **重放消息未追加到表格** — 传递 `EmitReplayMessage` 回调
3. **重放进度标签不更新** — 添加 `emitProgress` 方法
4. **重放结果页签无法接收数据** — 字段名统一为 `"latest_msg"`
5. **重放结果页签双事件通道修复** — 新增 `replay:result` 事件通道
6. **工会战动态变量提取失效（2026-06-15）** — `TeamSelectGuildCityReq` 的 `cityId` 始终用写死值，未从 `TransportRawNtf` 动态提取。根因：`prepareVariableContext` 在连接建立后立即调用 `ExtractVariableValues` 全量提取，而 Ntf 仅在客户端发送前置 `GuildCityWarDataReq` 后才由服务器推送，提取时序过早导致 `WaitMsg` 超时、`variableStore` 为空。修复：改为**惰性按需提取**——`prepareVariableContext` 只启动 `FrameMux` 不提取，发送循环中对每条依赖变量的消息发送前调用 `ExtractVariablesForMessage` 按需提取（复用已有值，仅对缺失变量 `WaitMsg`）。同时修复连接池 drain 阶段不再缓存积压帧（防止读到上次会话残留），提取失败时跳过该消息发送并计为失败（而非用写死值静默兜底）。验证：`TestLazyExtract_GuildWarTimeline` 用 FakeConn 模拟完整工会战时序通过。

## 待解决问题

1. **Service 未拆分**：当前所有方法集中在 `wails.go` 的 `StreamProxyService` 中
2. **wails3 dev bindings rename 失败（Windows）**：见 `docs/Wails开发注意事项.md`
3. **readDrainer 只推送 Proto 消息**：框架消息（MsgID < 1000）不会被推送到前端
4. **CLI 账号范围迭代**：`cmd/tests/streamproxy/main.go` 需增加 `-range-start` / `-range-end` 参数
5. **多线程并发迭代**：当前账号范围循环是串行的，可改为 goroutine pool
6. **账号迭代进度粒度**：`replay:progress` 的 total 未反映多账号总量
7. **~~readDrainer 强制回收~~** — ✅ 已修复（2026-06-17）
   - 中断取消时 `readDrainer` goroutine 可能泄漏
   - 修复：`readDrainer` 增加 `done` channel，`cleanup()` 设置 `SetReadDeadline(time.Now())` 强制中断 `io.ReadFull` 并等待 `done` 同步退出
   - 相关代码：`[readDrainer](msg/transport.go:78)`、`[cleanup](msg/replay.go:192)`、`[prepareVariableContext](msg/replay.go:300)`
8. **多账号重放创角成功但登录仍提示创建角色** — 根因：FaSDK 与 Giant SDK 认证体系的 accountid 不匹配。验证：`cmd/tests/replay_case` 工具执行后账号在游戏客户端中角色存在且可正常登录。

## 拦截模式相关限制（2026-06-10 调查更新）

> **完整调查与改造计划**：[intercept-batch-release-investigation.md](intercept-batch-release-investigation.md)（待单独会话实现）

当前实现与目标设计存在显著差距：

1. **放行路径错误**：「重发」走 `SendMessages` 新开连接，未在原代理 `serverConn` 上注入；`conn_id` 前端未使用
2. **无后端 pending 队列**：读循环不阻塞，拦截帧从客户端 socket 消费但未转发；仅前端 `interceptedSeqIDs` 做 UI 标记
3. **登录后批量 Req**：客户端异步连发时 toast 刷屏、选中行总跳到最后一条
4. **RecordLoginPayload 未生效**：`interceptAndParse` 中条件 `MsgID >= 1000` 导致 LoginReq 未写入 `login_payload_b64`
5. **filePath 隐式临时文件**：自动生成 `record_{ts}.json` 于 worktree 根目录，二次录制覆盖同路径，UI 不展示

目标规则（已定，待实现）：Ping P0 透传；LoginReq 透传+可选录制；Proto Req 入 pending；重发=快照批量放行（下一批边界）；见调查文档 §4。

## 代码审查优化建议（2026-06-11 /simplify 审查，待后续处理）

来源：code quality + performance + reuse 三组并行审查 agent

### 高优先级

1. **Proto payload 提取逻辑重复**：`framePayloadJSON`（record_worker.go:632）和 `emitInterceptedMessage`（record_worker.go:1043）中有逐字复制的 proto→JSON 提取逻辑，应提取为公共函数
2. **RelayAndParse 每帧 goroutine**：`decoder.go:144` 热路径中每帧 spawn 一个 goroutine 处理消息，高流量下有调度开销，可改为 worker pool 或同步处理
3. **HTTP body 重写 O(n*m) 字符串拼接**：`record_worker.go:484` 的 `replaceAddrInBody` 在循环中拼接字符串，应使用 `strings.Builder`

### 中优先级

4. **连接状态遍历模式重复 5 次**：`GetActiveAccounts`、`HasAccountConnection`、`InjectMessages`、`GetStateForAccount`（已删除）、`tryHandoffConnection` 中都有"遍历 connStates + lock-check-unlock"模式，可提取为 `forEachActiveState(callback)` 辅助
5. **run/runSend/runInject 三重复 run-done-select 模式**：`replay_worker.go` 中三个方法的 goroutine 启动+结果收集+取消处理逻辑几乎相同，可提取为 `runWithCancel(fn) (error, bool)` 框架
6. **forwardFrame 一次性辅助**：`record_worker.go` 的 `forwardFrame` 仅在两个地方调用，逻辑简单可内联
7. **ConnAge 字段未填充**：`conn_pool.go` 的 `ConnPoolEntry.ConnAge` 在 `List()` 中始终为空字符串

### 低优先级

8. **decoder.go 每帧三次分配**：`RelayAndParse` 热路径中每帧分配 header/body/decrypted 三个 slice，可使用 sync.Pool
9. **http.DefaultClient 无超时**：`record_worker.go:454` 使用默认 HTTP 客户端，应设置请求超时
10. **WriteLocker 接口使用不一致**：`decoder.go` 定义了 `WriteLocker` 接口用于保护 `conn.Write`，但 `record_worker.go` 中 3 处 `writeMu.Lock()+Write` 未通过该接口
