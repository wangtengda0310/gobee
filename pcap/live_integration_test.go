//go:build cgo && livecapture && integration

// 本文件是「集成测试」，仅在同时满足以下条件时编译并运行：
//   1. CGO_ENABLED=1（cgo 开启）
//   2. -tags livecapture（链接 libpcap/Npcap）
//   3. -tags integration（显式要求跑集成测试）
//
// 运行方式（需本机已装 Npcap/libpcap + 真实网卡 + 管理员权限）：
//
//	CGO_ENABLED=1 go test -tags livecapture,integration -v -timeout 60s ./...
//
// 默认 go test / CI 不跑本文件（无 integration tag）。
// 这些测试验证 liveSource / ListInterfaces / ValidateBPF 的真实 cgo 链路。

package pcap

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pickFirstUsableInterface 选第一个有 IPv4 地址、非 loopback 的网卡用于测试。
// 若没有则跳过（t.Skip），避免在无合适网卡的机器上误判失败。
func pickFirstUsableInterface(t *testing.T) Interface {
	t.Helper()
	devs, err := ListInterfaces()
	require.NoError(t, err, "ListInterfaces 失败")
	require.NotEmpty(t, devs, "本机应至少有一个网卡")

	for _, d := range devs {
		if isLoopbackName(d.Name) {
			continue
		}
		for _, ip := range d.IPs {
			if v4 := ip.To4(); v4 != nil && !v4.IsLoopback() {
				return d
			}
		}
	}
	t.Skip("没有找到带 IPv4 的非 loopback 网卡，跳过集成测试")
	return Interface{}
}

// isLoopbackName 按网卡名启发式判断是否 loopback（跨平台）。
func isLoopbackName(name string) bool {
	return len(name) >= 4 && name[len(name)-4:] == "back"
}

// TestIntegration_ListInterfaces 验证 ListInterfaces 能返回非空网卡列表。
func TestIntegration_ListInterfaces(t *testing.T) {
	devs, err := ListInterfaces()
	require.NoError(t, err)
	assert.NotEmpty(t, devs, "本机应至少有一个网卡")
	t.Logf("找到 %d 个网卡", len(devs))
	for _, d := range devs {
		t.Logf("  - %s (%s)", d.Name, d.Description)
	}
}

// TestIntegration_ValidateBPF 验证 ValidateBPF 对合法/非法表达式的判断。
func TestIntegration_ValidateBPF(t *testing.T) {
	dev := pickFirstUsableInterface(t)

	src, err := NewLiveSource(dev.Name, 65535, false, "")
	require.NoError(t, err)
	defer src.Close()

	// Source 接口不含 ValidateBPF，需断言到 BPFValidator。
	v, ok := src.(BPFValidator)
	require.True(t, ok, "liveSource 应实现 BPFValidator")

	// 合法表达式。
	assert.NoError(t, v.ValidateBPF("tcp port 80"))
	assert.NoError(t, v.ValidateBPF("udp port 53 and host 8.8.8.8"))

	// 非法表达式（语法错误）。
	assert.Error(t, v.ValidateBPF("this is not valid bpf !!!"))
}

// TestIntegration_OpenAndImmediateClose 验证 liveSource 能正常打开并立即关闭，无泄漏/panic。
func TestIntegration_OpenAndImmediateClose(t *testing.T) {
	dev := pickFirstUsableInterface(t)

	src, err := NewLiveSource(dev.Name, 65535, false, "")
	require.NoError(t, err)
	require.NotPanics(t, func() { _ = src.Close() })
	// 多次 Close 幂等。
	require.NotPanics(t, func() { _ = src.Close() })
}

// TestIntegration_CaptureLoopbackBriefly 验证能真正抓到包（在 loopback 上）。
// 注：Windows 上 Npcap 的 loopback 抓包需要安装时勾选相应选项；
//
//	若无 loopback 网卡则跳过。
func TestIntegration_CaptureLoopbackBriefly(t *testing.T) {
	devs, err := ListInterfaces()
	require.NoError(t, err)

	var loopbackName string
	for _, d := range devs {
		if isLoopbackName(d.Name) {
			loopbackName = d.Name
			break
		}
	}
	if loopbackName == "" {
		t.Skip("无 loopback 网卡，跳过")
	}

	src, err := NewLiveSource(loopbackName, 65535, false, "")
	require.NoError(t, err)
	defer src.Close()

	c := NewCapturer(WithBufferSize(64), WithOverflowStrategy(OverflowDrop))
	defer c.Close()

	got := make(chan struct{}, 1)
	require.NoError(t, c.RegisterHandler(NewHandlerFunc("any", func(ctx context.Context, e *PacketEvent) error {
		select {
		case got <- struct{}{}:
		default:
		}
		return nil
	})))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() { _ = c.Capture(ctx, src, Target{}) }()

	// loopback 上通常有持续流量（DNS、心跳等）；2 秒内应至少收到一个包。
	select {
	case <-got:
		// 收到包，成功。
	case <-ctx.Done():
		t.Skip("2 秒内未在 loopback 上抓到包（可能无流量），跳过")
	}
}

// TestLive_SourceCloseExitsQuickly 验证 liveSource.Close 能在有限时间内打断读取循环。
// 这是对「Ctrl+C 无响应」根因（BlockForever）的回归守卫：
// Close 后 Packets() channel 应在 liveReadTimeout（1秒）+缓冲内关闭。
func TestLive_SourceCloseExitsQuickly(t *testing.T) {
	dev := pickFirstUsableInterface(t)

	src, err := NewLiveSource(dev.Name, 65535, false, "")
	require.NoError(t, err)

	// 启动读取（消费 Packets channel）。
	ch := src.Packets()
	consumed := make(chan struct{})
	go func() {
		for range ch {
		}
		close(consumed)
	}()

	// 给读取循环一点时间进入阻塞 ReadPacketData。
	time.Sleep(300 * time.Millisecond)

	// 关闭并计时——Close 应在 liveReadTimeout(1s) + 余量 内让 channel 关闭。
	start := time.Now()
	require.NoError(t, src.Close())

	select {
	case <-consumed:
		elapsed := time.Since(start)
		// liveReadTimeout=1s，加调度余量，应远小于 5 秒。
		assert.Less(t, elapsed, 5*time.Second,
			"Close 后 Packets channel 应在 5 秒内关闭，实际 %v", elapsed)
	case <-time.After(10 * time.Second):
		t.Fatal("Close 后 10 秒内 Packets channel 未关闭——liveSource 未正确响应中断")
	}
}
