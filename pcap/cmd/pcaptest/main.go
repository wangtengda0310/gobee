//go:build livecapture

// 命令 pcaptest：实时抓取 itsnot.fun 的 HTTP 请求（演示/调试用）。
//
// 构建 & 运行（需先安装 Npcap/libpcap）：
//
//	CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡名或设备路径>
//
// 说明：
//   - 本文件带 //go:build livecapture 标签，默认不参与编译，避免在无 Npcap 的环境下链接失败。
//   - 这是 README 中「TODO：真实抓包工具」的最小可用实现，便于在装了 Npcap 的机器上
//     验证整条实时抓包链路。单元测试（itsnotfun_test.go）已覆盖离线等价场景。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wangtengda0310/gobee/pcap"
)

func main() {
	// 所有逻辑放在 run 中，确保 defer 正常执行；osExit 只在 main 顶层调用。
	os.Exit(run())
}

// run 执行命令行逻辑，返回进程退出码。
// 把 os.Exit 隔离在 main 里，是为了让本函数中的 defer（src.Close / lc.Close）
// 能在退出前正常执行——直接 os.Exit 会跳过 defer（exitAfterDefer 缺陷）。
func run() int {
	iface := flag.String("iface", "", "网卡名 (Linux: eth0) 或设备路径 (Windows: \\Device\\Npcap_{...})")
	snaplen := flag.Int("snaplen", 65535, "每个包最大截获字节数")
	promisc := flag.Bool("promisc", false, "是否开启混杂模式")
	host := flag.String("host", "itsnot.fun", "应用层 host 过滤（HTTP Host 头 / IP 子串）")
	flag.Parse()

	if *iface == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -iface")
		return 2
	}
	const maxSnaplen = 1<<31 - 1 // math.MaxInt32，pcap.OpenLive 的 snaplen 是 int32
	if *snaplen <= 0 || *snaplen > maxSnaplen {
		// snaplen 会被转成 int32 传给 pcap.OpenLive，这里前置校验避免溢出（gosec G115）。
		fmt.Fprintf(os.Stderr, "snaplen 超出有效范围 [1, %d]\n", maxSnaplen)
		return 2
	}

	// 打开网卡（需要 livecapture 构建标签 + libpcap/Npcap）。
	// snaplen 已在上方校验落在 int32 正数范围内，转换安全。
	src, err := pcap.NewLiveSource(*iface, int32(*snaplen), //nolint:gosec // G115: 已校验范围
		*promisc, "tcp port 80")
	if err != nil {
		fmt.Fprintln(os.Stderr, "open live source:", err)
		return 1
	}
	defer src.Close()

	c := pcap.NewCapturer(
		pcap.WithBufferSize(2048),
		pcap.WithOverflowStrategy(pcap.OverflowDrop),
	)
	// NewCapturer 返回的 capturer 同时实现 LifeCycler；显式关闭 worker。
	if lc, ok := c.(pcap.LifeCycler); ok {
		defer lc.Close()
	}

	// 注册一个打印型处理函数。
	if err := c.RegisterHandler(pcap.NewHandlerFunc("printer", func(ctx context.Context, e *pcap.PacketEvent) error {
		app := e.Packet.ApplicationLayer()
		payload := ""
		if app != nil {
			payload = string(app.Payload())
		}
		fmt.Printf("[%s] %s -> %s | %s\n",
			e.Timestamp.Format(time.RFC3339Nano),
			e.NetworkFlow.Src(), e.NetworkFlow.Dst(), firstLine(payload))
		return nil
	})); err != nil {
		fmt.Fprintln(os.Stderr, "register handler:", err)
		return 1
	}

	// 监听信号，优雅退出。
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("开始在 %s 上抓取 host=%s 的 HTTP 流量，Ctrl+C 退出...\n", *iface, *host)
	if err := c.Capture(ctx, src, pcap.Target{Host: *host}); err != nil {
		fmt.Fprintln(os.Stderr, "capture:", err)
		return 1
	}

	st := c.Stats()
	fmt.Printf("结束。captured=%d\n", st.Captured)
	for _, hs := range st.Handlers {
		fmt.Printf("  handler %s: received=%d processed=%d dropped=%d errors=%d\n",
			hs.Name, hs.Received, hs.Processed, hs.Dropped, hs.Errors)
	}
	return 0
}

// firstLine 返回 payload 的第一行，便于单行打印 HTTP 请求行。
func firstLine(s string) string {
	for i, r := range s {
		if r == '\r' || r == '\n' {
			return s[:i]
		}
	}
	return s
}
