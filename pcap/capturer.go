package pcap

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopacket/gopacket" // 纯 Go（Packet/Layer/Flow），不需 Npcap
)

// =============================================================================
// 关于「数据包数量巨大、处理函数来不及消费」的应对设计
// =============================================================================
// 这是经典的多生产者-多消费者过载（背压）问题。核心矛盾：
//   - 抓包循环是热路径，绝对不能阻塞（阻塞 = 丢真实流量）。
//   - 处理函数可能很慢（写库、网络、复杂解析），二者速度不匹配。
//
// 本实现采用分层解耦策略：
//
//  1. 广播解耦：抓包 goroutine 只负责读包 + 投递到「广播分发器」，
//     不直接调用处理函数。分发器为「每个处理函数」维护一条独立的有界队列，
//     互不影响——某个慢处理函数不会拖垮其他处理函数，也不会拖垮抓包。
//
//  2. 每个处理函数独占一个 worker goroutine 串行消费自己的队列
//     （for pkt := range ch），处理函数之间天然并行。
//
//  3. 三种可配置的过载策略（OverflowStrategy）：
//       - OverflowDrop      队列满则丢弃新包，绝不阻塞抓包（默认，适合实时）。
//       - OverflowBlock     阻塞等待空位（适合离线、不能丢包）。
//       - OverflowDropOldest 丢最旧的包（滑动窗口，适合只看最新）。
//
//  4. 全量统计（Stats）：每个队列的 Received / Processed / Dropped / Errors，
//     全部用 atomic 维护，便于监控告警。
//
// 结论：抓包永不阻塞；处理函数互相隔离；过载行为可预测、可观测、可配置。
// =============================================================================

// handlerSlot 是单个处理函数的运行时上下文：队列 + worker + 统计。
// 每个 RegisterHandler 创建一个独立的 slot，worker 在注册时启动并持续运行，
// 直到 UnregisterHandler 或 Close() 关闭其队列。
type handlerSlot struct {
	handler PacketHandler
	ch      chan *PacketEvent

	// workerWG 用于等待该 slot 的 worker 退出。
	// 注册时启动 worker；UnregisterHandler / Close 时关闭 channel 触发退出。
	workerWG sync.WaitGroup

	// flushWG 用于实现 Capture 退出时的「优雅 flush」。
	// flushAll 向队列投递一个 nil 哨兵（Add(1)），worker 见到 nil 时 Done(1)。
	// 当哨兵被消费，说明在此之前的所有真实包都已被处理完。
	// 这比「逐包 Add/Done 配对」更稳健——尤其当 OverflowDropOldest 会把包踢出队列时，
	// 哨兵方案不依赖"每个包恰好 Done 一次"的脆弱不变量。
	flushWG sync.WaitGroup

	// 统计字段，全部 atomic。
	received  atomic.Int64
	processed atomic.Int64
	dropped   atomic.Int64
	errors    atomic.Int64
}

// stats 快照当前 slot 的统计。持有 capturer 的读锁后调用。
func (s *handlerSlot) stats() HandlerStats {
	return HandlerStats{
		Name:      s.handler.Name(),
		Received:  s.received.Load(),
		Processed: s.processed.Load(),
		Dropped:   s.dropped.Load(),
		Errors:    s.errors.Load(),
	}
}

// capturer 是 Capturer 接口的默认实现。
type capturer struct {
	opts *Config

	mu       sync.RWMutex
	handlers map[string]*handlerSlot

	// captureCtx 持有当前 Capture 的 ctx，供 worker 传给 HandlePacket。
	// Capture 开始时 set，结束时 clear（nil 时 worker 用 context.Background()）。
	// 用 atomic.Pointer 保证 worker 无锁读取。
	captureCtx atomic.Pointer[context.Context]

	// captured 记录本次/历次 Capture 累计分发数（跨多次 Capture 累加）。
	captured atomic.Int64
	// startedAt 记录最近一次 Capture 开始时间。
	startedAt atomic.Value // time.Time
}

// NewCapturer 创建一个抓包器，opts 用于覆盖默认配置。
func NewCapturer(opts ...Option) Capturer {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	c := &capturer{
		opts:     cfg,
		handlers: make(map[string]*handlerSlot),
	}
	return c
}

// RegisterHandler 注册一个处理函数。
// - 重名返回 ErrHandlerExists。
// - 注册后立即启动该 handler 的 worker goroutine，开始消费队列。
// - 可在 Capture 前或 Capture 运行期调用，并发安全。
func (c *capturer) RegisterHandler(handler PacketHandler) error {
	if handler == nil {
		return ErrHandlerNil
	}
	name := handler.Name()
	if name == "" {
		return ErrEmptyName
	}

	c.mu.Lock()
	if _, ok := c.handlers[name]; ok {
		c.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrHandlerExists, name)
	}

	slot := &handlerSlot{
		handler: handler,
		ch:      make(chan *PacketEvent, c.opts.BufferSize),
	}
	c.handlers[name] = slot
	c.mu.Unlock()

	// 在锁外启动 worker，避免潜在的锁顺序问题。
	slot.workerWG.Add(1)
	go c.runWorker(slot)
	return nil
}

// UnregisterHandler 注销一个处理函数。
// - 关闭其队列、等待 worker 排空在途包后退出。
// - 不存在返回 ErrHandlerNotFound。
// - 并发安全，可在运行期调用。
func (c *capturer) UnregisterHandler(name string) error {
	c.mu.Lock()
	slot, ok := c.handlers[name]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrHandlerNotFound, name)
	}
	delete(c.handlers, name)
	c.mu.Unlock()

	// 关闭队列触发 worker 退出；等待在途包处理完。
	close(slot.ch)
	slot.workerWG.Wait()
	return nil
}

// runWorker 是单个处理函数的消费循环。
// 从自己的队列读取包，串行调用 handler.HandlePacket。
// 遇到 nil 包表示 flush 哨兵：不调用 handler，只 Done flushWG，表示「到此前所有真实包已处理完」。
// 队列关闭（UnregisterHandler / Close）后 range 自然退出。
func (c *capturer) runWorker(slot *handlerSlot) {
	defer slot.workerWG.Done()
	for pkt := range slot.ch {
		if pkt == nil {
			// flush 哨兵：见 flushAll 的注释。不调用 handler，只放行 flush。
			slot.flushWG.Done()
			continue
		}
		// 取出当前 Capture 的 ctx 传给 handler，使 handler 能感知取消。
		// 若不在 Capture 期间（如 Unregister 排空），回退到 context.Background()。
		ctx := context.Background()
		if captured := c.captureCtx.Load(); captured != nil {
			ctx = *captured
		}
		if err := slot.handler.HandlePacket(ctx, pkt); err != nil {
			slot.errors.Add(1)
			c.fireError(fmt.Errorf("handler %q: %w", slot.handler.Name(), err))
		}
		slot.processed.Add(1)
	}
}

// Capture 接收 source 与 target，开始抓包并广播给所有已注册处理函数。
//
// 阻塞直到：ctx 取消 / 数据源 EOF / 发生致命错误。
// 返回前会 flush 所有处理函数的在途包（等待它们被处理完），保证优雅退出。
// 处理函数不会被注销，可在后续 Capture 中继续复用；真正释放请调用 Close()。
//
// 注意：本方法会复用传入 target 与全局 WithBPFFilter 二者中较严格的过滤组合。
// 若 Source 不支持 BPF 但 target/opts 指定了 BPF，返回 ErrBPFNotSupported。
func (c *capturer) Capture(ctx context.Context, source Source, target Target) error {
	if source == nil {
		return ErrSourceNil
	}

	// 合并全局 BPF 与 target.BPF（target 优先，若为空则回退全局）。
	bpf := target.BPF
	if bpf == "" {
		bpf = c.opts.BPFFilter
	}
	if bpf != "" {
		// 尝试在 Source 上设置 BPF；不支持则报错，避免静默放过非预期流量。
		if bs, ok := source.(BPFCapable); ok {
			if err := bs.SetBPFFilter(bpf); err != nil {
				return fmt.Errorf("pcap: set bpf filter %q: %w", bpf, err)
			}
		} else {
			return fmt.Errorf("%w: %q", ErrBPFNotSupported, bpf)
		}
	}

	c.captured.Store(0)
	c.startedAt.Store(time.Now())

	// 把本次 Capture 的 ctx 暴露给 worker，使 HandlePacket 能感知取消。
	// Capture 返回后清空，避免旧 ctx 泄漏到后续 Capture 或非 Capture 期间的调用。
	c.captureCtx.Store(&ctx)
	defer c.captureCtx.Store(nil)

	if c.opts.Hooks.OnStart != nil {
		c.opts.Hooks.OnStart(ctx)
	}
	if c.opts.Hooks.OnStop != nil {
		defer c.opts.Hooks.OnStop()
	}

	// 启动抓包循环。这里直接 range source.Packets() —— 该 channel 在 EOF/出错时
	// 由 gopacket 关闭，循环自然退出。
	pkts := source.Packets()
	for {
		select {
		case <-ctx.Done():
			// 被外部取消：返回前 flush 在途包（不注销 handler，便于复用）。
			c.flushAll()
			return ctx.Err()
		case pkt, ok := <-pkts:
			if !ok {
				// 数据源结束（EOF / 错误）。gopacket 把读取错误记录在 Packet.Error() 上。
				c.flushAll()
				return nil
			}
			if err := pkt.ErrorLayer(); err != nil {
				// 解析错误不致命，上报后继续。
				c.fireError(fmt.Errorf("decode error: %w", err.Error()))
			}

			event := toEvent(pkt, source)
			if !c.matchTarget(event, target) {
				continue
			}

			if c.opts.Hooks.OnPacket != nil {
				c.opts.Hooks.OnPacket(event)
			}
			c.captured.Add(1)
			c.broadcast(ctx, event)
		}
	}
}

// broadcast 把一个包按当前所有已注册处理函数的快照逐个投递。
// 使用读锁拷贝快照，避免长持锁阻塞 Register/Unregister。
// 投递本身按 c.opts.Overflow 策略进行，保证抓包循环不因慢 handler 阻塞（除 OverflowBlock 外）。
func (c *capturer) broadcast(ctx context.Context, event *PacketEvent) {
	// 「快照 + 锁外处理」模式（sync.Map.Range 同款思路）：
	// 先在读锁内把 handlers 的指针拷到切片（纳秒级），立刻释放锁，再在无锁状态下遍历投递。
	//
	// 为什么不合并成一次持锁遍历（for + dispatch）？
	//   1. dispatch 可能阻塞（OverflowBlock 等队列空位），持锁期间阻塞会卡死
	//      RegisterHandler/UnregisterHandler（它们需要写锁），导致整个 capturer 僵死。
	//   2. dispatch 间接调用 HandlePacket，若 handler 内部回调 capturer（如 Stats），
	//      会因重入锁而死锁。
	//
	// 代价：多一次微小内存分配（拷指针，handler 通常个位数）。相比持锁阻塞的风险，完全值得。
	c.mu.RLock()
	slots := make([]*handlerSlot, 0, len(c.handlers))
	for _, s := range c.handlers {
		slots = append(slots, s)
	}
	c.mu.RUnlock()

	for _, s := range slots {
		c.dispatch(ctx, s, event)
	}
}

// dispatch 按 OverflowStrategy 把包投递到单个 slot 的队列。
// 这是过载保护的核心。统计上只区分 received（成功入队）与 dropped（被丢）。
func (c *capturer) dispatch(ctx context.Context, s *handlerSlot, event *PacketEvent) {
	switch c.opts.Overflow {
	case OverflowBlock:
		// 阻塞投递，但仍然响应 ctx 取消，避免无法退出。
		select {
		case s.ch <- event:
			s.received.Add(1)
		case <-ctx.Done():
			s.dropped.Add(1)
		}
	case OverflowDropOldest:
		select {
		case s.ch <- event:
			s.received.Add(1)
		default:
			// 队列满：丢最旧的，给最新的腾位置。
			select {
			case <-s.ch:
				s.dropped.Add(1)
			default:
				// 极端竞态：刚满又被别的并发消费空了，此时无需丢。
			}
			// 再次尝试投递（仍可能失败，失败则计入 dropped）。
			select {
			case s.ch <- event:
				s.received.Add(1)
			default:
				s.dropped.Add(1)
			}
		}
	default: // OverflowDrop
		select {
		case s.ch <- event:
			s.received.Add(1)
		default:
			s.dropped.Add(1)
		}
	}
}

// flushAll 等待所有 slot 处理完「当前已投递但尚未消费」的包。
// 用于 Capture 退出前的优雅收尾。
//
// 实现机制（哨兵方案，替代早期版本的 inFlight WaitGroup）：
//   - 对每个 slot，先 Add(1) 到 flushWG，再向其 channel 投递一个 nil 哨兵；
//   - worker 在 range 中遇到 nil 包即 Done(1)，表示「在此之前的真实包都已处理完」；
//   - 本函数 Wait 所有 slot 的 flushWG 归零。
//
// 为什么不用「逐包 Add/Done 配对」？因为 OverflowDropOldest 会把已入队的包踢出队列，
// 导致"每个包恰好被 Done 一次"的约束极难维护（实测会 panic: negative WaitGroup counter）。
// 哨兵方案把"等待清空"这一事件从"逐包追踪"降为"一次水位标记"，正确性显然。
//
// 注意：
//   - 投递哨兵时，若 slot 恰被 UnregisterHandler 关闭了 channel，send 会 panic；
//     此处用 recover 容错（已注销的 handler 视为无需 flush）。
//   - 它不注销 handler，worker 不退出，handler 可跨多次 Capture 复用。
//   - 真正关闭 worker 由 UnregisterHandler / Close 负责。
func (c *capturer) flushAll() {
	// 快照模式：同 broadcast（读锁内拷指针 → 锁外操作）。
	// 投递哨兵可能阻塞（等队列空位），不能持锁。
	c.mu.RLock()
	slots := make([]*handlerSlot, 0, len(c.handlers))
	for _, s := range c.handlers {
		slots = append(slots, s)
	}
	c.mu.RUnlock()

	for _, s := range slots {
		s.flushWG.Add(1)
		// 哨兵走阻塞投递：worker 持续消费会腾出空位。
		// 若 channel 已被 UnregisterHandler 关闭，send 会 panic；
		// 此时 worker 已退出、不会再 Done，必须手动 Done 平衡刚才的 Add，否则 Wait 死锁。
		if !sendSentinel(s) {
			s.flushWG.Done()
		}
	}
	for _, s := range slots {
		s.flushWG.Wait()
	}
}

// sendSentinel 向 slot 的队列投递一个 nil 哨兵包。
// 返回 true 表示投递成功（worker 将消费它并 Done flushWG）；
// 返回 false 表示 channel 已关闭（slot 已被 UnregisterHandler 注销），
// 此时调用方必须手动 Done flushWG 以平衡前置的 Add。
func sendSentinel(s *handlerSlot) (sent bool) {
	defer func() {
		if r := recover(); r != nil {
			// send on closed channel：worker 已退出。
			sent = false
		}
	}()
	s.ch <- nil
	return true
}

// Close 关闭所有处理函数的 worker 并清空注册表。
// 这是对 Capturer 接口的扩展方法（接口未包含），用于 capturer 生命周期终结时释放资源。
// 可被多次调用。
func (c *capturer) Close() {
	// 快照模式（同 broadcast/flushAll），但用写锁：边拷指针边从 map 删除。
	// close(s.ch) 在锁外做——close 后 worker 会退出并可能访问 slot，无需持锁保护。
	c.mu.Lock()
	slots := make([]*handlerSlot, 0, len(c.handlers))
	for name, s := range c.handlers {
		slots = append(slots, s)
		delete(c.handlers, name)
	}
	c.mu.Unlock()

	for _, s := range slots {
		close(s.ch)
		s.workerWG.Wait()
	}
}

// Stats 实现 Capturer。
func (c *capturer) Stats() *CaptureStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := &CaptureStats{
		Handlers: make([]HandlerStats, 0, len(c.handlers)),
		Captured: c.captured.Load(),
	}
	if v, ok := c.startedAt.Load().(time.Time); ok {
		out.StartedAt = v
	}
	for _, s := range c.handlers {
		out.Handlers = append(out.Handlers, s.stats())
	}
	return out
}

// fireError 触发 OnError 钩子，若无钩子则忽略（非致命错误不影响抓包）。
func (c *capturer) fireError(err error) {
	if c.opts.Hooks.OnError != nil {
		c.opts.Hooks.OnError(err)
	}
}

// matchTarget 判断包是否命中 Target 的用户态过滤条件（目前只支持 Host）。
// BPF 过滤已在 Source 层完成，这里只做应用层/网络层的二次过滤。
func (c *capturer) matchTarget(event *PacketEvent, target Target) bool {
	if target.Host == "" {
		return true
	}
	return hostMatches(event, target.Host)
}

// hostMatches 检查包的应用层（明文 HTTP Host 头）或网络层 IP 是否包含目标 host。
// 用于在 BPF 不可用时（如读已抓好的 pcap 文件）做内容过滤。
// 说明：这里有意只做「payload 字面子串匹配」，不强行依赖 layers.HTTP 解码器，
// 这样对未被 gopacket 解析为结构化 HTTP 的明文 payload 同样有效，过滤更稳健。
func hostMatches(event *PacketEvent, host string) bool {
	if event == nil || event.Packet == nil {
		return false
	}
	pkt := event.Packet

	// 1) 应用层 payload 字面包含 host（命中 "Host: itsnot.fun" 等）。
	if app := pkt.ApplicationLayer(); app != nil {
		if payload := app.Payload(); len(payload) > 0 {
			if strings.Contains(string(payload), host) {
				return true
			}
		}
	}

	// 2) 网络层 src/dst IP 字符串包含 host。
	if net := pkt.NetworkLayer(); net != nil {
		flow := net.NetworkFlow()
		src, dst := flow.Src().String(), flow.Dst().String()
		if src != "" || dst != "" {
			if strings.Contains(src, host) || strings.Contains(dst, host) {
				return true
			}
		}
	}
	return false
}

// toEvent 把 gopacket.Packet 转成轻量的 PacketEvent，预解析常用字段。
func toEvent(pkt gopacket.Packet, source Source) *PacketEvent {
	event := &PacketEvent{
		Packet:   pkt,
		Source:   source.String(),
		LinkType: source.LinkType(),
	}
	if ci := pkt.Metadata().CaptureInfo; true {
		event.Timestamp = ci.Timestamp
		event.Length = ci.Length
	}
	if net := pkt.NetworkLayer(); net != nil {
		event.NetworkFlow = net.NetworkFlow()
	}
	if tp := pkt.TransportLayer(); tp != nil {
		event.TransportFlow = tp.TransportFlow()
	}
	return event
}

// BPFCapable 由支持 BPF 过滤的 Source 实现（通常是实时网卡句柄）。
// readerSource（离线 pcap）不实现该接口，因此对其使用 BPF 会返回 ErrBPFNotSupported。
type BPFCapable interface {
	SetBPFFilter(expr string) error
}

// BPFValidator 由能预校验 BPF 表达式的 Source 实现（通常是实时网卡句柄）。
// 与 BPFCapable 分离：校验 BPF 和应用 BPF 是两个能力，某些 Source 可能只支持其一。
// liveSource 同时实现两者：先用 ValidateBPF 检查表达式合法性，再 SetBPFFilter 应用。
//
// 用法：
//
//	if v, ok := src.(pcap.BPFValidator); ok {
//	    if err := v.ValidateBPF(expr); err != nil { return err }
//	}
type BPFValidator interface {
	ValidateBPF(expr string) error
}

// Ensure capturer satisfies Capturer at compile time.
var _ Capturer = (*capturer)(nil)
