# pcap

基于 [gopacket](https://github.com/gopacket/gopacket) 的网络抓包库，提供「数据源 → 抓包器 → 处理函数」的广播式抓包能力。

## 核心特性

- **广播式分发**：一次抓包，多个处理函数并行消费，互不拖累。
- **过载保护**：抓包循环绝不因慢处理函数而阻塞；三种可配置的背压策略。
- **并发安全**：运行期动态增删处理函数，全量原子统计。
- **纯 Go 核心**：离线 pcap 重放与单元测试不依赖 cgo / Npcap，任意平台可运行。
- **实时抓包可选**：基于 cgo 的网卡抓包用构建标签 `livecapture` 隔离。

## 功能对 Npcap/libpcap 的依赖关系

不是所有功能都需要 Npcap。下表帮你判断「我想做 X，要不要装 Npcap」：

| 功能 / API | 是否需要 Npcap/libpcap | 说明 |
|---|---|---|
| `NewCapturer` / `RegisterHandler` / `Capture` / `Stats` / `Close` | ❌ 不需要 | 核心库，纯 Go |
| `NewReaderSource`（离线重放 pcap 文件） | ❌ 不需要 | 基于 `pcapgo`，纯 Go |
| `NewMergedSource`（多网卡 fan-in 合并） | ❌ 不需要 | 纯 Go channel 合并 |
| `NewHTTPRequestHandler` / `NewHTTPResponseHandler`（流重组+HTTP 解析） | ❌ 不需要 | 基于 `tcpassembly` + 标准库 `net/http`，纯 Go |
| `NewTCPStreamHandler`（通用 TCP 流重组，protobuf/自定义协议） | ❌ 不需要 | 基于 `tcpassembly`，使用者提供解析逻辑，纯 Go |
| `PacketEvent` / 所有 `OverflowStrategy` / `Hooks` | ❌ 不需要 | 纯 Go 逻辑 |
| `Target.Host`（用户态 host 过滤） | ❌ 不需要 | 纯 Go 字符串匹配 |
| **单元测试**（`go test ./...`） | ❌ 不需要 | 全部纯 Go，CI 友好 |
| `NewLiveSource`（实时打开网卡抓包） | ✅ **需要** | 调用 `pcap.OpenLive`（cgo） |
| `ListInterfaces`（列出本机网卡） | ✅ **需要** | 调用 `pcap.FindAllDevs`（cgo） |
| `Target.BPF` / `WithBPFFilter`（内核态 BPF 过滤） | ✅ **需要** | 调用 `pcap.SetBPFFilter`，且仅对 `liveSource` 生效 |
| `SetBPFFilter` / `ValidateBPF`（BPF 热重载/校验） | ✅ **需要** | 仅 `liveSource` 实现 |
| `cmd/pcaptest`（实时抓包 CLI） | ✅ **需要** | 依赖 `NewLiveSource` / `ListInterfaces` |
| `go test -race`（并发安全检测） | ⚠️ 需要 **gcc**，但不需要 Npcap | race 依赖 cgo，但不链接 libpcap |

**一句话总结**：只要不调用 `NewLiveSource` / `ListInterfaces` / BPF，就不需要 Npcap——包括所有离线分析、多源合并、单元测试、核心 API。

```bash
go get github.com/wangtengda0310/gobee/pcap
```

### 只用离线 pcap 重放（纯 Go，无需任何额外组件）

不需要安装任何东西。核心库 + 单元测试在任意平台直接可用：

```bash
go build ./...          # 纯 Go
go test ./...           # 纯 Go 测试
```

### 需要实时抓网卡（cgo + libpcap/Npcap）

实时抓包走 `gopacket/pcap`（cgo 子包），编译时需要 C 编译器 + libpcap/Npcap 库，并带上 `-tags livecapture`。**按平台分别配置：**

#### Linux / macOS

```bash
# 安装 libpcap 开发包
sudo apt install libpcap-dev      # Debian/Ubuntu
sudo yum install libpcap-devel    # RHEL/CentOS
brew install libpcap              # macOS

# 编译（gcc 通常系统自带）
CGO_ENABLED=1 go build -tags livecapture ./...

# 运行需要 CAP_NET_RAW 权限或 root
sudo ./myapp
# 或赋予权限：sudo setcap cap_net_raw=eip ./myapp
```

#### Windows（Npcap + MSYS2 gcc）

Windows 上实时抓包需要三样东西：**Npcap 运行库** + **gcc（C 编译器）** + **Npcap SDK 头文件**。

**第 1 步：安装 Npcap**

1. 从 https://npcap.com/dist/ 下载 `npcap-x.x.x-installer.exe`。
2. **以管理员身份运行**安装包。
3. **务必勾选**：
   - ✅ **Install Npcap in WinPcap API-compatible Mode**（关键！gopacket 链接的就是这个兼容层）
   - ❌ 不要勾「Restrict Npcap driver's access to Administrators only」（否则非管理员抓不到包）
4. 验证：`C:\Windows\System32\wpcap.dll` 和 `Packet.dll` 应存在。

**第 2 步：安装 gcc（MSYS2 + mingw-w64）**

1. 从 https://www.msys2.org/ 安装 MSYS2（路径建议 `C:\msys64`，不含空格/中文）。
2. 打开「MSYS2 MINGW64」（绿色图标，不是 UCRT/MSYS），执行：
   ```bash
   pacman -Syu
   pacman -S --needed mingw-w64-x86_64-gcc
   gcc --version   # 应输出版本号
   ```
3. **把 `C:\msys64\mingw64\bin` 加入系统 PATH**（Win 键搜「环境变量」→ 编辑系统变量 `Path` → 新增此路径），然后**重开所有终端**。

**第 3 步：准备 Npcap SDK 头文件**

gopacket/pcap 编译时需要 `pcap.h` 等头文件。两种方式任选：

- **方式 A（推荐，用源码包）**：从 https://github.com/nmap/nmap/tags 下载 Npcap 源码包（如 `npcap-1.88.zip`），解压到无空格路径（如 `D:\npcap-1.88`）。头文件在 `wpcap\libpcap\` 下，已自包含齐全。
- **方式 B（用预编译 SDK）**：从 https://npcap.com/dist/ 下载 `npcap-sdk-x.x.x.zip`，解压后用其 `Include\` 目录。

**第 4 步：生成 mingw 导入库（关键，否则 race/链接会失败）**

mingw 的链接器需要 `.a` 导入库才能解析 `wpcap` 符号。System32 里的 `wpcap.dll` 不够，需手动生成（一次性操作）：

```bash
# 在 MSYS2 MINGW64 终端中：
pacman -S --needed mingw-w64-x86_64-tools   # 提供 gendef

mkdir /tmp/wpcaplib && cd /tmp/wpcaplib
cp /c/Windows/System32/wpcap.dll .
cp /c/Windows/System32/Packet.dll .
gendef wpcap.dll
gendef Packet.dll
dlltool -d wpcap.def -l libwpcap.a -k
dlltool -d Packet.def -l libPacket.a -k
cp libwpcap.a libPacket.a /c/msys64/mingw64/x86_64-w64-mingw32/lib/
```

**第 5 步：配置 cgo 环境变量**

设置以下**用户环境变量**（用 `setx` 或图形界面，**永久生效**）：

```bash
setx CGO_CFLAGS "-I D:\npcap-1.88\wpcap\libpcap"
setx CGO_LDFLAGS "-l wpcap"
# PATH 已在第 2 步加入 C:\msys64\mingw64\bin
```

> ⚠️ `setx` 只对**之后新开的终端**生效，当前终端需手动 `export` 或重开。

**第 6 步：验证**（新开一个终端）

```bash
gcc --version                                          # 应输出版本
CGO_ENABLED=1 go build -tags livecapture ./...        # 实时抓包代码编译
go test ./... -race -tags livecapture -timeout 180s   # 完整链路（含 race）
```

> **Windows 实时抓包需管理员权限**：右键「以管理员身份运行」终端再执行抓包程序。

## 快速开始

### 场景 1：离线重放 pcap 文件（纯 Go，任意平台）

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gopacket/gopacket/pcapgo"
	"github.com/wangtengda0310/gobee/pcap"
)

func main() {
	f, err := os.Open("dump.pcap")
	if err != nil { panic(err) }
	defer f.Close()

	reader, err := pcapgo.NewReader(f)
	if err != nil { panic(err) }

	// 把 pcapgo.Reader 适配成本包的 Source。
	src := pcap.NewReaderSource(reader, reader.LinkType(), "dump.pcap")

	// 创建抓包器，配置每个 handler 的队列容量。
	c := pcap.NewCapturer(pcap.WithBufferSize(2048))
	defer c.Close() // 关闭所有 worker

	// 注册处理函数。
	c.RegisterHandler(pcap.NewHandlerFunc("printer", func(ctx context.Context, e *pcap.PacketEvent) error {
		fmt.Printf("[%s] %s -> %s\n", e.Timestamp, e.NetworkFlow.Src(), e.NetworkFlow.Dst())
		return nil
	}))

	// 抓包并广播。ctx 取消或数据源 EOF 时返回。
	c.Capture(context.Background(), src, pcap.Target{})
}
```

### 场景 2：实时抓取网卡（需 `livecapture` + Npcap/libpcap）

```bash
CGO_ENABLED=1 go build -tags livecapture -o myapp
# Windows 需管理员权限；Linux 需 CAP_NET_RAW 或 root
```

```go
src, err := pcap.NewLiveSource(`\Device\Npcap_{xxxx}`, 65535, false, "tcp port 80")
if err != nil { panic(err) }
defer src.Close()

c := pcap.NewCapturer(
	pcap.WithBPFFilter("tcp port 80"),          // 内核态过滤
	pcap.WithOverflowStrategy(pcap.OverflowDrop), // 满即丢，不阻塞抓包
)
c.RegisterHandler(pcap.NewHandlerFunc("http", func(ctx context.Context, e *pcap.PacketEvent) error {
    if app := e.Packet.ApplicationLayer(); app != nil {
        fmt.Println(string(app.Payload()))
    }
    return nil
})

c.Capture(ctx, src, pcap.Target{Host: "itsnot.fun"}) // 用户态二次过滤
```

### 场景 3：用 CLI 工具实时抓包 + HTTP 流重组

本包自带 `cmd/pcaptest` 命令行工具，覆盖常用抓包场景：

```bash
# 列出可用网卡
CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -list

# 抓包并打印每个包摘要
CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡名>

# 多网卡合并 + BPF 过滤 + 输出到 pcap 文件
CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡1>,<网卡2> -bpf "tcp port 80" -out dump.pcap

# 启用 HTTP 流重组，打印重组后的完整请求（Method/URL/Headers）
CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡名> -http
```

| Flag | 默认 | 说明 |
|---|---|---|
| `-list` | false | 列出本机网卡后退出 |
| `-iface` | 必填 | 网卡名，多个用逗号分隔 |
| `-bpf` | `tcp port 80` | BPF 过滤表达式 |
| `-http` | false | 启用 HTTP/1.x 流重组，打印重组后的完整请求 |
| `-out` | 空 | 输出到 pcap 文件 |
| `-host` | 空 | 用户态 host 过滤 |
| `-snaplen` | 65535 | 每包最大截获字节 |
| `-promisc` | false | 混杂模式 |

> Windows 抓包需管理员权限（右键以管理员运行终端）；Linux 需 root 或 `CAP_NET_RAW`。

## API 文档

### 核心类型

#### `Capturer` 接口

抓包器，两个核心方法：

```go
type Capturer interface {
    // RegisterHandler 注册处理函数，抓到的包会广播给它。
    // 重名返回 ErrHandlerExists；可在 Capture 前或运行期调用。
    RegisterHandler(handler PacketHandler) error

    // Capture 接收 source（数据源）与 target（过滤目标），开始抓包并广播。
    // 阻塞直到 ctx 取消 / 数据源 EOF / 致命错误。
    // 返回前会等待所有已投递的包被处理函数消费完（优雅退出）。
    Capture(ctx context.Context, source Source, target Target) error

    // UnregisterHandler 按名注销处理函数，并排空其队列。
    UnregisterHandler(name string) error

    // Stats 返回运行时统计（并发安全）。
    Stats() *CaptureStats
}
```

#### 生命周期 `Close()`

`Close()` 是 `Capturer` 接口的方法，关闭所有处理函数的 worker 并清空注册表。建议 `defer` 调用：

```go
c := pcap.NewCapturer(...)
defer c.Close()
```

### 处理函数

实现 `PacketHandler` 接口，或用便利函数 `NewHandlerFunc`：

```go
type PacketHandler interface {
    HandlePacket(ctx context.Context, pkt *PacketEvent) error
    Name() string
}

// 便利用法
pcap.NewHandlerFunc("my-handler", func(ctx context.Context, e *pcap.PacketEvent) error {
    // 处理包
    return nil
})
```

**并发约定**：
- 同一 handler 的 `HandlePacket` 调用是**串行**的（由各自的 worker 消费），单个 handler 实现无需加锁。
- 不同 handler 之间是**并行**的，若共享外部资源需自行保证安全。

### 数据源 `Source` & 构造函数

| 构造函数 | 用途 | 依赖 |
|---|---|---|
| `NewReaderSource(ds, link, name)` | 离线重放 pcap 文件 / 任意 `PacketDataSource` | 纯 Go |
| `NewLiveSource(iface, snaplen, promisc, bpf)` | 实时网卡抓包 | cgo + libpcap/Npcap + `-tags livecapture` |
| `NewMergedSource(sources...)` | 多网卡 fan-in 合并（要求子源 LinkType 一致） | 纯 Go |

**实时抓包相关函数**（均需 `-tags livecapture`）：

| 函数 | 用途 |
|---|---|
| `ListInterfaces()` | 列出本机网卡（`[]Interface{Name, Description, IPs}`） |
| `Source.(BPFCapable).SetBPFFilter(expr)` | 运行期热重载 BPF（并发安全） |
| `Source.(BPFValidator).ValidateBPF(expr)` | 预校验 BPF 表达式合法性（不应用） |

自定义数据源：实现 `Source` 接口的四个方法 `Packets()` / `LinkType()` / `Close()` / `String()` 即可。

#### 多网卡合并抓包

```go
// 打开多个网卡，合并成一个 Source 喂给同一个 Capturer。
src1, _ := pcap.NewLiveSource("eth0", 65535, false, "tcp port 80")
src2, _ := pcap.NewLiveSource("eth1", 65535, false, "tcp port 80")
merged, err := pcap.NewMergedSource(src1, src2) // 要求 LinkType 一致
if err != nil { panic(err) }
defer merged.Close()

capturer.Capture(ctx, merged, pcap.Target{Host: "itsnot.fun"})
```

> 合并后 `PacketEvent.Source` 统一为 `"merged:n"`，无法精确标识某个包来自哪个子源。
> 若需区分子源，handler 可依据包的网络层信息（IP/端口）区分。

### 过滤目标 `Target`

```go
type Target struct {
    BPF  string // BPF 表达式，如 "tcp port 80 and host itsnot.fun"。内核态过滤，性能最佳。
    Host string // 应用层 host，如 "itsnot.fun"。按 HTTP Host 头 / IP 子串匹配。BPF 不可用时用这个。
}
```

二者可组合：BPF 在内核态先过滤，Host 在用户态二次过滤。留空表示不过滤。

> ⚠️ **Host 子串匹配局限**：`Target.Host` 使用 `strings.Contains` 子串匹配，可能产生误匹配——如 `Host:"1.2"` 会命中 IP `10.1.2.3`（"1.2" 是其子串），或匹配到 payload 里的任意子串。这是 BPF 不可用时的降级方案，需要精确匹配请用 `Target.BPF`。

> **BPF 限制**：仅 `liveSource`（实时抓包）支持 BPF。离线 `readerSource` 不支持，指定 BPF 会返回 `ErrBPFNotSupported`。离线过滤请用 `Target.Host`。

### 配置选项（functional options）

| 选项 | 默认 | 说明 |
|---|---|---|
| `WithBufferSize(n)` | `1024` | 每个 handler 的有界队列容量 |
| `WithOverflowStrategy(s)` | `OverflowDrop` | 队列满时的背压策略 |
| `WithBPFFilter(expr)` | 空 | 全局 BPF 过滤（与 `Target.BPF` 二选一，后者优先） |

> ⚠️ **BPF 三入口覆盖语义**：BPF 有 3 个设置入口——`NewLiveSource(bpf)`（构造时）、`WithBPFFilter`（全局选项）、`Target.BPF`（每次 Capture）。三者同时设置时是**覆盖**而非合并，优先级：`Target.BPF` > `WithBPFFilter` > `NewLiveSource(bpf)`。即 Target.BPF 非空时完全忽略其它两者。
| `WithHooks(h)` | 空 | 生命周期回调 |

### 过载策略 `OverflowStrategy`

| 策略 | 行为 | 适用场景 |
|---|---|---|
| `OverflowDrop`（默认） | 队列满即丢新包，**绝不阻塞抓包** | 实时抓包（热路径） |
| `OverflowBlock` | 阻塞等待空位（响应 ctx 取消） | 离线分析、不能丢包 |
| `OverflowDropOldest` | 丢最旧的包（滑动窗口） | 只关心最新流量 |

> 选择建议：实时抓包用 `OverflowDrop`；读 pcap 文件重放用 `OverflowBlock`；监控最新流量用 `OverflowDropOldest`。

### 事件 `PacketEvent`

每个包的轻量封装，预解析了常用字段：

```go
type PacketEvent struct {
    Packet        gopacket.Packet // 原始包，可做深度解析
    Timestamp     time.Time       // 抓包时间
    Length        int             // 原始线上长度
    NetworkFlow   gopacket.Flow   // 网络层（src/dst IP）
    TransportFlow gopacket.Flow   // 传输层（src/dst port）
    LinkType      layers.LinkType
    Source        string          // 数据源标识（网卡名 / 文件路径）
}
```

常用解析：
```go
// 应用层 payload（HTTP 明文等）
if app := e.Packet.ApplicationLayer(); app != nil {
    payload := string(app.Payload())
}

// TCP 层细节
if tcp := e.Packet.Layer(layers.LayerTypeTCP); tcp != nil {
    t := tcp.(*layers.TCP)
    _ = t.SrcPort
}
```

### 统计 `Stats()`

```go
st := capturer.Stats()
fmt.Println("抓取总数:", st.Captured, "开始时间:", st.StartedAt)
for _, hs := range st.Handlers {
    fmt.Printf("  %s: 收=%d 处理=%d 丢=%d 错=%d\n",
        hs.Name, hs.Received, hs.Processed, hs.Dropped, hs.Errors)
}
```

可周期性采样用于监控告警（如 `Dropped` 持续增长说明处理函数过慢，需调大 `BufferSize` 或优化逻辑）。

### 回调 `Hooks`

```go
pcap.NewCapturer(pcap.WithHooks(pcap.Hooks{
    OnStart:  func(ctx context.Context) { log.Println("开始抓包") },
    OnPacket: func(e *pcap.PacketEvent) {}, // 每个包分发前（全局计数/采样）
    OnError:  func(err error) { log.Warn(err) }, // 非致命错误（解析失败等）
    OnStop:   func() { log.Println("停止抓包") },
}))
```

> 回调在抓包主 goroutine 中同步执行，请保持轻量；耗时操作转发到其他 goroutine。

### 流重组 `HTTPRequestHandler` / `HTTPResponseHandler`

普通 `PacketHandler` 看到的是**零散的 TCP 包**。若要拿到**重组后的完整 HTTP 请求/响应**，用这两个内置 handler——它们实现 `PacketHandler`，内部用 `tcpassembly` 做 TCP 重组 + 标准库 `net/http` 解析。

```go
// 重组并解析 HTTP 请求
h := pcap.NewHTTPRequestHandler("http-req", func(flow pcap.FlowKey, req *http.Request) error {
    fmt.Printf("[%s] %s %s Host=%s\n", flow, req.Method, req.URL.RequestURI(), req.Host)
    return nil
})
capturer.RegisterHandler(h)
// Capture 运行...
defer h.Close() // flush 残留流 + 等待 per-flow goroutine 退出

// HTTP 响应重组（对称）
hr := pcap.NewHTTPResponseHandler("http-resp", func(flow pcap.FlowKey, resp *http.Response) error {
    fmt.Printf("[%s] status=%d\n", flow, resp.StatusCode)
    return nil
})
```

| 类型 | 回调签名 | 解析内容 |
|---|---|---|
| `NewHTTPRequestHandler` | `func(flow FlowKey, req *http.Request) error` | Method / URL / Headers / Body |
| `NewHTTPResponseHandler` | `func(flow FlowKey, resp *http.Response) error` | StatusCode / Headers / Body |

**关键点**：
- **HTTP/1.x only**：基于标准库 `net/http`，不支持 HTTP/2（记为 TODO）。
- **明文 only**：不解密 TLS；抓 HTTPS 需先用 SSLKEYLOGFILE 预解密。
- **回调中的 `req`/`resp` 复用底层缓冲**，返回后可能被改写；如需保留请深拷贝。
- **⚠️ 必须手动 `Close()`**：这三个 handler 内部持有 `tcpassembly.Assembler` 和 per-flow goroutine，**`c.Close()` 不会替你关闭它们**。必须在 Capture 返回后调用 `defer h.Close()`，否则：(1) goroutine 泄漏；(2) 最后一个未 FIN 的流永远 flush 不出来（数据丢失）。
- **必须 `Close()`**：Capture 返回后调用，确保缓冲中的不完整流被 flush（否则可能丢失最后一个未 FIN 的请求）。
- **`FlowKey`** 标识 TCP 流（`NetworkFlow` + `TransportFlow`），便于区分请求来自哪个连接。

> 这两个 handler 完全是纯 Go 实现，**不依赖 Npcap**——可离线重放 pcap 测试，也可接入实时抓包。

### 通用 TCP 流重组 `TCPStreamHandler`（protobuf / 自定义二进制协议）

`HTTPRequestHandler` 内置了 HTTP 解析。若你要重组**其它基于 TCP 的协议**（protobuf、Redis、自定义二进制等），用 `NewTCPStreamHandler`——它把重组后的字节流（`io.Reader`）交给你，你自己解析消息。

**protobuf 示例**（假设协议是「4字节大端长度前缀 + protobuf 消息」）：

```go
h := pcap.NewTCPStreamHandler("proto", func(flow pcap.FlowKey, r io.Reader) error {
    for {
        // 1. 读 4 字节长度前缀
        var lenBuf [4]byte
        if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
            return err // EOF / 流结束
        }
        msgLen := binary.BigEndian.Uint32(lenBuf[:])

        // 2. 读对应长度的消息体
        payload := make([]byte, msgLen)
        if _, err := io.ReadFull(r, payload); err != nil {
            return err
        }

        // 3. 用 proto.Unmarshal 解析（yourpb 是你从 .proto 生成的包）
        msg := &yourpb.YourMessage{}
        if err := proto.Unmarshal(payload, msg); err != nil {
            continue // 不是合法 protobuf，跳过（可能是握手/其它帧）
        }
        fmt.Printf("[%s] %s\n", flow, proto.CompactTextString(msg))
    }
})
capturer.RegisterHandler(h)
defer h.Close() // Capture 返回后 flush 残留流
```

**关键点**：
- `r` 是**重组后的完整字节流**——TCP 层的拆包/乱序已被处理，你按协议格式流式读取即可。
- 一个 TCP 流可能含**多条协议消息**（keep-alive），回调内应**循环读取**直到 `io.EOF`。
- 消息边界格式（长度前缀？分隔符？）取决于你的协议，需你自己实现读取逻辑。
- 同样不依赖 Npcap（纯 Go，基于 tcpassembly）。

## 最佳实践

### ✅ 推荐

- **实时抓包用 `OverflowDrop`**：保证不因慢 handler 丢真实流量。
- **优先用 BPF 过滤**：在内核态过滤远比用户态高效，是降低投递压力最有效的手段。
- **handler 保持轻量**：把 IO（写库/网络）异步化，或只做采集、用单独的消费者处理。
- **按 `Dropped` 调参**：监控统计，丢包多时调大 `BufferSize` 或拆分 handler。
- **用 `c.Close()` 收尾**：长跑服务退出前关闭 worker，避免 goroutine 泄漏。

### ❌ 避免

- **不要在 handler 里做长时间阻塞操作**：会占满该 handler 的队列，导致后续包被丢（Drop 策略）或阻塞抓包（Block 策略）。
- **不要对离线 `readerSource` 用 BPF**：会返回 `ErrBPFNotSupported`，改用 `Target.Host`。
- **不要假设不同 handler 看到同一份包的时机一致**：它们各自独立消费，时序不保证。

## FAQ

### Q: 单元测试怎么不抓真实网卡？

A: 测试用 `pcapgo` 在内存构造 pcap 文件（含 itsnot.fun HTTP 请求），复用与实时抓包完全相同的「读取 → 解析 → 广播 → 处理」链路，验证 gopacket 能正确解码。把离线 `readerSource` 换成 `liveSource` 即可实时抓取。这样测试不依赖网卡、特权或 Npcap，任意机器/CI 可运行。

### Q: Windows 实时抓包报 "cgo: C compiler gcc not found"？

A: gcc 不在 PATH。按「安装 → Windows」章节装好 MSYS2 mingw-w64 gcc，并把 `C:\msys64\mingw64\bin` 加入系统 PATH，然后**重开终端**。用 `gcc --version` 验证。

### Q: Windows 链接报 "cannot find -lwpcap"？

A: mingw 链接器缺少 wpcap 的导入库（`.a`）。按「安装 → Windows → 第 4 步」用 `gendef` + `dlltool` 生成 `libwpcap.a` 并复制到 `C:\msys64\mingw64\x86_64-w64-mingw32\lib\`。这是 Windows + cgo 抓包最容易卡住的一步。

### Q: `go test -race` 报 cgo / 链接错误？

A: race 检测依赖 cgo，需要 gcc 在 PATH。有两种情况：
- **核心库 race**（不带 `-tags livecapture`）：只需 gcc；**不**需要 Npcap。若仍报 `cannot find -lwpcap`，是因为你全局设了 `CGO_LDFLAGS=-l wpcap` 但没生成导入库（见上一问）。生成导入库后即可解决。
- **完整链路 race**（`-tags livecapture`）：需要 gcc + Npcap SDK 头文件（`CGO_CFLAGS`）+ 导入库。

### Q: 多个 handler 看到的包数量不一致？

A: 正常。每个 handler 有独立队列与过载统计，慢的 handler 在 `OverflowDrop` 下会丢更多包。用 `Stats()` 查看各自的 `Dropped`。

### Q: 如何抓 HTTPS（加密流量）？

A: 本库只读明文。抓 HTTPS 需配合 TLS 解密（如 SSLKEYLOGFILE + Wireshark 导出的解密 pcap），用离线重放方式读解密后的 pcap。

## 测试

```bash
cd pcap

go test ./... -v                                    # 全部单元测试（纯 Go，无需 Npcap）
go test ./... -race                                 # 并发安全检测（需 gcc/clang）
go test -tags livecapture ./...                     # 含 cgo 实时代码（需 Npcap）
go test -tags livecapture,integration -v ./...      # 集成测试（需 Npcap + 真实网卡 + 管理员权限）
```

## 相关文档

- [gopacket 文档](https://github.com/gopacket/gopacket) — 底层依赖
