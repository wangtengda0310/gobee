package pcap

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gopacket/gopacket"        // 纯 Go（SerializeLayers/NewPacket），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（Ethernet/IPv4/TCP/UDP），不需 Npcap
	"github.com/gopacket/gopacket/pcapgo" // 纯 Go（pcap 文件读写），不需 Npcap
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 流重组单元测试（纯 Go，不依赖网卡/Npcap）。
//
// 关键挑战：构造合法的多包 TCP 流（SYN/数据段/FIN，序列号正确），
// 验证 tcpassembly.Assembler 能重组出完整 HTTP 请求。
// =============================================================================

// tcpSegment 描述一个 TCP 段（一个包）。
type tcpSegment struct {
	srcIP, dstIP        net.IP
	srcPort, dstPort    uint16
	seq, ack            uint32
	syn, ackF, fin, psh bool
	payload             []byte
}

// writeTCPStreamPcap 把一系列 TCP 段写成 pcap 字节。
// 调用方负责保证 seq/ack 的正确性（SYN/FIN 各占 1 个序列号）。
func writeTCPStreamPcap(segments []tcpSegment) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := pcapgo.NewWriter(buf)
	if err := w.WriteFileHeader(65535, layers.LinkTypeEthernet); err != nil {
		return nil, err
	}
	base := time.Unix(1700000000, 0)
	mac := net.HardwareAddr{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	for i, seg := range segments {
		data, err := serializeTCPSegment(seg, mac)
		if err != nil {
			return nil, err
		}
		ci := gopacket.CaptureInfo{
			Timestamp:     base.Add(time.Duration(i) * time.Millisecond),
			CaptureLength: len(data),
			Length:        len(data),
		}
		if err := w.WritePacket(ci, data); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// serializeTCPSegment 构造一个 Eth→IPv4→TCP→payload 帧。
func serializeTCPSegment(seg tcpSegment, mac net.HardwareAddr) ([]byte, error) {
	eth := &layers.Ethernet{
		SrcMAC: mac, DstMAC: mac, EthernetType: layers.EthernetTypeIPv4,
	}
	ipv4 := &layers.IPv4{
		Version: 4, IHL: 5, TTL: 64, Protocol: layers.IPProtocolTCP,
		SrcIP: seg.srcIP, DstIP: seg.dstIP,
	}
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(seg.srcPort),
		DstPort: layers.TCPPort(seg.dstPort),
		Seq:     seg.seq, Ack: seg.ack,
		SYN: seg.syn, ACK: seg.ackF, FIN: seg.fin, PSH: seg.psh,
		Window: 65535,
	}
	if err := tcp.SetNetworkLayerForChecksum(ipv4); err != nil {
		return nil, err
	}
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, opts, eth, ipv4, tcp, gopacket.Payload(seg.payload)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildRequestStreamPcap 构造一个完整的 HTTP 请求 TCP 流。
// payload 被拆成 chunkSize 大小的多个段，验证跨包重组。
// outOfOrder=true 时打乱数据段顺序，测试乱序重组。
func buildRequestStreamPcap(t *testing.T, payload []byte, chunkSize int, outOfOrder bool) []byte {
	t.Helper()
	const (
		clientPort, serverPort uint16 = 51000, 80
	)
	cIP := net.ParseIP("10.0.0.2")
	sIP := net.ParseIP("10.0.0.3")

	var segs []tcpSegment
	clientSeq := uint32(1000)
	// SYN（占 1 个序列号）
	segs = append(segs, tcpSegment{
		srcIP: cIP, dstIP: sIP, srcPort: clientPort, dstPort: serverPort,
		seq: clientSeq, syn: true,
	})
	clientSeq++

	// 数据段（每个 chunkSize 字节，seq 递增）。
	for start := 0; start < len(payload); start += chunkSize {
		end := start + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		chunk := payload[start:end]
		segs = append(segs, tcpSegment{
			srcIP: cIP, dstIP: sIP, srcPort: clientPort, dstPort: serverPort,
			seq: clientSeq, ackF: true, psh: true, payload: chunk,
		})
		clientSeq += uint32(len(chunk)) //nolint:gosec // G115: chunk 大小受控（测试数据，远小于 int32 上限）
	}

	// FIN（占 1 个序列号）
	segs = append(segs, tcpSegment{
		srcIP: cIP, dstIP: sIP, srcPort: clientPort, dstPort: serverPort,
		seq: clientSeq, ackF: true, fin: true,
	})

	if outOfOrder && len(segs) > 3 {
		// 交换前两个数据段（索引 1 和 2）测试乱序重组。
		segs[1], segs[2] = segs[2], segs[1]
	}

	data, err := writeTCPStreamPcap(segs)
	require.NoError(t, err)
	return data
}

// feedHandler 把 Source 的每个包直接喂给 handler（不经 Capturer 广播层，更直接测重组）。
func feedHandler(t *testing.T, src Source, h PacketHandler) {
	t.Helper()
	ctx := context.Background()
	for pkt := range src.Packets() {
		e := &PacketEvent{
			Packet:    pkt,
			Timestamp: pkt.Metadata().CaptureInfo.Timestamp,
		}
		if net := pkt.NetworkLayer(); net != nil {
			e.NetworkFlow = net.NetworkFlow()
		}
		if tp := pkt.TransportLayer(); tp != nil {
			e.TransportFlow = tp.TransportFlow()
		}
		_ = h.HandlePacket(ctx, e)
	}
}

// collectedReq 收到的一次 HTTP 请求（深拷贝关键字段，避免被复用）。
type collectedReq struct {
	Method     string
	RequestURI string
	Host       string
	UserAgent  string
}

// -----------------------------------------------------------------------------
// 测试用例
// -----------------------------------------------------------------------------

// TestReassembly_SingleHTTPRequest：跨多包的 HTTP 请求被正确重组。
func TestReassembly_SingleHTTPRequest(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: itsnot.fun\r\nUser-Agent: test/1.0\r\nConnection: close\r\n\r\n")
	pcapData := buildRequestStreamPcap(t, payload, 10, false) // 10 字节一段，多段

	var (
		mu  sync.Mutex
		got []collectedReq
	)
	h := NewHTTPRequestHandler("http-req", func(flow FlowKey, req *http.Request) error {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, collectedReq{
			Method:     req.Method,
			RequestURI: req.URL.RequestURI(),
			Host:       req.Host,
			UserAgent:  req.UserAgent(),
		})
		return nil
	})

	src := newSourceFromBytes(t, "stream.pcap", pcapData)
	feedHandler(t, src, h)
	require.NoError(t, h.Close(), "Close 应 flush 残留流")

	require.Len(t, got, 1, "应解析出 1 个 HTTP 请求")
	r := got[0]
	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "/", r.RequestURI)
	assert.Equal(t, "itsnot.fun", r.Host)
	assert.Equal(t, "test/1.0", r.UserAgent)
}

// TestReassembly_OutOfOrder：乱序数据段仍能正确重组（含 body 完整性验证）。
func TestReassembly_OutOfOrder(t *testing.T) {
	payload := []byte("POST /api HTTP/1.1\r\nHost: itsnot.fun\r\nContent-Length: 5\r\n\r\nhello")
	pcapData := buildRequestStreamPcap(t, payload, 8, true) // 乱序

	var (
		mu      sync.Mutex
		got     []collectedReq
		gotBody string
	)
	h := NewHTTPRequestHandler("ooo", func(flow FlowKey, req *http.Request) error {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, collectedReq{Method: req.Method, RequestURI: req.URL.RequestURI(), Host: req.Host})
		// 读取 body 验证跨包重组的内容完整性（此前只 Close 未读，审查发现）。
		if req.Body != nil {
			body, _ := io.ReadAll(req.Body)
			gotBody = string(body)
			_ = req.Body.Close()
		}
		return nil
	})

	src := newSourceFromBytes(t, "ooo.pcap", pcapData)
	feedHandler(t, src, h)
	require.NoError(t, h.Close())

	require.Len(t, got, 1, "乱序下仍应解析出 1 个请求")
	assert.Equal(t, "POST", got[0].Method, "应解析出 POST 方法")
	assert.Equal(t, "/api", got[0].RequestURI, "应解析出 /api 路径")
	assert.Equal(t, "itsnot.fun", got[0].Host, "应解析出 Host 头")
	assert.Equal(t, "hello", gotBody, "body 应被正确重组（跨包乱序后仍完整）")
}

// TestReassembly_CloseFlush：不发 FIN，直接 Close，残留流应被 flush（回调仍触发）。
func TestReassembly_CloseFlush(t *testing.T) {
	// 构造只有 SYN + 数据段（无 FIN）的流。
	payload := []byte("GET / HTTP/1.1\r\nHost: itsnot.fun\r\n\r\n")
	// 手动构造，不调 buildRequestStreamPcap（它带 FIN）
	cIP := net.ParseIP("10.0.0.2")
	sIP := net.ParseIP("10.0.0.3")
	segs := []tcpSegment{
		{srcIP: cIP, dstIP: sIP, srcPort: 51000, dstPort: 80, seq: 1000, syn: true},
		{srcIP: cIP, dstIP: sIP, srcPort: 51000, dstPort: 80, seq: 1001, ackF: true, psh: true, payload: payload},
	}
	pcapData, err := writeTCPStreamPcap(segs)
	require.NoError(t, err)

	var (
		mu       sync.Mutex
		gotCount int
	)
	h := NewHTTPRequestHandler("flush", func(flow FlowKey, req *http.Request) error {
		mu.Lock()
		gotCount++
		mu.Unlock()
		return nil
	})

	src := newSourceFromBytes(t, "nofin.pcap", pcapData)
	feedHandler(t, src, h)
	// Close 前：流未结束（无 FIN），Assembler 内部缓冲着未 flush。
	// Close → FlushAll → 触发 ReassemblyComplete → per-flow goroutine 读到 EOF → 解析出请求。
	require.NoError(t, h.Close())
	assert.Equal(t, 1, gotCount, "Close 应 flush 出残留流的请求")
}

// TestReassembly_NonTCPPacket：UDP 包应静默跳过，不 panic。
func TestReassembly_NonTCPPacket(t *testing.T) {
	// 构造一个 UDP 包。
	eth := &layers.Ethernet{SrcMAC: []byte{1, 2, 3, 4, 5, 6}, DstMAC: []byte{6, 5, 4, 3, 2, 1}, EthernetType: layers.EthernetTypeIPv4}
	ip := &layers.IPv4{Version: 4, IHL: 5, TTL: 64, Protocol: layers.IPProtocolUDP, SrcIP: net.ParseIP("10.0.0.2"), DstIP: net.ParseIP("10.0.0.3")}
	udp := &layers.UDP{SrcPort: 5353, DstPort: 5353}
	require.NoError(t, udp.SetNetworkLayerForChecksum(ip))
	buf := gopacket.NewSerializeBuffer()
	require.NoError(t, gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}, eth, ip, udp, gopacket.Payload([]byte("dns"))))
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	h := NewHTTPRequestHandler("udp-test", func(flow FlowKey, req *http.Request) error {
		t.Error("UDP 包不应触发 HTTP 回调")
		return nil
	})
	e := &PacketEvent{Packet: pkt}
	require.NotPanics(t, func() { _ = h.HandlePacket(context.Background(), e) })
	require.NoError(t, h.Close())
}

// TestReassembly_CloseIdempotent：多次 Close 不 panic。
func TestReassembly_CloseIdempotent(t *testing.T) {
	h := NewHTTPRequestHandler("idem", func(flow FlowKey, req *http.Request) error { return nil })
	require.NoError(t, h.Close())
	require.NoError(t, h.Close())
	require.NotPanics(t, func() { _ = h.Close() })
}

// TestReassembly_WithCapturer：端到端——HTTP handler 注册到 Capturer，跑完整 Capture。
func TestReassembly_WithCapturer(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: itsnot.fun\r\n\r\n")
	pcapData := buildRequestStreamPcap(t, payload, 6, false)
	src := newSourceFromBytes(t, "e2e.pcap", pcapData)

	var (
		mu  sync.Mutex
		got []collectedReq
	)
	h := NewHTTPRequestHandler("e2e", func(flow FlowKey, req *http.Request) error {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, collectedReq{Method: req.Method, RequestURI: req.URL.RequestURI(), Host: req.Host})
		return nil
	})

	c := NewCapturer(WithBufferSize(64), WithOverflowStrategy(OverflowBlock))
	defer c.Close()
	require.NoError(t, c.RegisterHandler(h))

	// Capture 会消费完所有包后返回；之后 Close handler flush 残留流。
	require.NoError(t, c.Capture(context.Background(), src, Target{}))
	require.NoError(t, h.Close())

	require.Len(t, got, 1, "端到端应解析出 1 个请求")
	assert.Equal(t, "GET", got[0].Method)
	assert.Equal(t, "itsnot.fun", got[0].Host)
}

// TestReassembly_HTTPResponse：验证 HTTP 响应重组（对称结构）。
func TestReassembly_HTTPResponse(t *testing.T) {
	// 构造一个 HTTP 响应流。
	body := "hello world"
	respPayload := []byte("HTTP/1.1 200 OK\r\nContent-Length: " + strconv.Itoa(len(body)) + "\r\nContent-Type: text/plain\r\n\r\n" + body)
	pcapData := buildRequestStreamPcap(t, respPayload, 9, false) // 复用构造（方向无所谓，只测重组）

	var (
		mu       sync.Mutex
		statuses []int
	)
	h := NewHTTPResponseHandler("http-resp", func(flow FlowKey, resp *http.Response) error {
		mu.Lock()
		defer mu.Unlock()
		statuses = append(statuses, resp.StatusCode)
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil
	})

	src := newSourceFromBytes(t, "resp.pcap", pcapData)
	feedHandler(t, src, h)
	require.NoError(t, h.Close())

	require.Len(t, statuses, 1, "应解析出 1 个响应")
	assert.Equal(t, 200, statuses[0])
}

// =============================================================================
// TCPStreamHandler 测试（通用 TCP 流重组，不限于 HTTP）。
//
// 用「4 字节大端长度前缀 + payload」的消息格式模拟典型二进制协议（如 protobuf over TCP），
// 验证使用者能从 io.Reader 读出重组后的完整消息流。
// =============================================================================

// framedMessage 构造一条「4字节长度前缀 + payload」的消息。
func framedMessage(payload string) []byte {
	buf := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(payload))) //nolint:gosec // G115: 测试数据受控
	copy(buf[4:], payload)
	return buf
}

// TestTCPStreamHandler_FramedMessages：跨多包的「长度前缀+消息」被正确重组并读出。
func TestTCPStreamHandler_FramedMessages(t *testing.T) {
	// 3 条消息，拼接后拆成小段（每段 7 字节），验证跨包重组 + 多消息循环读取。
	msgs := []string{"hello", "itsnot.fun", "protobuf works"}
	var allBytes []byte
	for _, m := range msgs {
		allBytes = append(allBytes, framedMessage(m)...)
	}
	pcapData := buildRequestStreamPcap(t, allBytes, 7, false)

	var (
		mu      sync.Mutex
		gotMsgs []string
	)
	h := NewTCPStreamHandler("framed", func(flow FlowKey, r io.Reader) error {
		// 使用者的典型逻辑：循环读长度前缀 + 消息体。
		for {
			var lenBuf [4]byte
			if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
				return err // EOF / 流结束
			}
			msgLen := binary.BigEndian.Uint32(lenBuf[:])
			payload := make([]byte, msgLen)
			if _, err := io.ReadFull(r, payload); err != nil {
				return err
			}
			mu.Lock()
			gotMsgs = append(gotMsgs, string(payload))
			mu.Unlock()
		}
	})

	src := newSourceFromBytes(t, "framed.pcap", pcapData)
	feedHandler(t, src, h)
	require.NoError(t, h.Close())

	require.Len(t, gotMsgs, 3, "应读出 3 条消息")
	assert.Equal(t, []string{"hello", "itsnot.fun", "protobuf works"}, gotMsgs)
}

// TestTCPStreamHandler_NonTCPPacket：UDP 包静默跳过。
func TestTCPStreamHandler_NonTCPPacket(t *testing.T) {
	eth := &layers.Ethernet{SrcMAC: []byte{1, 2, 3, 4, 5, 6}, DstMAC: []byte{6, 5, 4, 3, 2, 1}, EthernetType: layers.EthernetTypeIPv4}
	ip := &layers.IPv4{Version: 4, IHL: 5, TTL: 64, Protocol: layers.IPProtocolUDP, SrcIP: net.ParseIP("10.0.0.2"), DstIP: net.ParseIP("10.0.0.3")}
	udp := &layers.UDP{SrcPort: 5353, DstPort: 5353}
	require.NoError(t, udp.SetNetworkLayerForChecksum(ip))
	buf := gopacket.NewSerializeBuffer()
	require.NoError(t, gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}, eth, ip, udp, gopacket.Payload([]byte("dns"))))
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	h := NewTCPStreamHandler("udp-test", func(flow FlowKey, r io.Reader) error {
		t.Error("UDP 包不应触发流回调")
		return nil
	})
	e := &PacketEvent{Packet: pkt}
	require.NotPanics(t, func() { _ = h.HandlePacket(context.Background(), e) })
	require.NoError(t, h.Close())
}

// TestTCPStreamHandler_CloseIdempotent：多次 Close 不 panic。
func TestTCPStreamHandler_CloseIdempotent(t *testing.T) {
	h := NewTCPStreamHandler("idem", func(flow FlowKey, r io.Reader) error { return nil })
	require.NoError(t, h.Close())
	require.NoError(t, h.Close())
}

// TestTCPStreamHandler_WithCapturer：端到端——经 Capturer 完整链路。
func TestTCPStreamHandler_WithCapturer(t *testing.T) {
	msgs := []string{"e2e-1", "e2e-2"}
	var allBytes []byte
	for _, m := range msgs {
		allBytes = append(allBytes, framedMessage(m)...)
	}
	pcapData := buildRequestStreamPcap(t, allBytes, 6, false)
	src := newSourceFromBytes(t, "e2e-framed.pcap", pcapData)

	var (
		mu      sync.Mutex
		gotMsgs []string
	)
	h := NewTCPStreamHandler("e2e", func(flow FlowKey, r io.Reader) error {
		for {
			var lenBuf [4]byte
			if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
				return err
			}
			payload := make([]byte, binary.BigEndian.Uint32(lenBuf[:]))
			if _, err := io.ReadFull(r, payload); err != nil {
				return err
			}
			mu.Lock()
			gotMsgs = append(gotMsgs, string(payload))
			mu.Unlock()
		}
	})

	c := NewCapturer(WithBufferSize(64), WithOverflowStrategy(OverflowBlock))
	defer c.Close()
	require.NoError(t, c.RegisterHandler(h))

	require.NoError(t, c.Capture(context.Background(), src, Target{}))
	require.NoError(t, h.Close())

	assert.Equal(t, []string{"e2e-1", "e2e-2"}, gotMsgs)
}
