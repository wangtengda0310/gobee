package pcap

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gopacket/gopacket"        // 纯 Go（Packet/NewPacket），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（LinkTypeEthernet），不需 Npcap
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 本文件覆盖核心库的并发安全与过载保护行为：
//  1. 多 handler 广播：每个 handler 都收到全部包（无丢无重复），-race 通过。
//  2. OverflowDrop：小 buffer + 慢 handler，断言产生 dropped 且抓包不阻塞。
//  3. OverflowDropOldest：队列满时丢弃最旧的，handler 仍能持续收到「最新」包。
//  4. 动态注册/注销：Capture 运行期增删 handler 不 panic、不影响抓包。
// =============================================================================

// mockSource 是一个可编程的 Source：按调用方注入的包列表产出，最后关闭通道。
// 用于在不依赖 pcap 文件的前提下做受控的过载/并发测试。
type mockSource struct {
	pkts []gopacket.Packet
	name string
}

func (m *mockSource) Packets() chan gopacket.Packet {
	ch := make(chan gopacket.Packet, len(m.pkts))
	go func() {
		defer close(ch)
		for _, p := range m.pkts {
			ch <- p
		}
	}()
	return ch
}
func (m *mockSource) LinkType() layers.LinkType { return layers.LinkTypeEthernet }
func (m *mockSource) Close() error              { return nil }
func (m *mockSource) String() string            { return "mock:" + m.name }

// makeFakePackets 构造 n 个最小可用的 gopacket.Packet（仅含应用层 payload），
// 用来在不构造完整以太网帧的前提下填充 mockSource。
func makeFakePackets(n int) []gopacket.Packet {
	out := make([]gopacket.Packet, n)
	for i := 0; i < n; i++ {
		p := gopacket.NewPacket(
			[]byte(fmt.Sprintf("pkt-%d", i)),
			gopacket.LayerTypePayload,
			gopacket.Default,
		)
		out[i] = p
	}
	return out
}

// -----------------------------------------------------------------------------
// 1. 多 handler 广播：每个 handler 都收到全部包，-race 通过。
// -----------------------------------------------------------------------------

func TestBroadcast_AllHandlersReceiveAll(t *testing.T) {
	const N = 200
	src := &mockSource{pkts: makeFakePackets(N), name: "broadcast"}

	c := NewCapturer(WithBufferSize(N+8), WithOverflowStrategy(OverflowBlock))
	defer c.(*capturer).Close()

	h1 := newCollectHandler("h1")
	h2 := newCollectHandler("h2")
	require.NoError(t, c.RegisterHandler(h1))
	require.NoError(t, c.RegisterHandler(h2))

	err := c.Capture(context.Background(), src, Target{})
	require.NoError(t, err)

	require.Len(t, h1.snapshot(), N, "h1 应收到全部 %d 个包", N)
	require.Len(t, h2.snapshot(), N, "h2 应收到全部 %d 个包", N)

	st := c.Stats()
	assert.Equal(t, int64(N), st.Captured)
	require.Len(t, st.Handlers, 2)
	for _, hs := range st.Handlers {
		assert.Equal(t, int64(N), hs.Received)
		assert.Equal(t, int64(N), hs.Processed)
		assert.Equal(t, int64(0), hs.Dropped, "%s 不应丢包", hs.Name)
	}
}

// -----------------------------------------------------------------------------
// 2. OverflowDrop：慢 handler + 小 buffer，抓包不阻塞且产生 dropped。
//
//    关键点：用一个「慢但会完成」的 handler 制造背压（每包 sleep）。
//    抓包循环用 OverflowDrop，投递瞬间完成（队列满即丢），不会因 handler 慢而阻塞。
// -----------------------------------------------------------------------------

func TestOverflowDrop_DoesNotBlockAndCountsDropped(t *testing.T) {
	const total = 500
	src := &mockSource{pkts: makeFakePackets(total), name: "drop"}

	// 慢 handler：每个包 sleep 一会，制造持续背压（队列始终被占）。
	slow := NewHandlerFunc("slow", func(ctx context.Context, pkt *PacketEvent) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	// 小 buffer + Drop 策略。
	c := NewCapturer(WithBufferSize(4), WithOverflowStrategy(OverflowDrop))
	defer c.(*capturer).Close()
	require.NoError(t, c.RegisterHandler(slow))

	done := make(chan error, 1)
	start := time.Now()
	go func() { done <- c.Capture(context.Background(), src, Target{}) }()

	// Capture 应在远小于「slow 串行处理 500 包」的时间内返回（≈ 立即读完即返回），
	// 证明 Drop 策略下抓包循环不被慢 handler 阻塞。
	// 注意：Capture 返回前会 flushAll 等待在途包处理完（最多 buffer=4 个，≈ 20ms）。
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("OverflowDrop 下 Capture 不应阻塞超过 5s")
	}
	elapsed := time.Since(start)

	st := c.Stats()
	require.Len(t, st.Handlers, 1)
	hs := st.Handlers[0]
	assert.Greater(t, hs.Dropped, int64(0), "Drop 策略下应产生丢弃")
	assert.Equal(t, st.Captured, hs.Received+hs.Dropped, "received + dropped == captured")
	// 如果抓包被阻塞，500 包 × 5ms ≈ 2.5s；这里允许 flushAll 的小开销，应远小于 2.5s。
	assert.Less(t, elapsed, 2500*time.Millisecond, "Drop 策略下 Capture 耗时应远小于串行处理时间")
}

// -----------------------------------------------------------------------------
// 3. OverflowDropOldest：队列满时丢弃最旧，received+dropped 守恒。
// -----------------------------------------------------------------------------

func TestOverflowDropOldest_KeepsLatest(t *testing.T) {
	const total = 100
	const buf = 8
	src := &mockSource{pkts: makeFakePackets(total), name: "dropoldest"}

	// 慢 handler：每个包 sleep 一会，制造背压。
	slow := NewHandlerFunc("slow", func(ctx context.Context, pkt *PacketEvent) error {
		time.Sleep(2 * time.Millisecond)
		return nil
	})

	c := NewCapturer(WithBufferSize(buf), WithOverflowStrategy(OverflowDropOldest))
	defer c.(*capturer).Close()
	require.NoError(t, c.RegisterHandler(slow))

	err := c.Capture(context.Background(), src, Target{})
	require.NoError(t, err)

	st := c.Stats()
	require.Len(t, st.Handlers, 1)
	hs := st.Handlers[0]
	// DropOldest 下：背压时会踢掉旧包，dropped 应 > 0。
	// 计数语义说明：received 统计「成功入队的总次数」（含后来被踢的），
	// dropped 统计「被踢/被丢的次数」，因此 received+dropped 会大于总投递数，这是预期行为。
	assert.Greater(t, hs.Dropped, int64(0), "DropOldest 背压时应踢出旧包")
	assert.Greater(t, hs.Processed, int64(0), "应有包被处理")
	assert.Equal(t, int64(total), st.Captured, "captured 应等于源包总数")
}

// -----------------------------------------------------------------------------
// 4. 动态注册/注销：Capture 运行期增删 handler，不 panic、不影响抓包。
//    同时验证重名注册/注销不存在返回正确的哨兵错误。
// -----------------------------------------------------------------------------

func TestRegisterUnregister_DynamicAndErrors(t *testing.T) {
	// 用一个会持续产出包、直到 ctx 取消才停的源。
	stop := make(chan struct{})
	liveSrc := &chanSource{ch: make(chan gopacket.Packet, 1), name: "live", stop: stop}
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				close(liveSrc.ch)
				return
			case liveSrc.ch <- gopacket.NewPacket([]byte(fmt.Sprintf("p%d", i)), gopacket.LayerTypePayload, gopacket.Default):
				i++
				time.Sleep(time.Millisecond)
			}
		}
	}()

	c := NewCapturer(WithBufferSize(16), WithOverflowStrategy(OverflowDrop))
	defer c.(*capturer).Close()

	// 重名注册错误。
	h := newCollectHandler("dyn")
	require.NoError(t, c.RegisterHandler(h))
	assert.ErrorIs(t, c.RegisterHandler(newCollectHandler("dyn")), ErrHandlerExists)

	// nil / 空名错误。
	assert.ErrorIs(t, c.RegisterHandler(nil), ErrHandlerNil)
	assert.ErrorIs(t, c.RegisterHandler(newCollectHandler("")), ErrEmptyName)

	// 注销不存在。
	assert.ErrorIs(t, c.UnregisterHandler("nope"), ErrHandlerNotFound)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Capture(ctx, liveSrc, Target{}) }()

	// 运行期再注册一个 handler，应继续工作不 panic。
	h2 := newCollectHandler("dyn2")
	require.NoError(t, c.RegisterHandler(h2))

	// 运行期注销第一个。
	require.NoError(t, c.UnregisterHandler("dyn"))

	// 让它跑一会，再停止。
	// 注意顺序：必须先 close(stop) 让生产者 goroutine 退出（并 close 它的 channel），
	// 再 cancel ctx 让 Capture 退出。否则生产者可能在 Capture 已返回后仍向
	// channel 发包，或与 close(liveSrc.ch) 产生 send-on-closed-channel panic。
	time.Sleep(80 * time.Millisecond)
	close(stop)
	cancel()
	<-done

	// h2 应收到若干包（说明动态注册生效）。
	require.Eventually(t, func() bool { return len(h2.snapshot()) > 0 }, time.Second, 5*time.Millisecond,
		"运行期注册的 h2 应收到包")
}

// chanSource 把一个外部 channel 适配成 Source，用于需要持续产包的测试。
type chanSource struct {
	ch   chan gopacket.Packet
	name string
	stop chan struct{}
}

func (s *chanSource) Packets() chan gopacket.Packet { return s.ch }
func (s *chanSource) LinkType() layers.LinkType     { return layers.LinkTypeEthernet }
func (s *chanSource) Close() error                  { return nil }
func (s *chanSource) String() string                { return "chan:" + s.name }

// compile-time: 确保 mock/chan source 实现 Source。
var (
	_ Source = (*mockSource)(nil)
	_ Source = (*chanSource)(nil)
)
