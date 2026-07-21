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
	"fmt"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
)

// liveSource 基于 gopacket/pcap（cgo）实现实时网卡抓包。
type liveSource struct {
	handle *pcap.Handle
	ps     *gopacket.PacketSource
	iface  string
	bpf    string
}

// NewLiveSource 打开指定网卡进行实时抓包。
//   - iface：网卡名（如 "eth0" / "\\Device\\Npcap_{...}"）。
//   - snaplen：每个包最大截获字节数（如 65535）。
//   - promisc：是否开启混杂模式。
//   - bpf：可选的 BPF 过滤表达式，留空则不过滤。
//
// 前置条件：本机已安装 Npcap（Windows）或 libpcap（Linux/macOS），且 CGO_ENABLED=1。
// 否则编译期即报错（cgo 子包不可用）。
func NewLiveSource(iface string, snaplen int32, promisc bool, bpf string) (Source, error) {
	handle, err := pcap.OpenLive(iface, snaplen, promisc, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("pcap: open live %q: %w", iface, err)
	}
	if bpf != "" {
		if err := handle.SetBPFFilter(bpf); err != nil {
			handle.Close()
			return nil, fmt.Errorf("pcap: set bpf %q on %q: %w", bpf, iface, err)
		}
	}
	ps := gopacket.NewPacketSource(handle, handle.LinkType())
	return &liveSource{handle: handle, ps: ps, iface: iface, bpf: bpf}, nil
}

// Packets 实现 Source。
func (s *liveSource) Packets() chan gopacket.Packet { return s.ps.Packets() }

// LinkType 实现 Source。
// 注意：pcap.LinkType 与 layers.LinkType 底层都是 LinkType/int，
// 这里做显式转换以满足 Source 接口签名。
func (s *liveSource) LinkType() layers.LinkType {
	return layers.LinkType(s.handle.LinkType())
}

// Close 实现 Source。
func (s *liveSource) Close() error {
	s.handle.Close()
	return nil
}

// String 实现 Source。
func (s *liveSource) String() string {
	if s.bpf != "" {
		return fmt.Sprintf("live:%s[%s]", s.iface, s.bpf)
	}
	return fmt.Sprintf("live:%s", s.iface)
}

// SetBPFFilter 实现 BPFCapable，支持在 Capture 时动态变更 BPF。
func (s *liveSource) SetBPFFilter(expr string) error {
	if err := s.handle.SetBPFFilter(expr); err != nil {
		return err
	}
	s.bpf = expr
	return nil
}

var _ Source = (*liveSource)(nil)
var _ BPFCapable = (*liveSource)(nil)
