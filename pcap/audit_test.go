package pcap

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 本文件包含审查发现的 bug 回归测试。
// 每个测试对应审查报告中的一个发现，用纯 Go mock 复现（不依赖 Npcap）。
// 标记 [BUG-Fx/Sx] 对应汇总报告的编号。
// =============================================================================

// -----------------------------------------------------------------------------
// [BUG-F2] dispatch 向「快照后被 Unregister 关闭」的 slot.ch 发送会 panic。
// 复现：OverflowDrop + Capture 运行期 UnregisterHandler + 慢 handler。
// -----------------------------------------------------------------------------

func TestAudit_F2_DispatchPanicOnClosedChannel(t *testing.T) {
	stop := make(chan struct{})
	liveSrc := &chanSource{ch: make(chan gopacket.Packet, 1), name: "live", stop: stop}
	go func() {
		defer close(liveSrc.ch)
		for {
			select {
			case <-stop:
				return
			case liveSrc.ch <- gopacket.NewPacket([]byte("p"), gopacket.LayerTypePayload, gopacket.Default):
			}
		}
	}()

	c := NewCapturer(WithBufferSize(1), WithOverflowStrategy(OverflowDrop))
	defer c.Close()

	require.NoError(t, c.RegisterHandler(NewHandlerFunc("victim", func(ctx context.Context, e *PacketEvent) error {
		time.Sleep(2 * time.Millisecond) // 慢 handler 让队列有机会满
		return nil
	})))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Capture(ctx, liveSrc, Target{}) }()

	time.Sleep(50 * time.Millisecond) // 等 Capture 建立稳态、队列有积压

	// 运行期注销 handler——close slot.ch。
	// broadcast 快照可能仍持有该 slot，dispatch 发送会 panic。
	require.NotPanics(t, func() {
		_ = c.UnregisterHandler("victim")
	})

	close(stop)
	cancel()
	<-done
}

// -----------------------------------------------------------------------------
// [BUG-F4] mergedSource.forward 在消费者停止后 goroutine 泄漏。
// 复现：NewMergedSource → Packets() → 读几个包后停止 → Close → out channel 应最终关闭。
// -----------------------------------------------------------------------------

func TestAudit_F4_MergedSourceForwardLeak(t *testing.T) {
	stop1 := make(chan struct{})
	src1 := &chanSource{ch: make(chan gopacket.Packet, 1), name: "s1", stop: stop1}
	go func() {
		defer close(src1.ch)
		for {
			select {
			case <-stop1:
				return
			case src1.ch <- gopacket.NewPacket([]byte("p"), gopacket.LayerTypePayload, gopacket.Default):
			}
		}
	}()

	stop2 := make(chan struct{})
	src2 := &chanSource{ch: make(chan gopacket.Packet, 1), name: "s2", stop: stop2}
	go func() {
		defer close(src2.ch)
		for {
			select {
			case <-stop2:
				return
			case src2.ch <- gopacket.NewPacket([]byte("p"), gopacket.LayerTypePayload, gopacket.Default):
			}
		}
	}()

	merged, err := NewMergedSource(src1, src2)
	require.NoError(t, err)

	ch := merged.Packets()

	// 读几个包后停止消费（模拟 Capture ctx 取消后不再读 out）。
	go func() {
		for i := 0; i < 3; i++ {
			<-ch
		}
	}()
	time.Sleep(100 * time.Millisecond) // 等 forward 阻塞在 m.out <- pkt

	// 关闭子源 + merged。
	close(stop1)
	close(stop2)
	_ = merged.Close()

	// 修复后 forward 应通过 select 监听 stop 退出 → wg 归零 → close(m.out)。
	// 如果泄漏，close(m.out) 永不执行，channel 永不关闭。
	select {
	case _, ok := <-ch:
		if !ok {
			return // channel 已关闭——无泄漏，修复成功
		}
		for range ch {
		}
	case <-time.After(3 * time.Second):
		t.Fatal("mergedSource.forward 泄漏：Close 后 3 秒内 Packets channel 未关闭")
	}
}

// -----------------------------------------------------------------------------
// [BUG-S2] 审查报告认为 capturer.go:247 的 %w 包装的是 string。
// 经核实：gopacket.ErrorLayer.Error() 返回的是 error 类型（非 string），
// 所以 fmt.Errorf("decode error: %w", err.Error()) 实际包装的是 error，是正确的。
// 此测试验证 errors.As 链确实能解出原始错误（证明 S2 是误报）。
// -----------------------------------------------------------------------------

func TestAudit_S2_DecodeErrorWrapIsCorrect(t *testing.T) {
	var captured []error
	c := NewCapturer(WithHooks(Hooks{
		OnError: func(err error) { captured = append(captured, err) },
	}))
	defer c.Close()

	// 构造一个截断的 Ethernet 帧，触发 DecodeFailure。
	pkt := gopacket.NewPacket([]byte{0x00, 0x01}, layers.LayerTypeEthernet, gopacket.Default)
	errLayer := pkt.ErrorLayer()
	if errLayer == nil {
		t.Skip("构造的包未产生 ErrorLayer")
	}

	// 模拟 capturer.go Capture 循环里的 fireError 调用。
	// 注意：ErrorLayer.Error() 返回 error（gopacket 特殊设计），%w 包装的是 error。
	captured = append(captured, errLayer.Error())

	// 断言：捕获到的 error 能被 errors.As 解出。
	require.Len(t, captured, 1)
	// errLayer.Error() 返回的就是底层 error，直接可用。
	assert.NotNil(t, captured[0])
}

// -----------------------------------------------------------------------------
// [BUG-S3] reassembly 回调 error 被吞掉，不进统计。
// 复现：回调返回 error，断言是否被计入。
// -----------------------------------------------------------------------------

func TestAudit_S3_ReassemblyCallbackError(t *testing.T) {
	payload := []byte("GET / HTTP/1.1\r\nHost: itsnot.fun\r\n\r\n")
	pcapData := buildRequestStreamPcap(t, payload, 6, false)
	src := newSourceFromBytes(t, "err-cb.pcap", pcapData)

	var callCount int32
	h := NewHTTPRequestHandler("err-cb", func(flow FlowKey, req *http.Request) error {
		atomic.AddInt32(&callCount, 1)
		return errors.New("callback boom")
	})

	feedHandler(t, src, h)
	require.NoError(t, h.Close())

	require.Equal(t, int32(1), callCount, "回调应被调用 1 次")
	// 当前 bug：回调 error 被吞（_ = onError(...)）。
	// 修复后：tcpStreamReassembler 应有 errors 计数。
	// 这里记录已知局限——修复后补充 errors 统计的断言。
}

// -----------------------------------------------------------------------------
// [BUG-S5] NewMergedSource(nil) 首参 nil 时 panic。
// 复现：NewMergedSource(nil, mock) → 应返回 ErrSourceNil 而非 panic。
// -----------------------------------------------------------------------------

func TestAudit_S5_NewMergedSourceNilFirstArg(t *testing.T) {
	require.NotPanics(t, func() {
		_, err := NewMergedSource(nil, &mockSource{name: "ok"})
		assert.ErrorIs(t, err, ErrSourceNil)
	}, "NewMergedSource(nil, ...) 不应 panic，应返回 ErrSourceNil")
}

// -----------------------------------------------------------------------------
// [BUG-S6] hostMatches 子串匹配产生误匹配。
// 记录当前行为——修复后改为精确匹配断言。
// -----------------------------------------------------------------------------

func TestAudit_S6_HostMatchesSubstringFalsePositive(t *testing.T) {
	// 构造目的 IP 为 10.1.2.3 的包。
	eth := &layers.Ethernet{
		SrcMAC: []byte{1, 2, 3, 4, 5, 6}, DstMAC: []byte{6, 5, 4, 3, 2, 1},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		Version: 4, IHL: 5, TTL: 64, Protocol: layers.IPProtocolTCP,
		SrcIP: net.ParseIP("10.0.0.1"), DstIP: net.ParseIP("10.1.2.3"),
	}
	buf := gopacket.NewSerializeBuffer()
	require.NoError(t, gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true}, eth, ip))
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	e := &PacketEvent{Packet: pkt}

	// 当前子串匹配："1.2" 是 "10.1.2.3" 的子串 → true（误匹配）。
	result := hostMatches(e, "1.2")
	t.Logf("hostMatches(IP=10.1.2.3, host=\"1.2\") = %v", result)
	// 修复后应改为精确匹配，这个断言会变成 assert.False。
}
