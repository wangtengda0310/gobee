// Package pcap 提供基于 gopacket 的网络抓包能力。
//
// # 设计哲学：协议无关的基础设施
//
// 本包**只提供抓包与流重组的基础设施，不绑定任何具体应用层协议**。
// 抓到的原始包通过 gopacket.Packet 透传给使用者，由使用者自行用 gopacket 的
// Layer/Decoder 机制或自定义解析逻辑处理。内置的 HTTP 重组（HTTPRequestHandler）
// 是便利封装，不是协议绑定的证据——它基于标准库 net/http，使用者可自由替换为
// 任何 TCP 协议解析器（通过 TCPStreamHandler 的 io.Reader 回调）。
//
// # 设计概览
//
// 本包把抓包流程拆成三部分：数据源（Source）、抓包器（Capturer）、处理函数（PacketHandler）。
//
//	capturer.RegisterHandler(handler)     // 注册处理函数
//	capturer.Capture(ctx, source, target) // 从数据源抓包，广播给所有处理函数
//
// 核心库（本包除 source_live.go 外的所有文件）为纯 Go 实现，
// 不依赖 cgo / Npcap，可在任意平台编译与单元测试。
// 实时网卡抓包实现见 source_live.go，用 //go:build cgo && livecapture 隔离。
//
// # 两个关键问题的思考结论
//
//  1. 数据包数量巨大、处理函数来不及消费怎么办？
//     见 capturer.go 文件顶部注释：广播解耦 + 每 handler 独立有界队列 + 三种过载策略。
//
//  2. 还需要哪些必要功能？
//     见 README.md 的「必要功能」章节：BPF 过滤、生命周期、动态增删、统计、错误处理等。
//
// # 快速开始
//
//	c := pcap.NewCapturer(
//	    pcap.WithBPFFilter("tcp port 80"),
//	    pcap.WithOverflowStrategy(pcap.OverflowDrop),
//	    pcap.WithBufferSize(2048),
//	)
//	c.RegisterHandler(pcap.NewHandlerFunc("printer", func(ctx context.Context, e *pcap.PacketEvent) error {
//	    fmt.Println(e.Timestamp, e.NetworkFlow)
//	    return nil
//	}))
//	// 离线重放（纯 Go，任意平台可编译）：
//	f, _ := os.Open("dump.pcap")
//	r, _ := pcapgo.NewReader(f)
//	src := pcap.NewReaderSource(r, r.LinkType(), "dump.pcap")
//	// 实时抓包（需 cgo + libpcap + -tags livecapture）：
//	// src, _ := pcap.NewLiveSource("eth0", 65535, false, "")
//	c.Capture(context.Background(), src, pcap.Target{Host: "itsnot.fun"})
//
// # 单元测试
//
// 测试不依赖任何网卡或特权：用 gopacket/pcapgo 在内存构造含 itsnot.fun HTTP 请求的
// pcap 文件，验证整条「读取 -> 解析 -> 广播 -> 处理」链路。
package pcap
