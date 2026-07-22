// 命令 pcapdump：解析 .pcap 文件并输出每层的结构化信息。
//
// 演示 pcap 包的「离线重放」能力——读取一个已有的 pcap 文件，
// 解析每个包的以太网/IP/TCP/UDP/HTTP 等层信息。
//
// 这是纯 Go 程序（不需要 Npcap/cgo），任意平台可运行。
//
// # 用法
//
//	# 基本用法：打印每个包的摘要
//	go run ./examples/pcapdump -file dump.pcap
//
//	# 只显示 TCP 包
//	go run ./examples/pcapdump -file dump.pcap -filter tcp
//
//	# 显示完整的应用层数据（HTTP 请求行等）
//	go run ./examples/pcapdump -file dump.pcap -app
//
//	# 统计模式：只输出协议分布统计
//	go run ./examples/pcapdump -file dump.pcap -stats
//
// # 获取测试用的 pcap 文件
//
// 你可以用 pcaptest 或 xcards 抓取真实流量并保存为 pcap 文件：
//
//	# 用 pcaptest 的 -out 参数保存
//	CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡> -out dump.pcap
//
// 或者用 Wireshark / tcpdump 抓取。
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/wangtengda0310/gobee/pcap"
)

func main() { os.Exit(run()) }

func run() int {
	file := flag.String("file", "", "pcap 文件路径（必填）")
	protoFilter := flag.String("filter", "", "只显示指定协议的包（tcp/udp/arp/icmp，留空=全部）")
	showApp := flag.Bool("app", false, "显示应用层数据（HTTP 请求行/响应行等明文内容）")
	statsOnly := flag.Bool("stats", false, "只输出协议分布统计，不逐包打印")
	flag.Parse()

	if *file == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -file")
		flag.Usage()
		return 2
	}

	// 打开 pcap 文件。
	f, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open file:", err)
		return 1
	}
	defer f.Close()

	reader, err := pcapgo.NewReader(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read pcap header:", err)
		return 1
	}

	fmt.Printf("文件: %s\n", *file)
	fmt.Printf("链路层: %s\n", reader.LinkType())
	fmt.Printf("SnapLen: %d\n\n", reader.Snaplen())

	// 用 pcap 包的 NewReaderSource 把 pcapgo.Reader 包装成 Source。
	// 这样可以直接用 Capturer + handler 的标准链路处理，
	// 也可以直接 range Packets() 手动处理。
	//
	// 这里演示直接手动 range（更直观），不用 Capturer：
	// 离线解析不需要广播/过载保护等特性。
	src := pcap.NewReaderSource(reader, reader.LinkType(), *file)

	// 协议统计。
	var stats packetStats
	var count int

	for pkt := range src.Packets() {
		count++
		stats.record(pkt)

		if *statsOnly {
			continue
		}

		// 协议过滤。
		if *protoFilter != "" && !hasProtocol(pkt, *protoFilter) {
			continue
		}

		printPacket(pkt, count, *showApp)
	}

	// 输出统计。
	fmt.Println("\n=== 统计 ===")
	fmt.Printf("总包数: %d\n", count)
	stats.print()

	return 0
}

// printPacket 打印单个包的各层信息。
func printPacket(pkt gopacket.Packet, num int, showApp bool) {
	ci := pkt.Metadata().CaptureInfo
	ts := ci.Timestamp.Format("15:04:05.000000")

	// 提取各层。
	var summary []string

	if eth := pkt.Layer(layers.LayerTypeEthernet); eth != nil {
		e, _ := eth.(*layers.Ethernet)
		summary = append(summary, fmt.Sprintf("ETH %s→%s", e.SrcMAC, e.DstMAC))
	}

	if ipv4 := pkt.Layer(layers.LayerTypeIPv4); ipv4 != nil {
		ip, _ := ipv4.(*layers.IPv4)
		summary = append(summary, fmt.Sprintf("IPv4 %s→%s", ip.SrcIP, ip.DstIP))
		summary = append(summary, fmt.Sprintf("proto=%s", ip.Protocol))
	}

	if tcp := pkt.Layer(layers.LayerTypeTCP); tcp != nil {
		t, _ := tcp.(*layers.TCP)
		flags := tcpFlagsString(t)
		summary = append(summary, fmt.Sprintf("TCP %d→%d %s", t.SrcPort, t.DstPort, flags))
	}

	if udp := pkt.Layer(layers.LayerTypeUDP); udp != nil {
		u, _ := udp.(*layers.UDP)
		summary = append(summary, fmt.Sprintf("UDP %d→%d", u.SrcPort, u.DstPort))
	}

	if arp := pkt.Layer(layers.LayerTypeARP); arp != nil {
		summary = append(summary, "ARP")
	}

	if icmp := pkt.Layer(layers.LayerTypeICMPv4); icmp != nil {
		summary = append(summary, "ICMP")
	}

	// 解析错误标记。
	if errLayer := pkt.ErrorLayer(); errLayer != nil {
		summary = append(summary, fmt.Sprintf("DECODE_ERR(%s)", errLayer.Error()))
	}

	fmt.Printf("#%04d [%s] len=%d  %s\n", num, ts, ci.Length, strings.Join(summary, " "))

	// 可选：显示应用层数据。
	if showApp {
		if app := pkt.ApplicationLayer(); app != nil {
			payload := string(app.Payload())
			// 只显示第一行（HTTP 请求行/响应行），截断长数据。
			firstLine := strings.SplitN(payload, "\r\n", 2)[0]
			if firstLine == "" {
				firstLine = strings.SplitN(payload, "\n", 2)[0]
			}
			if len(firstLine) > 80 {
				firstLine = firstLine[:80] + "..."
			}
			if firstLine != "" {
				fmt.Printf("      app: %s\n", firstLine)
			}
		}
	}
}

// tcpFlagsString 返回 TCP 标志的可读字符串（SYN/ACK/FIN/RST/PSH）。
func tcpFlagsString(t *layers.TCP) string {
	var flags []string
	if t.SYN {
		flags = append(flags, "SYN")
	}
	if t.ACK {
		flags = append(flags, "ACK")
	}
	if t.FIN {
		flags = append(flags, "FIN")
	}
	if t.RST {
		flags = append(flags, "RST")
	}
	if t.PSH {
		flags = append(flags, "PSH")
	}
	if len(flags) == 0 {
		return "."
	}
	return strings.Join(flags, "/")
}

// hasProtocol 检查包是否包含指定协议层。
func hasProtocol(pkt gopacket.Packet, proto string) bool {
	switch strings.ToLower(proto) {
	case "tcp":
		return pkt.Layer(layers.LayerTypeTCP) != nil
	case "udp":
		return pkt.Layer(layers.LayerTypeUDP) != nil
	case "arp":
		return pkt.Layer(layers.LayerTypeARP) != nil
	case "icmp":
		return pkt.Layer(layers.LayerTypeICMPv4) != nil
	case "ipv4":
		return pkt.Layer(layers.LayerTypeIPv4) != nil
	default:
		return false
	}
}

// packetStats 统计各协议的包数。
type packetStats struct {
	total                int
	ethernet, ipv4, ipv6 int
	tcp, udp, arp, icmp  int
	http                 int
	errors               int
}

func (s *packetStats) record(pkt gopacket.Packet) {
	s.total++
	if pkt.Layer(layers.LayerTypeEthernet) != nil {
		s.ethernet++
	}
	if pkt.Layer(layers.LayerTypeIPv4) != nil {
		s.ipv4++
	}
	if pkt.Layer(layers.LayerTypeIPv6) != nil {
		s.ipv6++
	}
	if pkt.Layer(layers.LayerTypeTCP) != nil {
		s.tcp++
	}
	if pkt.Layer(layers.LayerTypeUDP) != nil {
		s.udp++
	}
	if pkt.Layer(layers.LayerTypeARP) != nil {
		s.arp++
	}
	if pkt.Layer(layers.LayerTypeICMPv4) != nil {
		s.icmp++
	}
	if pkt.ApplicationLayer() != nil {
		s.http++
	}
	if pkt.ErrorLayer() != nil {
		s.errors++
	}
}

func (s *packetStats) print() {
	fmt.Printf("  Ethernet: %d\n", s.ethernet)
	fmt.Printf("  IPv4: %d, IPv6: %d\n", s.ipv4, s.ipv6)
	fmt.Printf("  TCP: %d, UDP: %d\n", s.tcp, s.udp)
	fmt.Printf("  ARP: %d, ICMP: %d\n", s.arp, s.icmp)
	fmt.Printf("  应用层数据: %d\n", s.http)
	if s.errors > 0 {
		fmt.Printf("  解析错误: %d\n", s.errors)
	}
}
