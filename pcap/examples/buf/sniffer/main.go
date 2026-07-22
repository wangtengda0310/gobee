// 命令 echo-sniffer：用 pcap 抓包并解析 Buf demo 的协议。
//
// 演示如何用 pcap 包的 TCPStreamHandler 配合 Buf 生成的 protobuf 代码，
// 实时抓取并结构化解析 echo-server ↔ echo-client 的通信。
//
// 这是 pcap 包「协议无关基础设施 + 使用者提供协议解析」设计哲学的完整展示：
//   - pcap 包负责：网卡抓包 + TCP 流重组 + 广播分发
//   - sniffer 负责：帧切分（frame 包）+ protobuf 反序列化（echopb）+ JSON 输出
//
// 构建 & 运行（需 Npcap + livecapture）：
//
//	CGO_ENABLED=1 go run -tags livecapture ./examples/buf/sniffer \
//	  -iface <网卡名> -port 19090
//
// 然后在另一个终端启动 server + client，sniffer 会抓到并解析全部消息。
//
//go:build livecapture

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/wangtengda0310/gobee/pcap"
	"github.com/wangtengda0310/gobee/pcap/examples/buf/frame"
	"github.com/wangtengda0310/gobee/pcap/examples/buf/proto/go/echopb"
	"google.golang.org/protobuf/proto"
)

func main() { os.Exit(run()) }

func run() int {
	iface := flag.String("iface", "", "网卡设备名")
	port := flag.Int("port", 19090, "要抓取的服务端口")
	flag.Parse()

	if *iface == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -iface")
		flag.Usage()
		return 2
	}

	src, err := pcap.NewLiveSource(*iface, 65535, false,
		fmt.Sprintf("tcp port %d", *port))
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

	// 用 TCPStreamHandler 重组 TCP 流，然后逐帧解析 + protobuf 反序列化。
	// 每个 TCP 连接（半连接）启动一个回调，r 是重组后的完整字节流。
	handler := pcap.NewTCPStreamHandler("echo-sniffer", func(flow pcap.FlowKey, r io.Reader) error {
		// 判断方向：源端口 > 目标端口 → 客户端→服务端。
		srcPort := flow.TransportFlow.Src().String()
		dstPort := flow.TransportFlow.Dst().String()
		dir := "服务端→客户端"
		if srcPort > dstPort {
			dir = "客户端→服务端"
		}
		fmt.Printf("\n=== 新 TCP 流 [%s] (%s) ===\n", flow, dir)

		// 循环读取帧，直到流结束。
		for {
			msgType, payload, err := frame.ReadFrame(r)
			if err != nil {
				fmt.Printf("  [%s] 流结束\n", dir)
				return err
			}
			printMessage(dir, msgType, payload)
		}
	})
	c.RegisterHandler(handler)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("开始抓包（端口 %d），Ctrl+C 退出...\n", *port)
	if err := c.Capture(ctx, src, pcap.Target{}); err != nil {
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "\n收到退出信号，停止抓包。")
		} else {
			fmt.Fprintln(os.Stderr, "capture:", err)
			return 1
		}
	}
	handler.Close()

	st := c.Stats()
	fmt.Printf("\n结束。captured=%d\n", st.Captured)
	return 0
}

// printMessage 解析一帧消息并打印结构化的 JSON 输出。
//
// 根据 msgType 选择对应的 protobuf 类型反序列化，
// 然后用 json.Marshal 输出可读的 JSON 格式。
// 这展示了「Buf 生成的类型安全 protobuf 代码」的便利——
// 不需要手动解析 wire format（如 xcards 示例中的 DumpProtobufRaw）。
func printMessage(dir string, msgType uint16, payload []byte) {
	name := frame.MsgTypeName(msgType)

	switch msgType {
	case frame.MsgTypeEchoRequest:
		var msg echopb.EchoRequest
		if err := proto.Unmarshal(payload, &msg); err != nil {
			fmt.Printf("  [%s] %s: unmarshal error: %v\n", dir, name, err)
			return
		}
		printJSON(dir, name, &msg)

	case frame.MsgTypeEchoResponse:
		var msg echopb.EchoResponse
		if err := proto.Unmarshal(payload, &msg); err != nil {
			fmt.Printf("  [%s] %s: unmarshal error: %v\n", dir, name, err)
			return
		}
		printJSON(dir, name, &msg)

	case frame.MsgTypeSumRequest:
		var msg echopb.SumRequest
		if err := proto.Unmarshal(payload, &msg); err != nil {
			fmt.Printf("  [%s] %s: unmarshal error: %v\n", dir, name, err)
			return
		}
		printJSON(dir, name, &msg)

	case frame.MsgTypeSumResponse:
		var msg echopb.SumResponse
		if err := proto.Unmarshal(payload, &msg); err != nil {
			fmt.Printf("  [%s] %s: unmarshal error: %v\n", dir, name, err)
			return
		}
		printJSON(dir, name, &msg)

	default:
		fmt.Printf("  [%s] %s: (unknown, %d bytes)\n", dir, name, len(payload))
	}
}

// printJSON 将 protobuf 消息序列化为 JSON 并打印。
// protobuf 的 jsonpb / protojson 可以更精确地序列化，
// 这里用标准库 encoding/json 做简化演示。
func printJSON(dir, name string, msg proto.Message) {
	jsonBytes, err := json.MarshalIndent(msg, "  ", "  ")
	if err != nil {
		fmt.Printf("  [%s] %s: json error: %v\n", dir, name, err)
		return
	}
	fmt.Printf("  [%s] %s: %s\n", dir, name, jsonBytes)
}
