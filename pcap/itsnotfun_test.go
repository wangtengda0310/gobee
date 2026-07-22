package pcap

import (
	"bytes"
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gopacket/gopacket"        // 纯 Go（SerializeLayers/CaptureInfo），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（Ethernet/IPv4/TCP），不需 Npcap
	"github.com/gopacket/gopacket/pcapgo" // 纯 Go（pcap 文件读写），不需 Npcap
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 测试辅助：在内存构造 pcap（含 HTTP 流量），不依赖网卡/Npcap。
// 这些辅助被 itsnotfun_test.go 与 broadcast_test.go 共享。
// =============================================================================

// pcapLinkType 是我们构造的离线 pcap 的链路层类型：以太网。
const pcapLinkType = layers.LinkTypeEthernet

// httpFlow 描述一条要写入 pcap 的 HTTP 流（一个请求包）。
type httpFlow struct {
	SrcMAC, DstMAC   net.HardwareAddr
	SrcIP, DstIP     net.IP
	SrcPort, DstPort uint16
	Payload          []byte // 应用层明文（HTTP 请求/响应）
}

// writePcap 把一组 HTTP 流序列化为一个内存中的 pcap 文件（返回字节）。
// 每条 flow 写成一个独立的以太网帧：Eth -> IPv4 -> TCP -> payload。
func writePcap(flows []httpFlow) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := pcapgo.NewWriter(buf)
	if err := w.WriteFileHeader(65535, pcapLinkType); err != nil {
		return nil, err
	}
	base := time.Unix(1700000000, 0)
	for i, f := range flows {
		data, err := serializeHTTP(f)
		if err != nil {
			return nil, err
		}
		ci := gopacket.CaptureInfo{
			Timestamp:     base.Add(time.Duration(i) * time.Second),
			CaptureLength: len(data),
			Length:        len(data),
		}
		if err := w.WritePacket(ci, data); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// serializeHTTP 用 gopacket 的 SerializeLayers 构造一个
// Eth -> IPv4 -> TCP -> 应用层 payload 的完整帧。
func serializeHTTP(f httpFlow) ([]byte, error) {
	eth := &layers.Ethernet{
		SrcMAC:       f.SrcMAC,
		DstMAC:       f.DstMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipv4 := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    f.SrcIP,
		DstIP:    f.DstIP,
	}
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(f.SrcPort),
		DstPort: layers.TCPPort(f.DstPort),
		Seq:     1,
		ACK:     true,
		Window:  65535,
	}
	// TCP 校验和需要 IPv4 上下文。
	if err := tcp.SetNetworkLayerForChecksum(ipv4); err != nil {
		return nil, err
	}

	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, opts,
		eth, ipv4, tcp, gopacket.Payload(f.Payload),
	); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// newSourceFromBytes 把一段 pcap 字节包装成 Source（基于 pcapgo.Reader）。
func newSourceFromBytes(t *testing.T, name string, data []byte) Source {
	t.Helper()
	r, err := pcapgo.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	return NewReaderSource(r, r.LinkType(), name)
}

// collectHandler 收集所有到达的 PacketEvent，供断言使用。并发安全。
type collectHandler struct {
	name string
	mu   sync.Mutex
	got  []*PacketEvent
}

func newCollectHandler(name string) *collectHandler { return &collectHandler{name: name} }

func (h *collectHandler) HandlePacket(_ context.Context, pkt *PacketEvent) error {
	h.mu.Lock()
	h.got = append(h.got, pkt)
	h.mu.Unlock()
	return nil
}
func (h *collectHandler) Name() string { return h.name }
func (h *collectHandler) snapshot() []*PacketEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]*PacketEvent, len(h.got))
	copy(out, h.got)
	return out
}

// =============================================================================
// ★ 核心验证测试：验证可通过 gopacket 抓取（解析） itsnot.fun 的 HTTP 请求。
//
// 背景：当前机器未安装 Npcap，无法实时抓包。本测试通过在内存构造一个
// 含 itsnot.fun HTTP 请求的 pcap 文件，复用与实时抓包相同的「读取 -> 解析 -> 广播」
// 链路（readerSource -> Capturer.Capture），验证 gopacket 能正确解码出
// 以 itsnot.fun 为目标的 HTTP 请求。链路打通后，把 readerSource 换成
// liveSource（Npcap）即可真正实时抓取。
// =============================================================================

func TestCapture_ItsNotFunHTTPRequest(t *testing.T) {
	// 准备测试用 MAC/IP。itsnot.fun 不在测试环境解析范围内，
	// 这里用一个稳定的占位 IP（100.64.1.2）作为 itsnot.fun 的目标 IP。
	const itsnotFunHost = "itsnot.fun"
	itsnotFunIP := net.IPv4(100, 64, 1, 2)

	httpReq := []byte("GET / HTTP/1.1\r\n" +
		"Host: " + itsnotFunHost + "\r\n" +
		"User-Agent: pcap-test/1.0\r\n" +
		"Accept: */*\r\n" +
		"Connection: close\r\n" +
		"\r\n")

	data, err := writePcap([]httpFlow{
		{
			// 一条干扰流量：example.org，验证过滤不误伤也不漏抓。
			SrcMAC:  net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			DstMAC:  net.HardwareAddr{0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb},
			SrcIP:   net.IPv4(10, 0, 0, 2),
			DstIP:   net.IPv4(93, 184, 216, 34),
			SrcPort: 51000,
			DstPort: 80,
			Payload: []byte("GET / HTTP/1.1\r\nHost: example.org\r\n\r\n"),
		},
		{
			// 目标流量：itsnot.fun 的 HTTP 请求。
			SrcMAC:  net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
			DstMAC:  net.HardwareAddr{0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb},
			SrcIP:   net.IPv4(10, 0, 0, 2),
			DstIP:   itsnotFunIP,
			SrcPort: 51001,
			DstPort: 80,
			Payload: httpReq,
		},
	})
	require.NoError(t, err, "构造 pcap 失败")

	src := newSourceFromBytes(t, "itsnotfun.pcap", data)

	// 收集型 handler：拿到所有命中的包。
	h := newCollectHandler("itsnot-collector")
	var seenErrs []error
	c := NewCapturer(
		WithBufferSize(64),
		WithOverflowStrategy(OverflowBlock), // 离线重放不丢包
		WithHooks(Hooks{OnError: func(err error) { seenErrs = append(seenErrs, err) }}),
	)
	require.NoError(t, c.RegisterHandler(h))

	// 用 Host 过滤，只投递包含 itsnot.fun 的包。
	// BPF 在离线 pcap 上不可用（readerSource 未实现 BPFCapable），
	// 这里走 Target.Host 的用户态过滤路径。
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { errCh <- c.Capture(ctx, src, Target{Host: itsnotFunHost}) }()

	// readerSource.Packets() 在 EOF 时关闭，Capture 随之返回 nil。
	select {
	case err := <-errCh:
		cancel()
		require.NoError(t, err, "Capture 返回错误")
	case <-time.After(3 * time.Second):
		cancel()
		t.Fatal("Capture 超时未返回")
	}

	got := h.snapshot()
	require.Len(t, got, 1, "应恰好投递 1 个 itsnot.fun 的包（example.org 被过滤掉）")

	evt := got[0]

	// 1) 网络层：目的 IP 应为 itsnot.fun 的 IP。
	dst := evt.NetworkFlow.Dst().String()
	assert.Equal(t, itsnotFunIP.String(), dst, "目的 IP 应为 itsnot.fun")

	// 2) 传输层：目的端口 80。
	assert.Equal(t, "80", evt.TransportFlow.Dst().String(), "目的端口应为 80")

	// 3) 应用层：gopacket 应能解码出 HTTP 请求，且 Host 头为 itsnot.fun。
	app := evt.Packet.ApplicationLayer()
	require.NotNil(t, app, "应用层不应为空，gopacket 应解码出 HTTP")
	payload := string(app.Payload())
	assert.Contains(t, payload, "GET / HTTP/1.1")
	assert.Contains(t, payload, "Host: itsnot.fun")

	// 4) 统计：抓取并分发 1 个包，处理 1 个，无丢弃。
	st := c.Stats()
	assert.Equal(t, int64(1), st.Captured, "Captured 应为 1")
	require.Len(t, st.Handlers, 1)
	assert.Equal(t, int64(1), st.Handlers[0].Received)
	assert.Equal(t, int64(1), st.Handlers[0].Processed)
	assert.Equal(t, int64(0), st.Handlers[0].Dropped)
}

// Ensure the Source/io interfaces we rely on are wired correctly at compile time.
