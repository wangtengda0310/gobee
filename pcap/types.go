package pcap

import (
	"context"
	"time"

	"github.com/gopacket/gopacket"        // 纯 Go（Packet/Flow），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（LinkType），不需 Npcap
)

// OverflowStrategy 定义当某个处理函数的缓冲队列已满（背压）时的应对策略。
//
// 抓包循环是热路径，绝对不能因为某个处理函数慢就阻塞整个抓包流程，
// 否则会造成真实流量丢失。不同业务场景对「丢包」的容忍度不同，
// 因此把策略做成可配置。
type OverflowStrategy int

const (
	// OverflowDrop 丢弃新包（默认）。
	// 适合抓包热路径：缓冲队列满时直接丢弃当前正在投递的包，绝不阻塞抓包。
	// 代价是会丢失部分数据，但保证了抓包的实时性。
	OverflowDrop OverflowStrategy = iota

	// OverflowBlock 阻塞等待队列空位。
	// 适合离线分析（读 pcap 文件）等「一个都不能少」的场景：
	// 投递会阻塞直到队列有空位或 ctx 被取消。
	// 注意：实时抓包场景下慎用，慢处理函数会导致抓包循环阻塞、丢真实流量。
	OverflowBlock

	// OverflowDropOldest 丢弃最旧的包（滑动窗口）。
	// 适合只关心「最新流量」的监控场景：队列满时弹出队首（最旧）的包再投递新包。
	// 保证了最新的包一定能进队列，但历史包会被淘汰。
	OverflowDropOldest
)

// String 返回策略名称，便于日志和统计输出。
func (s OverflowStrategy) String() string {
	switch s {
	case OverflowDrop:
		return "drop"
	case OverflowBlock:
		return "block"
	case OverflowDropOldest:
		return "drop_oldest"
	default:
		return "unknown"
	}
}

// PacketEvent 是投递给处理函数的包事件。
// 它在原始 gopacket.Packet 基础上做了轻量封装，
// 预解析了常用字段，避免每个处理函数都重复解码。
type PacketEvent struct {
	// Packet 原始 gopacket 包，处理函数可据此做深度解析。
	Packet gopacket.Packet

	// Timestamp 抓包时间戳。
	Timestamp time.Time

	// Length 包的原始长度（线上长度，可能大于截获长度）。
	Length int

	// NetworkFlow 网络层流标识（src/dst IP），没有则为 nil。
	NetworkFlow gopacket.Flow

	// TransportFlow 传输层流标识（src/dst port），没有则为 nil。
	TransportFlow gopacket.Flow

	// LinkType 链路层类型（Ethernet / Loopback / ...）。
	LinkType layers.LinkType

	// Source 描述本包来自哪个数据源（网卡名 / 文件路径），便于多源聚合时区分。
	Source string
}

// HandlerStats 是单个处理函数的运行时统计。
// 所有计数字段均通过 atomic 操作更新，可在任意时刻并发安全地读取。
type HandlerStats struct {
	// Name 处理函数名称。
	Name string

	// Received 该处理函数的队列累计收到的包数（投递成功）。
	Received int64

	// Processed 该处理函数已经处理完毕的包数。
	Processed int64

	// Dropped 因队列满而被丢弃的包数（仅 OverflowDrop / OverflowDropOldest 会产生）。
	Dropped int64

	// Errors 处理函数返回错误的次数。
	Errors int64
}

// CaptureStats 是整个抓包器的聚合统计。
type CaptureStats struct {
	// Handlers 每个处理函数的独立统计。
	Handlers []HandlerStats

	// Captured 抓包循环累计读取并分发的包数。
	Captured int64

	// StartedAt 本次 Capture 开始时间。
	StartedAt time.Time
}

// Config 抓包器配置，通过 functional option 构造。
type Config struct {
	// BufferSize 每个处理函数的有界队列容量。默认 1024。
	BufferSize int

	// Overflow 背压（队列满）时的应对策略。默认 OverflowDrop。
	Overflow OverflowStrategy

	// BPFFilter BPF 过滤表达式（如 "tcp port 80"），在数据源层过滤，
	// 大幅降低投递压力。空表示不过滤。
	BPFFilter string

	// Hooks 生命周期回调，所有字段可选（nil 表示不回调）。
	Hooks Hooks
}

// defaultConfig 返回带合理默认值的配置。
func defaultConfig() *Config {
	return &Config{
		BufferSize: 1024,
		Overflow:   OverflowDrop,
		BPFFilter:  "",
	}
}

// Hooks 是一组可选的生命周期回调函数字段。
// 沿用本仓库 agent 模块的回调风格：在调用点做 nil 检查后再触发。
//
// 所有回调都在抓包主 goroutine 中同步执行，因此实现应尽量轻量，
// 耗时操作请转发到其他 goroutine。
type Hooks struct {
	// OnStart 在 Capture 开始抓包前触发。
	OnStart func(ctx context.Context)

	// OnPacket 在每个包分发之前触发（在过载策略判定之前）。
	// 可用于全局计数或采样。
	OnPacket func(pkt *PacketEvent)

	// OnError 在出现非致命错误（解析失败、处理函数报错、丢包等）时触发。
	// 致命错误会直接从 Capture 返回，不经过这里。
	OnError func(err error)

	// OnStop 在 Capture 即将返回前触发。
	OnStop func()
}
