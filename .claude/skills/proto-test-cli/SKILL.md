---
name: proto-test-cli
description: |
  proto-test 协议测试 CLI 使用指南 — 用例查看、消息重放、GM 命令发送的命令行操作。

  **触发条件**（满足任一）：
  - 用户要求重放/发送协议消息（如"给账号发道具"、"重放用例"、"发 GM 命令"）
  - 用户要求查看协议测试用例（如"有哪些用例"、"看用例里的消息"）
  - 用户提到 proto-test、协议测试、协议重放、消息发送
  - 用户要求批量给多个账号发送相同的游戏操作
  - 用户要求查看服务器对某条消息的返回（Ack/Ntf）

  **不要触发**：
  - rain-qa-func的功能开发、bug修复
  - 用户要求录制新协议（录制依赖 GUI，不在 CLI 范围）
  - 用户要求实时拦截/改包（依赖 GUI 交互）
  - 用户操作 rain-robot 测试机器人（那是另一个项目）

  **判定关键**：涉及 `rain-qa-func proto-test` 子命令的命令行操作时使用。
---

# proto-test 协议测试 CLI 使用指南

> `rain-qa-func proto-test` CLI 提供协议测试核心能力：用例查看、消息重放、单条消息发送。适用于 AI Agent 自动化测试和批量操作场景。

## 可执行文件位置

本 skill 自带预编译的二进制，无需重新编译。根据运行平台选择对应的二进制：

| 平台 | 二进制路径 |
|------|-----------|
| Windows | `.claude/skills/proto-test-cli/bin/rain-qa-func.exe` |
| macOS (Apple Silicon) | `.claude/skills/proto-test-cli/bin/rain-qa-func-darwin-arm64` |
| macOS (Intel) | `.claude/skills/proto-test-cli/bin/rain-qa-func-darwin-amd64` |

所有命令示例中的 `rain-qa-func` 应替换为当前平台的二进制路径（或将其所在目录加入 PATH）。

> **注意**：所有平台均为纯 CLI 构建（入口 `cmd/rain-qa-func-cli/`，不含 GUI/Wails 依赖），proto-test CLI 子命令功能与完整应用一致。二进制经 UPX 压缩，单个约 3.5–4 MB。

### 重新编译

如需从源码重新构建 CLI 二进制（例如更新 proto-test 功能后）：

```bash
# Windows CLI
go build -o rain-qa-func.exe ./cmd/rain-qa-func-cli

# macOS CLI (Intel)
GOOS=darwin GOARCH=amd64 go build -o rain-qa-func-darwin-amd64 ./cmd/rain-qa-func-cli

# macOS CLI (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o rain-qa-func-darwin-arm64 ./cmd/rain-qa-func-cli
```

这些命令均基于 `cmd/rain-qa-func-cli/main.go`，不依赖 cgo，可在任意平台交叉编译。

**用例文件自动定位**：CLI 会按以下顺序查找用例目录，无需手动指定：
1. 若**当前工作目录**下存在 `cases/proto_cases/`（即在 rain-qa-func 项目目录中），使用项目的用例
2. 否则使用 skill 自带的用例（二进制同级的 `cases/` 目录）

因此从任意目录运行都能正常列出和重放用例。在项目目录中运行时使用最新项目用例；在项目外运行时使用 skill 打包时快照的用例。

## 命令总览

```
rain-qa-func proto-test
├── case list                # 列出所有用例
├── case show <name>         # 查看用例详情
├── replay <case-name>       # 重放用例（核心命令）
├── unity-log                # 解析 Unity NetRecorder 日志
│   ├── list [log-dir]       # 列出日志文件
│   ├── show <log-file>      # 显示日志摘要
│   └── replay <log-file>    # 重放日志中的 C→S 消息
├── send-msg                 # 发送单条自定义消息
├── case edit <name>         # 修改用例（待实现）
├── case create <name>       # 创建用例（待实现）
└── case delete <name>       # 删除用例（待实现）
```

**录制和实时拦截改包依赖 GUI 交互，不在 CLI 范围内。**

---

## 查看用例

### 列出所有用例

```bash
rain-qa-func proto-test case list
rain-qa-func proto-test case list --format json   # JSON 数组输出
```

### 查看用例详情

```bash
rain-qa-func proto-test case show <name>                  # summary 格式（默认）
rain-qa-func proto-test case show <name> --format json    # 完整 JSON（含 payload）
```

- **summary**（默认）：仅输出消息序号、名称、描述。token 消耗小，适合先看结构。
- **json**：输出完整 payload，适合需要看具体字段值时使用。

**关键约定：所有序号都是 1-based**。`case show` 输出的 `seq` 与 `replay --select` 的序号完全一致。

---

## 重放用例（核心命令）

```bash
rain-qa-func proto-test replay <case-name> [flags]
```

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--server` | 用例文件中的 server_addr | 目标 TCP 服务器 |
| `--openid` | （必填） | 账号前缀（如 `test`，实际账号为 test1、test2...） |
| `--range` | `1-1` | 账号范围（如 `1-10` 批量 10 个账号） |
| `--repeat` | `1` | 整个用例重复轮数（每轮重发所有选中消息）。配合 `--select` 选单条时等价于前端"重发 N 次" |
| `--select` | （空） | 按 1-based 序号选择消息子集，如 `3-7`、`3,5,9`、`1-3,7` |
| `--select-name` | （空） | 按消息名筛选（可多次指定），如 `GmCommandReq` |
| `--print-ack` | `false` | 输出服务端返回的每条消息（Ack/Ntf）为 NDJSON |
| `--concurrency` | `0` | 最大并发账号数（0=不限） |
| `--interval` | `1000` | 消息发送间隔（毫秒） |
| `--timeout` | `5m` | 整体超时 |

**`--select` 与 `--select-name` 取交集**（消息必须同时满足序号和名称条件）。不传任何 select flag 时默认重放全部消息。

### --print-ack 输出格式

开启 `--print-ack` 后，服务端返回的每条消息（方向为服务端→客户端）会以 NDJSON 输出到 stdout，每行一个 JSON：

```json
{"account":"test1","msg_name":"GmCommandAck","msg_id":1002,"seq_id":38,"payload":{}}
{"account":"test1","msg_name":"UpdateItemListNtf","msg_id":1103,"seq_id":37,"payload":{"itemList":[{"amount":10,"itemId":1000016}]}}
```

自己发出的消息（客户端→服务端）不会输出。多账号并发时输出用互斥锁保护，不会交错。

---

## 解析 Unity 客户端日志并重放

```bash
rain-qa-func proto-test unity-log [子命令]
```

解析 Unity 客户端 NetRecorder 输出的 `.log` 文件，提供日志浏览、摘要查看和消息重放能力。

```
unity-log
├── list [log-dir]      # 列出日志文件
├── show <log-file>     # 显示日志摘要
└── replay <log-file>   # 重放日志中的 C→S 消息
```

### unity-log list — 列出日志文件

```bash
rain-qa-func proto-test unity-log list [log-dir]
```

- 默认日志目录：`D:\work\client\Master\Card\Log\net`
- 递归列出所有 `.log` 文件，显示相对路径和文件大小

### unity-log show — 显示日志摘要

```bash
rain-qa-func proto-test unity-log show <log-file> [--format summary|json]
```

- `summary`（默认）：输出总记录数、最后一次登录位置、消息方向/MsgID/MsgName/SeqID 表格
- `json`：输出完整记录数组（含 payload）

### unity-log replay — 重放日志中的 C→S 消息

```bash
rain-qa-func proto-test unity-log replay <log-file> [flags]
```

解析 `.log` 文件，提取最后一次登录完成后的所有客户端→服务端游戏协议消息，在内存中构建为测试用例并重放。

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--server` | （必填） | 目标 TCP 服务器（客户端日志不记录服务器地址） |
| `--openid` | （必填） | 账号前缀（如 `test`，实际账号为 test1、test2...） |
| `--range` | `1-1` | 账号范围（如 `1-10` 批量 10 个账号） |
| `--repeat` | `1` | 整个用例重复轮数 |
| `--select` | （空） | 按 1-based 序号选择消息子集 |
| `--select-name` | （空） | 按消息名筛选 |
| `--print-ack` | `false` | 输出服务端返回的每条消息为 NDJSON |
| `--concurrency` | `0` | 最大并发账号数（0=不限） |
| `--interval` | `1000` | 消息发送间隔（毫秒） |
| `--timeout` | `5m` | 整体超时 |

#### 会话边界规则

- 以日志中最后一次 `LoginResp`（`MsgId=2`, `ReceiveSuccess`）作为会话边界。
- 只取该位置之后的 `C→S` 方向游戏协议消息（`MsgID >= 1000`）。
- 登录由 proto-test 自行完成（`AuthLogin` + `sendLoginReq`），不依赖客户端日志中的 `LoginReq`。

#### 示例

```bash
# 列出客户端日志
rain-qa-func proto-test unity-log list

# 查看单条日志摘要
rain-qa-func proto-test unity-log show "D:\work\client\Master\Card\Log\net\2026-06-16\2026-06-16-11-42-44.log"

# 单账号重放最近一次登录后的客户端请求
rain-qa-func proto-test unity-log replay "D:\work\client\Master\Card\Log\net\2026-06-16\2026-06-16-11-42-44.log" \
  --server 10.254.114.204:18000 \
  --openid test

# 批量 10 个账号并发重放
rain-qa-func proto-test unity-log replay "D:\work\client\Master\Card\Log\net\2026-06-16\2026-06-16-11-42-44.log" \
  --server 10.254.114.204:18000 \
  --openid test \
  --range 1-10 \
  --concurrency 5
```

---

## 发送单条消息

```bash
rain-qa-func proto-test send-msg [flags]
```

| Flag | 必填 | 说明 |
|------|------|------|
| `--msg-name` | 是 | 消息名（如 `GmCommandReq`） |
| `--openid` | 是 | 账号前缀（实际账号为 `<openid>1`） |
| `--server` | 是 | 目标 TCP 服务器（无用例文件可读取默认地址） |
| `--payload` | 否 | payload JSON 字符串（缺省 `{}`） |

> `send-msg` 的 `--server` 是必填，与 replay 不同（replay 可从用例文件读取默认地址）。

### GM 命令速查

GM 命令通过 `send-msg` 发送，`msg-name` 固定为 `GmCommandReq`，payload 的 `content` 字段为 GM 指令文本：

```bash
# 添加物品（黄金=1000001 碎银=1000002 铜钱=1000003 虎符=1000016）
rain-qa-func proto-test send-msg --msg-name GmCommandReq --payload '{"content":"//AddItem 1000016 10"}' --openid test --server 10.254.114.204:18000

# 游戏结束-胜利
rain-qa-func proto-test send-msg --msg-name GmCommandReq --payload '{"content":"//GameOverRoomWin"}' --openid test --server 10.254.114.204:18000
```

**不知道道具 ID 时**：先用 `case show 创号 --format json` 查看已有 GM 命令的 payload，里面包含常用道具 ID。

---

## 常见任务方案

### 给单个账号发道具

```bash
# 方法 1：用 send-msg 直接发
rain-qa-func proto-test send-msg --msg-name GmCommandReq --payload '{"content":"//AddItem 1000016 10"}' --openid test --server 10.254.114.204:18000

# 方法 2：用 replay 重放"创号"用例的第 7 条（添加虎符）
rain-qa-func proto-test case show 创号                  # 先确认序号
rain-qa-func proto-test replay 创号 --select 7 --openid test --print-ack
```

### 批量给多个账号发道具

```bash
# 给 test1-test5 五个账号都发虎符
rain-qa-func proto-test replay 创号 --select 7 --openid test --range 1-5 --concurrency 5 --print-ack
```

### 查看服务器返回

```bash
# 重放并观察服务端返回的 Ack/Ntf
rain-qa-func proto-test replay 创号 --select 7 --openid test --print-ack --timeout 30s

# 重放客户端 NetRecorder 日志并观察返回
rain-qa-func proto-test unity-log replay "D:\work\client\Master\Card\Log\net\2026-06-16\2026-06-16-11-42-44.log" \
  --server 10.254.114.204:18000 --openid test --print-ack --timeout 30s
```

NDJSON 输出可直接用文本工具筛选特定消息（如只看物品更新）：

```bash
rain-qa-func proto-test replay 创号 --select 7 --openid test --print-ack | grep UpdateItemListNtf
```

### 探索一个不熟悉的用例

```bash
# Step 1: 看摘要（小 token 消耗）
rain-qa-func proto-test case show 工会战

# Step 2: 看某条消息的 payload
rain-qa-func proto-test case show 工会战 --format json

# Step 3: 只重放关心的几条消息
rain-qa-func proto-test replay 工会战 --select 9-12 --openid test
```

---

## 行为约束

1. **先 show 再操作**：操作用例前先用 `case show <name>` 确认消息序号。序号是 1-based。
2. **send-msg 必填 --server**：send-msg 无用例文件可读取默认地址，必须显式指定服务器。
3. **--select 与 --select-name 是交集**：两个 flag 同时传时，消息必须同时满足序号和名称条件才会被选中。
4. **edit/create/delete 当前不可用**：这三个子命令还是占位实现。需要修改用例时直接编辑 `cases/proto_cases/<name>.json` 文件。
5. **服务器地址默认从用例文件读取**：replay 不传 `--server` 时用用例文件中的 `server_addr`（通常是 `10.254.114.204:18000`）。
6. **unity-log replay 需要 --server**：客户端 NetRecorder 日志不记录服务器地址，必须显式指定。
7. **unity-log replay 的 --openid 必填**：因为日志只记录消息内容，不记录发送账号。

---

## 与现有工具的关系

| CLI 命令 | 复用的后端能力 |
|----------|--------------|
| `case list/show` | `TestCaseService` + `RecordFileService` |
| `replay` | `msg.SendMessages`（支持账号范围/重复/消息子集/并发） |
| `send-msg` | `msg.SendMessages`（单条单账号） |

详细帮助文档见 `backend/pkg/proto-test/cobra-help.md`。
