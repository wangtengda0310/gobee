package pcap

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopacket/gopacket"        // 纯 Go（Packet/NewPacket），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（LinkType），不需 Npcap
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MergedSource 单元测试（纯 Go，不依赖网卡/Npcap）。
// 复用 broadcast_test.go 定义的 mockSource / chanSource / makeFakePackets。
// =============================================================================

func TestMergedSource_EmptySourcesError(t *testing.T) {
	_, err := NewMergedSource()
	require.ErrorIs(t, err, ErrNoSources)
}

func TestMergedSource_NilSourceError(t *testing.T) {
	_, err := NewMergedSource(&mockSource{name: "ok"}, nil)
	require.ErrorIs(t, err, ErrSourceNil)
}

func TestMergedSource_InconsistentLinkType(t *testing.T) {
	// 构造两个不同 LinkType 的 mock。
	a := &mockSource{name: "a"} // LinkType=Ethernet（mockSource 默认）
	b := &mockSourceWithLink{name: "b", link: layers.LinkTypeLinuxSLL}
	_, err := NewMergedSource(a, b)
	require.ErrorIs(t, err, ErrInconsistentLinkType)
}

func TestMergedSource_SingleSourcePassThrough(t *testing.T) {
	const N = 10
	src := &mockSource{pkts: makeFakePackets(N), name: "only"}
	merged, err := NewMergedSource(src)
	require.NoError(t, err)
	defer merged.Close()

	got := drainAll(t, merged.Packets())
	assert.Len(t, got, N, "单子源合并应原样转发全部包")
}

func TestMergedSource_FanInAllPackets(t *testing.T) {
	// 3 个子源，各 N 个包，合并后应收到 3N 个、无丢无重复。
	const perSrc = 20
	srcs := []Source{
		&mockSource{pkts: makeFakePackets(perSrc), name: "a"},
		&mockSource{pkts: makeFakePackets(perSrc), name: "b"},
		&mockSource{pkts: makeFakePackets(perSrc), name: "c"},
	}
	merged, err := NewMergedSource(srcs...)
	require.NoError(t, err)
	defer merged.Close()

	got := drainAll(t, merged.Packets())
	assert.Len(t, got, perSrc*3, "三个子源应合并出 %d 个包", perSrc*3)
}

func TestMergedSource_PartialEOFDoesNotCloseMerge(t *testing.T) {
	// 一个子源 EOF 后，合并 channel 不应立即关闭，另一个子源的包仍能收到。
	fast := &mockSource{pkts: makeFakePackets(5), name: "fast"} // 会很快 EOF
	slowStop := make(chan struct{})
	slow := &chanSource{ch: make(chan gopacket.Packet, 10), name: "slow", stop: slowStop}
	// slow 持续产包直到 stop。
	go func() {
		defer close(slow.ch)
		for {
			select {
			case <-slowStop:
				return
			case slow.ch <- gopacket.NewPacket([]byte("slow"), gopacket.LayerTypePayload, gopacket.Default):
				time.Sleep(time.Millisecond)
			}
		}
	}()

	merged, err := NewMergedSource(fast, slow)
	require.NoError(t, err)

	ch := merged.Packets()

	// 核心断言：fast（5 包）EOF 后，合并 channel 仍保持开启，并能继续收到包。
	// 收够 5（fast）+ 3（slow）= 8 个包，证明 partial EOF 未关闭合并 channel。
	received := 0
	timeout := time.After(2 * time.Second)
	for received < 8 {
		select {
		case _, ok := <-ch:
			if !ok {
				t.Fatalf("合并 channel 在 fast EOF 后不应关闭（已收到 %d 个包）", received)
			}
			received++
		case <-timeout:
			t.Fatalf("超时：只收到 %d 个包（期望 >=8，证明 partial EOF 后仍能收包）", received)
		}
	}
	// 走到这里说明：fast 的 5 个包 + 至少 3 个 slow 的包，合并 channel 仍开着。
	assert.GreaterOrEqual(t, received, 8, "partial EOF 后应仍持续收到包")

	// 关闭 slow，合并 channel 应随之关闭。
	close(slowStop)
	// 排空剩余，确认最终关闭。
	timeout2 := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return // 合并 channel 已关闭，符合预期
			}
		case <-timeout2:
			t.Fatal("所有子源 EOF 后合并 channel 应关闭")
		}
	}
}

func TestMergedSource_CloseClosesAllSources(t *testing.T) {
	a := &closeTrackingSource2{name: "a"}
	b := &closeTrackingSource2{name: "b"}
	merged, err := NewMergedSource(a, b)
	require.NoError(t, err)

	require.NoError(t, merged.Close())
	assert.True(t, a.closed.Load(), "子源 a 应被 Close")
	assert.True(t, b.closed.Load(), "子源 b 应被 Close")

	// 多次 Close 幂等，不 panic。
	assert.NotPanics(t, func() { _ = merged.Close() })
}

func TestMergedSource_WithCapturer(t *testing.T) {
	// 端到端：MergedSource 喂给 Capturer，验证多源包都被 handler 消费。
	const perSrc = 15
	merged, err := NewMergedSource(
		&mockSource{pkts: makeFakePackets(perSrc), name: "a"},
		&mockSource{pkts: makeFakePackets(perSrc), name: "b"},
	)
	require.NoError(t, err)
	defer merged.Close()

	c := NewCapturer(WithBufferSize(64), WithOverflowStrategy(OverflowBlock))
	defer c.Close()

	var count atomic.Int64
	require.NoError(t, c.RegisterHandler(NewHandlerFunc("counter", func(ctx context.Context, e *PacketEvent) error {
		count.Add(1)
		return nil
	})))

	require.NoError(t, c.Capture(context.Background(), merged, Target{}))
	assert.Equal(t, int64(perSrc*2), count.Load(), "handler 应收到两个子源的全部包")
	assert.Equal(t, "merged:2", merged.String())
}

// -----------------------------------------------------------------------------
// 测试辅助
// -----------------------------------------------------------------------------

// drainAll 读取一个 packet channel 直到关闭，返回收到的所有包。
func drainAll(t *testing.T, ch chan gopacket.Packet) []gopacket.Packet {
	t.Helper()
	var got []gopacket.Packet
	timeout := time.After(3 * time.Second)
	for {
		select {
		case pkt, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, pkt)
		case <-timeout:
			t.Fatalf("drainAll 超时，已收到 %d 个包", len(got))
			return got
		}
	}
}

// mockSourceWithLink 是可指定 LinkType 的 mock，用于测 LinkType 校验。
type mockSourceWithLink struct {
	name string
	link layers.LinkType
}

func (m *mockSourceWithLink) Packets() chan gopacket.Packet { return make(chan gopacket.Packet) }
func (m *mockSourceWithLink) LinkType() layers.LinkType     { return m.link }
func (m *mockSourceWithLink) Close() error                  { return nil }
func (m *mockSourceWithLink) String() string                { return "mocklink:" + m.name }

// closeTrackingSource2 跟踪 Close 调用的 Source（与 coverage_test.go 的 closeTrackingDataSource
// 不同：那个是 PacketDataSource，这个是 Source 接口）。
type closeTrackingSource2 struct {
	name   string
	closed atomic.Bool
	pkts   chan gopacket.Packet
}

func (s *closeTrackingSource2) Packets() chan gopacket.Packet { return s.pkts }
func (s *closeTrackingSource2) LinkType() layers.LinkType     { return layers.LinkTypeEthernet }
func (s *closeTrackingSource2) Close() error {
	s.closed.Store(true)
	// 注意：不 close pkts channel——mergedSource 只检查 Close 是否被调用，
	// 本测试不通过 Packets() 驱动转发。
	return nil
}
func (s *closeTrackingSource2) String() string { return "tracked:" + s.name }

var _ Source = (*mockSourceWithLink)(nil)
var _ Source = (*closeTrackingSource2)(nil)
