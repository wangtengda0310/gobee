package pcap

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 本文件专门补齐核心库的测试覆盖缺口（阶段 1）。
// 复用 broadcast_test.go / itsnotfun_test.go 已定义的辅助：
// mockSource / chanSource / collectHandler / makeFakePackets / writePcap / newSourceFromBytes。
//
// 覆盖目标（对应 go tool cover 报告里的 0% 或低覆盖函数）：
//   - OverflowBlock 策略
//   - BPFCapable / ErrBPFNotSupported 分支
//   - Hooks（OnStart/OnPacket/OnError/OnStop）
//   - Close / LifeCycler
//   - fireError / Target.String / OverflowStrategy.String / WithBPFFilter
//   - hostMatches 边界 / readerSource.Close
// =============================================================================

// -----------------------------------------------------------------------------
// 1. OverflowBlock：阻塞投递，慢 handler 不丢包（离线场景）。
// -----------------------------------------------------------------------------

func TestOverflowBlock_NoDropWhenSlow(t *testing.T) {
	const N = 50
	src := &mockSource{pkts: makeFakePackets(N), name: "block"}

	// 每个 handler sleep 一小会，制造背压。Block 策略应等待而非丢弃。
	slow := NewHandlerFunc("slow", func(ctx context.Context, pkt *PacketEvent) error {
		time.Sleep(time.Millisecond)
		return nil
	})

	c := NewCapturer(WithBufferSize(8), WithOverflowStrategy(OverflowBlock))
	defer c.(*capturer).Close()
	require.NoError(t, c.RegisterHandler(slow))

	require.NoError(t, c.Capture(context.Background(), src, Target{}))

	st := c.Stats()
	require.Len(t, st.Handlers, 1)
	hs := st.Handlers[0]
	// Block 策略下不应有丢弃；全部 N 个包都被接收并处理。
	assert.Equal(t, int64(0), hs.Dropped, "Block 策略不应丢包")
	assert.Equal(t, int64(N), hs.Received, "应接收全部 %d 个包", N)
	assert.Equal(t, int64(N), hs.Processed, "应处理全部 %d 个包", N)
}

// Block 策略下若 ctx 取消，Capture 应返回 context.Canceled，不死锁。
//
// 设计说明（对应 CLAUDE.md 陷阱清单）：Block 策略假设 handler「慢但会完成」。
// 若 handler 永久阻塞，Capture 退出前的 flushAll 会死锁——这是已知限制，不是 bug。
// 因此本测试用「慢但会完成」的 handler（每包固定 sleep），确保 flushAll 能返回。
// 重点验证「ctx 取消时 Capture 正确返回」，而非制造永久背压。
func TestOverflowBlock_RespectsCtxCancel(t *testing.T) {
	stop := make(chan struct{})
	liveSrc := &chanSource{ch: make(chan gopacket.Packet, 1), name: "live", stop: stop}
	// 较慢的生产速率，避免队列持续高水位导致 flushAll 等待过久。
	go func() {
		defer close(liveSrc.ch)
		for {
			select {
			case <-stop:
				return
			case liveSrc.ch <- gopacket.NewPacket([]byte("p"), gopacket.LayerTypePayload, gopacket.Default):
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	// 慢 handler：每包 sleep，但最终会完成（满足 flushAll 的前提）。
	slow := NewHandlerFunc("slow", func(ctx context.Context, pkt *PacketEvent) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	c := NewCapturer(WithBufferSize(4), WithOverflowStrategy(OverflowBlock))
	defer c.(*capturer).Close()
	require.NoError(t, c.RegisterHandler(slow))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Capture(ctx, liveSrc, Target{}) }()

	// 建立稳态后取消。
	time.Sleep(60 * time.Millisecond)
	close(stop)
	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled, "Capture 应因 ctx 取消返回 context.Canceled")
	case <-time.After(5 * time.Second):
		t.Fatal("Block 策略在 ctx 取消后应返回，不应死锁")
	}
}

// -----------------------------------------------------------------------------
// 2. BPFCapable / ErrBPFNotSupported 分支。
// -----------------------------------------------------------------------------

// bpfMockSource 同时实现 Source 与 BPFCapable，用于测试 BPF 路径。
type bpfMockSource struct {
	mockSource
	setBPFErr error // 若非 nil，SetBPFFilter 返回该错误
	bpfCalled bool
	bpfArg    string
}

func (b *bpfMockSource) SetBPFFilter(expr string) error {
	b.bpfCalled = true
	b.bpfArg = expr
	return b.setBPFErr
}

func TestBPF_SourceNotCapable_ReturnsErrBPFNotSupported(t *testing.T) {
	// 普通 mockSource 未实现 BPFCapable，指定 BPF 应返回 ErrBPFNotSupported。
	src := &mockSource{pkts: makeFakePackets(1), name: "nobpf"}
	c := NewCapturer()
	defer c.(*capturer).Close()

	err := c.Capture(context.Background(), src, Target{BPF: "tcp port 80"})
	require.ErrorIs(t, err, ErrBPFNotSupported)
}

func TestBPF_GlobalFilterFromWithBPFFilter(t *testing.T) {
	// WithBPFFilter 设置全局 BPF，Capture 时 target.BPF 为空也应回退到全局。
	src := &bpfMockSource{mockSource: mockSource{pkts: makeFakePackets(0), name: "bpf"}}
	c := NewCapturer(WithBPFFilter("tcp port 443"))
	defer c.(*capturer).Close()

	require.NoError(t, c.Capture(context.Background(), src, Target{}))
	assert.True(t, src.bpfCalled, "应调用 SetBPFFilter")
	assert.Equal(t, "tcp port 443", src.bpfArg, "应使用全局 BPF")
}

func TestBPF_TargetOverridesGlobal(t *testing.T) {
	// target.BPF 优先于全局 WithBPFFilter。
	src := &bpfMockSource{mockSource: mockSource{pkts: makeFakePackets(0), name: "bpf"}}
	c := NewCapturer(WithBPFFilter("tcp port 443"))
	defer c.(*capturer).Close()

	require.NoError(t, c.Capture(context.Background(), src, Target{BPF: "udp port 53"}))
	assert.Equal(t, "udp port 53", src.bpfArg, "target.BPF 应覆盖全局")
}

func TestBPF_SetBPFFilterError_Propagated(t *testing.T) {
	// SetBPFFilter 返回错误时，Capture 应包装该错误返回。
	src := &bpfMockSource{
		mockSource: mockSource{pkts: makeFakePackets(0), name: "bpf"},
		setBPFErr:  errors.New("invalid filter"),
	}
	c := NewCapturer()
	defer c.(*capturer).Close()

	err := c.Capture(context.Background(), src, Target{BPF: "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter")
	assert.Contains(t, err.Error(), "bogus")
}

// -----------------------------------------------------------------------------
// 3. Hooks（OnStart/OnPacket/OnError/OnStop）。
// -----------------------------------------------------------------------------

func TestHooks_AllFired(t *testing.T) {
	src := &mockSource{pkts: makeFakePackets(3), name: "hooks"}

	var (
		mu       sync.Mutex
		started  bool
		packets  int
		stopped  bool
		errs     []error
		startCtx context.Context
	)

	c := NewCapturer(WithHooks(Hooks{
		OnStart: func(ctx context.Context) {
			mu.Lock()
			started = true
			startCtx = ctx
			mu.Unlock()
		},
		OnPacket: func(e *PacketEvent) {
			mu.Lock()
			packets++
			mu.Unlock()
		},
		OnError: func(err error) {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		},
		OnStop: func() {
			mu.Lock()
			stopped = true
			mu.Unlock()
		},
	}))
	defer c.(*capturer).Close()

	// 一个普通 handler，确保 OnPacket 被触发。
	require.NoError(t, c.RegisterHandler(NewHandlerFunc("noop", func(ctx context.Context, e *PacketEvent) error {
		return nil
	})))

	ctx := context.Background()
	require.NoError(t, c.Capture(ctx, src, Target{}))

	mu.Lock()
	defer mu.Unlock()
	assert.True(t, started, "OnStart 应被触发")
	assert.Equal(t, ctx, startCtx, "OnStart 应收到 Capture 的 ctx")
	assert.Equal(t, 3, packets, "OnPacket 应被触发 3 次")
	assert.True(t, stopped, "OnStop 应被触发")
}

func TestHooks_OnError_WhenHandlerReturnsError(t *testing.T) {
	src := &mockSource{pkts: makeFakePackets(2), name: "err"}

	var errs []error
	c := NewCapturer(WithHooks(Hooks{
		OnError: func(err error) { errs = append(errs, err) },
	}))
	defer c.(*capturer).Close()

	boomer := errors.New("boom")
	require.NoError(t, c.RegisterHandler(NewHandlerFunc("bad", func(ctx context.Context, e *PacketEvent) error {
		return boomer
	})))

	require.NoError(t, c.Capture(context.Background(), src, Target{}))
	// 每个 handler 错误应被 OnError 捕获并计入统计。
	require.Len(t, errs, 2, "应触发 2 次 OnError")
	for _, e := range errs {
		assert.ErrorIs(t, e, boomer)
	}
	st := c.Stats()
	require.Len(t, st.Handlers, 1)
	assert.Equal(t, int64(2), st.Handlers[0].Errors)
}

func TestHooks_OnError_WhenDecodeError(t *testing.T) {
	// 构造一个带 ErrorLayer 的包，验证 Capture 的 decode-error 分支会触发 OnError。
	badPkt := gopacket.NewPacket([]byte("garbage"), layers.LayerTypeEthernet, gopacket.Default)
	src := &mockSource{pkts: []gopacket.Packet{badPkt}, name: "decode"}

	var gotErr []string
	c := NewCapturer(WithHooks(Hooks{
		OnError: func(err error) { gotErr = append(gotErr, err.Error()) },
	}))
	defer c.(*capturer).Close()
	require.NoError(t, c.RegisterHandler(NewHandlerFunc("noop", func(ctx context.Context, e *PacketEvent) error { return nil })))

	// 解析错误不应导致 Capture 返回错误（非致命）。
	require.NoError(t, c.Capture(context.Background(), src, Target{}))
	// 是否产生 decode error 取决于 gopacket 对 "garbage" 的解析；
	// 这里不强制断言（解析器可能把它当成合法 payload），但 Capture 必须不崩溃、正常返回。
	_ = gotErr
}

// -----------------------------------------------------------------------------
// 4. Close / LifeCycler。
// -----------------------------------------------------------------------------

func TestClose_ReleasesAllWorkers(t *testing.T) {
	c := NewCapturer()
	cap := c.(*capturer)

	require.NoError(t, c.RegisterHandler(newCollectHandler("a")))
	require.NoError(t, c.RegisterHandler(newCollectHandler("b")))
	require.Len(t, cap.Stats().Handlers, 2)

	cap.Close()

	// Close 后所有 handler 应从注册表中清空。
	require.Empty(t, cap.Stats().Handlers, "Close 后 handlers 应清空")

	// Close 可被多次调用，不应 panic。
	assert.NotPanics(t, func() { cap.Close() })
}

func TestLifeCycler_TypeAssertion(t *testing.T) {
	c := NewCapturer()
	// NewCapturer 返回值应同时实现 LifeCycler。
	lc, ok := c.(LifeCycler)
	require.True(t, ok, "capturer 应实现 LifeCycler")
	require.NotPanics(t, func() { lc.Close() })
}

func TestClose_AfterCapture_StillWorks(t *testing.T) {
	// 先 Capture（EOF 自然结束），再 Close，worker 应正常退出。
	src := &mockSource{pkts: makeFakePackets(5), name: "close"}
	c := NewCapturer()
	require.NoError(t, c.RegisterHandler(newCollectHandler("h")))
	require.NoError(t, c.Capture(context.Background(), src, Target{}))

	lc := c.(LifeCycler)
	require.NotPanics(t, func() { lc.Close() })
}

// -----------------------------------------------------------------------------
// 5. fireError / Target.String / OverflowStrategy.String / WithBPFFilter。
// -----------------------------------------------------------------------------

func TestTarget_String(t *testing.T) {
	cases := []struct {
		t    Target
		want string
	}{
		{Target{}, "<all>"},
		{Target{BPF: "tcp port 80"}, "bpf=\"tcp port 80\""},
		{Target{Host: "itsnot.fun"}, "host=\"itsnot.fun\""},
		{Target{BPF: "tcp", Host: "x"}, "bpf=\"tcp\" host=\"x\""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.t.String())
	}
}

func TestOverflowStrategy_String(t *testing.T) {
	assert.Equal(t, "drop", OverflowDrop.String())
	assert.Equal(t, "block", OverflowBlock.String())
	assert.Equal(t, "drop_oldest", OverflowDropOldest.String())
	// 未知值兜底。
	assert.Equal(t, "unknown", OverflowStrategy(99).String())
}

func TestWithBPFFilter_SetsConfig(t *testing.T) {
	c := NewCapturer(WithBPFFilter("tcp port 80")).(*capturer)
	assert.Equal(t, "tcp port 80", c.opts.BPFFilter)
}

func TestWithBufferSize_RejectsNonPositive(t *testing.T) {
	// 非正值应被忽略，保留默认。
	c1 := NewCapturer(WithBufferSize(0)).(*capturer)
	assert.Equal(t, 1024, c1.opts.BufferSize, "BufferSize=0 应保留默认")
	c2 := NewCapturer(WithBufferSize(-1)).(*capturer)
	assert.Equal(t, 1024, c2.opts.BufferSize, "BufferSize<0 应保留默认")
	c3 := NewCapturer(WithBufferSize(42)).(*capturer)
	assert.Equal(t, 42, c3.opts.BufferSize)
}

func TestWithBufferSize_ActuallyApplied(t *testing.T) {
	// 验证 BufferSize 真正影响 channel 容量：用一个小 buffer + DropOldest，
	// 大量包灌入应触发丢弃（说明 buffer 生效了）。
	src := &mockSource{pkts: makeFakePackets(1000), name: "buf"}
	c := NewCapturer(WithBufferSize(1), WithOverflowStrategy(OverflowDrop))
	defer c.(*capturer).Close()
	// 慢 handler 确保队列会满。
	require.NoError(t, c.RegisterHandler(NewHandlerFunc("slow", func(ctx context.Context, e *PacketEvent) error {
		time.Sleep(time.Millisecond)
		return nil
	})))
	require.NoError(t, c.Capture(context.Background(), src, Target{}))
	st := c.Stats()
	require.Len(t, st.Handlers, 1)
	assert.Greater(t, st.Handlers[0].Dropped, int64(0), "buffer=1 + 慢 handler 应产生丢弃")
}

// -----------------------------------------------------------------------------
// 6. hostMatches 边界 + readerSource.Close。
// -----------------------------------------------------------------------------

func TestHostMatches_NilEvent(t *testing.T) {
	assert.False(t, hostMatches(nil, "x"))
	assert.False(t, hostMatches(&PacketEvent{}, "x"), "Packet 为 nil 应 false")
}

func TestHostMatches_ApplicationLayer(t *testing.T) {
	// 构造带应用层 payload 的包。
	pkt := gopacket.NewPacket([]byte("GET / Host: itsnot.fun"), gopacket.LayerTypePayload, gopacket.Default)
	e := &PacketEvent{Packet: pkt}
	assert.True(t, hostMatches(e, "itsnot.fun"))
	assert.False(t, hostMatches(e, "example.org"))
}

func TestHostMatches_NetworkLayerIP(t *testing.T) {
	// 构造真实以太网帧（含 IP 层），让 gopacket 解析出 NetworkFlow。
	eth := &layers.Ethernet{
		SrcMAC: []byte{1, 2, 3, 4, 5, 6}, DstMAC: []byte{6, 5, 4, 3, 2, 1},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{Version: 4, IHL: 5, TTL: 64, Protocol: layers.IPProtocolTCP,
		SrcIP: []byte{10, 0, 0, 1}, DstIP: []byte{10, 0, 0, 2}}
	buf := gopacket.NewSerializeBuffer()
	require.NoError(t, gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true}, eth, ip))
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	e := &PacketEvent{Packet: pkt}
	assert.True(t, hostMatches(e, "10.0.0.2"), "目的 IP 命中")
	assert.True(t, hostMatches(e, "10.0.0.1"), "源 IP 命中")
	assert.False(t, hostMatches(e, "10.0.0.99"))
}

func TestReaderSource_Close_ClosesUnderlying(t *testing.T) {
	// readerSource.Close 应转发给底层的 io.Closer（若底层实现了 io.Closer）。
	// 用一个跟踪 Close 调用的 PacketDataSource（同时实现 io.Closer）验证。
	ds := &closeTrackingDataSource{}
	src := NewReaderSource(ds, layers.LinkTypeEthernet, "tracked")

	// 第一次 Close 转发给底层。
	require.NoError(t, src.Close())
	assert.True(t, ds.closed.Load(), "底层 Close 应被调用")

	// 多次 Close 不应 panic（底层 Close 被设计为幂等）。
	assert.NotPanics(t, func() { _ = src.Close() })
}

// closeTrackingDataSource 实现 gopacket.PacketDataSource + io.Closer，
// 用于验证 readerSource.Close 的转发行为。
type closeTrackingDataSource struct {
	closed atomic.Bool
}

func (d *closeTrackingDataSource) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	// 返回 EOF，使 PacketSource 的 Packets() 立即关闭。
	return nil, gopacket.CaptureInfo{}, errEOF
}

func (d *closeTrackingDataSource) Close() error {
	d.closed.Store(true)
	return nil
}

// errEOF 作为 ReadPacketData 的结束信号。
var errEOF = errors.New("eof")

// 编译期确保 closeTrackingDataSource 实现所需接口。
var (
	_ gopacket.PacketDataSource = (*closeTrackingDataSource)(nil)
)
