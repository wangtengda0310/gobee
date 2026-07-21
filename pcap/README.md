# pcap

基于 [gopacket](https://github.com/gopacket/gopacket) 的网络抓包库，提供「数据源 → 抓包器 → 处理函数」的广播式抓包能力。

> 维护者/开发指南请见 [CLAUDE.md](./CLAUDE.md)。

## 核心特性

- **广播式分发**：一次抓包，多个处理函数并行消费，互不拖累。
- **过载保护**：抓包循环绝不因慢处理函数而阻塞；三种可配置的背压策略。
- **并发安全**：运行期动态增删处理函数，全量原子统计。
- **纯 Go 核心**：离线 pcap 重放与单元测试不依赖 cgo / Npcap，任意平台可运行。
- **实时抓包可选**：基于 cgo 的网卡抓包用构建标签 `livecapture` 隔离。

## 安装

```bash
go get github.com/wangtengda0310/gobee/pcap
```

**实时抓包的额外要求**（仅当需要抓网卡流量时）：
- Windows：安装 [Npcap](https://npcap.com/)（安装时勾选 "Install Npcap in WinPcap API-compatible Mode"）。
- Linux/macOS：安装 libpcap（`apt install libpcap-dev` / `brew install libpcap`）。
- 编译时带上 `CGO_ENABLED=1 -tags livecapture`。

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
	if lc, ok := c.(pcap.LifeCycler); ok { defer lc.Close() }

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
}))

c.Capture(ctx, src, pcap.Target{Host: "itsnot.fun"}) // 用户态二次过滤
```

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

#### `LifeCycler` 接口（可选扩展）

`NewCapturer` 的返回值同时实现 `LifeCycler`，用于显式关闭所有 worker：

```go
if lc, ok := capturer.(pcap.LifeCycler); ok {
    defer lc.Close()
}
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

自定义数据源：实现 `Source` 接口的三个方法 `Packets()` / `LinkType()` / `Close()` 即可。

### 过滤目标 `Target`

```go
type Target struct {
    BPF  string // BPF 表达式，如 "tcp port 80 and host itsnot.fun"。内核态过滤，性能最佳。
    Host string // 应用层 host，如 "itsnot.fun"。按 HTTP Host 头 / IP 子串匹配。BPF 不可用时用这个。
}
```

二者可组合：BPF 在内核态先过滤，Host 在用户态二次过滤。留空表示不过滤。

> **BPF 限制**：仅 `liveSource`（实时抓包）支持 BPF。离线 `readerSource` 不支持，指定 BPF 会返回 `ErrBPFNotSupported`。离线过滤请用 `Target.Host`。

### 配置选项（functional options）

| 选项 | 默认 | 说明 |
|---|---|---|
| `WithBufferSize(n)` | `1024` | 每个 handler 的有界队列容量 |
| `WithOverflowStrategy(s)` | `OverflowDrop` | 队列满时的背压策略 |
| `WithBPFFilter(expr)` | 空 | 全局 BPF 过滤（与 `Target.BPF` 二选一，后者优先） |
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

## 最佳实践

### ✅ 推荐

- **实时抓包用 `OverflowDrop`**：保证不因慢 handler 丢真实流量。
- **优先用 BPF 过滤**：在内核态过滤远比用户态高效，是降低投递压力最有效的手段。
- **handler 保持轻量**：把 IO（写库/网络）异步化，或只做采集、用单独的消费者处理。
- **按 `Dropped` 调参**：监控统计，丢包多时调大 `BufferSize` 或拆分 handler。
- **用 `LifeCycler.Close()` 收尾**：长跑服务退出前关闭 worker，避免 goroutine 泄漏。

### ❌ 避免

- **不要在 handler 里做长时间阻塞操作**：会占满该 handler 的队列，导致后续包被丢（Drop 策略）或阻塞抓包（Block 策略）。
- **不要对离线 `readerSource` 用 BPF**：会返回 `ErrBPFNotSupported`，改用 `Target.Host`。
- **不要假设不同 handler 看到同一份包的时机一致**：它们各自独立消费，时序不保证。

## FAQ

### Q: 单元测试怎么不抓真实网卡？

A: 测试用 `pcapgo` 在内存构造 pcap 文件（含 itsnot.fun HTTP 请求），复用与实时抓包完全相同的「读取 → 解析 → 广播 → 处理」链路，验证 gopacket 能正确解码。把离线 `readerSource` 换成 `liveSource` 即可实时抓取。这样测试不依赖网卡、特权或 Npcap，任意机器/CI 可运行。

### Q: Windows 实时抓包报 "cgo: C compiler gcc not found"？

A: 安装 MinGW-w64（如通过 MSYS2：`pacman -S mingw-w64-x86_64-gcc`），并确保 `gcc` 在 PATH。

### Q: `go test -race` 报 cgo 错误？

A: race 检测依赖 cgo，需要 C 编译器（gcc/clang）。它**不会**链接 libpcap（race 不需要 `-tags livecapture`），只需 gcc 在 PATH。

### Q: 多个 handler 看到的包数量不一致？

A: 正常。每个 handler 有独立队列与过载统计，慢的 handler 在 `OverflowDrop` 下会丢更多包。用 `Stats()` 查看各自的 `Dropped`。

### Q: 如何抓 HTTPS（加密流量）？

A: 本库只读明文。抓 HTTPS 需配合 TLS 解密（如 SSLKEYLOGFILE + Wireshark 导出的解密 pcap），用离线重放方式读解密后的 pcap。

## 测试

```bash
cd pcap

go test ./... -v            # 全部单元测试（纯 Go，无需 Npcap）
go test ./... -race         # 并发安全检测（需 gcc/clang）
go test -tags livecapture   # 需 Npcap + 真实网卡（CI 不跑）
```

## 相关文档

- [CLAUDE.md](./CLAUDE.md) — 维护者/开发指南（架构、设计决策、扩展、陷阱）
- [gopacket 文档](https://github.com/gopacket/gopacket) — 底层依赖
