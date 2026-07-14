# 实时拦截改包 — 调查结论与改造计划

> **调查日期**：2026-06-10  
> **产品规则确认**：2026-06-10（第二轮，见 §4.5）  
> **状态**：**已实现**（2026-06-10）  
> **关联待办**：`docs/TODO.md` →「streamproxy 拦截改包批量放行改造」

## 1. 背景

用户在「发包改包」页签开启实时修改（`filterMode`）后录制并登录游戏，观察到：

1. 登录后短时间内弹出多条「已拦截 XXX，请编辑后点击重发放行」toast
2. 点击「重发」后，行为与「在原连接上放行」的预期不符
3. 客户端异步连续发送多条 Req（如 PostUserStatusReq、FriendListReq 等），需明确程序如何处理

本调查覆盖：当前实现、与产品预期的差距、已确认的设计规则、录制文件路径（`filePath`）问题。

---

## 2. 实测时序（2026-06-10 日志摘录）

```
11:17:06.664  LoginResp（框架消息，原样转发）
11:17:06.667~673  服务端批量推送 30+ 条 Ntf（S→C，不拦截）
11:17:06.673  LoginServerMsgOverNtf（登录流程结束）
11:17:06.759  拦截 PostUserStatusReq   SeqID=2
11:17:06.792  拦截 FriendListReq       SeqID=3
11:17:06.826  拦截 BlackListReq        SeqID=4
11:17:06.860  拦截 GiftShopListReq     SeqID=5
11:17:06.893  拦截 ShopQueryReq        SeqID=6
```

5 条 Req 在 **134ms** 内全部被拦截；Ping/Pong 仍按 S→C / 框架路径正常转发。

录制文件：`record_20260610_111702.json`（worktree 根目录，相对路径无目录前缀）。

---

## 3. 当前实现摘要

### 3.1 拦截路径（`record_worker.go` → `interceptAndParse`）

- `MsgID >= 1000` 的 C→S 消息：**不转发**到 `serverConn`，`emit record:intercepted`，`continue` 读下一帧
- `MsgID < 1000`：**原样转发**（含 LoginReq、Ping）
- S→C：独立 goroutine `RelayAndParse`，始终转发

**关键偏差**：读循环**不阻塞**；被拦截帧从客户端 socket 消费掉但未到达服务端，与设计文档「阻塞直到用户放行」描述不一致。

### 3.2 前端放行路径（`packet-tab.vue` → `handleRetryMessage`）

- 拦截消息点「重发」→ 调用 `ReplayControlService.sendMessages([targetMsg], ...)`
- `SendMessages`：**新开 TCP 连接** + HTTP AuthLogin + LoginReq + 用 `replayOpenID`（如 test1）发送
- **未使用** 后端推送的 `conn_id`；**未写入** 客户端代理连接的 `serverConn`

因此「重发」在拦截模式下本质是**离线重放单条**，不是**代理内放行**。

### 3.3 前端拦截队列

- `interceptedSeqIDs: Set<number>` 仅做 UI 橙色标记
- 每条 `record:intercepted` 触发 `message.info` toast + `selectedIndex = 最新消息`（易盖住前几条）
- 后端无 pending 队列

### 3.4 已知 Bug：`RecordLoginPayload` 未生效

`interceptAndParse` 中 `RecordLoginPayload` 写在 `frame.MsgID >= 1000` 分支内，LoginReq（MsgID=1）**永远不会被录制**。实测 `record_20260610_111702.json` 无 `login_payload_b64` 字段。

普通录制路径 `decoder.go` 存在相同逻辑错误。

---

## 4. 目标产品设计（用户确认，待实现）

### 4.1 核心流程

```
积累阶段：C→S 可读消息 → 入 pending、不转发 → 推前端追加表格 → 用户任意编辑
放行阶段：用户点「重发」→ 快照 pending → 按 SeqID 排序 → 同一代理连接写入 serverConn
响应阶段：S→C Ack/Ntf 原样转发 → 客户端可能再发 Req → 重新积累
```

### 4.2 已确认规则

| # | 规则 | 结论 |
|---|------|------|
| 批次边界 | 点击重发瞬间快照队列；执行期间新到的消息 | **下一批** |
| 放行顺序 | 按 `seq_id` 升序 | 是 |
| 编辑内容 | 有编辑用编辑后 payload，无编辑用原始 | 是 |
| 录制落盘 | 以**实际发送到服务端**的内容为准 | 是 |
| 客户端并行 | 程序不干涉客户端发 Req 顺序 | 不管 |
| Ping | **P0**：立即透传，不入 pending、不展示、不参与批量放行 | 已定 |
| LoginReq | **不拦截**，立即转发；可选 `RecordLoginPayload` 写入录制文件 | 已定 |

### 4.5 产品规则确认（2026-06-10 第二轮）

| # | 议题 | 结论 |
|---|------|------|
| 1 | `filePath` / 录制落盘路径 | **不主动改造**。原逻辑不影响启动与改包则保持现状；若实现中发现影响改包正常执行，仅用时间戳生成新文件名做最小修复 |
| 2 | 放行按钮与重发/迭代 UI | 实时改包模式下：**单独新增按钮**（批量放行，文案待定为「放行全部」类）；原「重发」、重复 N 次、迭代发送在改包状态下**不可见** |
| 3 | 「开始重放」 | 实时改包模式下 **禁用** |
| 4 | 编辑器「应用」 | 改包状态下 **不可见**（编辑仅内存，放行时一并提交） |
| 5 | 停录 / 关闭实时修改 | **已修改的按编辑内容发送服务端**，之后连接恢复 **透传**（不再拦截入队） |
| 6 | 多连接 `conn_id` | 本次不调查；见 `docs/TODO.md` 单独待办，另开会话处理 |
| 7 | 批量放行失败 | **fail-fast**（一条失败则中止后续） |
| 8 | pending 为空时点放行 | **不特殊处理**（空队列遍历即可） |

### 4.3 消息分流（定稿草案）

```
C→S + filterMode=true:
  MsgID == 1 (LoginReq)  → 立即转发 + RecordLoginPayload（可选，仅写文件）
  MsgID == 3 (Ping)      → 立即转发（P0）
  MsgID >= 1000          → 入 pending + emit record:intercepted
  其他 < 1000            → 立即转发 + 日志

S→C：全部立即转发（不变）
```

### 4.4 MsgID < 1000 说明

| MsgID | 名称 | 方向 | 处理 |
|-------|------|------|------|
| 1 | LoginReq | C→S | 透传，不拦截 |
| 2 | LoginResp | S→C | 透传 |
| 3 | Ping | C→S | P0 透传 |
| 4 | Pong | S→C | 透传 |
| 10 | KickOut | S→C | 透传 |

`HelloReq`(1001)、`CreateRoleReq`(1006) 等 **≥1000 的 Proto** 走 pending，与框架 Ping 无关。

LoginReq 字段（ByteStream）：Account、Token、UID、Version、Metadata、ExtData、ReqType、SeqID、Entity、Sign。实时改包场景**不需要拦截**；离线重放换 Token 等见 `sendLoginReq`。

---

## 5. 录制文件路径（`filePath`）调查

### 5.1 来源

非拦截功能新增；首见于 commit `83725bb`（`packet-tab.vue`）：

```typescript
if (!filePath.value) {
  filePath.value = `record_${timestamp}.json`
}
await recordControlService.startRecord(filePath.value, ...)
```

页面上**无路径输入框**；`2026-06-02` spec 曾计划 `Dialogs.OpenFile`/`SaveFile`，**未实现**。

### 5.2 传递链

`packet-tab.vue` → `StartRecord` → `RecordWorker.Start` → `NewRecorder(filePath)` → `os.WriteFile(r.filename)`

相对路径 → 应用工作目录（`wails3 dev` 下为 **worktree 根目录**），非 `cases/proto_cases/`。

### 5.3 问题清单

| 问题 | 说明 |
|------|------|
| 第二次录制复用路径 | `filePath` 仅首次生成，停录后不清空 → 覆盖同一文件 |
| `recordData` 跨会话污染 | 仅 `!recordData.value` 时初始化，第二次在旧表格上追加 |
| 录制中「应用」可能失败 | `updateMessagePayload` 读盘，但 `Recorder.Save()` 尚未执行 |
| UI 不展示路径 | 用户只能从日志看到 `保存到=record_xxx.json` |
| E2E 文档路径错误 | `e2e/CLAUDE.md` 写堆积在 `cases/proto_cases/`，实际在根目录 |

### 5.4 实施策略（§4.5 #1 已定）

- **默认**：保持现有 `record_{ts}.json` 相对路径与落盘时机，不为此单独做 UI/目录重构。
- **例外**：若联调发现 `filePath` 复用、录制中写盘等与改包冲突，仅做最小修复（如每次开始录制强制新时间戳文件名）。
- 批量放行**不依赖** `filePath`；改包编辑不经过「应用」写盘。

---

## 6. 推荐实现方案（供下一会话）

### 6.1 后端

1. **`connInterceptState`**：per-connection `pending []*pendingFrame`，保存原始 `header+body`、seq_id、msg_id、payloadJSON
2. **改造 `interceptAndParse`**：白名单透传（LoginReq、Ping）+ Proto 入队
3. **新增 API**：`ReleasePendingMessages(connID, edits map[uint32]json.RawMessage)`
   - 快照 pending → 按 seq_id 排序 → `serverConn.Write` → 清空已放行项
   - 录制以实际发送内容更新
4. **修复** `RecordLoginPayload` 到 LoginReq 透传分支

### 6.2 前端

1. **新增**改包专用按钮（批量放行），传 `conn_id` + pending 快照 + edits map
2. `filterMode` 下隐藏：原「重发」、重复 N 次、迭代发送、「应用」
3. `filterMode` 下禁用：「开始重放」
4. 待放行计数 UI；合并 toast（如「新增 5 条待放行」）
5. 拦截时**不强制** `selectedIndex = last`
6. 停录/关改包：触发「已修改内容发送 + 后续透传」（§4.5 #5）

### 6.3 不在本次范围

- 恢复完整路径浏览/另存为 UI（`2026-06-02` spec 遗留）
- Ping P1（透传+展示）

---

## 7. 相关代码索引

| 模块 | 文件 |
|------|------|
| 拦截读循环 | `record_worker.go` → `interceptAndParse` |
| 录制落盘 | `streamproto/record.go` → `Save`（纯数据类型在 `cases/record.go`） |
| 错误放行 | `packet-tab.vue` → `handleRetryMessage` |
| 新开连接重放 | `streamproto/record.go` → `SendMessages`（重放编排逻辑） |
| 自动生成路径 | `packet-tab.vue` → `handleStartRecord` |
| 原设计规格 | `docs/superpowers/specs/2026-06-09-stream-proxy-intercept-filter-mode.md` |

---

## 8. 验收要点（实现后）

1. 登录后多条 Req 入表格，toast 不刷屏
2. 改包专用按钮：同代理连接按 seq_id 批量发出，非 `SendMessages` 新连接
3. Ping 透传；改包模式下无「重发/N 次/迭代/应用」，「开始重放」禁用
4. `login_payload_b64` 修复后可在录制 JSON 中出现
5. 停录/关改包：已编辑内容发送后恢复透传
6. 放行失败 fail-fast；pending 为空时放行为空操作
7. 放行后 Ack 到达客户端，后续 Req 进入下一批 pending

## 9. 讨论收口说明

实时改包相关的架构、协议分流、UI 行为与异常策略均已确认（§4.2、§4.5）。**未收口项**：多连接 `conn_id` 策略 → 见 `docs/TODO.md` 独立待办，另开会话调查实现。
