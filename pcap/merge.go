package pcap

import (
	"fmt"
	"sync"

	"github.com/gopacket/gopacket"        // 纯 Go（Packet），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（LinkType），不需 Npcap
)

// =============================================================================
// MergedSource：多网卡 fan-in
// =============================================================================
// 把多个 Source 合并为一个逻辑 Source，喂给同一个 Capturer。
// 这样多个网卡的流量会被同一个抓包循环消费、广播给同一组处理函数。
//
// 设计要点：
//   - 不改 Capturer 接口：MergedSource 本身实现 Source，对 Capturer 透明。
//   - 要求所有子 Source 的 LinkType 一致：gopacket 用单一解码器解析合并后的包，
//     链路层类型不同会解码错乱。不一致时 NewMergedSource 返回 ErrInconsistentLinkType。
//   - 任一子 Source 的 Packet channel 关闭（EOF/出错），不影响其它子源继续投递；
//     只有当【所有】子源都关闭后，合并 channel 才关闭。
//   - PacketEvent.Source 字段在合并场景下统一填 "merged:n"（n 为子源总数），
//     无法精确标识某个包来自哪个子源（gopacket.Packet 不可变，附加上下文成本高）。
//     若需区分子源，handler 可依据包自身的网络层信息（IP/端口）区分。
//     精细的 per-packet 子源标记为后续 TODO。

// mergedSourceBufferPerSource 是每个子源在合并 channel 中的预缓冲系数。
const mergedSourceBufferPerSource = 16

// =============================================================================

// mergedSource 把多个 Source 合并为一个。
type mergedSource struct {
	sources []Source
	link    layers.LinkType

	// out 是合并后的 packet channel，由 start() 启动的 goroutine 填充。
	// 用 once 保证 start 只执行一次（Packets 可能被多次调用，虽然 Capturer 只调一次）。
	out  chan gopacket.Packet
	once sync.Once

	// closeOnce 保证 Close 幂等。
	closeOnce sync.Once

	// stop 由 Close 关闭，通知所有 forward goroutine 退出。
	// 解决「消费者停止后 forward 阻塞在 m.out <- pkt」的泄漏问题（BUG-F4）。
	stop chan struct{}

	// wg 跟踪所有转发 goroutine，用于 Close 时等待它们退出。
	wg sync.WaitGroup
}

// NewMergedSource 把多个 Source 合并为一个。
//
// 用法：
//
//	merged, err := pcap.NewMergedSource(src1, src2, src3)
//	if err != nil { ... }
//	defer merged.Close()
//	capturer.Capture(ctx, merged, target)
//
// 要求：
//   - 至少传入 1 个 Source，否则返回 ErrNoSources。
//   - 所有子 Source 的 LinkType() 必须一致，否则返回 ErrInconsistentLinkType。
//
// 返回的 Source 可像普通 Source 一样传给 Capturer。
// 合并后 PacketEvent.Source 统一为 "merged:n"。
func NewMergedSource(sources ...Source) (Source, error) {
	if len(sources) == 0 {
		return nil, ErrNoSources
	}
	// 先校验 nil，再取 LinkType（避免 sources[0] 为 nil 时 panic）。
	for i, s := range sources {
		if s == nil {
			return nil, fmt.Errorf("%w: source at index %d is nil", ErrSourceNil, i)
		}
	}
	link := sources[0].LinkType()
	for i, s := range sources {
		if lt := s.LinkType(); lt != link {
			return nil, fmt.Errorf("%w: source[%d]=%v source[0]=%v",
				ErrInconsistentLinkType, i, lt, link)
		}
	}
	return &mergedSource{
		sources: sources,
		link:    link,
		out:     make(chan gopacket.Packet, len(sources)*mergedSourceBufferPerSource),
		stop:    make(chan struct{}),
	}, nil
}

// Packets 实现 Source。
// 首次调用时启动所有转发 goroutine；后续调用返回同一个 channel。
// 当所有子 Source 的 channel 都关闭后，合并 channel 被关闭。
func (m *mergedSource) Packets() chan gopacket.Packet {
	m.once.Do(func() { m.start() })
	return m.out
}

// start 启动每个子 Source 的转发 goroutine。
// 每个 goroutine range 一个子源的 Packets() channel，把 packet 转投到 m.out。
// 所有 goroutine 退出后关闭 m.out。
func (m *mergedSource) start() {
	for _, src := range m.sources {
		m.wg.Add(1)
		go m.forward(src)
	}
	// 所有转发 goroutine 退出后关闭合并 channel。
	go func() {
		m.wg.Wait()
		close(m.out)
	}()
}

// forward 把单个子 Source 的 packet 转投到合并 channel。
// 用 select 监听 stop，避免消费者停止后阻塞在 m.out <- pkt（BUG-F4 修复）。
func (m *mergedSource) forward(src Source) {
	defer m.wg.Done()
	for pkt := range src.Packets() {
		select {
		case m.out <- pkt:
		case <-m.stop:
			return
		}
	}
}

// LinkType 实现 Source。返回所有子源统一的链路层类型（构造时校验过）。
func (m *mergedSource) LinkType() layers.LinkType { return m.link }

// Close 实现 Source。关闭所有子 Source 并等待转发 goroutine 退出，幂等。
// 关闭 stop channel 让阻塞在 m.out <- pkt 的 forward 退出（BUG-F4 修复）。
func (m *mergedSource) Close() error {
	m.closeOnce.Do(func() {
		close(m.stop) // 通知 forward 退出（即使消费者已停止，forward 也能退出）
		for _, s := range m.sources {
			_ = s.Close() // 关闭子源，让 range 退出
		}
		m.wg.Wait() // 等所有 forward 退出 → start 的协调 goroutine 随后 close(m.out)
	})
	return nil
}

// String 实现 Source。
// 合并场景下 PacketEvent.Source 会填本返回值，标识「这是合并源」。
func (m *mergedSource) String() string {
	return fmt.Sprintf("merged:%d", len(m.sources))
}

// 编译期保证 mergedSource 实现 Source 接口。
var _ Source = (*mergedSource)(nil)
