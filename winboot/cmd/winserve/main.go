package main

// winserve: winboot 更新部署服务 (deploy/serve.py 的 Go 复刻, 单文件零依赖).
//
// 用法:   winserve [-http 8760] [-udp 58765] [-dir <服务目录, 默认当前目录>]
// 部署:   同目录放 winboot.exe + winboot-manifest.json (构建产物)
// 客户端: UPDATE_BASE 指向 http://<部署机IP>:<端口>; 客户端也可不配 ——
//         本服务的 UDP 应答会把地址告诉它们 (DHCP 自愈).
//
// 本服务同时是发现协议的一个"对等方": 应答 WINBOOT_DISCOVER, 携带
//   tool (版本/sha256/size, 读 winboot-manifest.json) -> 客户端自治更新可直接选中它
//   files (目录内若有安装包缓存, 顺带成为缓存源)
//   update_base (自己的运行时地址) -> 客户端经此学到中心地址
// UDP 端口被占则静默跳过 (部署机同时跑着 winboot 实例时, 由实例转发中心地址).

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const discoverMsg = "WINBOOT_DISCOVER"

// 与 lanshare 的 SHAREABLE 对齐: 目录里放了这些名字的文件就会被宣告为缓存源.
var shareableNames = []string{"DockerDesktopInstaller.exe", "wsl-x64.msix"}

func lanIP() string {
	conn, err := net.Dial("udp", "10.255.255.255:1")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "?"
	}
	return h
}

// buildPeerManifest 构造发现应答 (与 winboot 实例的 manifest 同 schema).
func buildPeerManifest(dir string, httpPort int) map[string]interface{} {
	m := map[string]interface{}{
		"app":         "winboot",
		"proto":       1,
		"host":        hostname(),
		"port":        httpPort,
		"files":       []interface{}{},
		"update_base": fmt.Sprintf("http://%s:%d", lanIP(), httpPort),
	}
	// tool: 读部署清单 (version/sha256/size)
	if b, err := os.ReadFile(filepath.Join(dir, "winboot-manifest.json")); err == nil {
		var mf map[string]interface{}
		if json.Unmarshal(b, &mf) == nil {
			tool := map[string]interface{}{}
			if v, ok := mf["version"]; ok {
				tool["version"] = fmt.Sprintf("%v", v)
			}
			if v, ok := mf["sha256"]; ok {
				tool["sha256"] = fmt.Sprintf("%v", v)
			}
			if v, ok := mf["size"]; ok {
				tool["size"] = v
			}
			if tool["version"] != nil {
				m["tool"] = tool
			}
		}
	}
	// files: 目录内若有安装包, 顺带宣告为缓存源 (size 校验用, 不算 sha)
	files := []interface{}{}
	for _, name := range shareableNames {
		if st, err := os.Stat(filepath.Join(dir, name)); err == nil && !st.IsDir() {
			files = append(files, map[string]interface{}{"name": name, "size": st.Size()})
		}
	}
	m["files"] = files
	return m
}

func udpResponder(dir string, httpPort, udpPort int) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: udpPort})
	if err != nil {
		fmt.Printf("UDP %d 被占, 跳过发现应答 (本机跑着 winboot 实例时会由它转发中心地址)\n", udpPort)
		return
	}
	defer conn.Close()
	fmt.Printf("UDP 发现应答已开启 (%d)\n", udpPort)
	buf := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		if strings.TrimSpace(string(buf[:n])) != discoverMsg {
			continue
		}
		reply, _ := json.Marshal(buildPeerManifest(dir, httpPort))
		_, _ = conn.WriteToUDP(reply, addr) // 单播回应: 消费方防火墙有状态放行
	}
}

// noCache 静态文件 + 禁缓存头 (更新检查必须拿到新鲜清单).
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		h.ServeHTTP(w, r)
	})
}

func main() {
	httpPort := flag.Int("http", 8760, "HTTP 服务端口")
	udpPort := flag.Int("udp", 58765, "UDP 发现应答端口 (被占则静默跳过)")
	dir := flag.String("dir", ".", "服务目录 (放 winboot.exe + winboot-manifest.json)")
	flag.Parse()

	abs, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Println("目录解析失败:", err)
		os.Exit(1)
	}
	go udpResponder(abs, *httpPort, *udpPort)
	fmt.Printf("更新服务: http://<本机IP>:%d  (目录: %s)\nCtrl+C 停止\n", *httpPort, abs)
	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", *httpPort),
		Handler:           noCache(http.FileServer(http.Dir(abs))),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("服务退出:", err)
		os.Exit(1)
	}
}
