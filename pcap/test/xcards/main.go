// 命令 xcards：pcap 抓包 + Unity 游戏协议解析（完整示例）。
//
// 演示如何用 pcap 包解析 go-service 的游戏消息格式（4字节包头 + msgID + seqID + body）。
//
// 构建 & 运行（需 Npcap + livecapture）：
//
//	# 流重组模式（需在 Unity 连接前启动，完整重组）
//	CGO_ENABLED=1 go run -tags livecapture ./test/xcards -iface <网卡名> -port 18000
//
//	# 逐包模式（支持先连接后抓包，不依赖 SYN，但大消息可能解析不全）
//	CGO_ENABLED=1 go run -tags livecapture ./test/xcards -iface <网卡名> -port 18000 -raw
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
	"github.com/wangtengda0310/gobee/pcap/test/xcards/gameproto"
)

func main() { os.Exit(run()) }

func run() int {
	iface := flag.String("iface", "", "网卡名")
	port := flag.Int("port", 18000, "游戏服务器端口（0=不过滤端口）")
	rawMode := flag.Bool("raw", false, "逐包模式：不依赖 TCP 流重组，支持先连接后抓包")
	flag.Parse()

	if *iface == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -iface")
		flag.Usage()
		return 2
	}

	bpf := fmt.Sprintf("tcp port %d", *port)
	if *port == 0 {
		bpf = "tcp"
	}

	src, err := pcap.NewLiveSource(*iface, 65535, false, bpf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open source:", err)
		return 1
	}
	defer src.Close()

	c := pcap.NewCapturer(
		pcap.WithBufferSize(2048),
		pcap.WithOverflowStrategy(pcap.OverflowDrop),
	)
	defer c.Close()

	if *rawMode {
		registerRawHandler(c)
	} else {
		registerStreamHandler(c)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("开始抓包（端口 %d, 模式=%s），Ctrl+C 退出...\n",
		*port, modeName(*rawMode))
	if err := c.Capture(ctx, src, pcap.Target{}); err != nil {
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "\n收到退出信号，停止抓包。")
		} else {
			fmt.Fprintln(os.Stderr, "capture:", err)
			return 1
		}
	}

	// flush 流重组的残留流（raw 模式无此需要）。
	if !*rawMode {
		if h, ok := c.(interface{ Close() }); ok {
			_ = h
		}
		gameHandler.Close()
	}

	st := c.Stats()
	fmt.Printf("结束。captured=%d\n", st.Captured)
	for _, hs := range st.Handlers {
		fmt.Printf("  handler %s: received=%d processed=%d dropped=%d errors=%d\n",
			hs.Name, hs.Received, hs.Processed, hs.Dropped, hs.Errors)
	}
	return 0
}

func modeName(raw bool) string {
	if raw {
		return "逐包(-raw)"
	}
	return "流重组"
}

// gameHandler 是流重组模式的全局 handler（用于 Close）。
var gameHandler *pcap.TCPStreamHandler

// registerStreamHandler 注册 TCP 流重组 handler（需要从连接建立时开始抓包）。
func registerStreamHandler(c pcap.Capturer) {
	gameHandler = pcap.NewTCPStreamHandler("game-stream", func(flow pcap.FlowKey, r io.Reader) error {
		fromClient, dir := detectDirection(flow)
		fmt.Printf("\n=== 新 TCP 流 [%s] (%s) ===\n", flow, dir)

		reader := gameproto.NewFrameReader(r, fromClient)
		for {
			msgID, seqID, body, err := reader.ReadMessage()
			if err != nil {
				fmt.Printf("  [%s] 流结束: %v\n", dir, err)
				return err
			}
			printMessage(dir, msgID, seqID, body)
		}
	})
	c.RegisterHandler(gameHandler)
}

// registerRawHandler 注册逐包 handler（支持先连接后抓包，不依赖 SYN）。
// 直接从每个 TCP 包的应用层 payload 解析消息——大多数游戏消息在单个 TCP 段内完整。
func registerRawHandler(c pcap.Capturer) {
	c.RegisterHandler(pcap.NewHandlerFunc("game-raw", func(ctx context.Context, e *pcap.PacketEvent) error {
		app := e.Packet.ApplicationLayer()
		if app == nil {
			return nil
		}
		data := app.Payload()
		if len(data) == 0 {
			return nil
		}

		fromClient, dir := detectDirectionFromEvent(e)
		// 从 payload 逐条切分消息（一个 TCP 包可能含多条消息）。
		buf := bytes.NewReader(data)
		reader := gameproto.NewFrameReader(buf, fromClient)
		for {
			msgID, seqID, body, err := reader.ReadMessage()
			if err != nil {
				return nil // 一个包里的剩余字节不够一条消息，正常跳过
			}
			printMessage(dir, msgID, seqID, body)
		}
	}))
}

// detectDirection 根据 TCP 端口判断方向。
// 客户端用临时高端口，服务端用固定端口（如 18000）。
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

// printMessage 打印解析出的游戏消息。
func printMessage(dir string, msgID uint16, seqID uint32, body []byte) {
	name := gameproto.MsgName(msgID)
	fmt.Printf("  [%s] %s ID=%d Seq=%d BodyLen=%d\n", dir, name, msgID, seqID, len(body))
	// 框架消息（ID<1000）的 body 不是 protobuf（如 Pong 是原始时间戳），跳过解析。
	// 游戏消息（ID>=1000）的 body 是 protobuf，用 protowire 解析（类似 protoc --decode_raw）。
	if msgID >= 1000 && len(body) > 0 {
		fmt.Print(gameproto.DumpProtobufRaw(body))
	}
}

// detectDirectionFromEvent 从 PacketEvent 的 TransportFlow 判断方向（raw 模式用）。
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
