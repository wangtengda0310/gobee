// 流量代理工具 — 透明转发 TCP/HTTP 流量并显示通信内容
//
// 用法:
//
//	# 代理+录制模式
//	streamproxy -tcp :18000:10.254.114.204:18000 -http :20144:10.254.114.204:20144 -record session.json
//
//	# 重放模式（TCP 地址从录制文件读取）
//	streamproxy -replay session.json -replay-openid test1
//
// module: git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/cmd/tests/streamproxy
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	streamproto "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// tcpRedirectMap 远程 TCP 地址 → 本地重定向地址
// HTTP 代理用此映射替换响应中的 server_addr
var tcpRedirectMap sync.Map // key: "remoteHost:remotePort" → value: "127.0.0.1:localPort"

var (
	// globalRecorder 全局录制器（录制模式时设置）
	globalRecorder *streamproto.Recorder
	// globalRecordFile 录制文件保存路径（录制模式时设置）
	globalRecordFile string
)

func main() {
	log.SetFlags(0)

	tcpAddr := flag.String("tcp", "", "TCP 代理，格式 :localPort:remoteHost:remotePort")
	httpAddr := flag.String("http", "", "HTTP 代理，格式 :localPort:remoteHost:remotePort")
	recordFile := flag.String("record", "", "录制文件路径（JSON格式）")
	replayFile := flag.String("replay", "", "重放文件路径（JSON格式）")
	replayOpenID := flag.String("replay-openid", "test1", "重放时使用的登录账号")
	flag.Parse()

	// 重放模式
	if *replayFile != "" {
		// 读取录制文件获取 server_addr
		recData, err := os.ReadFile(*replayFile)
		if err != nil {
			log.Fatalf("[重放] 读取录制文件失败: %v", err)
		}
		var rec struct {
			ServerAddr string `json:"server_addr"`
		}
		if err := json.Unmarshal(recData, &rec); err != nil {
			log.Fatalf("[重放] 解析录制文件失败: %v", err)
		}
		if rec.ServerAddr == "" {
			log.Fatalf("[重放] 录制文件中缺少 server_addr")
		}

		tcpRemote := rec.ServerAddr
		// HTTP 地址：从 -http 参数或从 TCP 地址推导（同 IP，端口 20144）
		httpRemote := ""
		if *httpAddr != "" {
			_, httpRemote = parseAddrPair(*httpAddr, "http")
		} else if *tcpAddr != "" {
			_, httpRemote = parseAddrPair(*tcpAddr, "http")
		} else {
			// 从 TCP 地址推导：提取 IP，拼接 :20144
			colonIdx := strings.LastIndex(tcpRemote, ":")
			if colonIdx > 0 {
				httpRemote = tcpRemote[:colonIdx] + ":20144"
			}
		}

		if err := streamproto.Replay(*replayFile, tcpRemote, httpRemote, *replayOpenID, nil, nil); err != nil {
			log.Fatalf("[重放] 失败: %v", err)
		}
		return
	}

	if *tcpAddr == "" && *httpAddr == "" {
		fmt.Fprintln(os.Stderr, "用法: streamproxy -tcp :localPort:remoteHost:remotePort [-http :localPort:remoteHost:remotePort]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "示例:")
		fmt.Fprintln(os.Stderr, "  streamproxy -tcp :18000:10.254.114.204:18000 -http :20144:10.254.114.204:20144")
		fmt.Fprintln(os.Stderr, "  streamproxy -tcp :18000:10.254.114.204:18000")
		fmt.Fprintln(os.Stderr, "  streamproxy -http :20144:10.254.114.204:20144")
		os.Exit(1)
	}

	if *tcpAddr != "" {
		listen, remote := parseAddrPair(*tcpAddr, "tcp")
		// 注册 TCP 地址映射，供 HTTP 代理替换 server_addr 使用
		localAddr := "127.0.0.1" + listen // listen 格式为 ":port"
		tcpRedirectMap.Store(remote, localAddr)
		log.Printf("[MAP] TCP %s → %s", remote, localAddr)

		// 初始化录制器
		if *recordFile != "" {
			globalRecorder = streamproto.NewRecorder(remote)
			globalRecordFile = *recordFile
			globalRecorder.Start()
			log.Printf("[录制] 已启动，保存到 %s", *recordFile)
		}

		go startTCPProxy(listen, remote)
	}

	if *httpAddr != "" {
		listen, remote := parseAddrPair(*httpAddr, "http")
		go startHTTPProxy(listen, remote)
	}

	select {}
}

// parseAddrPair 解析 ":localPort:remoteHost:remotePort" 格式
func parseAddrPair(s, proto string) (listen, remote string) {
	if len(s) == 0 || s[0] != ':' {
		log.Fatalf("-%s 格式错误，应以 : 开头 (如 :9000:10.0.0.1:9000)", proto)
	}

	rest := s[1:] // "9000:10.0.0.1:9000"

	idx := 0
	for i, c := range rest {
		if c == ':' {
			idx = i
			break
		}
	}
	if idx == 0 {
		log.Fatalf("-%s 格式错误: %q", proto, s)
	}

	listenPort := rest[:idx]
	remoteAddr := rest[idx+1:]

	if listenPort == "" || remoteAddr == "" {
		log.Fatalf("-%s 格式错误: %q", proto, s)
	}

	return ":" + listenPort, remoteAddr
}

// ==================== TCP 代理 ====================

func startTCPProxy(listen, remote string) {
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("[TCP] 监听失败 %s → %s: %v", listen, remote, err)
	}
	log.Printf("[TCP] 监听 %s, 转发到 %s", listen, remote)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[TCP] 接受连接失败: %v", err)
			continue
		}
		go handleTCPConn(conn, remote)
	}
}

func handleTCPConn(clientConn net.Conn, remoteAddr string) {
	id := streamproto.ConnIDCounter.Add(1)
	startTime := time.Now()

	log.Printf("[TCP] #%d 新连接 %s → %s", id, clientConn.RemoteAddr(), remoteAddr)

	remoteConn, err := net.DialTimeout("tcp", remoteAddr, 5*time.Second)
	if err != nil {
		log.Printf("[TCP] #%d 连接目标失败: %v", id, err)
		clientConn.Close()
		return
	}

	log.Printf("[TCP] #%d 已连接目标 %s", id, remoteAddr)

	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端 → 服务端
	go func() {
		defer wg.Done()
		n := streamproto.RelayAndParse(clientConn, remoteConn, id, streamproto.DirClientToServer, true, globalRecorder, nil, nil, nil)
		log.Printf("[TCP] #%d 客户端→服务端 完成, %d 字节, 耗时 %s", id, n, time.Since(startTime).Truncate(time.Millisecond))
		remoteConn.Close()
	}()

	// 服务端 → 客户端
	go func() {
		defer wg.Done()
		n := streamproto.RelayAndParse(remoteConn, clientConn, id, streamproto.DirServerToClient, false, nil, nil, nil, nil)
		log.Printf("[TCP] #%d 服务端→客户端 完成, %d 字节, 耗时 %s", id, n, time.Since(startTime).Truncate(time.Millisecond))
		clientConn.Close()
	}()

	wg.Wait()

	// 连接关闭时保存录制文件
	if globalRecorder != nil {
		if err := streamproto.SaveRecordingToFile(globalRecordFile, globalRecorder.ToRecording()); err != nil {
			log.Printf("[录制] 保存失败: %v", err)
		} else {
			log.Printf("[录制] 已保存到 %s", globalRecordFile)
		}
	}

	log.Printf("[TCP] #%d 连接关闭", id)
}

// ==================== HTTP 代理 ====================

func startHTTPProxy(listen, remote string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ts := time.Now().Format("15:04:05.000")
		qs := ""
		if req.URL.RawQuery != "" {
			qs = "?" + req.URL.RawQuery
		}
		log.Printf("[HTTP] [%s] → %s %s%s", ts, req.Method, req.URL.Path, qs)

		// 构建到目标服务器的请求
		outURL := fmt.Sprintf("http://%s%s%s", remote, req.URL.Path, qs)
		if req.URL.RawQuery != "" {
			outURL = fmt.Sprintf("http://%s%s?%s", remote, req.URL.Path, req.URL.RawQuery)
		}

		outReq, err := http.NewRequest(req.Method, outURL, req.Body)
		if err != nil {
			log.Printf("[HTTP] [%s] ← 构建请求失败: %v", ts, err)
			http.Error(w, "代理请求失败", http.StatusBadGateway)
			return
		}

		// 复制请求头
		for k, vs := range req.Header {
			for _, v := range vs {
				outReq.Header.Add(k, v)
			}
		}
		outReq.Host = remote

		resp, err := http.DefaultClient.Do(outReq)
		if err != nil {
			log.Printf("[HTTP] [%s] ← 请求目标失败: %v", ts, err)
			http.Error(w, "目标不可达", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 读取响应体到内存（需要完整读取才能做地址替换）
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[HTTP] [%s] ← 读取响应体失败: %v", ts, err)
			http.Error(w, "读取响应失败", http.StatusBadGateway)
			return
		}

		// 替换 server_addr / server_list 中指向远程服务器的地址为本地代理地址
		// 服务端可能返回多个端口（如 18001, 18002），都需要重定向到本地 TCP 端口
		bodyStr := string(bodyBytes)
		replaced := false
		tcpRedirectMap.Range(func(key, value any) bool {
			remoteAddr := key.(string)  // "10.254.114.204:18000"
			localAddr := value.(string) // "127.0.0.1:18000"
			// 提取远程服务器 IP 部分（去掉端口）
			colonIdx := strings.LastIndex(remoteAddr, ":")
			if colonIdx < 0 {
				return true
			}
			remoteIP := remoteAddr[:colonIdx] // "10.254.114.204"
			// 替换所有该 IP 出现的地方（无论端口是什么）
			if strings.Contains(bodyStr, remoteIP) {
				// 用正则风格替换：remoteIP:任意端口 → localAddr
				// 简单实现：按 "remoteIP:" 前缀逐个替换
				prefix := remoteIP + ":"
				for {
					idx := strings.Index(bodyStr, prefix)
					if idx < 0 {
						break
					}
					// 找到端口结束位置（下一个非数字字符或字符串结尾）
					portStart := idx + len(prefix)
					portEnd := portStart
					for portEnd < len(bodyStr) && bodyStr[portEnd] >= '0' && bodyStr[portEnd] <= '9' {
						portEnd++
					}
					oldAddr := bodyStr[idx:portEnd]
					bodyStr = bodyStr[:idx] + localAddr + bodyStr[portEnd:]
					replaced = true
					log.Printf("[HTTP] [MAP] 地址替换: %s → %s", oldAddr, localAddr)
				}
			}
			return true
		})
		if replaced {
			bodyBytes = []byte(bodyStr)
			log.Printf("[HTTP] [DEBUG] 替换后响应体: %s", bodyStr)
		}

		// 复制响应头（Content-Length 由 Write 自动设置，跳过原始值）
		for k, vs := range resp.Header {
			if strings.EqualFold(k, "Content-Length") {
				continue // 替换后长度变了，让 Go 的 http 自动计算
			}
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// 写入（可能已替换的）响应体
		written, _ := w.Write(bodyBytes)

		ts2 := time.Now().Format("15:04:05.000")
		extra := ""
		if replaced {
			extra = " [已重定向]"
		}
		log.Printf("[HTTP] [%s] ← %d %d字节%s", ts2, resp.StatusCode, written, extra)
	})

	log.Printf("[HTTP] 监听 %s, 转发到 http://%s", listen, remote)
	if err := http.ListenAndServe(listen, handler); err != nil {
		log.Fatalf("[HTTP] 监听失败: %v", err)
	}
}
