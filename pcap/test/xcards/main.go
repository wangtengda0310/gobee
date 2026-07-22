// 命令 xcards：pcap 抓包 + Unity 游戏协议解析（完整示例）。
//
// 演示如何用 pcap 包的 TCPStreamHandler + 自定义 Layer 解析
// go-service 的游戏消息格式（4字节包头 + msgID + seqID + body）。
//
// 构建 & 运行（需 Npcap + livecapture）：
//
//	CGO_ENABLED=1 go run -tags livecapture ./test/xcards -iface <网卡名>
//
// 注意：本文件带 //go:build livecapture 标签（依赖 NewLiveSource）。
// 如用离线 pcap 重放，去掉标签并改用 NewReaderSource。
//
//go:build livecapture

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/wangtengda0310/gobee/pcap"
	"github.com/wangtengda0310/gobee/pcap/test/xcards/gameproto"
	// 你的 protobuf 生成包（如果有，取消注释并替换路径）
	// pb "yourmodule/protos"
)

func main() {
	os.Exit(run())
}

func run() int {
	// 参数：网卡名（必填）、端口（默认 20144）。
	iface := ""
	port := 20144
	args := os.Args[1:]
	if len(args) > 0 {
		iface = args[0]
	}
	if iface == "" {
		fmt.Fprintln(os.Stderr, "用法: xcards <网卡名> [端口]")
		fmt.Fprintln(os.Stderr, "示例: xcards \"\\Device\\NPF_{FDEA18E5-...}\" 20144")
		return 2
	}
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%d", &port)
	}

	// 打开网卡抓包。
	src, err := pcap.NewLiveSource(
		iface, 65535, false,
		fmt.Sprintf("tcp port %d", port), // BPF 过滤游戏服务器端口
	)
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

	// ★ 核心：用 TCPStreamHandler + 自定义 Layer 解析游戏消息。
	gameHandler := pcap.NewTCPStreamHandler("game-proto", func(flow pcap.FlowKey, r io.Reader) error {
		fmt.Printf("\n=== 新 TCP 流 [%s] ===\n", flow)

		// 用 FrameReader 从重组后的字节流逐条切分消息。
		reader := gameproto.NewFrameReader(r)
		for {
			msgID, seqID, body, err := reader.ReadMessage()
			if err != nil {
				// io.EOF / 连接关闭 / 消息格式错误 / 加密消息（跳过）
				return err
			}

			// msgID 对应你的 proto 消息号，body 是 protobuf 字节流。
			// 用 proto.Unmarshal 解析具体的消息体：
			//
			//   switch msgID {
			//   case 1001: // 假设 1001 是 AuthLogin
			//       msg := &pb.AuthLogin{}
			//       if err := proto.Unmarshal(body, msg); err == nil {
			//           fmt.Printf("  AuthLogin: %+v\n", msg)
			//       }
			//   case 1002:
			//       msg := &pb.Move{}
			//       ...
			//   }

			fmt.Printf("  消息 ID=%d, Seq=%d, BodyLen=%d, Body=%x\n",
				msgID, seqID, len(body), body[:min(len(body), 32)])
		}
	})
	if err := c.RegisterHandler(gameHandler); err != nil {
		fmt.Fprintln(os.Stderr, "register handler:", err)
		return 1
	}

	// 信号监听，优雅退出。
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("开始抓包（端口 %d），Ctrl+C 退出...\n", port)
	if err := c.Capture(ctx, src, pcap.Target{}); err != nil {
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "\n收到退出信号，停止抓包。")
		} else {
			fmt.Fprintln(os.Stderr, "capture:", err)
			return 1
		}
	}

	// Capture 返回后 flush 残留流。
	gameHandler.Close()

	st := c.Stats()
	fmt.Printf("结束。captured=%d\n", st.Captured)
	for _, hs := range st.Handlers {
		fmt.Printf("  handler %s: received=%d processed=%d dropped=%d errors=%d\n",
			hs.Name, hs.Received, hs.Processed, hs.Dropped, hs.Errors)
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
