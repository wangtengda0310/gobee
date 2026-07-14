# cmd/tests — 测试与调试 CLI 工具集

测试和调试用的命令行工具集合。每个子目录包含一个独立的 `main.go`，可直接 `go run` 执行。

## 工具列表

| 工具 | 用途 | 用法 |
|------|------|------|
| **auth_check** | HTTP 认证诊断，对比多个 UID 的 AuthLogin 响应（open_id、token、server_addr） | `go run ./cmd/tests/auth_check` |
| **create_role_test** | 创建角色流程验证，使用真实 proto 库构造 CreateRoleReq 并检查 Ack | `go run ./cmd/tests/create_role_test [openID]` |
| **diag_server** | 服务器地址分配诊断，对比 AuthLogin 分配地址与前端配置地址的差异 | `go run ./cmd/tests/diag_server` |
| **full_test** | 完整登录创角流程验证（登录→创角→断开→重新登录→重复创角），验证 UID 一致性和 ErrCode | `go run ./cmd/tests/full_test` |
| **login_debug** | LoginReq payload 对比工具，输出字段级分析和 hex dump，用于排查登录协议差异 | `go run ./cmd/tests/login_debug [openID]` |
| **proto_test** | 使用真实 proto 库构造消息 + ByteStream 包装的创角验证，验证重登录 UID 一致性 | `go run ./cmd/tests/proto_test [openID]` |
| **replay_case** | 从 JSON 录制文件重放消息，使用 streamproto 相同的发送逻辑 | `go run ./cmd/tests/replay_case [openID]` |
| **replay_guild_war** | 工会城战用例重放，解析 TransportRawNtf 中的 PveGuildCityDataNtf 动态改写 cityId | `go run ./cmd/tests/replay_guild_war -openid test1 [-case 工会战.json] [-repeat 100]` |
| **sim_wails3** | 模拟 wails3 前端行为（使用固定 serverAddr 而非 AuthLogin 分配地址），批量账号登录重放 | `go run ./cmd/tests/sim_wails3` |
| **streamproxy** | 流量代理工具，支持录制/重放模式，详见 [CLAUDE.md](streamproxy/CLAUDE.md) | 见单独文档 |
| **uid_diag** | 同一账号两次登录对比 UID 一致性（验证 accountid 映射） | `go run ./cmd/tests/uid_diag [openID]` |
| **uid_test** | 最小化 UID 诊断，不做假设只输出两次登录的 UID 对比结果 | `go run ./cmd/tests/uid_test` |
| **verify_flow** | 多用例顺序重放验证（如：先创号→再添加黄金），验证完整流程 | `go run ./cmd/tests/verify_flow [openID]` |

## 依赖关系

```
cmd/tests/
├── 依赖 pkg/proto-test/msg/ (协议编解码)
│   ├── create_role_test
│   ├── proto_test
│   ├── replay_case
│   ├── replay_guild_war (额外依赖 pkg/proto-test/ 和 rain-robot xcard_pb)
│   ├── sim_wails3
│   ├── streamproxy
│   └── verify_flow
├── 依赖 pkg/proto-test/ (重放编排)
│   └── replay_guild_war
├── 自包含（内嵌网络层代码，不依赖 pkg）
│   ├── auth_check
│   ├── diag_server
│   ├── full_test
│   ├── login_debug
│   ├── uid_diag
│   └── uid_test
└── streamproxy (详见 streamproxy/CLAUDE.md)
```

## 通用模式

所有工具共享以下模式：

- **HTTP 认证**：POST `/authlogin` 获取 token 和 server_addr
- **TCP 登录**：构造 ByteStream 格式的 LoginReq payload，XOR 加密后发送
- **默认服务器**：`10.254.114.204:20144`（HTTP）、`10.254.114.204:18000`（TCP）
- **测试账号**：使用 FakeSDK（`platform:13, sdk:0`），UID 格式如 `test1`、`ut_new_xxx`

## 与生产代码的关系

这些工具是调试和验证用的辅助程序，不参与正式构建（`wails3 task build` 不包含它们）。它们用于：

- 验证 `pkg/proto-test/msg/` 的编解码逻辑是否与服务端兼容
- 诊断登录、创角、重放流程中的协议问题
- 作为新功能开发时的原型验证工具
