# Buf Demo：客户端 ↔ 服务端通信 + pcap 抓包

一个完整的端到端示例，展示如何用 **Buf** 工具链管理 proto 定义、生成 Go 代码，
并用 **pcap 包**实时抓取和结构化解析通信协议。

## 目录结构

```
examples/buf/
├── proto/
│   ├── echo.proto              # Echo 服务的 proto 定义
│   └── go/echopb/echo.pb.go    # buf generate 生成的 Go 代码
├── frame/frame.go              # 共享的帧编解码（4字节包头 + protobuf）
├── server/main.go              # TCP 服务端（Echo + Sum）
├── client/main.go              # TCP 客户端（发送请求 + 打印响应）
├── sniffer/main.go             # pcap 抓包工具（流重组 + protobuf 解析 → JSON）
├── buf.yaml                    # Buf 配置（lint）
├── buf.gen.yaml                # Buf 代码生成配置
└── go.mod                      # 独立的 Go 模块
```

## 快速开始

### 前置条件

- **Go 1.25+**
- **Buf** CLI（`go install github.com/bufbuild/buf/cmd/buf@latest`）
- **Npcap + gcc**（仅 sniffer 需要，server/client 不需要）

### 1. 生成 protobuf 代码

```bash
cd examples/buf
buf generate    # 读取 proto/echo.proto → 生成 proto/go/echopb/echo.pb.go
```

> 生成的 `echo.pb.go` 已提交到仓库，如果你不修改 proto 文件可以跳过此步。

### 2. 启动服务端

```bash
go run ./examples/buf/server -port 19090
```

### 3. 启动客户端

```bash
go run ./examples/buf/client -addr localhost:19090
```

客户端输出：
```
→ EchoRequest: message="hello from client #1" seq=1
← EchoResponse: message="hello from client #1" seq=1 server_time=1789...
→ EchoRequest: message="hello from client #2" seq=2
← EchoResponse: message="hello from client #2" seq=2 server_time=1789...
→ SumRequest: numbers=[10 20 30 40 50]
← SumResponse: sum=150 count=5
```

### 4. 启动抓包工具（需要 Npcap + livecapture）

```bash
CGO_ENABLED=1 go run -tags livecapture ./examples/buf/sniffer \
  -iface <网卡名> -port 19090
```

然后重新运行 client，sniffer 输出：
```
=== 新 TCP 流 [127.0.0.1:54321->127.0.0.1:19090] (客户端→服务端) ===
  [客户端→服务端] EchoRequest: {
    "message": "hello from client #1",
    "seq": 1
  }
  [客户端→服务端] EchoRequest: {
    "message": "hello from client #2",
    "seq": 2
  }
  [客户端→服务端] SumRequest: {
    "numbers": [10, 20, 30, 40, 50]
  }

=== 新 TCP 流 [127.0.0.1:19090->127.0.0.1:54321] (服务端→客户端) ===
  [服务端→客户端] EchoResponse: {
    "message": "hello from client #1",
    "seq": 1,
    "server_time": 1789...
  }
  [服务端→客户端] SumResponse: {
    "sum": 150,
    "count": 5
  }
```

## 手工运行 Guide（完整 step-by-step）

以下假设你在 `pcap/` 目录下操作，使用 3 个终端窗口。

### 步骤 1：生成 protobuf 代码（首次运行或修改 proto 后）

```bash
cd pcap/examples/buf
buf generate
# → 生成 proto/go/echopb/echo.pb.go
```

如果提示 `buf: command not found`：
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
export PATH="$HOME/go/bin:$PATH"
buf --version  # 确认安装成功
```

### 步骤 2：编译所有程序

```bash
cd pcap/examples/buf

# server / client / frame（纯 Go，无需 cgo）
CGO_ENABLED=0 go build ./server/... ./client/... ./frame/...

# sniffer（需要 cgo + Npcap）
CGO_ENABLED=1 go build -tags livecapture ./sniffer/...
```

### 步骤 3：启动服务端（终端 1）

```bash
cd pcap/examples/buf
go run ./server -port 19090
```

输出：
```
2026/07/23 10:00:00 Echo server listening on :19090
```

### 步骤 4（可选）：启动抓包工具（终端 2）

> 需要在 client 连接之前启动 sniffer，否则 TCP 流重组无法跟踪序列号。

```bash
cd pcap
CGO_ENABLED=1 go run -tags livecapture ./examples/buf/sniffer \
  -iface "\Device\NPF_{...}" -port 19090
```

> 如果在 localhost 上测试，网卡用 loopback 设备（Windows 的 `\Device\NPF_Loopback`）。
> `-iface` 的值通过 `go run -tags livecapture ./cmd/pcaptest -list` 查看。

输出：
```
开始抓包（端口 19090），Ctrl+C 退出...
```

### 步骤 5：启动客户端（终端 3）

```bash
cd pcap/examples/buf
go run ./client -addr localhost:19090
```

### 观察输出

**客户端终端**：
```
→ EchoRequest: message="hello from client #1" seq=1
← EchoResponse: message="hello from client #1" seq=1 server_time=1789...
→ EchoRequest: message="hello from client #2" seq=2
← EchoResponse: message="hello from client #2" seq=2 server_time=1789...
→ SumRequest: numbers=[10 20 30 40 50]
← SumResponse: sum=150 count=5
```

**服务端终端**：
```
[127.0.0.1:54321] connected
[127.0.0.1:54321] EchoRequest: message="hello from client #1" seq=1
[127.0.0.1:54321] EchoRequest: message="hello from client #2" seq=2
[127.0.0.1:54321] SumRequest: [10 20 30 40 50] → sum=150
[127.0.0.1:54321] disconnected
```

**抓包终端**（JSON 格式的消息内容）：
```
=== 新 TCP 流 [127.0.0.1:54321->127.0.0.1:19090] (客户端→服务端) ===
  [客户端→服务端] EchoRequest: {
    "message": "hello from client #1",
    "seq": 1
  }
  [客户端→服务端] EchoRequest: {
    "message": "hello from client #2",
    "seq": 2
  }

=== 新 TCP 流 [127.0.0.1:19090->127.0.0.1:54321] (服务端→客户端) ===
  [服务端→客户端] EchoResponse: {
    "message": "hello from client #1",
    "seq": 1,
    "server_time": 1789...
  }
  [服务端→客户端] SumResponse: {
    "sum": 150,
    "count": 5
  }
```

### 步骤 6：停止

- client 运行完自动退出
- server / sniffer 按 `Ctrl+C` 停止

## 设计要点

### Buf vs 传统 protoc

| 传统 protoc | Buf |
|---|---|
| 手动安装 protoc + 插件 | `buf generate` 一条命令 |
| 版本管理靠人 | buf.lock 锁定远程插件版本 |
| 无 lint | `buf lint` 自动检查 proto 规范 |
| 无 breaking change 检测 | `buf breaking` 对比 git 历史 |

### 帧协议

```
┌─── 4 字节包头 ───┐ ┌────── payload（变长）──────┐
│ msgType(2B LE)  │ │ msgLen(2B LE) │ protobuf... │
└─────────────────┘ └───────────────┴─────────────┘
```

刻意简单（无加密/压缩），聚焦展示 Buf + pcap 的完整链路。

### pcap 的角色

pcap 包负责「抓包 + TCP 流重组」，不绑定任何具体协议。
sniffer 提供「帧切分 + protobuf 反序列化」——这正是 pcap 包
「协议无关基础设施」设计哲学的体现。
