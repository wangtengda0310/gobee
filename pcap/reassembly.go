package pcap

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/tcpassembly"
	"github.com/gopacket/gopacket/tcpassembly/tcpreader"
)

// =============================================================================
// TCP 流重组 + HTTP/1.x 结构化解析
// =============================================================================
// 本文件把 gopacket 的 tcpassembly（TCP 重组）+ 标准库 net/http（HTTP 解析）
// 封装成一个 PacketHandler，让 handler 拿到「重组后的完整 HTTP 请求/响应」，
// 而非零散的 TCP 包。
//
// 设计要点（Design B，详见 CLAUDE.md 决策 7）：
//   - 实现 PacketHandler 接口，不改 Capturer。
//   - 关键契合点：tcpassembly.Assembler.Assemble() 不是并发安全的，
//     但 capturer 的 per-handler worker 已保证 HandlePacket 串行调用，
//     所以串行约束自动满足，零新锁。
//   - 每个 TCP 流（半连接）在 StreamFactory.New 时启动一个 goroutine，
//     drain ReaderStream（否则 Reassembled 会阻塞）并用 net/http 解析。
//   - HTTP/1.x only：标准库 net/http 不解析 HTTP/2；HTTP/2 支持记为 TODO。
// =============================================================================

// FlowKey 标识一个 TCP 流（由网络层流 + 传输层流共同确定一个半连接）。
// 传给使用者的回调，便于区分请求来自哪个连接。
type FlowKey struct {
	NetworkFlow   gopacket.Flow // src/dst IP
	TransportFlow gopacket.Flow // src/dst port
}

// HTTPRequestHandler 重组 TCP 流并解析出 HTTP/1.x 请求。
// 它实现 PacketHandler，可直接用 capturer.RegisterHandler 注册。
//
// 用法：
//
//	h := pcap.NewHTTPRequestHandler("http-req", func(flow pcap.FlowKey, req *http.Request) error {
//	    fmt.Printf("%s %s Host=%s\n", req.Method, req.URL.RequestURI(), req.Host)
//	    return nil
//	})
//	capturer.RegisterHandler(h)
//	// ... Capture 运行 ...
//	defer h.Close() // 或在 Capture 返回后调用，flush 残留流
//
// 注意：
//   - req 复用底层缓冲，回调返回后可能被复用；如需保留请深拷贝 Header/Body。
//   - 非 TCP 包静默跳过。
//   - Close 后不再产生新的回调。
type HTTPRequestHandler struct {
	ras *tcpStreamReassembler
}

// NewHTTPRequestHandler 创建一个 HTTP 请求重组处理函数。
//
//   - name：handler 名称（实现 PacketHandler.Name）。
//   - onRequest：每解析出一个完整 HTTP 请求时调用。返回 error 仅用于统计，
//     不会中断重组（由 capturer 的 errors 计数吸收）。
func NewHTTPRequestHandler(name string, onRequest func(flow FlowKey, req *http.Request) error) *HTTPRequestHandler {
	return &HTTPRequestHandler{
		ras: newTCPStreamReassembler(name, func(flow FlowKey, r io.Reader) {
			buf := bufio.NewReader(r)
			for {
				req, err := http.ReadRequest(buf)
				if err != nil {
					// io.EOF / 连接关闭 / 解析错误：本流结束。
					return
				}
				_ = onRequest(flow, req)
			}
		}),
	}
}

// HandlePacket 实现 PacketHandler。提取 TCP 层喂给 Assembler；非 TCP 包跳过。
func (h *HTTPRequestHandler) HandlePacket(_ context.Context, e *PacketEvent) error {
	return h.ras.handlePacket(e)
}

// Name 实现 PacketHandler。
func (h *HTTPRequestHandler) Name() string { return h.ras.name }

// Close flush 剩余流并等待所有 per-flow goroutine 退出。幂等。
// 建议在 Capture 返回后调用，确保缓冲中的不完整流也被处理。
func (h *HTTPRequestHandler) Close() error { return h.ras.close() }

// HTTPResponseHandler 重组 TCP 流并解析出 HTTP/1.x 响应。结构与 HTTPRequestHandler 对称。
//
// 用法：
//
//	h := pcap.NewHTTPResponseHandler("http-resp", func(flow pcap.FlowKey, resp *http.Response) error {
//	    fmt.Printf("status=%d\n", resp.StatusCode)
//	    return nil
//	})
type HTTPResponseHandler struct {
	ras *tcpStreamReassembler
}

// NewHTTPResponseHandler 创建一个 HTTP 响应重组处理函数。
func NewHTTPResponseHandler(name string, onResponse func(flow FlowKey, resp *http.Response) error) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		ras: newTCPStreamReassembler(name, func(flow FlowKey, r io.Reader) {
			buf := bufio.NewReader(r)
			for {
				resp, err := http.ReadResponse(buf, nil)
				if err != nil {
					return
				}
				_ = onResponse(flow, resp)
			}
		}),
	}
}

// HandlePacket 实现 PacketHandler。
func (h *HTTPResponseHandler) HandlePacket(_ context.Context, e *PacketEvent) error {
	return h.ras.handlePacket(e)
}

// Name 实现 PacketHandler。
func (h *HTTPResponseHandler) Name() string { return h.ras.name }

// Close flush 并等待 per-flow goroutine 退出。幂等。
func (h *HTTPResponseHandler) Close() error { return h.ras.close() }

// TCPStreamHandler 重组 TCP 流，把重组后的字节流（io.Reader）交给使用者解析。
//
// 这是「通用 TCP 流重组」入口：HTTP 用 NewHTTPRequestHandler/NewHTTPResponseHandler
// （内置 net/http 解析）；其它基于 TCP 的协议（protobuf、Redis RESP、自定义二进制等）
// 用本 handler——你只需提供如何从字节流读出消息的逻辑。
//
// 用法（以 protobuf 为例）：
//
//	h := pcap.NewTCPStreamHandler("proto", func(flow pcap.FlowKey, r io.Reader) error {
//	    // r 是重组后的完整字节流，按你的协议格式读消息。
//	    // 常见模式：先读 4 字节长度前缀，再读对应长度的 protobuf 消息。
//	    for {
//	        var lenBuf [4]byte
//	        if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
//	            return err // EOF / 流结束
//	        }
//	        msgLen := binary.BigEndian.Uint32(lenBuf[:])
//	        payload := make([]byte, msgLen)
//	        if _, err := io.ReadFull(r, payload); err != nil {
//	            return err
//	        }
//	        msg := &yourpb.YourMessage{}
//	        if err := proto.Unmarshal(payload, msg); err != nil {
//	            continue // 不是合法 protobuf，跳过（可能是握手/其它协议）
//	        }
//	        fmt.Println(msg)
//	    }
//	})
//	capturer.RegisterHandler(h)
//	defer h.Close() // Capture 返回后 flush 残留流
//
// 注意：
//   - r 在本流的整个生命周期内有效，直到流结束（FIN/RST/超时/Close）。
//     Read 返回 io.EOF 表示本流结束，你的回调应直接返回。
//   - 一个流可能含多条协议消息（keep-alive），回调内应循环读取。
//   - 如果协议消息边界与 TCP 段不对齐，r 已帮你重组好，你按流式读取即可。
//   - Close 后不再产生新的回调。
type TCPStreamHandler struct {
	ras *tcpStreamReassembler
}

// NewTCPStreamHandler 创建一个通用 TCP 流重组处理函数。
//
//   - name：handler 名称（实现 PacketHandler.Name）。
//   - onStream：每个 TCP 流启动一个回调，r 是该流重组后的字节流。
//     回调返回时本流的 goroutine 退出。返回 error 仅用于统计（capturer 的 errors）。
//
// onStream 的常见实现：循环从 r 读「长度前缀 + 消息体」，用 proto.Unmarshal 解析。
func NewTCPStreamHandler(name string, onStream func(flow FlowKey, r io.Reader) error) *TCPStreamHandler {
	return &TCPStreamHandler{
		ras: newTCPStreamReassembler(name, func(flow FlowKey, r io.Reader) {
			_ = onStream(flow, r) // error 由内部统计吸收，不中断其它流。
		}),
	}
}

// HandlePacket 实现 PacketHandler。提取 TCP 层喂给 Assembler；非 TCP 包跳过。
func (h *TCPStreamHandler) HandlePacket(_ context.Context, e *PacketEvent) error {
	return h.ras.handlePacket(e)
}

// Name 实现 PacketHandler。
func (h *TCPStreamHandler) Name() string { return h.ras.name }

// Close flush 剩余流并等待所有 per-flow goroutine 退出。幂等。
// 建议在 Capture 返回后调用，确保缓冲中的不完整流也被处理。
func (h *TCPStreamHandler) Close() error { return h.ras.close() }

// -----------------------------------------------------------------------------
// 内部实现
// -----------------------------------------------------------------------------

// tcpStreamReassembler 是 HTTPRequestHandler / HTTPResponseHandler /
// TCPStreamHandler 共用的核心。持有 tcpassembly.Assembler + StreamPool，
// 管理 per-flow goroutine 生命周期。
//
// 它是「TCP 流重组 + 自定义解析」的通用引擎：
// 使用者只需提供一个 parse 函数，从重组后的 io.Reader 读出自己的协议消息。
type tcpStreamReassembler struct {
	name string

	pool *tcpassembly.StreamPool
	asm  *tcpassembly.Assembler

	// parse 在每个流的首个包时启动（由 factory.New），负责 drain ReaderStream
	// 并解析协议。返回时本 per-flow goroutine 退出（wg 由本结构内部管理，
	// 使用者无需关心）。
	parse func(flow FlowKey, r io.Reader)

	// wg 跟踪所有活跃 per-flow goroutine。
	wg sync.WaitGroup

	// closed 标记 Close 已调用，阻止新流创建（factory.New 检查）。
	closed atomic.Bool

	// mu 保护 factory.New 与 Close 的竞争（Close 设 closed 后不再 Add）。
	mu sync.Mutex
}

// newTCPStreamReassembler 构造核心。parse 函数在每个流上运行，从 r 读出协议消息。
func newTCPStreamReassembler(name string, parse func(flow FlowKey, r io.Reader)) *tcpStreamReassembler {
	ras := &tcpStreamReassembler{name: name, parse: parse}
	ras.pool = tcpassembly.NewStreamPool(ras) // ras 实现 StreamFactory
	ras.asm = tcpassembly.NewAssembler(ras.pool)
	return ras
}

// handlePacket 提取 TCP 层并喂给 Assembler。
func (ras *tcpStreamReassembler) handlePacket(e *PacketEvent) error {
	if ras.closed.Load() {
		return nil // 已 Close，静默丢弃。
	}
	if e == nil || e.Packet == nil {
		return nil
	}
	tcpLayer := e.Packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil // 非 TCP 包跳过（如 UDP）。
	}
	tcp, ok := tcpLayer.(*layers.TCP)
	if !ok {
		return nil
	}
	// Assemble 内部会查找/创建流，并同步调用 StreamFactory.New（首个包）。
	// 由于 HandlePacket 由 per-handler worker 串行调用，Assemble 的串行约束自动满足。
	ras.asm.AssembleWithTimestamp(e.NetworkFlow, tcp, e.Timestamp)
	return nil
}

// close flush 剩余流并等待所有 per-flow goroutine 退出。幂等。
func (ras *tcpStreamReassembler) close() error {
	ras.mu.Lock()
	if !ras.closed.CompareAndSwap(false, true) {
		ras.mu.Unlock()
		return nil // 已 Close，幂等返回。
	}
	// FlushAll 让 Assembler 触发所有未完成流的 ReassemblyComplete，
	// 进而使 ReaderStream 的 Read 返回 EOF，per-flow goroutine 退出。
	ras.asm.FlushAll()
	ras.mu.Unlock()

	// 等待所有 per-flow goroutine 退出（它们在 ReaderStream EOF 后 Done）。
	ras.wg.Wait()
	return nil
}

// New 实现 tcpassembly.StreamFactory。每个新流首次出现时由 Assembler 调用。
//
// 这里启动 per-flow goroutine：drain ReaderStream（必须，否则 Reassembled 阻塞）
// 并用 parse 函数解析协议。wg 由本结构内部管理，parse 无需关心。
func (ras *tcpStreamReassembler) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	ras.mu.Lock()
	defer ras.mu.Unlock()
	if ras.closed.Load() {
		// 已 Close：返回一个空的 ReaderStream，其 ReassemblyComplete 会被立即调用。
		// 不启动 goroutine，不 Add wg。
		r := tcpreader.NewReaderStream()
		return &r
	}
	ras.wg.Add(1)
	r := tcpreader.NewReaderStream()
	flow := FlowKey{NetworkFlow: netFlow, TransportFlow: tcpFlow}
	// per-flow goroutine：drain + parse。内部管 wg，parse 只管读协议。
	go func() {
		defer ras.wg.Done()
		ras.parse(flow, &r)
	}()
	return &r
}

// 编译期保证实现 StreamFactory。
var _ tcpassembly.StreamFactory = (*tcpStreamReassembler)(nil)

// 编译期保证两个 Handler 实现 PacketHandler。
var (
	_ PacketHandler = (*HTTPRequestHandler)(nil)
	_ PacketHandler = (*HTTPResponseHandler)(nil)
)

// String 用于日志/调试，返回 "src:sport->dst:dport" 形式。
func (k FlowKey) String() string {
	src, dst := k.NetworkFlow.Src().String(), k.NetworkFlow.Dst().String()
	if src == "" && dst == "" {
		return "<empty>"
	}
	return fmt.Sprintf("%s:%s->%s:%s",
		src, k.TransportFlow.Src(),
		dst, k.TransportFlow.Dst())
}
