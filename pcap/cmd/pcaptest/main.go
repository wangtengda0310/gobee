//go:build livecapture

// 命令 pcaptest：实时抓取 HTTP 请求（演示/调试用）。
//
// 构建 & 运行（需先安装 Npcap/libpcap，详见 README「安装」）：
//
//	# 列出可用网卡
//	CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -list
//
//	# 单网卡抓包
//	CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡名>
//
//	# 多网卡合并抓包 + BPF 过滤 + 输出到文件
//	CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡1>,<网卡2> -bpf "tcp port 80" -out dump.pcap
//
//	# 启用 HTTP 流重组，打印重组后的完整请求（Method/URL/Headers）
//	CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest -iface <网卡名> -http
//
// 说明：
//   - 本文件带 //go:build livecapture 标签，默认不参与编译，避免在无 Npcap 的环境下链接失败。
//   - 单元测试（itsnotfun_test.go / merge_test.go / reassembly_test.go）已覆盖离线等价场景。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gopacket/gopacket/layers" // 纯 Go（LinkType），本身不需 Npcap
	"github.com/gopacket/gopacket/pcapgo" // 纯 Go（pcap 文件读写），本身不需 Npcap
	"github.com/wangtengda0310/gobee/pcap"
	// ⚠️ 本文件整体需要 Npcap：通过本地 pcap 包的 NewLiveSource/ListInterfaces
	//    传递依赖 gopacket/pcap（cgo）。故带 //go:build livecapture 标签。
)

func main() {
	// 所有逻辑放在 run 中，确保 defer 正常执行；os.Exit 只在 main 顶层调用。
	os.Exit(run())
}

// run 执行命令行逻辑，返回进程退出码。
// 把 os.Exit 隔离在 main 里，是为了让本函数中的 defer（src.Close / lc.Close）
// 能在退出前正常执行——直接 os.Exit 会跳过 defer（exitAfterDefer 缺陷）。
func run() int {
	list := flag.Bool("list", false, "列出本机可用网卡后退出")
	iface := flag.String("iface", "", "网卡名，多个用逗号分隔（如 \"eth0,eth1\" 或 Windows 的 \"\\\\Device\\\\Npcap_{a},\\\\Device\\\\Npcap_{b}\"）")
	snaplen := flag.Int("snaplen", 65535, "每个包最大截获字节数")
	promisc := flag.Bool("promisc", false, "是否开启混杂模式")
	bpf := flag.String("bpf", "tcp port 80", "BPF 过滤表达式（留空则不过滤）")
	host := flag.String("host", "", "应用层 host 过滤（HTTP Host 头 / IP 子串，留空则不过滤）")
	out := flag.String("out", "", "输出到 pcap 文件路径（留空则不落盘）")
	httpMode := flag.Bool("http", false, "启用 HTTP/1.x 流重组，打印重组后的完整请求（Method/URL/Headers）")
	flag.Parse()

	// -list：列出网卡后退出。
	if *list {
		return listInterfaces()
	}

	if *iface == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -iface（或用 -list 查看可用网卡）")
		return 2
	}
	const maxSnaplen = 1<<31 - 1 // math.MaxInt32，pcap.OpenLive 的 snaplen 是 int32
	if *snaplen <= 0 || *snaplen > maxSnaplen {
		// snaplen 会被转成 int32 传给 pcap.OpenLive，这里前置校验避免溢出（gosec G115）。
		fmt.Fprintf(os.Stderr, "snaplen 超出有效范围 [1, %d]\n", maxSnaplen)
		return 2
	}

	// 解析 -iface（逗号分隔 → 多个网卡）。
	ifaceNames := splitAndTrim(*iface)

	// 打开一个或多个网卡。
	src, err := openSources(ifaceNames, int32(*snaplen), *promisc, *bpf) //nolint:gosec // G115: 已校验范围
	if err != nil {
		fmt.Fprintln(os.Stderr, "open source:", err)
		return 1
	}
	defer src.Close()

	c := pcap.NewCapturer(
		pcap.WithBufferSize(2048),
		pcap.WithOverflowStrategy(pcap.OverflowDrop),
		pcap.WithBPFFilter(*bpf),
	)
	// NewCapturer 返回的 capturer 同时实现 LifeCycler；显式关闭 worker。
	if lc, ok := c.(pcap.LifeCycler); ok {
		defer lc.Close()
	}

	// 注册打印型 handler（始终启用）。
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
		fmt.Fprintln(os.Stderr, "register printer:", err)
		return 1
	}

	// 可选：启用 HTTP 流重组，打印重组后的完整请求。
	var httpHandler *pcap.HTTPRequestHandler
	if *httpMode {
		httpHandler = pcap.NewHTTPRequestHandler("http", func(flow pcap.FlowKey, req *http.Request) error {
			fmt.Printf("\n=== HTTP 请求 [%s] ===\n", flow)
			fmt.Printf("%s %s HTTP/%d.%d\n", req.Method, req.URL.RequestURI(), req.ProtoMajor, req.ProtoMinor)
			fmt.Printf("Host: %s\n", req.Host)
			for k, vs := range req.Header {
				if k == "Host" {
					continue
				}
				for _, v := range vs {
					fmt.Printf("%s: %s\n", k, v)
				}
			}
			return nil
		})
		if err := c.RegisterHandler(httpHandler); err != nil {
			fmt.Fprintln(os.Stderr, "register http handler:", err)
			return 1
		}
	}

	// 可选：输出到 pcap 文件。
	if *out != "" {
		if err := registerFileWriter(c, *out, src.LinkType()); err != nil {
			fmt.Fprintln(os.Stderr, "register filewriter:", err)
			return 1
		}
	}

	// 监听信号，优雅退出。
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	target := pcap.Target{Host: *host} // BPF 已在 NewLiveSource 应用，这里只用 Host 做用户态二次过滤
	fmt.Printf("开始在 %s 上抓取（bpf=%q host=%q out=%q），Ctrl+C 退出...\n",
		src, *bpf, *host, *out)
	if err := c.Capture(ctx, src, target); err != nil {
		// 区分「被信号取消」（Ctrl+C，正常优雅退出）与「真实错误」。
		// ctx 被 SIGINT/SIGTERM 取消时 Capture 返回 context.Canceled，视为正常退出。
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "\n收到退出信号，停止抓包。")
		} else {
			fmt.Fprintln(os.Stderr, "capture:", err)
			return 1
		}
	}

	// Capture 返回后 flush HTTP 流重组的残留流（未 FIN 的请求会被 Close 触发出回调）。
	if httpHandler != nil {
		if err := httpHandler.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "close http handler:", err)
		}
	}

	st := c.Stats()
	fmt.Printf("结束。captured=%d\n", st.Captured)
	for _, hs := range st.Handlers {
		fmt.Printf("  handler %s: received=%d processed=%d dropped=%d errors=%d\n",
			hs.Name, hs.Received, hs.Processed, hs.Dropped, hs.Errors)
	}
	return 0
}

// listInterfaces 列出本机网卡并返回退出码。
func listInterfaces() int {
	devs, err := pcap.ListInterfaces()
	if err != nil {
		fmt.Fprintln(os.Stderr, "list interfaces:", err)
		return 1
	}
	fmt.Printf("找到 %d 个网卡：\n", len(devs))
	for _, d := range devs {
		fmt.Printf("  - %s", d.Name)
		if d.Description != "" {
			fmt.Printf("  (%s)", d.Description)
		}
		fmt.Println()
		for _, ip := range d.IPs {
			fmt.Printf("      IP: %s\n", ip)
		}
	}
	return 0
}

// openSources 打开一个或多个网卡，返回（可能是合并的）Source。
// 单网卡直接返回该网卡 Source；多网卡用 NewMergedSource 合并。
func openSources(ifaces []string, snaplen int32, promisc bool, bpf string) (pcap.Source, error) {
	if len(ifaces) == 1 {
		return pcap.NewLiveSource(ifaces[0], snaplen, promisc, bpf)
	}
	// 多网卡：逐个打开后合并。
	sources := make([]pcap.Source, 0, len(ifaces))
	for _, name := range ifaces {
		s, err := pcap.NewLiveSource(name, snaplen, promisc, bpf)
		if err != nil {
			// 打开失败：关闭已打开的，返回错误。
			for _, opened := range sources {
				_ = opened.Close()
			}
			return nil, fmt.Errorf("open %q: %w", name, err)
		}
		sources = append(sources, s)
	}
	merged, err := pcap.NewMergedSource(sources...)
	if err != nil {
		for _, s := range sources {
			_ = s.Close()
		}
		return nil, err
	}
	return merged, nil
}

// registerFileWriter 注册一个把包写入 pcap 文件的 handler。
// linkType 来自 Source.LinkType()，用于 pcap 文件头。
func registerFileWriter(c pcap.Capturer, path string, linkType layers.LinkType) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	writer := pcapgo.NewWriter(f)
	if err := writer.WriteFileHeader(65535, linkType); err != nil {
		f.Close()
		return fmt.Errorf("write pcap header: %w", err)
	}
	return c.RegisterHandler(pcap.NewHandlerFunc("filewriter", func(ctx context.Context, e *pcap.PacketEvent) error {
		ci := e.Packet.Metadata().CaptureInfo
		return writer.WritePacket(ci, e.Packet.Data())
	}))
	// 注意：f 不在此关闭——handler 生命周期与 Capture 一致，Capture 结束后进程通常退出。
	// 长跑场景若需优雅关闭文件，应改用 LifeCycler 钩子（当前简化处理）。
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

// splitAndTrim 按逗号分隔并去除空白。
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
