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

### 1. 生成 protobuf 代码（需要 buf）

```bash
cd examples/buf
buf generate    # 读取 proto/echo.proto → 生成 proto/go/echopb/echo.pb.go
```

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
