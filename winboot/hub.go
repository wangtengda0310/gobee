package main

// hub.go: 汇聚层 —— 日志环形缓冲(供 /logs 报告) + 局域网服务编排 + 自动更新调度.
// 所有后台 goroutine 的日志都经 emitLog 汇入: 既推前端事件, 也进环形缓冲;
// 前端零改动即可显示全部新信息 (LAN 状态/告警/更新进度都走 boot:log 事件).

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const logRingMax = 500

var (
	logRing   []string
	logRingMu sync.Mutex
)

func pushLog(line string) {
	logRingMu.Lock()
	if len(logRing) >= logRingMax {
		logRing = logRing[1:]
	}
	logRing = append(logRing, line)
	logRingMu.Unlock()
}

// emitLog 后台统一日志出口 (前端事件 + 环形缓冲).
func emitLog(line string) {
	pushLog(line)
	if wailsApp != nil {
		wailsApp.Event.Emit("boot:log", line)
	}
}

// guiLogSnapshot /logs 报告的数据源 (环形缓冲全文).
func guiLogSnapshot() string {
	logRingMu.Lock()
	defer logRingMu.Unlock()
	return strings.Join(logRing, "\n")
}

// ---------------------------------------------------------------------------

type Hub struct {
	provider  *LanProvider
	discovery *LanDiscovery

	mu           sync.Mutex
	providers    []providerInfo // 最新一轮发现结果 (install 注入 -LanBase 用)
	lanSourceURL string         // 变更检测: 避免每 5s 刷一行日志
	prefetchAttempted map[string]bool
	prefetchBusy      bool
	anomalyLast  map[string]time.Time // 异常广播冷却
}

var hub = &Hub{
	prefetchAttempted: map[string]bool{},
	anomalyLast:       map[string]time.Time{},
}

func (h *Hub) Start() {
	h.startProvider()
	h.discovery = NewLanDiscovery(h.onProviders)
	h.discovery.Start()
	go h.updateWorker()
}

func (h *Hub) Stop() {
	if h.discovery != nil {
		h.discovery.Stop()
	}
	if h.provider != nil {
		h.provider.Stop()
	}
}

func (h *Hub) startProvider() {
	p := NewLanProvider(h.onAnomaly, guiLogSnapshot, ConfiguredBase())
	if err := p.Start(); err != nil {
		emitLog("未开启局域网共享: " + err.Error())
		return
	}
	h.provider = p
	m := buildManifest(p.httpPort, p.updateBase)
	names := make([]string, 0, len(m.Files))
	for _, f := range m.Files {
		names = append(names, fmt.Sprintf("%s (%dMB)", f.Name, f.Size>>20))
	}
	files := strings.Join(names, ", ")
	if files == "" {
		files = "无缓存"
	}
	emitLog("已开启局域网共享: " + p.BaseURL() + " [" + files + "]; 排查入口: " + p.BaseURL() + "/logs (浏览器直开)")
}

// onProviders 发现结果 (discovery goroutine 回调): 状态只记变更 + 调度预取.
func (h *Hub) onProviders(providers []providerInfo) {
	h.mu.Lock()
	h.providers = providers
	h.mu.Unlock()
	if len(providers) == 0 {
		h.mu.Lock()
		changed := h.lanSourceURL != ""
		h.lanSourceURL = ""
		h.mu.Unlock()
		if changed {
			emitLog("--- 局域网源已失联 (安装时正常下载) ---")
		}
		return
	}
	p := PickInstallSource(providers)
	h.mu.Lock()
	changed := p != nil && p.URL != h.lanSourceURL
	if changed {
		h.lanSourceURL = p.URL
	}
	h.mu.Unlock()
	if changed {
		var total int64
		for _, f := range p.Files {
			total += f.Size
		}
		emitLog(fmt.Sprintf("--- 局域网源: %s (可提供 %d 个文件, 共 %dMB) ---",
			p.URL, len(p.Files), total>>20))
	}
	h.maybePrefetch(p)
}

// installLanBase install 命令的 -LanBase 注入源 (无源返回 "").
func (h *Hub) installLanBase() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.providers) == 0 {
		return ""
	}
	if p := PickInstallSource(h.providers); p != nil {
		return p.URL
	}
	return ""
}

// onAnomaly 收到他人异常广播 (provider 的 UDP goroutine 回调).
func (h *Hub) onAnomaly(a anomalyMsg) {
	tail := ""
	if a.Tail != "" {
		tail = "  " + a.Tail
	}
	emitLog(fmt.Sprintf("⚠ [组内告警] %s: %s 失败 (exit %d)%s", a.Host, a.Command, a.Code, tail))
}

// maybeBroadcastAnomaly 命令失败时广播摘要; 同一命令 5 分钟冷却 (防失败循环刷屏).
func (h *Hub) maybeBroadcastAnomaly(command string, code int, output string) {
	h.mu.Lock()
	if time.Since(h.anomalyLast[command]) < 5*time.Minute {
		h.mu.Unlock()
		return
	}
	h.anomalyLast[command] = time.Now()
	h.mu.Unlock()
	lines := strings.Split(output, "\n")
	if len(lines) > 5 {
		lines = lines[len(lines)-5:]
	}
	go broadcastAnomaly(command, code, strings.Join(lines, "\n"))
}

// maybePrefetch 本机缺的缓存从发现的源后台预取 (每文件一次机会, 失败下次启动再试).
func (h *Hub) maybePrefetch(source *providerInfo) {
	h.mu.Lock()
	if h.prefetchBusy {
		h.mu.Unlock()
		return
	}
	var missing []string
	for _, s := range shareable {
		if h.prefetchAttempted[s.Name] {
			continue
		}
		if st, err := os.Stat(lanSharePath(s.Name)); err == nil && st.Size() >= s.MinBytes {
			continue // 本机已有
		}
		for _, f := range source.Files {
			if f.Name == s.Name {
				missing = append(missing, s.Name)
				break
			}
		}
	}
	if len(missing) == 0 {
		h.mu.Unlock()
		return
	}
	for _, m := range missing {
		h.prefetchAttempted[m] = true
	}
	h.prefetchBusy = true
	h.mu.Unlock()
	go func() {
		defer func() {
			h.mu.Lock()
			h.prefetchBusy = false
			h.mu.Unlock()
		}()
		prefetchMissing(source, emitLog)
	}()
}

// updateWorker 后台检查+下载新版本; 热替换后重启.
// 源策略: 中心清单 (UPDATE_BASE) 是权威; 局域网同类作镜像 (命中同 version+sha 才用);
// 中心失联时局域网自治 (同类中版本最新者, 信任=局域网).
func (h *Hub) updateWorker() {
	time.Sleep(8 * time.Second) // 等第一轮发现
	central := CheckUpdate("")
	centralBase := ConfiguredBase()
	if central == nil {
		// 本机配置缺失/失效 (如 DHCP 换网段): 用发现协议学到的中心地址再试一次.
		// 信任级别 = 局域网 (与失联自治同级); 代码内配置仍是首选锚点.
		if learned := DiscoveredUpdateBase(h.discovery.Snapshot()); learned != "" && learned != centralBase {
			central = CheckUpdate(learned)
			if central != nil {
				centralBase = learned
				emitLog("(更新中心地址经局域网发现: " + learned + ")")
			}
		}
	}
	var target *updateInfo
	if central != nil {
		target = central
	} else {
		peer := NewestToolPeer(h.discovery.Snapshot())
		if peer == nil || peer.Tool == nil {
			return // 中心失联且局域网无同类: 静默
		}
		target = &updateInfo{Version: peer.Tool.Version, SHA256: peer.Tool.SHA256, Size: peer.Tool.Size}
		emitLog("更新中心不可达, 采用局域网源 v" + target.Version + " (" + peer.URL + ")")
	}
	if !VersionGt(target.Version, TOOL_VERSION) {
		return // 不比当前新 (含不回滚)
	}
	var sources []*updateInfo
	if m := FindToolMirror(h.discovery.Snapshot(), target.Version, target.SHA256); m != nil {
		sources = append(sources, &updateInfo{Version: target.Version,
			URL: m.URL + "/" + lanToolFile, SHA256: target.SHA256, Size: target.Size})
	}
	if central != nil {
		c := *central
		// 相对 url 在此绝对化: 中心地址可能来自"发现学到的 base",
		// 而 DownloadUpdate 的缺省 base 是本机配置 (两者可能不同)
		if c.URL != "" && !strings.Contains(c.URL, "://") {
			c.URL = centralBase + "/" + c.URL
		}
		sources = append(sources, &c)
	}
	if len(sources) == 0 {
		return
	}
	emitLog(fmt.Sprintf("发现新版本 v%s (当前 v%s), 后台下载...", target.Version, TOOL_VERSION))
	var path string
	for _, src := range sources {
		p, err := DownloadUpdate(src, emitLog)
		if err == nil {
			path = p
			break
		}
		emitLog("下载源失败 (" + src.URL + "): " + err.Error())
	}
	if path == "" {
		emitLog("新版本下载失败 (所有源), 继续使用当前版本")
		return
	}
	if ApplyUpdate(path) {
		emitLog("已热替换到 v" + target.Version + ", 正在重启...")
		RestartSelf()
		quitApp()
	} else {
		emitLog("新版本热替换失败, 下次启动重试")
	}
}
