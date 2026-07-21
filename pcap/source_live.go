//go:build cgo && livecapture

// 本文件仅在同时满足以下两个条件时编译：
//   1. CGO_ENABLED=1（cgo 开启）。
//   2. 构建标签 livecapture（如 -tags livecapture）。
//
// 即：用户必须显式带上 livecapture 标签，才会把基于 gopacket/pcap（cgo）的实时抓包实现
// 编译进来。这把「cgo 用于 race 检测」与「链接 libpcap/Npcap」两件事解耦：
//   - 默认（无 tag）：核心库纯 Go，任意平台可编译、可跑 go test -race（cgo 由 race 开启但不链 libpcap）。
//   - -tags livecapture：链接本机的 libpcap(Npcap)，启用实时网卡抓包。
//
// 前置条件（仅 livecapture 模式）：本机已安装 Npcap(Windows) 或 libpcap(Linux/macOS)。

package pcap

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
)

// liveReadTimeout 是 liveSource 读取循环的单次超时。
// 不用 pcap.BlockForever——那样读取会永久阻塞，Close() / ctx 取消无法打断。
// 用有限超时后，每次 ReadPacketData 在超时时返回 NextErrorTimeoutExpired，
// 读取循环得以周期性地检查 stop 信号，使 Close 能在至多 liveReadTimeout 内退出。
const liveReadTimeout = 1 * time.Second

// liveSource 基于 gopacket/pcap（cgo）实现实时网卡抓包。
//
// 并发安全：handle 和 bpf 受 mu 保护，支持在 Capture 运行期从其它 goroutine
// 调用 SetBPFFilter / ValidateBPF 热重载过滤器。
//
// 可中断：Packets() 的读取循环由 stop channel 控制，Close() 可在 liveReadTimeout 内
// 打断读取，避免 Ctrl+C 无响应。
type liveSource struct {
	mu     sync.RWMutex // 保护 handle / bpf / stop / out
	handle *pcap.Handle
	iface  string
	bpf    string

	// stop 由 Close 关闭，通知读取 goroutine 退出。
	stop chan struct{}
	// stopOnce 保证 stop 只关闭一次。
	stopOnce sync.Once
}

// NewLiveSource 打开指定网卡进行实时抓包。
//   - iface：网卡名（如 "eth0" / "\\Device\\Npcap_{...}"）。
//   - snaplen：每个包最大截获字节数（如 65535）。
//   - promisc：是否开启混杂模式。
//   - bpf：可选的 BPF 过滤表达式，留空则不过滤。
//
// 前置条件：本机已安装 Npcap（Windows）或 libpcap（Linux/macOS），且 CGO_ENABLED=1。
// 否则编译期即报错（cgo 子包不可用）。
//
// 读取超时为固定的 liveReadTimeout（1秒），保证 Close 可在秒级内打断读取。
func NewLiveSource(iface string, snaplen int32, promisc bool, bpf string) (Source, error) {
	handle, err := pcap.OpenLive(iface, snaplen, promisc, liveReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("pcap: open live %q: %w", iface, err)
	}
	if bpf != "" {
		if err := handle.SetBPFFilter(bpf); err != nil {
			handle.Close()
			return nil, fmt.Errorf("pcap: set bpf %q on %q: %w", bpf, iface, err)
		}
	}
	return &liveSource{
		handle: handle,
		iface:  iface,
		bpf:    bpf,
		stop:   make(chan struct{}),
	}, nil
}

// Packets 实现 Source。
//
// 不使用 gopacket.PacketSource.Packets()——它内部用 BlockForever 读取时不响应取消。
// 这里自己驱动读取循环：循环 ReadPacketData，超时则 continue，收到 stop 信号则退出。
// 这样 Capture 的 ctx 取消 → Close 能在 liveReadTimeout 内打断读取。
//
// 注意：本方法启动的 goroutine 在 Close 后退出并关闭 out channel。
func (s *liveSource) Packets() chan gopacket.Packet {
	out := make(chan gopacket.Packet, 128)
	go s.readLoop(out)
	return out
}

// readLoop 驱动读取循环。
// 每次 ReadPacketData 超时（NextErrorTimeoutExpired）则 continue；
// 成功读到包则解码后投递到 out；
// 收到 stop 信号或 handle 出错则关闭 out 并退出。
func (s *liveSource) readLoop(out chan gopacket.Packet) {
	defer close(out)
	s.mu.RLock()
	handle := s.handle
	// LinkType 本身实现了 Decoder 接口，直接用即可（与 gopacket.NewPacketSource 的做法一致）。
	// 不要用 DecodersByLayerName[name]——link type 的 String() 可能不在 map 里（如 loopback 的 Null/Raw）。
	decoder := gopacket.Decoder(handle.LinkType())
	s.mu.RUnlock()

	for {
		// 先检查 stop（在每次阻塞读取前），保证 Close 能及时退出。
		select {
		case <-s.stop:
			return
		default:
		}

		data, ci, err := handle.ReadPacketData()
		if err != nil {
			if errors.Is(err, pcap.NextErrorTimeoutExpired) {
				// 读超时：正常，继续下一轮（会重新检查 stop）。
				continue
			}
			// 其它错误（handle 关闭/设备移除等）：终止读取。
			return
		}

		pkt := gopacket.NewPacket(data, decoder, gopacket.Default)
		pkt.Metadata().CaptureInfo = ci

		select {
		case out <- pkt:
		case <-s.stop:
			return
		}
	}
}

// LinkType 实现 Source。
// 注意：pcap.LinkType 与 layers.LinkType 底层都是 LinkType/int，
// 这里做显式转换以满足 Source 接口签名。
func (s *liveSource) LinkType() layers.LinkType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return layers.LinkType(s.handle.LinkType())
}

// Close 实现 Source。幂等。
// 先关闭 stop 通知读取 goroutine 退出（至多 liveReadTimeout 内生效），
// 再关闭 pcap handle 释放底层资源。
func (s *liveSource) Close() error {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handle != nil {
		s.handle.Close()
	}
	return nil
}

// String 实现 Source。
func (s *liveSource) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.bpf != "" {
		return fmt.Sprintf("live:%s[%s]", s.iface, s.bpf)
	}
	return fmt.Sprintf("live:%s", s.iface)
}

// SetBPFFilter 实现 BPFCapable，支持在 Capture 运行期动态变更 BPF（热重载）。
//
// 并发安全：本方法加写锁，可与 Capture 循环并发调用。
// 底层 pcap.SetBPFFilter 通过 libpcap 的 pcap_setfilter 原子替换内核 BPF 程序，
// 不会中断正在进行的抓包。
//
// 建议先调用 ValidateBPF 校验表达式合法性，再调用本方法应用，避免应用了非法 BPF。
func (s *liveSource) SetBPFFilter(expr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.handle.SetBPFFilter(expr); err != nil {
		return fmt.Errorf("pcap: set bpf %q on %q: %w", expr, s.iface, err)
	}
	s.bpf = expr
	return nil
}

// ValidateBPF 校验 BPF 表达式是否合法，但不应用。
// 用于热重载前的预检查：避免把一个非法表达式 SetBPFFilter 进去导致抓包异常。
// 返回 nil 表示表达式可被内核接受。
func (s *liveSource) ValidateBPF(expr string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, err := s.handle.CompileBPFFilter(expr); err != nil {
		return fmt.Errorf("pcap: invalid bpf %q: %w", expr, err)
	}
	return nil
}

// 编译期保证 liveSource 实现相关接口。
var (
	_ Source       = (*liveSource)(nil)
	_ BPFCapable   = (*liveSource)(nil)
	_ BPFValidator = (*liveSource)(nil)
)

// =============================================================================
// 网卡列举
// =============================================================================

// Interface 是对 pcap.Interface 的精简封装，只暴露跨平台稳定的字段。
type Interface struct {
	// Name 网卡设备名，用于传给 NewLiveSource 的 iface 参数。
	// Windows 形如 "\\Device\\Npcap_{xxxx}"；Linux 形如 "eth0"。
	Name string

	// Description 人类可读的网卡描述（可能为空）。
	Description string

	// IPs 该网卡上的 IPv4/IPv6 地址列表（可能为空）。
	IPs []net.IP
}

// ListInterfaces 列出本机所有可用于抓包的网卡。
// 包装 pcap.FindAllDevs，转换为精简的 Interface 切片。
//
// 前置条件：已安装 Npcap(Windows) / libpcap(Linux/macOS)。
// 常见用法：配合 cmd/pcaptest -list 查看，或程序化选择网卡。
func ListInterfaces() ([]Interface, error) {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("pcap: find all devs: %w", err)
	}
	out := make([]Interface, 0, len(devs))
	for _, d := range devs {
		iface := Interface{
			Name:        d.Name,
			Description: d.Description,
		}
		for _, a := range d.Addresses {
			if a.IP != nil {
				iface.IPs = append(iface.IPs, a.IP)
			}
		}
		out = append(out, iface)
	}
	return out, nil
}
