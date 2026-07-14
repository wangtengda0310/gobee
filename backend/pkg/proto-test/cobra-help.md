# proto-test — Proto 协议测试工具

## 简介

proto-test 子命令提供协议测试的核心 CLI 能力：用例管理（增删改查）、用例重放、单条消息发送。适用于 AI Agent 自动化测试和批量压测场景。

录制和实时拦截改包依赖 GUI 交互，不在 CLI 范围内。

## 用法

```
rain-qa-func proto-test [子命令] [选项]
```

## 子命令

```
proto-test
├── case                    # 用例管理
│   ├── list                # 列出所有用例
│   ├── show <name>         # 查看用例详情
│   ├── edit <name>         # 修改已有用例
│   ├── create <name>       # 从文件/标准输入创建用例
│   └── delete <name>       # 删除用例
├── replay <case-name>      # 重放用例（核心命令）
├── unity-log               # 解析 Unity NetRecorder 日志
│   ├── list [log-dir]      # 列出日志文件
│   ├── show <log-file>     # 显示日志摘要
│   └── replay <log-file>   # 重放日志中的 C→S 消息
└── send-msg                # 发送单条自定义消息
```

## proto-test case list — 列出所有用例

```bash
rain-qa-func proto-test case list [--format table|json]
```

- `--format table`（默认）：表格输出用例名、消息数、服务器地址、录制时间
- `--format json`：JSON 数组输出

## proto-test case show — 查看用例详情

```bash
rain-qa-func proto-test case show <name> [--format json|summary]
```

- `--format summary`（默认）：仅输出消息序号、名称、描述（不含 payload，减少 token 消耗）
- `--format json`：输出完整 JSON（含 payload）

两种格式都使用 **1-based 序号**（`seq` 字段），与 `replay --select`、`case edit --msg` 的索引完全一致。

json 格式输出示例（`seq` 是 1-based）：

```json
{
  "server_addr": "10.254.114.204:18000",
  "recorded_at": "2026-06-08T17:03:55.985Z",
  "message_count": 7,
  "messages": [
    { "seq": 1, "msg_id": 1006, "msg_name": "CreateRoleReq", "descript": "-", "payload": {} },
    { "seq": 3, "msg_id": 1001, "msg_name": "GmCommandReq", "descript": "添加黄金", "payload": {"content":"//AddItem 1000001 1"} }
  ]
}
```

## proto-test case edit — 修改已有用例（待实现）

> **当前为占位实现**，flag 尚未注册。以下为目标设计，实现后可用。

```bash
# 修改第 N 条消息的 payload 字段（RFC 7396 合并补丁，深度合并）
rain-qa-func proto-test case edit <name> --msg <N> --merge '<json>'

# 整条 payload 替换
rain-qa-func proto-test case edit <name> --msg <N> --payload '<json>'

# 设置/修改消息描述
rain-qa-func proto-test case edit <name> --msg <N> --desc '<text>'

# 追加一条消息
rain-qa-func proto-test case edit <name> --append --msg-id <id> --msg-name <name> --payload '<json>'

# 删除第 N 条消息
rain-qa-func proto-test case edit <name> --remove <N>
```

| Flag | 说明 |
|------|------|
| `--msg <N>` | 1-based 消息索引（与 `case show` 输出序号一致） |
| `--merge '<json>'` | RFC 7396 合并补丁，只覆盖指定字段，保留其余（嵌套对象深度合并） |
| `--payload '<json>'` | 整条 payload 替换 |
| `--desc '<text>'` | 设置消息描述 |
| `--append` | 追加新消息（配合 `--msg-id`/`--msg-name`/`--payload`） |
| `--remove <N>` | 删除指定序号的消息 |

## proto-test case create — 创建用例（待实现）

> **当前为占位实现**。以下为目标设计。

```bash
rain-qa-func proto-test case create <name> --file <path.json>
rain-qa-func proto-test case create <name> --stdin  # 从标准输入读取
```

## proto-test case delete — 删除用例（待实现）

> **当前为占位实现**。以下为目标设计。

```bash
rain-qa-func proto-test case delete <name> [--force]
```

`--force` 跳过确认提示。AI Agent 使用前应向用户确认。

## proto-test replay — 重放用例（核心命令）

```bash
rain-qa-func proto-test replay <case-name> [flags]
```

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--server` | string | 用例文件中的 `server_addr` | 目标 TCP 服务器 |
| `--http` | string | 从 `--server` 推导（同IP:20144） | HTTP 认证地址 |
| `--openid` | string | （必填） | 账号前缀（如 `test`） |
| `--range` | string | `1-1` | 账号范围（如 `1-10` 批量 10 个账号） |
| `--repeat` | int | `1` | 整个用例重复轮数（每轮重发所有选中消息） |
| `--interval` | int | `1000` | 消息发送间隔（毫秒） |
| `--ack-wait` | int | `2000` | 等待 Ack 时间（毫秒） |
| `--concurrency` | int | `0` | 最大并发账号数（0=使用默认上限 10） |
| `--max-retries` | int | `3` | 因登录限流（code=-600）失败时的最大重试次数 |
| `--retry-interval` | int | `500` | 重试间隔（毫秒） |
| `--select` | string | （空） | 按 1-based 序号选择消息子集，如 `3-7`、`3,5,9`、`1-3,7` |
| `--select-name` | string | （空） | 按消息名筛选（可多次指定），如 `GmCommandReq` |
| `--print-ack` | bool | `false` | 输出服务端返回的每条消息（Ack/Ntf）为 NDJSON |
| `--timeout` | duration | `5m` | 整体超时 |

`--select` 和 `--select-name` 可组合使用，取**交集**（消息必须同时满足序号和名称条件）。不传任何 select flag 时默认重放全部消息。

## proto-test send-msg — 发送单条自定义消息

```bash
rain-qa-func proto-test send-msg [flags]
```

| Flag | 必填 | 说明 |
|------|------|------|
| `--msg-name` | 是 | 消息名（如 `GmCommandReq`） |
| `--openid` | 是 | 账号前缀（实际账号为 `<openid>1`） |
| `--server` | 是 | 目标 TCP 服务器地址（如 `10.254.114.204:18000`） |
| `--msg-id` | 否 | 消息 ID（如 `1001`） |
| `--payload` | 否 | payload JSON 字符串（缺省 `{}`） |
| `--payload-file` | 否 | 从文件读取 payload（避免 shell 转义） |
| `--http` | 否 | HTTP 认证地址（默认从 `--server` 推导，同 IP:20144） |
| `--interval` | 否 | 消息发送间隔毫秒（默认 1000） |
| `--ack-wait` | 否 | 发送后等待 Ack 毫秒（默认 2000） |
| `--timeout` | 否 | 整体超时（默认 5m） |

> `--server` 在 send-msg 中是必填（无用例文件可读取默认地址），与 replay 不同。

### GM 命令速查（通过 send-msg 发送）

```bash
# 添加物品（1=黄金 2=碎银 3=铜钱）
rain-qa-func proto-test send-msg --msg-id 1001 --msg-name GmCommandReq --payload '{"content":"//AddItem 1000001 999"}' --openid test1

# 游戏结束-胜利
rain-qa-func proto-test send-msg --msg-id 1001 --msg-name GmCommandReq --payload '{"content":"//GameOverRoomWin"}' --openid test1
```

## 示例

```bash
# 列出所有用例
rain-qa-func proto-test case list

# 查看工会战用例摘要（summary 是默认格式）
rain-qa-func proto-test case show 工会战 --format summary

# 查看创号用例的完整 payload（json 格式，seq 是 1-based）
rain-qa-func proto-test case show 创号 --format json

# 重放创号用例（单账号）
rain-qa-func proto-test replay 创号 --openid test --server 10.254.114.204:18000

# 批量重放 10 个账号，每个重复 5 次
rain-qa-func proto-test replay 工会战 --openid test --range 1-10 --repeat 5 --server 10.254.114.204:18000

# 只重放工会战用例的第 9-12 条消息
rain-qa-func proto-test case show 工会战 --format summary  # 先看序号
rain-qa-func proto-test replay 工会战 --openid test --select 9-12 --server 10.254.114.204:18000

# 直接发送单条 GM 命令（不依赖用例文件）
rain-qa-func proto-test send-msg --msg-id 1001 --msg-name GmCommandReq --payload '{"content":"//AddItem 8000003 10"}' --openid test --server 10.254.114.204:18000

# 查看帮助
rain-qa-func proto-test --help
```

> **注意**：`case edit`/`create`/`delete` 当前为占位实现，暂不可用。修改用例请直接编辑 `cases/proto_cases/<name>.json` 文件。

## proto-test unity-log — 解析 Unity NetRecorder 日志

解析 Unity 客户端 NetRecorder 输出的 `.log` 文件，提供日志浏览、摘要查看和消息重放能力。

```
proto-test unity-log
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
- 可通过 `--log-dir` 或位置参数指定其他目录

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

| Flag | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--server` | string | （必填） | 目标 TCP 服务器（客户端日志不记录服务器地址） |
| `--http` | string | 从 `--server` 推导（同 IP:20144） | HTTP 认证地址 |
| `--openid` | string | （必填） | 账号前缀（如 `test`） |
| `--range` | string | `1-1` | 账号范围（如 `1-10` 批量 10 个账号） |
| `--repeat` | int | `1` | 整个用例重复轮数 |
| `--interval` | int | `1000` | 消息发送间隔（毫秒） |
| `--ack-wait` | int | `2000` | 等待 Ack 时间（毫秒） |
| `--concurrency` | int | `0` | 最大并发账号数（0=使用默认上限 10） |
| `--max-retries` | int | `3` | 因登录限流（code=-600）失败时的最大重试次数 |
| `--retry-interval` | int | `500` | 重试间隔（毫秒） |
| `--print-ack` | bool | `false` | 输出服务端返回的每条消息为 NDJSON |
| `--timeout` | duration | `5m` | 整体超时 |

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

## 更多信息

详见 `backend/pkg/proto-test/` 目录及 `CLAUDE.md`。
