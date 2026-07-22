# pcap AI 开发指南

> 本文档指导 AI 助手与人类维护者进行 pcap 模块的开发和维护。
> 使用者文档见 [README.md](./README.md)。

## First of all
- 对gopacket的调用需要使用注释标注是否需要安装npcap

## 项目定位

pcap 是一个基于 [gopacket](https://github.com/gopacket/gopacket) 的网络抓包库，
提供「数据源（Source）→ 抓包器（Capturer）→ 处理函数（PacketHandler）」的广播式抓包能力。

**核心价值**：
- 广播解耦（一次抓包，多 handler 并行消费，互不拖累）
- 过载保护（抓包热路径绝不阻塞；三种可配置背压策略）
- 纯 Go 核心（离线 pcap 重放与单元测试不依赖 cgo / Npcap）

**关键约束**：核心库**必须保持纯 Go 可编译**。实时抓包（cgo + libpcap/Npcap）必须用构建标签隔离，绝不能让默认构建依赖本机 C 库。

## 快速参考

| 文档 | 路径 | 说明 |
|------|------|------|
| 使用者文档 | @pcap/README.md | API、使用示例、FAQ |
| 包文档 | @pcap/doc.go | Package pcap 概览（含两段思考结论索引） |
| 核心实现 | @pcap/capturer.go | 广播 + 过载 + 并发安全（顶部含思考结论注释） |
| 实时抓包 | @pcap/source_live.go | `//go:build cgo && livecapture`，Npcap/libpcap |

## 架构总览

### 数据流

```
   Source.Packets() ──────────────►  Capture 循环（抓包主 goroutine）
   (pcap 文件 / 网卡 / mock)              │
                                          ▼ matchTarget（Host 用户态过滤）
                                     broadcast（取 handler 快照，读锁）
                                          │
              ┌───────────────────────────┼───────────────────────────┐
              ▼                           ▼                           ▼
        handlerSlot A               handlerSlot B               handlerSlot C
        ch chan (有界)              ch chan (有界)              ch chan (有界)
              │                           │                           │
         dispatch                    dispatch                    dispatch
        (按 Overflow 策略)           (按 Overflow 策略)           (按 Overflow 策略)
              ▼                           ▼                           ▼
         worker A                    worker B                    worker C
      for pkt := range ch        for pkt := range ch        for pkt := range ch
         HandlePacket               HandlePacket               HandlePacket
```

**关键性质**：
- 抓包主 goroutine 只负责「读包 + 过滤 + 投递」，**不调用 HandlePacket**。
- 每个 handler 独占一个 worker goroutine + 一条独立的有界 channel。慢 handler 只会填满自己的队列，不影响其他 handler，也不阻塞抓包。
- `dispatch` 是过载保护的核心（见 `capturer.go` 的 `dispatch` 函数）：按 `OverflowStrategy` 决定投递行为。

### 构建模式（重要）

本模块有**两种构建形态**，维护时必须同时考虑：

| 模式 | 命令 | 包含的文件 | 用途 |
|------|------|----------|------|
| 纯 Go（默认） | `CGO_ENABLED=0 go build` | 除 `source_live.go` 外全部 | 离线重放、单元测试、CI |
| 实时抓包 | `CGO_ENABLED=1 go build -tags livecapture` | 全部（含 `source_live.go`） | 网卡抓包（需 Npcap/libpcap） |

#### Windows 开发环境配置（cgo 实时抓包）

在 Windows 上开发/验证实时抓包功能，需要一次性配置（详见 README「安装 → Windows」）。本节列出维护者需要的**关键配置摘要**，便于新环境快速复刻：

1. **gcc**：MSYS2 mingw-w64，`C:\msys64\mingw64\bin` 加入系统 PATH。
2. **Npcap 运行库**：`C:\Windows\System32\wpcap.dll` + `Packet.dll`。
3. **Npcap SDK 头文件**：用源码包的 `wpcap\libpcap\` 目录（如 `D:\npcap-1.88\wpcap\libpcap`），头文件自包含齐全。
4. **mingw 导入库**（关键，否则 `-lwpcap` 失败）：用 `gendef`+`dlltool` 从 `wpcap.dll`/`Packet.dll` 生成 `libwpcap.a`/`libPacket.a`，复制到 `C:\msys64\mingw64\x86_64-w64-mingw32\lib\`。
5. **cgo 环境变量**（用户级，`setx` 永久化）：
   ```
   CGO_CFLAGS  = -I D:\npcap-1.88\wpcap\libpcap
   CGO_LDFLAGS = -l wpcap
   ```

**陷阱**：持久化 `CGO_LDFLAGS=-l wpcap` 后，**所有** cgo 构建（含不带 `livecapture` 的 race 测试）都会尝试链接 wpcap。只有生成了导入库（第 4 步）才能让 `-lwpcap` 永远成功，否则 race 测试会因 `cannot find -lwpcap` 失败。

`source_live.go` 顶部的 `//go:build cgo && livecapture` 是**双重门槛**：
- `cgo`：把「race 检测需要 cgo」与「链接 libpcap」解耦——`go test -race`（开启 cgo 但无 tag）不会拉入 libpcap。
- `livecapture`：用户必须显式带上才链接本机抓包库。

> 改动 `source_live.go` 时，确保它在默认构建下完全不参与编译。

## 文件职责

| 文件 | 职责 | 维护要点 |
|------|------|---------|
| `doc.go` | 包文档概览 | 改动对外行为时同步更新 |
| `interface.go` | `Capturer` / `PacketHandler` / `Source` / `Target` / `LifeCycler` 接口 | 接口变更 = 破坏性改动，需谨慎 |
| `types.go` | `PacketEvent` / `Stats` / `Config` / `Hooks` / `OverflowStrategy` | 统计字段必须 atomic |
| `options.go` | functional options | 新增配置项时加 `WithXxx` |
| `errors.go` | 哨兵错误 | 用 `errors.Is` 友好的命名错误 |
| `capturer.go` | **核心实现** | 改动需重点 review 并发安全 |
| `source.go` | 纯 Go：`pcapgo.Reader` → `Source` | 不依赖 cgo |
| `source_live.go` | cgo：网卡实时抓包 → `Source` + `ListInterfaces` + BPF 热重载 | `//go:build cgo && livecapture` |
| `merge.go` | 纯 Go：`MergedSource` 多网卡 fan-in 合并 | 不依赖 cgo；要求子源 LinkType 一致 |
| `reassembly.go` | 纯 Go：TCP 流重组 + HTTP/通用协议解析（`HTTPRequestHandler`/`HTTPResponseHandler`/`TCPStreamHandler`） | 基于 `tcpassembly`；实现 `PacketHandler`，串行约束由 worker 满足 |
| `itsnotfun_test.go` | ★ 验证 itsnot.fun HTTP 抓取 + 测试辅助（共享） | 含 `writePcap` / `collectHandler` |
| `merge_test.go` | MergedSource fan-in 单测（纯 Go） | 改 merge.go 必改这里 |
| `broadcast_test.go` | 并发安全 + 过载策略 + 动态增删测试 | 改 dispatch 必改这里 |
| `coverage_test.go` | 覆盖缺口补齐（OverflowBlock/BPF/Hooks/Close 等） | 详见各子测试 |
| `reassembly_test.go` | 流重组单测（含 `writeTCPStreamPcap` 构造多包 TCP 流） | 改 reassembly.go 必改这里 |
| `live_integration_test.go` | 真实网卡集成测试 | `//go:build cgo && livecapture && integration` |
| `cmd/pcaptest/main.go` | 实时抓包 CLI（`-list`/`-bpf`/`-out`/多 `-iface`/`-http` 流重组） | `//go:build livecapture` |

## 核心设计决策（必读）

### 决策 1：广播解耦 + 每 handler 独立队列

**问题**：数据包数量巨大、处理函数来不及消费怎么办？

**答案**（实现见 `capturer.go` 顶部注释与 `broadcast`/`dispatch`）：
1. 抓包 goroutine 不直接调用处理函数，而是把包投递给广播分发器。
2. 分发器为**每个 handler 维护独立的有界 channel**——慢 handler 只堵自己，不传染。
3. 每个 handler 独占一个 worker goroutine 串行消费（`for pkt := range ch`）。
4. 三种过载策略（`OverflowStrategy`）：Drop（默认，不阻塞抓包）/ Block（不丢包）/ DropOldest（滑动窗口）。

**结论**：抓包永不阻塞；handler 互相隔离；过载可预测、可观测、可配置。

### 决策 2：哨兵（sentinel）flush 方案 + `flushAll` 优雅退出

**问题**：Capture 返回时如何保证已投递的包被处理完？

**当前实现**（`capturer.go` 的 `handlerSlot.flushWG` / `runWorker` / `flushAll` / `sendSentinel`）：
- `handlerSlot.flushWG sync.WaitGroup`：仅用于 flush，不追踪每个包。
- `flushAll` 对每个 slot `flushWG.Add(1)` 后投递一个 **nil 哨兵**到其 channel；worker 在 `range` 中遇到 nil 包即 `flushWG.Done()`，表示「到此之前的真实包都已处理完」。最后 `flushAll` 等 `flushWG.Wait()`。
- `sendSentinel` 用 `recover` 容错 send-on-closed-channel（slot 恰被 Unregister 时），并返回是否成功；失败时调用方手动 `Done` 平衡刚才的 `Add`。

**历史教训（已修复，但记录于此以防回退）**：
早期版本用「逐包 inFlight WaitGroup：dispatch Add(1)、worker Done(1)」。这在 `OverflowDropOldest` 下极难正确——踢出旧包时需要手动 Done 平衡，但与 worker 的消费并发后会产生 **`panic: sync: negative WaitGroup counter`**（实测从 google/gopacket 迁移到 gopacket/gopacket 后多次运行暴露）。
哨兵方案把"等待清空"从"逐包追踪"降为"一次水位标记"，正确性显然，不再依赖"每个包恰好 Done 一次"的脆弱不变量。

`broadcast_test.go::TestOverflowDropOldest_KeepsLatest` 是回归守卫。

### 决策 3：handler 注册持久化，`Capture` 不注销

**问题**：为什么 `Capture` 结束后 `Stats()` 还能看到 handler 统计？

**实现**：早期版本 `Capture` 退出时调用 `drainAll`（关闭队列 + 从 map 删除 handler），导致：
- `Stats()` 在 Capture 后看不到历史 handler 统计。
- handler 无法跨多次 Capture 复用。

**修正**：
- `Capture` 退出只调用 `flushAll`（等在途包，**不动 handler 注册**）。
- worker 在 `RegisterHandler` 时启动，持续运行，直到 `UnregisterHandler` 或 `Close()`。
- 真正关闭 worker 由 `Close()`（`LifeCycler` 接口）负责。

> 改动生命周期逻辑时，确认这三件事的边界：`Capture` = 喂包 + flush；`UnregisterHandler` = 单个注销；`Close` = 全部关闭。

### 决策 4：BPF 能力探测，不支持则显式失败

**实现**（`capturer.go` 的 `Capture` BPF 合并段 + `BPFCapable` 接口）：
- `BPFCapable` 接口：只有 `liveSource` 实现（`SetBPFFilter`）。
- `Capture` 时若指定了 BPF 但 Source 未实现 `BPFCapable`，返回 `ErrBPFNotSupported`，而非静默放过。
- 离线 `readerSource` 不实现 BPF → 离线过滤只能用 `Target.Host`（用户态）。

**原因**：BPF 是安全/性能关键，静默降级会让用户误以为过滤生效实则全量投递。

### 决策 5：MergedSource 多网卡 fan-in（纯 Go）

**实现**（`merge.go`）：
- `NewMergedSource(sources ...Source)` 把多个 Source 合并为一个，`Capturer` 接口不变。
- 每个 Source 一个转发 goroutine，把 `Packets()` 的包投到合并 channel；所有子源 EOF 后合并 channel 关闭。
- **要求所有子源 `LinkType()` 一致**，否则返回 `ErrInconsistentLinkType`（gopacket 用单一解码器，异构会解码错乱）。

**局限（TODO）**：合并后 `PacketEvent.Source` 统一为 `"merged:n"`，无法精确标识某包来自哪个子源（gopacket.Packet 不可变，附加上下文成本高）。handler 若需区分子源，靠包内容（IP/端口）。

> **TODO（待后续评估）**：
> - 支持**异构 LinkType** 合并：把 LinkType 从 Source 级移到 per-packet（`PacketEvent.LinkType` 已存在，但 `toEvent` 当前用 Source 的 LinkType 填充）。需要调整 `toEvent` + 解码逻辑，改动面较大。
> - **per-packet 子源标记**：在合并时给每个包附上来源标识。可用 `gopacket.Packet` 的 metadata 或 wrapper，但会增加复杂度。

### 决策 6：BPF 热重载的并发安全

**实现**（`source_live.go` 的 `liveSource`）：
- `liveSource.mu sync.RWMutex` 保护 `handle` / `bpf`。
- `SetBPFFilter`（热重载）加写锁，底层 `pcap.SetBPFFilter` 通过 libpcap 的 `pcap_setfilter` **原子替换**内核 BPF 程序，不中断抓包。
- `ValidateBPF` 加读锁，用 `CompileBPFFilter` 预编译校验（不应用）。
- `BPFValidator` 接口与 `BPFCapable` 分离：校验与应用是两个能力，某些 Source 可能只支持其一。

**推荐用法**：热重载前先 `ValidateBPF` 校验，通过后再 `SetBPFFilter` 应用，避免把非法 BPF 应用进去导致抓包异常。

### 决策 7：流重组为何实现为 PacketHandler 而非新接口（Design B）

**实现**（`reassembly.go`）：
- `HTTPRequestHandler` / `HTTPResponseHandler` 都实现 `PacketHandler`，内部持有 `tcpassembly.Assembler` + `StreamPool`。
- `HandlePacket` 提取 `*layers.TCP`，喂给 `Assembler.AssembleWithTimestamp`。
- 每个 TCP 流在 `StreamFactory.New` 时启动一个 goroutine，drain `tcpreader.ReaderStream` 并用标准库 `net/http` 解析。
- `Close()` 调 `asm.FlushAll()` 触发所有未完成流的 `ReassemblyComplete`，等待 per-flow goroutine 退出。

**为什么不用新的 `StreamHandler` 接口（Design A）而用 `PacketHandler`（Design B）**：

关键架构契合点——`tcpassembly.Assembler.Assemble()` **不是并发安全的**，必须串行调用。而现有 `capturer.runWorker`（per-handler 独占 worker goroutine）**已保证单个 handler 的 `HandlePacket` 串行执行**。把 Assembler 放进一个 `PacketHandler`：
- 串行约束**自动满足**，零新锁。
- `Capturer` 的 2 方法接口**不变**。
- 复用现有 `OverflowStrategy`（包队列背压）和 `flushAll`（Capture 退出收尾）。

**HTTP 解析路径**：gopacket v1.7.0 **没有** `layers.HTTP` 类型。HTTP 解析走标准库 `net/http.ReadRequest/ReadResponse`，喂给它 `tcpreader.ReaderStream`（既是 Stream 又是 io.Reader）。这与 gopacket 官方 `examples/httpassembly` 示例完全一致。

**生命周期要点**：
- per-flow goroutine 在 `StreamFactory.New` 启动（drain ReaderStream 是硬性要求，否则 `Reassembled` 阻塞）。
- `Close()` 必须被调用：`FlushAll` 触发未完成流结束，否则缓冲中的不完整流会丢失，且 goroutine 泄漏。
- `closed atomic.Bool` 阻止 Close 后新流创建；`factory.New` 在 closed 状态下返回空 ReaderStream（不启动 goroutine）。

**局限（TODO）**：
- **HTTP/2 不支持**：标准库 `net/http` 不解析 HTTP/2；如需支持需引入专门的 HTTP/2 解析器。
- **TLS 不解密**：只读明文；HTTPS 需配合 SSLKEYLOGFILE 预解密。
- **req/resp 缓冲复用**：回调返回后底层缓冲可能被改写，需保留时调用方需深拷贝。

## 并发安全契约

**严格遵守**，改动核心代码时逐条核对：

| 资源 | 保护方式 | 位置 |
|------|---------|------|
| `handlers` map（增删 handler） | `sync.RWMutex`（`Register`/`Unregister`/`Close` 写锁，`broadcast`/`Stats`/`flushAll` 读锁） | `capturer.go` |
| 分发时的 handler 列表 | **快照**：读锁拷贝到本地切片后释放锁，再遍历投递（避免长持锁阻塞注册） | `broadcast` |
| 统计计数 | `atomic.Int64`（`received`/`processed`/`dropped`/`errors`/`captured`） | `handlerSlot` / `capturer` |
| 每个 slot 的 channel | 单写（dispatch）单读（worker），channel 本身线程安全 | `dispatch` / `runWorker` |
| flush 哨兵计数 | 每个 slot 一个 `sync.WaitGroup`（`flushWG`），`flushAll` Add、worker 见 nil Done | 见决策 2 |

**禁止**：在持锁状态下调用 handler 的 `HandlePacket` 或做 IO。

## 注释标准

沿用本仓库 `agent` 模块风格：**双语注释**（中文叙述 + 英文标识符），所有导出符号必须有文档注释。

### 函数注释模板

```go
// Capture 接收 source 与 target，开始抓包并广播给所有已注册处理函数。
//
// 阻塞直到：ctx 取消 / 数据源 EOF / 发生致命错误。
// 返回前会 flush 所有处理函数的在途包（等待它们被处理完），保证优雅退出。
// 处理函数不会被注销，可在后续 Capture 中继续复用；真正释放请调用 Close()。
func (c *capturer) Capture(ctx context.Context, source Source, target Target) error
```

### 设计决策注释（解释「为什么」）

```go
// 哨兵走阻塞投递：worker 持续消费会腾出空位。
// 若 channel 已被 UnregisterHandler 关闭，send 会 panic；
// 此时 worker 已退出、不会再 Done，必须手动 Done 平衡刚才的 Add，否则 Wait 死锁。
if !sendSentinel(s) {
    s.flushWG.Done()
}
```

> 性能/并发相关的「非显然」决策必须有 Why 注释，便于下一位维护者（或 AI）不破坏不变量。

## 测试指南

### 测试矩阵

| 测试文件 | 覆盖 | 关键不变量 |
|---------|------|----------|
| `itsnotfun_test.go::TestCapture_ItsNotFunHTTPRequest` | ★ 验证 gopacket 解析 itsnot.fun HTTP | 目的 IP、80 端口、`Host: itsnot.fun`、统计守恒 |
| `broadcast_test.go::TestBroadcast_AllHandlersReceiveAll` | 多 handler 全量广播、无丢无重复 | 每个 handler 收到 N 个，-race 友好 |
| `broadcast_test.go::TestOverflowDrop_DoesNotBlockAndCountsDropped` | Drop 不阻塞抓包 + dropped 计数 | `received + dropped == captured`，耗时 < 串行 |
| `broadcast_test.go::TestOverflowDropOldest_KeepsLatest` | DropOldest flush 哨兵不破坏（回归守卫） | 不死锁，dropped > 0 |
| `broadcast_test.go::TestRegisterUnregister_DynamicAndErrors` | 动态增删 + 哨兵错误 | 运行期注册生效，`ErrHandlerExists`/`ErrHandlerNil`/`ErrEmptyName`/`ErrHandlerNotFound` |
| `coverage_test.go` | 补齐覆盖缺口：OverflowBlock/BPF/Hooks/Close/LifeCycler/String/Option/hostMatches | 详见各子测试，覆盖率 80% → 96% |

### 测试数据策略（重要）

**所有单元测试都不依赖网卡/特权/Npcap**。原理：
1. `itsnotfun_test.go` 顶部定义共享辅助：`writePcap`（用 `gopacket.SerializeLayers` 内存构造 Eth→IPv4→TCP→HTTP 帧）+ `collectHandler`（并发安全的收集型 handler）。
2. 用 `pcapgo.NewReader` 读回内存 pcap，经 `NewReaderSource` 接入与实时抓包**完全相同**的链路。
3. 因此「验证 itsnot.fun HTTP 抓取」等价于验证了实时路径（把 `readerSource` 换成 `liveSource` 即可）。

### 运行命令

```bash
cd pcap

# 1. 全量单元测试（纯 Go，CI 默认跑这个，覆盖率 ~96%）
CGO_ENABLED=0 go test ./... -v

# 2. 并发安全检测（需 gcc/clang）
#    核心库 race：只需 gcc，不链接 libpcap
go test ./... -race -count=1
#    完整链路 race（含 cgo 实时代码）：需 gcc + Npcap SDK + 导入库（见「构建模式 → Windows」）
go test ./... -race -tags livecapture -count=1 -timeout 180s

# 3. vet 两种模式都要过
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=1 go vet -tags livecapture ./...

# 4. lint 两种模式（可选，严格集）
CGO_ENABLED=0 golangci-lint run --enable errcheck,gocritic,unparam,gosec ./...
CGO_ENABLED=1 golangci-lint run --build-tags livecapture --enable errcheck,gocritic,unparam,gosec ./...
```

### 新增测试的检查清单

- [ ] 不依赖网卡/特权/Npcap（用 mockSource / writePcap）
- [ ] 有 race 友好的同步（`require.Eventually` + mutex/atomic，勿裸 sleep）
- [ ] 断言统计守恒，而非仅「不 panic」

## 常见任务

### 添加新的 Source 类型

1. 实现 `Source` 接口（`Packets()` / `LinkType()` / `Close()` / `String()`）。
2. 若支持 BPF，额外实现 `BPFCapable`（`SetBPFFilter`）。
3. 纯 Go 源放 `source_xxx.go`；cgo 源放 `source_xxx_live.go` + `//go:build cgo && <tag>`。
4. 加测试：用 mock 注入，验证它能被 `Capture` 正确消费。

### 自定义流解析器（非 HTTP 协议）

`reassembly.go` 已封装 HTTP/1.x（`NewHTTPRequestHandler`/`NewHTTPResponseHandler`）。
对于其它基于 TCP 的协议（protobuf、Redis RESP、自定义二进制），直接用 **`NewTCPStreamHandler`**：

```go
pcap.NewTCPStreamHandler("proto", func(flow pcap.FlowKey, r io.Reader) error {
    // r 是重组后的字节流，按你的协议格式读消息（如长度前缀 + protobuf）
    // 循环读取直到 io.EOF（流结束）。详见 README 的 protobuf 示例。
})
```

内部实现：`TCPStreamHandler` 与 HTTP handler 共用 `tcpStreamReassembler`
（持 `tcpassembly.Assembler` + `StreamPool`）。`onStream` 回调在 per-flow goroutine 中运行，
wg 由内部管理，使用者无需关心。

若需要更底层的控制（如自定义 StreamFactory、Accept gating、方向标记），
可直接基于 `github.com/gopacket/gopacket/tcpassembly` 自己实现，完全绕过本包的封装。

### 添加新的过载策略

1. 在 `types.go` 的 `OverflowStrategy` 枚举加常量 + `String()`。
2. 在 `capturer.go::dispatch` 的 switch 加 case。
3. **核对 flush 哨兵路径**：新策略下丢弃的包不应影响 flush（哨兵方案天然不受丢包影响），但若你新增的分支可能丢弃 nil 哨兵，需保证 `flushWG` 仍能 Done（否则 `flushAll` 死锁）。
4. 在 `broadcast_test.go` 加测试（覆盖「不死锁 + 计数正确」）。

### 修改 Capture 生命周期

重读决策 2、3。确认：
- `Capture` 退出不删 handler（只 flush）。
- `flushAll` 能在合理时间内返回（哨兵未被新代码路径意外吞掉）。
- `Close` 关闭所有 worker 且可多次调用。

### 添加配置项

1. `types.go::Config` 加字段 + `defaultConfig()` 设默认。
2. `options.go` 加 `WithXxx` Option。
3. 文档（README 表格 + doc.go 示例）同步。

## 陷阱清单

- **`OverflowDropOldest` 不要丢掉 nil 哨兵**：哨兵是 flush 的水位标记，被丢会导致 `flushAll` 死锁。当前实现中哨兵走阻塞投递不经过 DropOldest 分支，保持这一约定。回归测试 `TestOverflowDropOldest_KeepsLatest`。
- **`OverflowBlock` 下 handler 永久阻塞会卡死 `flushAll`**：Block 策略假设 handler 是「慢但会完成」，不是「永久卡」。测试中用 sleep 模拟慢 handler，不要用永久阻塞的 gate。
- **持锁调用 HandlePacket**：禁止。会导致死锁（若 handler 内部又回调 capturer）或性能塌陷。
- **修改 `source_live.go` 却忘了构建标签**：会导致默认构建拉入 libpcap，CI（无 Npcap）编译失败。
- **`doc.go` / `source.go` 注释里的构建标签表述**：保持与 `source_live.go` 实际 tag 一致（`cgo && livecapture`）。

## 禁止的操作

- ❌ 让核心库（除 `source_live.go`）依赖 cgo / libpcap / Npcap。
- ❌ 在 `Capture` 主循环中直接调用 `HandlePacket`（破坏广播解耦）。
- ❌ 在持 `capturer.mu` 锁时调用 handler 或做 IO。
- ❌ 在 dispatch 里丢弃 nil 哨兵（会破坏 flushAll，死锁）。
- ❌ 移除过载策略测试（尤其 DropOldest 的回归测试）。
- ❌ 改动 `Source` / `Capturer` / `PacketHandler` 接口签名而不更新测试与文档。
- ❌ 假设 `MergedSource` 场景下 `PacketEvent.Source` 是具体子源名（实际为 `"merged:n"`）。
- ❌ 对 `liveSource` 并发调用 `SetBPFFilter` 而不加锁（已由 `liveSource.mu` 保护，勿破坏）。

## 推荐的操作

- ✅ 改动 dispatch/生命周期后，跑 `-race`（有 gcc 时）+ 全量 `-v` 测试。
- ✅ 新增过载策略/Source 时，先写测试再实现。
- ✅ 并发相关决策写 Why 注释。
- ✅ 保持 README（使用者）与 CLAUDE（维护者）的职责边界：API 变更两边都更新。

## 联系方式

- 问题反馈：提交 GitHub Issue
- 代码审查：参考本文件的并发安全契约与陷阱清单
- 使用问题：指向 README.md
