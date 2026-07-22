// 命令 xcards：pcap 抓包 + Unity 游戏协议解析（完整示例）。
//
// 本程序演示如何用 pcap 包的 TCPStreamHandler / PacketHandler
// 配合自定义的 gameproto 包，实时解析 go-service 的游戏消息。
//
// # 架构
//
//	网卡（Npcap）→ pcap.NewLiveSource → pcap.Capturer → gameproto.FrameReader → 可读输出
//
// # 两种工作模式
//
//  1. 流重组模式（默认）：用 pcap.TCPStreamHandler 做 TCP 流重组，
//     能正确处理跨 TCP 段的大消息。但需要从连接建立时开始抓包（tcpassembly 依赖 SYN）。
//
//  2. 逐包模式（-raw）：直接从每个 TCP 包的 payload 解析消息，
//     不依赖 SYN，支持「先连接后抓包」。代价是跨 TCP 段的大消息可能解析不全。
//
// # 构建 & 运行
//
// 需要 Npcap + MSYS2 gcc + CGO_ENABLED=1 + -tags livecapture（详见 README）。
//
//	# 流重组模式（需在 Unity 连接前启动）
//	CGO_ENABLED=1 go run -tags livecapture ./examples/xcards -iface <网卡名> -port 18000
//
//	# 逐包模式（支持先连接后抓包）
//	CGO_ENABLED=1 go run -tags livecapture ./examples/xcards -iface <网卡名> -port 18000 -raw
//
// 网卡名通过 `xcards -list` 或 `pcaptest -list` 查看。
// 端口 20144 = HTTP 登录，端口 18000 = 游戏二进制协议。
//
//go:build livecapture

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/wangtengda0310/gobee/pcap"
	"github.com/wangtengda0310/gobee/pcap/examples/xcards/gameproto"
)

// main 把 run() 的返回值传给 os.Exit，确保 defer 正常执行。
// （直接在 main 里 os.Exit 会跳过 defer。）
func main() { os.Exit(run()) }

// run 是主逻辑入口，返回进程退出码（0=成功，1=错误，2=参数错误）。
func run() int {
	// --- 参数解析 ---
	iface := flag.String("iface", "", "网卡设备名（如 \"\\\\Device\\\\Npcap_{...}\" 或 Linux 的 \"eth0\"）")
	port := flag.Int("port", 18000, "游戏服务器 TCP 端口（20144=HTTP登录，18000=游戏协议，0=不过滤端口）")
	rawMode := flag.Bool("raw", false, "逐包模式：不依赖 TCP 流重组，支持先连接后抓包")
	flag.Parse()

	if *iface == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -iface（用 pcaptest -list 查看可用网卡）")
		flag.Usage()
		return 2
	}

	// 构建 BPF 过滤表达式。
	// BPF 在网卡内核层过滤，比用户态过滤高效得多。
	// port=0 表示不过滤端口（抓所有 TCP 流量）。
	bpf := fmt.Sprintf("tcp port %d", *port)
	if *port == 0 {
		bpf = "tcp"
	}

	// --- 打开网卡 ---
	// NewLiveSource 调用 Npcap 的 pcap.OpenLive（cgo），需要管理员权限。
	src, err := pcap.NewLiveSource(*iface, 65535, false, bpf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open source:", err)
		return 1
	}
	defer src.Close()

	// --- 创建抓包器 ---
	// OverflowDrop：队列满时丢弃新包，保证抓包循环不阻塞。
	// BufferSize=2048：每个 handler 的队列容量，高流量时可调大。
	c := pcap.NewCapturer(
		pcap.WithBufferSize(2048),
		pcap.WithOverflowStrategy(pcap.OverflowDrop),
	)
	defer c.Close()

	// --- 注册消息解析 handler ---
	// 根据模式选择不同的 handler：
	//   流重组模式 → TCPStreamHandler（完整重组，需从 SYN 开始）
	//   逐包模式   → 普通 PacketHandler（直接解析每个包的 payload）
	if *rawMode {
		registerRawHandler(c)
	} else {
		registerStreamHandler(c)
	}

	// --- 启动抓包 ---
	// 监听 Ctrl+C / SIGTERM，优雅退出。
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("开始抓包（端口 %d, 模式=%s），Ctrl+C 退出...\n",
		*port, modeName(*rawMode))
	if err := c.Capture(ctx, src, pcap.Target{}); err != nil {
		// ctx 被信号取消是正常的优雅退出，不算错误。
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "\n收到退出信号，停止抓包。")
		} else {
			fmt.Fprintln(os.Stderr, "capture:", err)
			return 1
		}
	}

	// --- 清理：flush 流重组的残留流 ---
	// 流重组模式下，Capture 返回后 TCPStreamHandler 内部可能还有未完成的流。
	// 调用 Close 触发 FlushAll，确保缓冲中的最后一条消息被输出。
	// raw 模式无此需要（没有流状态）。
	if !*rawMode {
		gameHandler.Close()
	}

	// --- 打印统计 ---
	st := c.Stats()
	fmt.Printf("结束。captured=%d\n", st.Captured)
	for _, hs := range st.Handlers {
		fmt.Printf("  handler %s: received=%d processed=%d dropped=%d errors=%d\n",
			hs.Name, hs.Received, hs.Processed, hs.Dropped, hs.Errors)
	}
	return 0
}

// modeName 返回模式的中文名称（用于日志输出）。
func modeName(raw bool) string {
	if raw {
		return "逐包(-raw)"
	}
	return "流重组"
}

// gameHandler 是流重组模式的全局 handler 引用（用于 Capture 返回后 Close）。
// 之所以用全局变量而非局部变量，是因为 registerStreamHandler 是独立函数，
// 不能返回值给 run()。
var gameHandler *pcap.TCPStreamHandler

// registerStreamHandler 注册 TCP 流重组 handler。
//
// 流重组模式下，pcap.TCPStreamHandler 把同一个 TCP 连接的所有包按序列号拼回完整字节流，
// 然后通过 io.Reader 回调交给使用者。使用者用 FrameReader 从中逐条切分游戏消息。
//
// 约束：tcpassembly 需要看到 TCP 握手（SYN）才能正确跟踪序列号，
// 因此必须在 Unity 连接服务器之前就启动抓包。
//
// 每个 TCP 连接（半连接）启动一个回调 goroutine，r 是该流重组后的字节流。
func registerStreamHandler(c pcap.Capturer) {
	gameHandler = pcap.NewTCPStreamHandler("game-stream", func(flow pcap.FlowKey, r io.Reader) error {
		// 根据端口判断方向（客户端→服务端 or 服务端→客户端）。
		// 方向决定解密时用哪个密钥（decryptKey vs encryptKey）。
		fromClient, dir := detectDirection(flow)
		fmt.Printf("\n=== 新 TCP 流 [%s] (%s) ===\n", flow, dir)

		// FrameReader 从重组后的字节流逐条切分游戏消息。
		// 它会自动处理：消息切分 → 解密 → 解压 → 提取 msgID/seqID/body。
		reader := gameproto.NewFrameReader(r, fromClient)
		for {
			msgID, seqID, body, err := reader.ReadMessage()
			if err != nil {
				// io.EOF 或连接关闭——正常退出。
				fmt.Printf("  [%s] 流结束: %v\n", dir, err)
				return err
			}
			printMessage(dir, msgID, seqID, body)
		}
	})
	c.RegisterHandler(gameHandler)
}

// registerRawHandler 注册逐包 handler。
//
// 逐包模式下，不使用 TCP 流重组——直接从每个 TCP 包的应用层 payload 解析消息。
// 优势：不需要从连接建立时开始抓包，支持「先连接后抓包」。
// 劣势：如果一个游戏消息被 TCP 分段（跨多个包），无法完整解析。
// 但游戏消息通常很小（心跳 10 字节、移动几十字节），大多数情况下一个包就是一条完整消息。
//
// 工作原理：每个 TCP 包触发一次回调，从 payload 中尝试逐条切分消息。
// 一个包可能包含 0 条、1 条或多条消息（TCP 的 Nagle 可能合并小包）。
func registerRawHandler(c pcap.Capturer) {
	c.RegisterHandler(pcap.NewHandlerFunc("game-raw", func(ctx context.Context, e *pcap.PacketEvent) error {
		// 提取 TCP payload（应用层数据）。
		app := e.Packet.ApplicationLayer()
		if app == nil {
			return nil // 非 TCP 包或无 payload（如纯 ACK），跳过。
		}
		data := app.Payload()
		if len(data) == 0 {
			return nil
		}

		fromClient, dir := detectDirectionFromEvent(e)

		// 从 payload 逐条切分消息。
		// 用 bytes.Reader 包装 payload，使其实现 io.Reader 接口。
		// FrameReader 会尝试从当前位置切分消息——如果剩余字节不够一条完整消息，
		// ReadMessage 返回 error，我们静默跳过（返回 nil）。
		buf := bytes.NewReader(data)
		reader := gameproto.NewFrameReader(buf, fromClient)
		for {
			msgID, seqID, body, err := reader.ReadMessage()
			if err != nil {
				// 一个包里的剩余字节不够一条消息——正常情况（TCP 分段），跳过。
				return nil
			}
			printMessage(dir, msgID, seqID, body)
		}
	}))
}

// detectDirection 根据 TCP 流的 FlowKey 判断方向（流重组模式用）。
//
// 判断逻辑：客户端使用临时高端口（如 54315），服务端使用固定端口（如 18000）。
// 源端口 > 目标端口 → 客户端→服务端（fromClient=true，解密用 decryptKey）。
// 源端口 < 目标端口 → 服务端→客户端（fromClient=false，解密用 encryptKey）。
//
// 返回 fromClient（方向标志）和 dir（中文方向描述，用于日志）。
func detectDirection(flow pcap.FlowKey) (fromClient bool, dir string) {
	srcPort := flow.TransportFlow.Src().String()
	dstPort := flow.TransportFlow.Dst().String()
	fromClient = srcPort > dstPort
	if fromClient {
		dir = "客户端→服务端"
	} else {
		dir = "服务端→客户端"
	}
	return
}

// detectDirectionFromEvent 从 PacketEvent 的 TransportFlow 判断方向（逐包模式用）。
// 逻辑同 detectDirection，只是数据来源不同（PacketEvent vs FlowKey）。
func detectDirectionFromEvent(e *pcap.PacketEvent) (fromClient bool, dir string) {
	srcPort := e.TransportFlow.Src().String()
	dstPort := e.TransportFlow.Dst().String()
	fromClient = srcPort > dstPort
	if fromClient {
		dir = "客户端→服务端"
	} else {
		dir = "服务端→客户端"
	}
	return
}

// printMessage 打印解析出的游戏消息（统一格式，两种模式共用）。
//
// body 格式因消息类型而异（见 gameproto 包文档）：
//   - LoginReq(1):  自定义 ByteStream（DumpLoginReq 解析）
//   - LoginResp(2): 自定义 ByteStream（DumpLoginResp 解析）
//   - Ping(3):      body 为空，不解析
//   - Pong(4):      body 是原始时间戳（非 protobuf），输出十六进制
//   - 游戏消息(>=1000): [2B len][protobuf] 子信封格式（DumpProtobufRaw 解析）
//   - 其它:         输出原始十六进制
func printMessage(dir string, msgID uint16, seqID uint32, body []byte) {
	name := gameproto.MsgName(msgID)
	fmt.Printf("  [%s] %s ID=%d Seq=%d BodyLen=%d\n", dir, name, msgID, seqID, len(body))
	if len(body) == 0 {
		return
	}

	switch {
	case msgID == 1:
		// LoginReq：自定义 ByteStream 格式。
		fmt.Print(gameproto.DumpLoginReq(body))
	case msgID == 2:
		// LoginResp：自定义 ByteStream 格式。
		fmt.Print(gameproto.DumpLoginResp(body))
	case msgID == 3:
		// Ping：body 为空（已在上方 len==0 时跳过）。
	case msgID == 4:
		// Pong：body 是服务端时间戳（uint64），不是 protobuf。
		fmt.Printf("         (timestamp: %x)\n", body)
	case msgID >= 1000:
		// 游戏消息：body 是 [2字节长度LE][protobuf数据]（子信封）。
		// 先剥掉 2 字节长度前缀，再用 protowire 解析 protobuf wire format。
		protoData := body
		if len(body) >= 2 {
			dataLen := int(body[0]) | int(body[1])<<8
			if dataLen <= len(body)-2 && dataLen > 0 {
				protoData = body[2 : 2+dataLen]
			}
		}
		if len(protoData) > 0 {
			fmt.Print(gameproto.DumpProtobufRaw(protoData))
		}
	default:
		fmt.Printf("         (raw: %x)\n", body)
	}
}
