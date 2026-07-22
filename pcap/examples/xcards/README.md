# xcards：Unity 游戏协议抓包解析示例

用 pcap 包实时抓取 go-service 游戏服务器的 TCP 通信，
自动完成 **TCP 流重组 → 消息切分 → 解密 → protobuf 解析**，输出可读的结构化消息。

## 目录结构

```
examples/xcards/
├── gameproto/layer.go    # 自定义 Layer + FrameReader + 加解密 + protobuf 解析
└── main.go               # 抓包程序（流重组 / 逐包 两种模式）
```

## 前置条件

1. **Npcap** 已安装（勾选 WinPcap 兼容模式）
2. **MSYS2 gcc** 已安装且在 PATH
3. **环境变量**已配置（详见 pcap/README.md「安装」章节）：
   ```
   CGO_CFLAGS=-I D:\npcap-1.88\wpcap\libpcap
   CGO_LDFLAGS=-l wpcap
   PATH 含 C:\msys64\mingw64\bin
   ```

## 构建

```bash
cd pcap
CGO_ENABLED=1 go build -tags livecapture -o xcards.exe ./examples/xcards/
```

## 手工运行 Guide

### 步骤 1：查看可用网卡

```bash
./pcaptest -list
```

找到你的物理网卡，如 `\Device\NPF_{FDEA18E5-...}`。

### 步骤 2：确认游戏服务器端口

| 端口 | 用途 |
|---|---|
| 20144 | HTTP 登录（POST /authlogin） |
| 18000 | 游戏二进制协议（LoginReq/Ping/Pong/GameMsg） |

xcards 抓的是 **18000 端口**的游戏协议。

### 步骤 3：选择运行模式

#### 模式 A：流重组模式（默认，完整重组）

**必须在 Unity 连接服务器之前启动 xcards**，否则 tcpassembly 无法跟踪序列号。

```bash
# 终端 1：启动抓包
xcards.exe -iface "\Device\NPF_{FDEA18E5-...}" -port 18000

# 终端 2：启动 Unity 并登录
```

输出示例：
```
=== 新 TCP 流 [192.168.20.15:54315->10.254.114.204:18000] (客户端→服务端) ===
  [客户端→服务端] LoginReq ID=1 Seq=0 BodyLen=463
    Account: "test1"
    Token: "ddd54050afea97ac..."
    Version: "0.5.3013.1.0"
  [服务端→客户端] LoginResp ID=2 Seq=1 BodyLen=97
    UID: 4294967952
    Result: 0 (成功)
  [服务端→客户端] GameMsg(1016) ID=1016 Seq=2 BodyLen=3581
    field 1 (bytes): { ... }
    field 2 (varint): 16
```

#### 模式 B：逐包模式（`-raw`，支持先连接后抓包）

**Unity 可以先登录，之后再启动 xcards**——适合临时检查。

```bash
# 先启动 Unity 并登录
# 然后启动抓包
xcards.exe -iface "\Device\NPF_{FDEA18E5-...}" -port 18000 -raw
```

输出示例（心跳为主）：
```
  [客户端→服务端] Ping ID=3 Seq=125 BodyLen=0
  [服务端→客户端] Pong ID=4 Seq=164 BodyLen=8
          (timestamp: 62436f9d9f010000)
```

### 步骤 4：在游戏里操作

登录后在 Unity 里移动角色 / 打开菜单 / 使用技能，xcards 会实时解析出对应的 GameMsg。

### 步骤 5：停止抓包

按 `Ctrl+C`，会输出统计信息：
```
结束。captured=123
  handler game-stream: received=123 processed=123 dropped=0 errors=0
```

## 两种模式对比

| | 流重组模式（默认） | 逐包模式（`-raw`） |
|---|---|---|
| 先连接后抓包 | ❌ 不行（需要 SYN） | ✅ **可以** |
| 先抓包后连接 | ✅ 完整重组 | ✅ 小消息 OK |
| 跨 TCP 段的大消息 | ✅ 正确重组 | ⚠️ 可能解析不全 |
| 加密消息 | ✅ 自动解密 | ✅ 自动解密 |
| 方向判断 | ✅ | ✅ |
| 适用场景 | 完整会话录制 | 临时检查/心跳监控 |

## 参数说明

| 参数 | 默认值 | 说明 |
|---|---|---|
| `-iface` | （必填） | 网卡设备名 |
| `-port` | 18000 | 游戏服务器端口（0=不过滤端口） |
| `-raw` | false | 逐包模式 |

## 消息格式

详见 `gameproto/layer.go` 的包文档。核心格式：

```
[4字节包头: msgSize(3B LE) + flag(1B)] [msgData: msgID(2B) + seqID(4B) + body]
```

- flag bit1=1 表示已加密（自动异或+循环移位解密）
- flag bit0=1 表示已压缩（自动 snappy 解压）
- msgID < 1000 = 框架消息（LoginReq/Ping/Pong，自定义 ByteStream 格式）
- msgID >= 1000 = 游戏消息（body 是 [2B len][protobuf] 子信封格式）
