# cmd/tests/streamproxy — 流量代理命令行工具

## 用法

### 命令行参数

```bash
# 录制模式：代理 TCP + HTTP 流量，录制客户端消息到 JSON 文件
streamproxy -tcp :18000:10.254.114.204:18000 -http :20144:10.254.114.204:20144 -record session.json

# 重放模式：从录制文件重放客户端消息（TCP 地址从录制文件 server_addr 字段读取）
streamproxy -replay session.json -replay-openid test1

# TODO: 账号范围迭代重放（CLI 尚未实现）
# streamproxy -replay session.json -replay-openid test -range-start 1 -range-end 3
```

### 启动示例

```
streamproxy -http :20144:10.254.114.204:20144 -tcp :18000:10.254.114.204:18000
```

日志输出 `[MAP] TCP 10.254.114.204:18000 -> 127.0.0.1:18000`，HTTP 响应中自动替换 server_addr。

## 录制 JSON 格式

详见 [cases/proto_cases/CLAUDE.md](../../../cases/proto_cases/CLAUDE.md)。

`login_payload_b64`：录制时保存客户端 LoginReq 的**解密后 payload**（ByteStream 格式）。重放时替换其中的 Token 字段后重新加密发送，确保 LoginReq 内容与真实客户端一致。

## 客户端登录流程与 server_addr 劫持

### 登录流程

1. 客户端从 `ServersTemplate` 配表获取 HTTP 认证地址
2. POST `/authlogin` -> 服务端返回 JSON `LoginRespToClient`
3. 客户端解析 `LoginRespToClient.server_addr`
4. 客户端 TCP 连接 `server_addr` 发送游戏消息

客户端关键代码见 `D:/work/client/Master/Card/Assets/Scripts/HotUpdate/Logic/Net/Common/Common.cs`。

### streamproxy 劫持机制

HTTP 代理在转发 `/authlogin` 响应时，将 `server_addr` 中的真实地址替换为本地代理地址。

- `tcpRedirectMap`（`sync.Map`）：启动 TCP 代理时注册 `远程地址 -> 本地地址` 映射
- 示例：`10.254.114.204:18000` -> `127.0.0.1:18000`

### 服务器配表

- **Excel 源文件**: `D:\work\config\excel\服务器配置表.xlsx`
- **JSON 版本**: `D:\work\client\Master/Card/Assets/Bundles/Config/Table/out_json/ServersTemplate.json`

### HTTP /authlogin 请求格式

```json
{"uid":"test1","platform":13,"sdk":0}
```

字段名是 `uid`（不是 `open_id`）。`sdk:0` = FakeSDK（测试用）。

## 已知问题与踩坑记录

### GOOS 默认为 linux
Bash 工具（Git Bash）中 `go build` 默认 `GOOS=linux`，会生成 ELF 格式 exe。必须显式指定：
```bash
GOOS=windows GOARCH=amd64 go build -o streamproxy.exe .
```

### 同账号重放限制
重放时 `-replay-openid` 必须和录制时的 Account 一致（因为使用录制的 LoginReq payload，只替换 Token）。

### LoginReq payload 大小差异
客户端 LoginReq payload 约 440 字节（含 ExtData 设备信息），而自行序列化的最小 payload 约 78 字节。必须使用录制的完整 payload 才能通过服务端校验。

## 依赖

- `backend/pkg/proto-test/streamproto/` -- 协议解析、录制、重放
- `backend/pkg/proto-test/protoMsg/` -- Protobuf 消息定义

## 参考文件

| 内容 | 路径 |
|------|------|
| 协议定义 | `backend/pkg/proto-test/streamproto/` |
| 编码/解码 | `rain-robot/project/xcard/xcard_net_lib/` |
| MsgID 常量 | `rain-robot/project/xcard/xcard_msg_def/msg_const.go` |
| 设计决策 | `backend/pkg/proto-test/docs/design-decisions.md` |
