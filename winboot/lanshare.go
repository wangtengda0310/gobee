package main

// lanshare.go: LAN 下载源 + 排查通道 (参考 PyQt 实现移植, 协议互通).
//
// 协议与 PyQt 版完全兼容 (同一局域网可互通: 互发现/互为镜像/互收告警):
//   UDP 58765: "WINBOOT_DISCOVER" -> 单播 JSON manifest; 异常广播 JSON (type=anomaly)
//   HTTP 8765: GET /<缓存白名单文件> (Range 续传) + /winboot.exe + /logs + /logs.zip
//
// 与 Python 版的刻意差异:
//   - Range 续传交给 http.ServeContent (原生支持, 比 Python 手写的还全)
//   - 端口绑定失败即报错 (Go 默认无 SO_REUSEADDR, 不存在 Windows 端口抢夺坑)

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
)

const (
	lanHTTPPort   = 8765
	lanUDPPort    = 58765
	lanProto      = 1
	lanAppID      = "winboot"
	lanToolFile   = "winboot.exe"
	lanContainer  = "sw_server_dev"
	lanDiscoverTx = "WINBOOT_DISCOVER"
)

// shareable 缓存白名单 (与 scripts/lib/acquire.ps1 各组件 FileName/MinBytes 约定一致,
// 缓存都在 %TEMP%). 新组件加入共享只需在此登记一行.
var shareable = []struct {
	Name     string
	MinBytes int64
}{
	{"DockerDesktopInstaller.exe", 200 << 20},
	{"wsl-x64.msix", 100 << 20},
}

// ---- manifest / anomaly 的 JSON 结构 (字段名与 Python 版逐字对齐) ----

type fileEntry struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256,omitempty"`
}

type toolEntry struct {
	Version string `json:"version"`
	Size    int64  `json:"size"`
	SHA256  string `json:"sha256"`
}

type lanManifest struct {
	App        string      `json:"app"`
	Proto      int         `json:"proto"`
	Host       string      `json:"host"`
	Port       int         `json:"port"`
	Files      []fileEntry `json:"files"`                 // 必须初始化为空切片: null 会让 Python 端迭代崩溃
	Tool       *toolEntry  `json:"tool,omitempty"`        // 工具本体宣告 (自治更新/镜像)
	UpdateBase string      `json:"update_base,omitempty"` // 本机知晓的更新中心地址 (转发给发现方)
	Flavor     string      `json:"flavor,omitempty"`      // 栈身份 (异构 peer 互不采纳为更新源)
	ToolURL    string      `json:"tool_url,omitempty"`    // TOS 门票: 工具本体的预签名直读 URL (TTL 900s, 现场签发)
	ToolHost   string      `json:"tool_host,omitempty"`   // 门票下载必带的 Host 头值 (网关按 Host 路由)
}

type anomalyMsg struct {
	App     string  `json:"app"`
	Type    string  `json:"type"`
	Proto   int     `json:"proto"`
	Host    string  `json:"host"`
	Version string  `json:"version"`
	Command string  `json:"command"`
	Code    int     `json:"code"`
	Tail    string  `json:"tail"`
	TS      float64 `json:"ts"`
}

// ---- sha256 缓存 (600MB 级文件, 按 mtime+size 失效; 与 Python 版同策略) ----

var (
	shaCache   = map[string]shaCacheEntry{}
	shaCacheMu sync.Mutex
)

type shaCacheEntry struct {
	mtimeNano int64
	size      int64
	digest    string
}

func sha256OfFile(path string) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	shaCacheMu.Lock()
	if c, ok := shaCache[path]; ok && c.mtimeNano == st.ModTime().UnixNano() && c.size == st.Size() {
		shaCacheMu.Unlock()
		return c.digest, nil
	}
	shaCacheMu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	digest := hex.EncodeToString(h.Sum(nil))
	shaCacheMu.Lock()
	shaCache[path] = shaCacheEntry{st.ModTime().UnixNano(), st.Size(), digest}
	shaCacheMu.Unlock()
	return digest, nil
}

// ---- 公共小工具 ----

func lanCacheDir() string { return os.TempDir() }

func lanSharePath(name string) string { return filepath.Join(lanCacheDir(), name) }

func localIP() string {
	conn, err := net.Dial("udp", "10.255.255.255:1")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// broadcastTargets 广播目标: 受限广播 + 各本机 IPv4 的 /24 定向广播 (启发式, 与 Python 版一致).
func broadcastTargets() []string {
	ips := map[string]bool{}
	if ip := localIP(); !strings.HasPrefix(ip, "127.") {
		ips[ip] = true
	}
	if host, err := os.Hostname(); err == nil {
		if addrs, err := net.LookupHost(host); err == nil {
			for _, a := range addrs {
				if ip := net.ParseIP(a); ip != nil && ip.To4() != nil && !strings.HasPrefix(a, "127.") {
					ips[a] = true
				}
			}
		}
	}
	targets := []string{"255.255.255.255"}
	for ip := range ips {
		targets = append(targets, ip[:strings.LastIndex(ip, ".")]+".255")
	}
	return targets
}

// resolveShared 返回可服务的本地路径; 白名单外返回 "".
func resolveShared(name string) string {
	if name == lanToolFile {
		if exe, err := os.Executable(); err == nil {
			return exe
		}
		return ""
	}
	for _, s := range shareable {
		if s.Name == name {
			return lanSharePath(name)
		}
	}
	return ""
}

// buildManifest 当前可共享清单 (files 必为非 nil 数组, 见 lanManifest 注释).
// updateBase 非空时随应答转发: 消费方在自身配置失效/缺失时经此学到中心地址 (DHCP 自愈).
func buildManifest(port int, updateBase string) lanManifest {
	m := lanManifest{
		App: lanAppID, Proto: lanProto,
		Host: hostname(), Port: port,
		Files:      []fileEntry{},
		UpdateBase: updateBase,
		Flavor:     TOOL_FLAVOR,
	}
	for _, s := range shareable {
		p := lanSharePath(s.Name)
		if st, err := os.Stat(p); err == nil && st.Size() >= s.MinBytes {
			e := fileEntry{Name: s.Name, Size: st.Size()}
			if d, err := sha256OfFile(p); err == nil {
				e.SHA256 = d
			}
			m.Files = append(m.Files, e)
		}
	}
	if exe, err := os.Executable(); err == nil {
		if st, err := os.Stat(exe); err == nil {
			t := toolEntry{Version: TOOL_VERSION, Size: st.Size()}
			if d, err := sha256OfFile(exe); err == nil {
				t.SHA256 = d
			}
			m.Tool = &t
		}
	}
	// TOS 门票 (引导场景 "首份 exe" 获取链的中继级): 本机持有 TOS 凭证时随应答
	// 现场签发短时效直读 URL —— 凭证零传播, 消费方 (引导脚本) 只做一个带 Host
	// 头的 GET. 应答是逐请求构造的, 门票始终新鲜; 未配置 TOS 渠道则字段留空.
	if u, h := TOSPresignURL(lanToolFile, 900*time.Second); u != "" {
		m.ToolURL, m.ToolHost = u, h
	}
	return m
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "?"
	}
	return h
}

// ---------------------------------------------------------------------------
// 提供方: HTTP 文件服务 + 排查报告 + UDP 发现应答/异常接收
// ---------------------------------------------------------------------------

type LanProvider struct {
	httpPort   int
	udpPort    int
	onAnomaly  func(anomalyMsg) // 收到他人异常广播 (UDP goroutine 回调, 须自行投递回 UI 线程)
	guiLogFn   func() string    // GUI 日志快照 (供 /logs 报告)
	updateBase string           // 本机知晓的更新中心地址 (非空则随应答转发)

	listener net.Listener
	udpConn  *net.UDPConn
	stopCh   chan struct{}
	wg       sync.WaitGroup

	// mu 保护 udpConn/stopped: 重试接管 goroutine 与 Stop 并发访问
	mu     sync.Mutex
	stopped bool
}

func NewLanProvider(onAnomaly func(anomalyMsg), guiLogFn func() string, updateBase string) *LanProvider {
	return &LanProvider{
		httpPort:   lanHTTPPort,
		udpPort:    lanUDPPort,
		onAnomaly:  onAnomaly,
		guiLogFn:   guiLogFn,
		updateBase: updateBase,
		stopCh:     make(chan struct{}),
	}
}

func (p *LanProvider) BaseURL() string {
	return fmt.Sprintf("http://%s:%d", localIP(), p.httpPort)
}

// Start 启动服务; HTTP 端口被占/无可共享内容时返回错误 (调用方静默处理).
//
// UDP 58765 被占不视为失败 (部署机同时跑着 winserve 的应答器是常态):
// 降级为 HTTP-only 继续服务, 后台每 15s 重试接管 —— 对端检测到本机 8765
// 有 provider 后会主动让出.
func (p *LanProvider) Start() error {
	m := buildManifest(p.httpPort, p.updateBase)
	if len(m.Files) == 0 && m.Tool == nil {
		return fmt.Errorf("本地无可共享内容 (检查 %%TEMP%% 缓存)")
	}
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", p.httpPort))
	if err != nil {
		return fmt.Errorf("HTTP 端口 %d 绑定失败: %w", p.httpPort, err)
	}
	p.listener = ln
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleHTTP)
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		_ = http.Serve(ln, mux)
	}()

	if !p.tryBindUDP() {
		go p.udpRetryLoop()
	}
	return nil
}

// tryBindUDP 绑定 UDP 发现端口; 失败返回 false.
func (p *LanProvider) tryBindUDP() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.udpConn != nil || p.stopped {
		return p.udpConn != nil
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: p.udpPort})
	if err != nil {
		return false
	}
	p.udpConn = conn
	p.wg.Add(1)
	go p.udpLoop()
	return true
}

// udpRetryLoop UDP 被占时的后台接管循环 (对端让出后本方最终持有).
func (p *LanProvider) udpRetryLoop() {
	for {
		time.Sleep(15 * time.Second)
		p.mu.Lock()
		stopped := p.stopped
		hasUDP := p.udpConn != nil
		p.mu.Unlock()
		if stopped || hasUDP {
			return
		}
		p.tryBindUDP()
	}
}

func (p *LanProvider) Stop() {
	close(p.stopCh)
	p.mu.Lock()
	p.stopped = true
	if p.listener != nil {
		p.listener.Close()
	}
	udp := p.udpConn
	p.udpConn = nil
	p.mu.Unlock()
	if udp != nil {
		udp.Close()
	}
	p.wg.Wait()
}

func (p *LanProvider) udpLoop() {
	defer p.wg.Done()
	buf := make([]byte, 2048)
	for {
		p.mu.Lock()
		conn := p.udpConn
		p.mu.Unlock()
		if conn == nil {
			return
		}
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return // Stop() 关闭了连接
		}
		msg := bytes.TrimSpace(buf[:n])
		if string(msg) == lanDiscoverTx {
			reply, _ := json.Marshal(buildManifest(p.httpPort, p.updateBase))
			_, _ = conn.WriteToUDP(reply, addr) // 单播回应
			continue
		}
		// 其余: 异常广播 —— 只处理合法 JSON 且非自己的
		var a anomalyMsg
		if json.Unmarshal(msg, &a) != nil {
			continue
		}
		if a.App != lanAppID || a.Type != "anomaly" || a.Host == "" || a.Host == hostname() {
			continue
		}
		if p.onAnomaly != nil {
			func() {
				defer func() { _ = recover() }() // 回调异常不影响接收循环
				p.onAnomaly(a)
			}()
		}
	}
}

func (p *LanProvider) handleHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	switch name {
	case "logs":
		body := buildLogsHTML(p.guiLogFn)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(body))
		return
	case "logs.zip":
		body := buildLogsZip(p.guiLogFn)
		w.Header().Set("Content-Type", "application/zip")
		w.Write(body)
		return
	}
	path := resolveShared(name)
	if path == "" {
		http.Error(w, "not shared", http.StatusNotFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "not shared", http.StatusNotFound)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		http.Error(w, "not shared", http.StatusNotFound)
		return
	}
	// ServeContent 原生处理 Range/断点续传 (acquire.ps1 的 curl -C - 依赖它)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, name, st.ModTime(), f)
}

// ---------------------------------------------------------------------------
// 异常广播 (fire-and-forget; 摘要 <1.4KB, 全量取证靠 GET /logs)
// ---------------------------------------------------------------------------

func broadcastAnomaly(command string, code int, tail string) {
	if len(tail) > 300 {
		tail = tail[len(tail)-300:]
	}
	payload := anomalyMsg{
		App: lanAppID, Type: "anomaly", Proto: lanProto,
		Host: hostname(), Version: TOOL_VERSION,
		Command: truncateStr(command, 40), Code: code, Tail: tail,
		TS: float64(time.Now().UnixNano()) / 1e9,
	}
	data, _ := json.Marshal(payload)
	if len(data) > 1400 {
		payload.Tail = ""
		data, _ = json.Marshal(payload)
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return
	}
	defer conn.Close()
	for _, t := range broadcastTargets() {
		if ip := net.ParseIP(t); ip != nil {
			_, _ = conn.WriteToUDP(data, &net.UDPAddr{IP: ip, Port: lanUDPPort})
		}
	}
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// ---------------------------------------------------------------------------
// 消费方: 周期广播发现, 收集提供方 (15s 过期)
// ---------------------------------------------------------------------------

type providerInfo struct {
	URL        string      `json:"url"`
	Host       string      `json:"host"`
	Files      []fileEntry `json:"files"`
	Tool       *toolEntry  `json:"tool"`
	UpdateBase string      `json:"update_base,omitempty"`
	Flavor     string      `json:"flavor,omitempty"`
	lastSeen   time.Time
}

type LanDiscovery struct {
	interval   time.Duration
	expire     time.Duration
	onProviders func([]providerInfo)

	mu        sync.Mutex
	providers map[string]*providerInfo // ip -> info
	running   bool
}

func NewLanDiscovery(onProviders func([]providerInfo)) *LanDiscovery {
	return &LanDiscovery{
		interval:    5 * time.Second,
		expire:      15 * time.Second,
		onProviders: onProviders,
		providers:   map[string]*providerInfo{},
	}
}

func (d *LanDiscovery) Start() {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return
	}
	d.running = true
	d.mu.Unlock()
	go d.loop()
}

func (d *LanDiscovery) Stop() {
	d.mu.Lock()
	d.running = false
	d.mu.Unlock()
}

func (d *LanDiscovery) loop() {
	for {
		d.mu.Lock()
		running := d.running
		d.mu.Unlock()
		if !running {
			return
		}
		d.round()
		time.Sleep(d.interval)
	}
}

func (d *LanDiscovery) round() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(d.interval))
	for _, t := range broadcastTargets() {
		if ip := net.ParseIP(t); ip != nil {
			_, _ = conn.WriteToUDP([]byte(lanDiscoverTx), &net.UDPAddr{IP: ip, Port: lanUDPPort})
		}
	}
	buf := make([]byte, 65536)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			break // 本轮超时
		}
		d.onReply(buf[:n], from.IP.String())
	}
	d.expireAndEmit()
}

func (d *LanDiscovery) onReply(data []byte, ip string) {
	var m lanManifest
	if json.Unmarshal(data, &m) != nil || m.App != lanAppID || m.Proto > lanProto {
		return
	}
	allowed := map[string]bool{}
	for _, s := range shareable {
		allowed[s.Name] = true
	}
	files := []fileEntry{}
	for _, f := range m.Files {
		if allowed[f.Name] {
			files = append(files, f)
		}
	}
	port := m.Port
	if port == 0 {
		port = lanHTTPPort
	}
	updateBase := m.UpdateBase
	if !strings.HasPrefix(updateBase, "http") {
		updateBase = "" // 非法地址丢弃 (防注入)
	}
	d.mu.Lock()
	d.providers[ip] = &providerInfo{
		URL:        fmt.Sprintf("http://%s:%d", ip, port),
		Host:       m.Host,
		Files:      files,
		Tool:       m.Tool,
		UpdateBase: updateBase,
		Flavor:     m.Flavor,
		lastSeen:   time.Now(),
	}
	d.mu.Unlock()
}

func (d *LanDiscovery) expireAndEmit() {
	d.mu.Lock()
	now := time.Now()
	for ip, p := range d.providers {
		if now.Sub(p.lastSeen) > d.expire {
			delete(d.providers, ip)
		}
	}
	list := make([]providerInfo, 0, len(d.providers))
	for _, p := range d.providers {
		list = append(list, *p)
	}
	d.mu.Unlock()
	sort.SliceStable(list, func(i, j int) bool { return len(list[i].Files) > len(list[j].Files) })
	if d.onProviders != nil {
		d.onProviders(list)
	}
}

// Snapshot 当前提供方快照 (稳定排序: 文件多者在前).
func (d *LanDiscovery) Snapshot() []providerInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	list := make([]providerInfo, 0, len(d.providers))
	for _, p := range d.providers {
		list = append(list, *p)
	}
	sort.SliceStable(list, func(i, j int) bool { return len(list[i].Files) > len(list[j].Files) })
	return list
}

// PickInstallSource 选安装用源: 优先有 DD 安装器的, 多个取文件最多者.
func PickInstallSource(providers []providerInfo) *providerInfo {
	for i := range providers {
		for _, f := range providers[i].Files {
			if f.Name == "DockerDesktopInstaller.exe" {
				return &providers[i]
			}
		}
	}
	if len(providers) > 0 {
		return &providers[0]
	}
	return nil
}

// NewestToolPeer 局域网里工具版本最新的【同 flavor】提供方 (中心失联时的自治更新源).
// 异构栈 (无 flavor 字段) 不采纳 —— 防跨栈蚕食.
func NewestToolPeer(providers []providerInfo) *providerInfo {
	var best *providerInfo
	for i := range providers {
		if providers[i].Flavor != TOOL_FLAVOR {
			continue
		}
		t := providers[i].Tool
		if t == nil {
			continue
		}
		if best == nil || VersionGt(t.Version, best.Tool.Version) {
			best = &providers[i]
		}
	}
	return best
}

// FindToolMirror 找能提供指定 version(+sha 一致) 工具本体的【同 flavor】局域网镜像.
func FindToolMirror(providers []providerInfo, version, sha256 string) *providerInfo {
	for i := range providers {
		if providers[i].Flavor != TOOL_FLAVOR {
			continue
		}
		t := providers[i].Tool
		if t == nil || t.Version != version {
			continue
		}
		if sha256 != "" && !strings.EqualFold(t.SHA256, sha256) {
			continue // 镜像哈希与权威清单不一致: 弃用 (防篡改)
		}
		return &providers[i]
	}
	return nil
}

// DiscoveredUpdateBase 从发现结果里学更新中心地址 —— 仅信任【同 flavor】应答方
// (异构部署中心会发来别栈的 exe).
func DiscoveredUpdateBase(providers []providerInfo) string {
	for i := range providers {
		if providers[i].Flavor == TOOL_FLAVOR && providers[i].UpdateBase != "" {
			return providers[i].UpdateBase
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// 后台预取: 缺失缓存提前拉到本机 %TEMP% (.prefetch.part 独立命名, 与 install 的 .part 并发安全)
// ---------------------------------------------------------------------------

func prefetchMissing(source *providerInfo, log func(string)) []string {
	var got []string
	for _, s := range shareable {
		cache := lanSharePath(s.Name)
		if st, err := os.Stat(cache); err == nil && st.Size() >= s.MinBytes {
			continue // 本机已有
		}
		var entry *fileEntry
		for i := range source.Files {
			if source.Files[i].Name == s.Name {
				entry = &source.Files[i]
				break
			}
		}
		if entry == nil {
			continue // 对方也没有
		}
		part := cache + ".prefetch.part"
		url := source.URL + "/" + s.Name
		n, err := downloadResumable(url, part, entry.Size, log)
		if err != nil {
			log(fmt.Sprintf("%s 预取失败: %v", s.Name, err))
			continue
		}
		if err := os.Rename(part, cache); err != nil {
			log(fmt.Sprintf("%s 预取落盘失败: %v", s.Name, err))
			continue
		}
		got = append(got, s.Name)
		log(fmt.Sprintf("%s 预取完成 (%dMB) <- %s", s.Name, n>>20, source.URL))
	}
	return got
}

func downloadResumable(url, partPath string, expectedSize int64, log func(string)) (int64, error) {
	var start int64
	if st, err := os.Stat(partPath); err == nil {
		start = st.Size()
	}
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	if start > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if start > 0 && resp.StatusCode != http.StatusPartialContent {
		start = 0 // 服务端不支持续传: 从头来
	}
	f, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if start == 0 {
		f.Truncate(0)
		f.Seek(0, 0)
	} else {
		f.Seek(start, 0)
	}
	var total int64 = start
	nextLog := int64(50 << 20)
	buf := make([]byte, 1<<20)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return 0, werr
			}
			total += int64(n)
			if total >= nextLog {
				log(fmt.Sprintf("预取中... %dMB", total>>20))
				nextLog += 50 << 20
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	if expectedSize > 0 && total != expectedSize {
		return 0, fmt.Errorf("尺寸不符: 期望 %d, 实得 %d", expectedSize, total)
	}
	return total, nil
}

// ---------------------------------------------------------------------------
// 排查报告: /logs 与 /logs.zip
// ---------------------------------------------------------------------------

func dockerSnapshot() string {
	exe := dockerPath()
	type step struct {
		label   string
		args    []string
		timeout time.Duration
	}
	steps := []step{
		{"docker version", []string{"version"}, 10 * time.Second},
		{"容器状态 (" + lanContainer + ")", []string{"ps", "-a", "--filter", "name=" + lanContainer,
			"--format", "table {{.Names}}\t{{.Status}}\t{{.Image}}"}, 10 * time.Second},
		{"镜像列表", []string{"images", "--format", "{{.Repository}}:{{.Tag}}\t{{.Size}}"}, 15 * time.Second},
		{"容器日志尾 (最近 50 行)", []string{"logs", "--tail", "50", lanContainer}, 15 * time.Second},
	}
	var sb strings.Builder
	for _, s := range steps {
		cmd := exec.Command(exe, s.args...)
		cmd.SysProcAttr = hideWindow()
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		done := make(chan error, 1)
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(&sb, "=== %s ===\n失败: %v\n\n", s.label, err)
			continue
		}
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(s.timeout):
			cmd.Process.Kill()
			<-done
		}
		text := strings.TrimSpace(out.String())
		if text == "" {
			text = "(无输出)"
		}
		fmt.Fprintf(&sb, "=== %s ===\n%s\n\n", s.label, text)
	}
	return sb.String()
}

func reportSections(guiLogFn func() string) [][2]string {
	secs := [][2]string{
		{"基本信息", fmt.Sprintf("主机: %s\n用户: %s\n工具版本: %s\n时间: %s",
			hostname(), currentUser(), TOOL_VERSION, time.Now().Format("2006-01-02 15:04:05"))},
		{"Docker 快照", dockerSnapshot()},
	}
	gui := "(不可用)"
	if guiLogFn != nil {
		if g := guiLogFn(); g != "" {
			gui = g
		}
	}
	lines := strings.Split(gui, "\n")
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	secs = append(secs, [2]string{"GUI 日志尾 (最近 200 行)", strings.Join(lines, "\n")})
	return secs
}

func currentUser() string {
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	return "?"
}

// transcriptFiles logs 目录下的 transcript 列表 (新->旧).
func transcriptFiles() []string {
	dir := winbootLogsDir()
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		si, _ := os.Stat(out[i])
		sj, _ := os.Stat(out[j])
		return si.ModTime().After(sj.ModTime())
	})
	return out
}

func buildLogsHTML(guiLogFn func() string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<html><head><meta charset=\"utf-8\"><title>winboot 排查报告 - %s</title></head><body>",
		html.EscapeString(hostname()))
	sb.WriteString("<h2>winboot 排查报告</h2><p>完整打包下载: <a href=\"/logs.zip\">/logs.zip</a></p>")
	for _, sec := range reportSections(guiLogFn) {
		fmt.Fprintf(&sb, "<h3>%s</h3><pre>%s</pre>",
			html.EscapeString(sec[0]), html.EscapeString(sec[1]))
	}
	trans := transcriptFiles()
	fmt.Fprintf(&sb, "<h3>运行留档 (transcripts, 共 %d 份, 展示最近 3 份尾部)</h3>", len(trans))
	for _, f := range trans[:min(3, len(trans))] {
		content := "(读取失败)"
		if b, err := os.ReadFile(f); err == nil {
			c := string(b)
			if len(c) > 3000 {
				c = c[len(c)-3000:]
			}
			content = c
		}
		fmt.Fprintf(&sb, "<h4>%s</h4><pre>%s</pre>",
			html.EscapeString(filepath.Base(f)), html.EscapeString(content))
	}
	for _, f := range trans {
		st, _ := os.Stat(f)
		fmt.Fprintf(&sb, "<p>%s (%d KB)</p>", html.EscapeString(filepath.Base(f)), st.Size()/1024)
	}
	sb.WriteString("</body></html>")
	return sb.String()
}

func buildLogsZip(guiLogFn func() string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	var rpt strings.Builder
	for _, sec := range reportSections(guiLogFn) {
		fmt.Fprintf(&rpt, "=== %s ===\n%s\n\n", sec[0], sec[1])
	}
	w, _ := zw.Create("report.txt")
	w.Write([]byte(rpt.String()))
	for _, f := range transcriptFiles() {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		if len(b) > 512<<10 {
			b = b[:512<<10]
		}
		w, _ := zw.Create("transcripts/" + filepath.Base(f))
		w.Write(b)
	}
	zw.Close()
	return buf.Bytes()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// 防火墙放行: provider 的入站需要程序级 allow 规则 (Profile Any ——
// 公司网常被 Windows 归类为公用, 限定专用/域的规则实测不生效).
// ---------------------------------------------------------------------------

const fwRuleName = "winboot 局域网共享"

// fwRuleExists 按程序路径查规则 (标准用户即可查询).
// PS 5.1 大坑 (实测): Get-NetFirewallRule -Enabled True 会静默过滤掉所有规则,
// 必须走程序过滤且不用 -Enabled 参数.
func fwRuleExists(exe string) bool {
	q := "if (Get-NetFirewallRule -ErrorAction SilentlyContinue | Where-Object { " +
		"(Get-NetFirewallApplicationFilter -AssociatedNetFirewallRule $_).Program -eq '" + exe + "' }) " +
		"{ exit 0 } else { exit 1 }"
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", q)
	cmd.SysProcAttr = hideWindow()
	return cmd.Run() == nil
}

// utf16leBytes PowerShell -EncodedCommand 要求 base64(UTF-16LE).
func utf16leBytes(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	out := make([]byte, len(u16)*2)
	for i, v := range u16 {
		out[i*2] = byte(v)
		out[i*2+1] = byte(v >> 8)
	}
	return out
}

// ensureFirewallRule 缺规则则弹一次 UAC 添加; 异常处理: UAC 取消 -> 日志说明
// 可重试, 组策略/EDR 拦截 -> 手动路径指引.
func ensureFirewallRule(log func(string)) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if fwRuleExists(exe) {
		return
	}
	inner := "$ErrorActionPreference='Stop'; New-NetFirewallRule -DisplayName '" + fwRuleName +
		"' -Direction Inbound -Action Allow -Program '" + exe + "' -Profile Any | Out-Null"
	encoded := base64.StdEncoding.EncodeToString(utf16leBytes(inner))
	outer := "$p = Start-Process powershell -Verb RunAs -Wait -PassThru -ArgumentList " +
		"'-NoProfile','-EncodedCommand','" + encoded + "'; exit $p.ExitCode"
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", outer)
	cmd.SysProcAttr = hideWindow()
	out, err := cmd.CombinedOutput()
	if err == nil && fwRuleExists(exe) {
		log("防火墙已放行 (winboot 入站, 全部 profile): 局域网共享/排查/镜像就绪")
		return
	}
	text := string(out)
	if strings.Contains(text, "取消") {
		log("未放行防火墙 (UAC 被取消): 本机入站不可用 —— 别人拉不了你的日志/缓存; 重开程序会再给一次机会")
	} else {
		log("防火墙规则添加失败 (可能被组策略/EDR 拦截): 手动放行 = Windows 安全中心 > 防火墙和网络保护 > 允许应用通过防火墙 > 勾选 winboot")
	}
}
