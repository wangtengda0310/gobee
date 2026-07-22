package pcap

import (
	"fmt"
	"io"

	"github.com/gopacket/gopacket"        // 纯 Go（PacketSource/PacketDataSource），不需 Npcap
	"github.com/gopacket/gopacket/layers" // 纯 Go（LinkType），不需 Npcap
)

// 本文件提供 Source 接口的纯 Go 实现，不依赖任何 cgo / Npcap。
// 主要用途：
//  1. 从已抓好的 pcap 文件重放（单元测试 / 离线分析）。
//  2. 作为 Source 接口的参考实现，演示如何把 gopacket.PacketSource 适配进来。
//
// 实时抓包（网卡）的实现见 source_live.go，用 //go:build cgo && livecapture 隔离，
// 仅当本机装了 Npcap/libpcap、启用 cgo 且带上 -tags livecapture 时才编译。

// readerSource 把任何实现了 gopacket.PacketDataSource 的对象
// （特别是 pcapgo.Reader）适配成 Source 接口。
//
// pcapgo.Reader 本身就实现了 ReadPacketData() ([]byte, CaptureInfo, error)，
// 因此可以直接喂给 gopacket.NewPacketSource。
type readerSource struct {
	ps     *gopacket.PacketSource
	link   layers.LinkType
	closer io.Closer
	name   string
}

// NewReaderSource 基于一个 gopacket.PacketDataSource 构造 Source。
// ds 通常为 *pcapgo.Reader；link 为其 LinkType；name 用于日志标识。
// 若 ds 同时实现了 io.Closer，Close 时会一并关闭。
//
// 这是把「离线 pcap 文件」接入抓包流程的统一入口。
func NewReaderSource(ds gopacket.PacketDataSource, link layers.LinkType, name string) Source {
	ps := gopacket.NewPacketSource(ds, link)
	rs := &readerSource{ps: ps, link: link, name: name}
	if c, ok := ds.(io.Closer); ok {
		rs.closer = c
	}
	return rs
}

// Packets 实现 Source。
func (s *readerSource) Packets() chan gopacket.Packet {
	return s.ps.Packets()
}

// LinkType 实现 Source。
func (s *readerSource) LinkType() layers.LinkType { return s.link }

// Close 实现 Source。
// ⚠️ 已知局限 [S9]：本方法转发给底层 closer，不保证幂等——若底层 closer
//（如 os.File）的 Close 第二次返回 "already closed" error，本方法也会返回该 error。
// 未来可用 sync.Once 包装以实现真正的幂等。
func (s *readerSource) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

// String 实现 Source。
func (s *readerSource) String() string {
	if s.name != "" {
		return fmt.Sprintf("reader:%s", s.name)
	}
	return "reader"
}

// Ensure readerSource satisfies Source at compile time.
var _ Source = (*readerSource)(nil)
