# 延迟（惰性）变量提取修复计划

> 日期: 2026-06-15
> 状态: 已审核（4 路并行审核通过，反馈已纳入 v2）
> 关联设计: docs/superpowers/specs/2026-06-11-replay-variable-extraction-design.md §3.1/§6
> 关联历史: dc2208d (被 8deee7a revert 的修复), 9e13ea0 (单独重新落地的 Recorder 重构)
>
> 审核记录（2026-06-15）:
> - 架构正确性: 1 Critical(acquireConnection DrainConn) / 2 Major(降级策略前端反馈、设计文档勘误) / 3 Minor
> - 并发安全: 无数据竞争/死锁; 2 Major(drain 丢弃窗口契约、WaitMsg 取消延迟) / 2 Minor
> - 测试质量: 基建扎实; P0 缺口(Ntf 延迟唤醒、DrainAndStart 池路径测试); P1 虚假测试未清理
> - 合并兼容性: 核心文件可直接 apply; 遗漏 excel.go worktree 层级修复 + 3 文档

## 1. 问题

重放工会战用例时，`TeamSelectGuildCityReq` 的 `cityId` 始终发送录制时的固定值（如 3014），未从服务端 `PveGuildCityDataNtf` 动态提取。

### 根因

`prepareVariableContext` (replay.go:256-277) 在发送循环**开始前**就调用 `ExtractVariableValues` 预提取所有 watchedID。而 `PveGuildCityDataNtf` 是服务端在收到 `TeamMatchGuildCityReq` 之后才推送的，此时还未发送任何 Req → `WaitMsg(5s)` 全部超时 → `variableStore` 为空 → `resolveVariablePayload` 因 `len(variableStore)==0` 直接返回录制固定值。

### 偏离设计

设计文档 §3.1/§6 明确定义的是**惰性按需提取**：每条 Req 发送前才检查变量依赖、查缓存或阻塞等待 Ntf。当前实现退化为"发送前一次性预提取"，违背设计意图。

### 测试盲区

- `replay_test.go` 含虚假测试（重新实现纯函数测自己），不覆盖真实重放链路
- `variable_test.go`/`variable_runtime_test.go` 只测孤立叶子函数，零集成测试覆盖时序
- `resolveVariables`（字符串占位符路径）在生产链路根本没被调用，是孤儿函数

## 2. 修复目标

1. 把变量提取从"发送前预提取"改回"发送循环内惰性按需提取"
2. 提取失败（如 Ntf 超时、变量未注册）改为报错并跳过该消息，而非静默用写死值兜底（QA 工具最怕"测试了错误数据却以为成功"）
3. 引入 FakeConn 测试基建 + 真正的集成测试，覆盖惰性提取全路径
4. 修正 `DrainAndStart` 的缓存语义：drain 只丢弃积压帧不缓存（避免池复用连接读到过期数据）
5. 不带入 Recorder 重构（已在 9e13ea0 独立落地），只做变量提取相关改动

## 3. 改动清单

### 3.1 变更文件

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `msg/variable_runtime.go` | 修改 | 新增 `msgNeedsVariable` + `ExtractVariablesForMessage`（惰性按需），删除旧 `ExtractVariableValues` 及孤儿函数 `resolveVariables`（审核 Minor-4：统一提取逻辑，消除混淆） |
| `msg/replay.go` | 修改 | `prepareVariableContext` 只启动 FrameMux 不预提取；发送循环改为按需触发 `ExtractVariablesForMessage`；`acquireConnection` 有变量时不调用 `DrainConn`（审核 Critical） |
| `msg/frame_mux.go` | 修改 | `DrainAndStart` 改为仅丢弃积压帧不缓存 |
| `msg/frame_mux_test.go` | 修改 | 对齐 drain 不缓存的新契约 |
| `msg/fakconn.go` | 新增 | FakeConn(net.Conn 内存实现)，供集成测试脚本化编排协议交互 |
| `msg/lazy_extract_test.go` | 新增 | 惰性提取全路径集成测试（含审核 P0 补充场景） |
| `msg/replay_test.go` | 修改 | 删除虚假测试 `openIDWithIndex`/`itoa`/`TestAccountCountCalculation`（审核 P1） |
| `msg/variable_runtime_test.go` | 修改 | 删除孤儿函数 `resolveVariables` 的测试 `TestResolveVariables`/`TestResolveVariablesFallback`（审核 P1） |
| `pkg/common/game/excel.go` | 修改 | worktree 目录层级 3→5（审核合并兼容性遗漏，被 revert 后从未落地） |
| `pkg/proto-test/docs/known-issues.md` | 修改 | 新增第 6 条"工会战动态变量提取失效"记录 |
| `pkg/proto-test/msg/CLAUDE.md` | 新增 | msg 包完整文档（目录结构、数据流时序图、测试索引） |
| `docs/superpowers/specs/2026-06-11-replay-variable-extraction-design.md` | 修改 | 勘误 §5.6/§9：drain 不缓存（审核 Major-3）；补充降级策略裁定说明 |

### 3.2 不变更的文件

- `record.go` / `record_worker.go` / `wails_record_control.go` — Recorder 重构已在 9e13ea0 落地，不重复（审核合并兼容性确认）
- `llm.go` — **绝对不取 dc2208d 版本**（会回退 `save_case` 到读文件模式，与 9e13ea0 内存驻留冲突）
- `variable.go`（协议解码 `extractGuildCityID` 等）— 解析逻辑本身正确，无需改
- `wails_replay_control.go` — 接口不变
- `cmd/tests/streamproxy/main.go` — 9e13ea0 已覆盖
- 前端代码 — 接口不变

## 4. 详细实现

### 4.0 acquireConnection 修改（审核 Critical）

当前 `acquireConnection`（replay.go:218）在池路径 `Borrow()` 后立即 `DrainConn(100ms)` 盲清。改为：**有变量上下文时跳过 `DrainConn`**，把 drain 职责完全交给 `DrainAndStart`（它能按 watchedIDs 做精细处理）。

由于 `acquireConnection` 在 `prepareVariableContext` 之前调用，此时还不知道是否有变量依赖。解决方案：**先扫描变量依赖，再获取连接**。调整 `sendMessagesOnce` 的阶段顺序：

```
阶段 0（新增）: ScanFieldValuesForVariables(messages) → hasVariable, watchedIDs
阶段 1: acquireConnection(..., skipDrain=hasVariable)  // 有变量时跳过 DrainConn
阶段 2: prepareVariableContext(conn, messages, ..., hasVariable, watchedIDs)
阶段 3: 发送循环
```

`acquireConnection` 签名增加 `skipDrain bool` 参数；池路径 `if !skipDrain { DrainConn(...) }`。

**说明**：对工会战场景（Req-触发型变量），原有的双重 drain 实际不影响修复有效性（积压帧是上次会话残留，本会话 Ntf 在 Req 之后才到）。但消除双重 drain 提升架构清晰度，且为未来"登录推送型变量"留出正确性空间。

### 4.1 variable_runtime.go

新增 `msgNeedsVariable` + `ExtractVariablesForMessage`，删除旧 `ExtractVariableValues` 和孤儿 `resolveVariables`：

```go
// msgNeedsVariable 判断单条消息是否依赖变量
func msgNeedsVariable(msg RecordMessage) bool {
	for _, fv := range msg.FieldValues {
		if fv.InputType == "variable" && fv.VariableName != "" {
			return true
		}
	}
	return false
}

// ExtractVariablesForMessage 按需提取单条消息所需的变量（惰性提取核心）
// 与 ExtractVariableValues 的区别：
//   - 只处理该消息 FieldValues 中声明的变量，而非全局 watchedIDs
//   - 先查 variableStore，跳过已有值的变量（避免重复 WaitMsg 5s 阻塞）
//   - 仅对缺失的变量调用 WaitMsg
// 返回错误时，variableStore 可能已填充部分成功提取的变量
func ExtractVariablesForMessage(msg RecordMessage, mux *FrameMux, variableStore map[string]any) error
```

`ExtractVariablesForMessage` 的行为：
- `mux == nil` 或消息不依赖变量 → 返回 nil
- 收集该消息需要的、且尚未提取的变量名 → VariableDef（复用 `params.FindVariableByShortName`，审核 Minor-5）
- 未注册变量 → 返回 error（阻止用写死值发送）
- 对每个缺失变量，逐个 WaitMsg 其 WatchMsgIDs（任一命中即提取成功）
- 全部超时 → 返回 error

### 4.2 replay.go 发送循环

`prepareVariableContext`：删除 `ExtractVariableValues(mux, watchedIDs, variableStore)` 调用，只启动 FrameMux + 创建空 `variableStore`。

`sendMessagesOnce` 发送循环（审核 Major-2：提取失败跳过时回传前端反馈）：

```go
hasVariableContext := sess.mux != nil  // 替代原 hasVariable := sess.variableStore != nil（语义等价，仅更精确，审核 Minor-6）

// 在发送每条消息前：
if hasVariableContext && msgNeedsVariable(msg) {
	if extractErr := ExtractVariablesForMessage(msg, sess.mux, sess.variableStore); extractErr != nil {
		log.Printf("...变量提取失败-跳过...")
		// 审核反馈：通过 onMessage 回传跳过原因，让前端可见
		if onMessage != nil {
			onMessage(msg.MsgName+"(跳过:变量提取失败)", msg.MsgID, msg.SeqID, "", msg.OffsetMs, DirClientToServer, accountID)
		}
		failed++
		roundFailed++
		continue
	}
	payloadToSend = resolveVariablePayload(msg, sess.variableStore)
}
```

### 4.3 frame_mux.go DrainAndStart

drain 循环中删除 watchedID 缓存逻辑，改为纯丢弃：

```go
// 原：检查 watchedID → 缓存 + log
// 新：仅 log 丢弃
log.Printf("[FrameMux] drain 丢弃积压: MsgID=%d, SeqID=%d, %dB", ...)
```

理由：池复用连接的积压帧属于上一次会话残留，缓存会让惰性变量提取（WaitMsg 命中 cache）读到过期数据。变量提取的数据源必须是 readLoop 启动后当前会话收到的新帧。

**契约约束**（审核并发 Major-1）：此设计要求 `VariableDef.WatchMsgIDs` 必须是**请求触发型**（由本会话发出的 Req 触发服务端推送），而非登录自发推送型。当前唯一变量 `cityId`（由 `TeamMatchGuildCityReq` 触发 `PveGuildCityDataNtf`）满足此约束。此契约需在 `variable_defs.go` 的 `VariableDef` 注释和 `msg/CLAUDE.md` 中固化。

### 4.4 frame_mux_test.go

3 个测试断言反转（drain 不再缓存）：
- `TestDrainAndStart`：验证 drain 后 cache 为空
- `TestDrainDecodingError`：验证 valid 帧也不缓存
- `TestDrainAndStartWithPipeEndClose`：验证 drain 不缓存

### 4.5 fakconn.go（新增）

`FakeConn` 实现 `net.Conn`，内存双向管道 + 脚本化写入。供 `lazy_extract_test.go` 模拟"发送 Req → 服务端推 Ntf → 提取变量"的完整时序。

### 4.6 lazy_extract_test.go（新增/审核 P0 补充）

覆盖场景（★ 为审核 P0/P1 补充）：
1. `msgNeedsVariable` 真假分支
2. `ExtractVariablesForMessage` 缓存命中（不重复 WaitMsg）
3. `ExtractVariablesForMessage` WaitMsg 超时
4. `ExtractVariablesForMessage` 未注册变量报错
5. 工会战完整时序：FakeConn → readLoop → 发送依赖变量的消息 → 验证提取的 cityId 注入 payload
6. ★ **Ntf 延迟到达→WaitMsg 阻塞唤醒**（审核 P0 盲区 A）：先启动提取（阻塞）→ 再推 Ntf → 验证唤醒后命中
7. ★ **DrainAndStart 池路径集成测试**（审核 P0 盲区 B）：用 DrainAndStart 启动 → 推过期帧（验证丢弃）→ 启动后推 Ntf → 验证惰性提取命中
8. ★ **提取失败后 continue 语义**（审核 P1 缺口 #6）：Ntf 缺失→变量消息跳过→后续非变量消息仍正常发送
9. 同名 Ntf 多次推送取最新值（审核 P2 缺口 #3）

## 5. 合并兼容性分析

dc2208d 被整体 revert 后，9e13ea0 单独重新落地了 Recorder 重构。当前代码状态：
- Recorder 已是内存驻留设计（`NewRecorder(serverAddr)` 无 filename，`StartRecord` 无 filePath）
- variable_runtime.go 仍是 buggy 的预提取版本（`ExtractVariableValues`）

因此本次修复**不能直接 cherry-pick dc2208d**（会与 9e13ea0 的 record.go/record_worker.go 冲突）。需要手工只取变量相关 diff：
- variable_runtime.go diff — 直接可用
- replay.go diff — 直接可用（replay.go 未被 9e13ea0 改动）
- frame_mux.go diff — 直接可用
- frame_mux_test.go diff — 直接可用
- fakconn.go / lazy_extract_test.go — 新文件，直接可用

## 6. 验证计划

1. `go build ./...` 编译通过
2. `go test ./backend/pkg/proto-test/...` 全部通过
3. 重点验证 lazy_extract_test 覆盖的场景（含审核 P0 补充的 Ntf 延迟唤醒、DrainAndStart 池路径）
4. `wails3 generate bindings` 重新生成 bindings（审核合并兼容性）
5. `wails3 task build` 构建产物生成
6. 运行时验证（用户手动）：重放工会战用例，确认 TeamSelectGuildCityReq 的 cityId 被替换为动态提取值

## 7. 风险与后续 TODO

- `ExtractVariablesForMessage` 提取失败时跳过消息而非兜底，可能改变既有行为。但这是设计文档 §6 降级策略的有意取舍（审核架构 Major-2 裁定：QA 工具应默认跳过 + 显式上报，比静默用写死值更安全）。已通过 onMessage 回传跳过原因。
- DrainAndStart 不缓存的改动影响池复用连接场景，已通过新连接路径的 readLoop 覆盖（审核并发确认：drain 间隙不丢帧，OS 缓冲兜底）。
- **WatchMsgIDs 契约**（审核并发 Major-1）：变量必须是请求触发型。已在 §4.3 固化，需同步写入 `variable_defs.go` 注释和 `msg/CLAUDE.md`。
- **TODO（审核并发 Major-2，不阻塞本次落地）**：给 `sendMessagesOnce` 透传 `ctx`，在每条消息发送前探测 `ctx.Done()`，或给 `WaitMsg` 加 context 参数，让取消在变量等待期（最长 5s）也能即时生效。当前取消响应延迟 bounded（5s/条变量消息）。
- **TODO（审核测试 P2）**：为 FakeConn 增加响应式推送能力（`OnClientWrite` 回调），提升时序保真度，支持更精确的"Ntf 在 Req 写出后自动推送"测试。
