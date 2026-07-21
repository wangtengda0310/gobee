package pcap

import (
	"context"
	"fmt"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
)

// PacketHandler 处理抓到的单个数据包。
//
// 并发约定：
//   - 同一个 handler 的 HandlePacket 调用是串行的（由各自的 worker goroutine 消费），
//     因此单个 handler 实现本身不需要加锁。
//   - 不同 handler 之间是并行的，它们若共享外部资源，需自行保证安全。
//
// Name 用于注册时的唯一标识：重名注册会返回错误，可用 UnregisterHandler(name) 注销。
type PacketHandler interface {
	// HandlePacket 处理一个包。返回 error 仅用于统计，不会中断抓包。
	HandlePacket(ctx context.Context, pkt *PacketEvent) error

	// Name 返回 handler 的唯一名称。
	Name() string
}

// HandlerFunc 把普通函数适配成 PacketHandler，便于函数式注册。
// 用法：capturer.RegisterHandler(pcap.HandlerFunc("printer", func(ctx, pkt) error {...}))
type HandlerFunc struct {
	name string
	fn   func(ctx context.Context, pkt *PacketEvent) error
}

// NewHandlerFunc 创建一个具名处理函数包装。
func NewHandlerFunc(name string, fn func(ctx context.Context, pkt *PacketEvent) error) *HandlerFunc {
	return &HandlerFunc{name: name, fn: fn}
}

// HandlePacket 实现 PacketHandler。
func (h *HandlerFunc) HandlePacket(ctx context.Context, pkt *PacketEvent) error {
	return h.fn(ctx, pkt)
}

// Name 实现 PacketHandler。
func (h *HandlerFunc) Name() string { return h.name }

// Source 描述抓包的数据来源。
//
// 它把底层 gopacket.PacketDataSource（可能是 pcap 文件、网卡实时句柄、
// 或内存 mock）抽象成一个统一的只读通道来源。
// 这样核心抓包逻辑与具体的抓包后端（纯 Go / cgo）完全解耦，
// 单元测试可注入任意 mock 数据源。
type Source interface {
	// Packets 返回一个会持续产出 gopacket.Packet 的只读通道。
	// 当底层读取结束（EOF）或出错时，该通道被关闭。
	Packets() chan gopacket.Packet

	// LinkType 返回数据源的链路层类型，供解码器使用。
	LinkType() layers.LinkType

	// Close 释放底层资源（文件句柄 / 网卡句柄）。
	// 可被多次调用。
	Close() error

	// String 返回数据源的可读标识（如网卡名、文件路径），用于 PacketEvent.Source。
	String() string
}

// Target 描述对抓包结果的过滤目标。
//
// 支持两种语义（可组合）：
//   - BPF 表达式（如 "tcp port 80 and host itsnot.fun"）：在数据源层过滤。
//   - 应用层 host 过滤（如 "itsnot.fun"）：在分发前按应用层 Host 头/目标 IP 过滤，
//     便于在 BPF 不可用（如读已抓好的 pcap）时做内容过滤。
//
// 空值（零值 Target）表示不过滤，投递所有包。
type Target struct {
	// BPF 是 BPF 过滤表达式。为空表示不使用 BPF。
	// 注意：并非所有 Source 都支持 BPF（如内存 mock / 已解码的 pcap 文件
	// 只能重放，无法回退到内核过滤）。不支持时 Capture 会返回 ErrBPFNotSupported。
	BPF string

	// Host 是应用层 host 过滤。为空表示不过滤。
	// 命中规则：包的应用层（HTTP Host 头）或网络层目的/源 IP 字符串包含本字段。
	Host string
}

// Capturer 抓包器。
//
// 这是本包对外暴露的核心接口，包含题目要求的两个方法：
//
//   - Capture(source, target)：接收数据源与过滤目标，开始抓包并把每个包
//     广播给所有已注册的处理函数；ctx 取消时优雅退出。
//
//   - RegisterHandler(handler)：注册处理函数，抓到的包会广播给它。
//     支持运行期动态增删，并发安全。
//
// 另外补充了 UnregisterHandler / Stats 两个必要方法（详见 README 的必要功能分析）。
type Capturer interface {
	// RegisterHandler 注册一个处理函数，抓到的包会广播给它。
	// 重名注册返回 ErrHandlerExists。可在 Capture 前或运行期调用，并发安全。
	RegisterHandler(handler PacketHandler) error

	// UnregisterHandler 按名称注销一个处理函数，并排空其队列。
	// 不存在则返回 ErrHandlerNotFound。
	UnregisterHandler(name string) error

	// Capture 接收 source（数据源）与 target（过滤目标），开始抓包。
	// 阻塞直到 ctx 被取消、数据源 EOF 或发生致命错误。
	// 返回前会等待所有处理函数排空在途包。
	Capture(ctx context.Context, source Source, target Target) error

	// Stats 返回当前抓包器与各处理函数的运行时统计，并发安全。
	Stats() *CaptureStats
}

// LifeCycler 是 Capturer 的可选扩展，用于显式终结抓包器生命周期。
// capturer（NewCapturer 的返回值）同时实现 Capturer 和 LifeCycler；
// 调用方可通过类型断言或直接断言到本接口来调用 Close。
// 不把 Close 放进 Capturer 是为了让接口保持最小（题目只要求两个核心方法）。
type LifeCycler interface {
	// Close 关闭所有处理函数的 worker 并清空注册表，释放资源。可被多次调用。
	Close()
}

// String 返回 Target 的可读表示。
func (t Target) String() string {
	if t.BPF == "" && t.Host == "" {
		return "<all>"
	}
	if t.BPF != "" && t.Host != "" {
		return fmt.Sprintf("bpf=%q host=%q", t.BPF, t.Host)
	}
	if t.BPF != "" {
		return fmt.Sprintf("bpf=%q", t.BPF)
	}
	return fmt.Sprintf("host=%q", t.Host)
}
