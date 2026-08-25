// go-boids — Go 端 WebRTC 鸟群节点，与 webrtc-boids.html 互联（协议完全对齐）
// 信令：公共 MQTT（EMQX）· 数据：WebRTC DataChannel · 仿真：实体归属权威（只仿真自己生成的鸟）
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pion/webrtc/v4"
)

// ================= 参数（与 boids-core.js 对齐） =================
const (
	worldW, worldH   = 800.0, 600.0
	maxSpeed         = 165.0
	minSpeed         = 55.0
	maxForce         = 430.0
	perception       = 72.0
	separationDist   = 26.0
	wSep, wAli, wCoh = 1.55, 1.0, 0.9
)

// ================= 实体与仿真 =================
type Entity struct {
	ID   string  `json:"id"`
	Kind string  `json:"k"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	VX   float64 `json:"vx"`
	VY   float64 `json:"vy"`
}

var (
	mu       sync.Mutex
	localEnt = map[string]*Entity{}
	remoteN  = map[string]int{} // peerId -> 最近一次 state 里的实体数
)

func spawnBoids(n int, cx, cy float64) {
	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < n; i++ {
		a := rand.Float64() * math.Pi * 2
		sp := minSpeed + rand.Float64()*(maxSpeed-minSpeed)*0.6
		id := fmt.Sprintf("%s-%d", myId, i+1)
		localEnt[id] = &Entity{
			ID: id, Kind: "boid",
			X: cx + (rand.Float64()-0.5)*40, Y: cy + (rand.Float64()-0.5)*40,
			VX: math.Cos(a) * sp, VY: math.Sin(a) * sp,
		}
	}
}

func clamp(x, max float64) float64 {
	if x > max {
		return max
	}
	return x
}

// 极简 Reynolds：分离/对齐/凝聚（邻居仅本地实体；远端实体不参与 Go 端邻居计算——演示足够）
func stepSim(dt float64) {
	mu.Lock()
	defer mu.Unlock()
	sepR2 := separationDist * separationDist
	perR2 := perception * perception
	for _, e := range localEnt {
		var sepX, sepY, aliX, aliY, cohX, cohY float64
		var sepN, aliN, cohN float64
		for _, n := range localEnt {
			if n == e {
				continue
			}
			dx, dy := wrapDelta(e.X-n.X, worldW), wrapDelta(e.Y-n.Y, worldH)
			d2 := dx*dx + dy*dy
			if d2 < sepR2 {
				d := math.Sqrt(d2)
				if d == 0 {
					d = 1
				}
				sepX += dx / d / d
				sepY += dy / d / d
				sepN++
			}
			if d2 < perR2 {
				aliX += n.VX
				aliY += n.VY
				aliN++
				cohX += n.X
				cohY += n.Y
				cohN++
			}
		}
		var ax, ay float64
		if sepN > 0 {
			ax += sepX * wSep * 60
			ay += sepY * wSep * 60
		}
		if aliN > 0 {
			ax += (aliX/aliN - e.VX) * wAli
			ay += (aliY/aliN - e.VY) * wAli
		}
		if cohN > 0 {
			ax += (cohX/cohN - e.X) * wCoh
			ay += (cohY/cohN - e.Y) * wCoh
		}
		m := math.Hypot(ax, ay)
		if m > maxForce {
			ax, ay = ax*maxForce/m, ay*maxForce/m
		}
		e.VX += ax * dt
		e.VY += ay * dt
		sp := math.Hypot(e.VX, e.VY)
		if sp > maxSpeed {
			e.VX *= maxSpeed / sp
			e.VY *= maxSpeed / sp
		} else if sp < minSpeed && sp > 0 {
			e.VX *= minSpeed / sp
			e.VY *= minSpeed / sp
		}
		e.X += e.VX * dt
		e.Y += e.VY * dt
		e.X = wrapNorm(e.X, worldW)
		e.Y = wrapNorm(e.Y, worldH)
	}
}

func wrapDelta(d, size float64) float64 {
	if d > size/2 {
		d -= size
	} else if d < -size/2 {
		d += size
	}
	return d
}
func wrapNorm(v, size float64) float64 {
	v = math.Mod(v, size)
	if v < 0 {
		v += size
	}
	return v
}

// 快照：坐标归一化到 [0,w)×[0,h)，保留 1 位小数（与 boids-core.js snapshot 一致）
func snapshot() []Entity {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Entity, 0, len(localEnt))
	for _, e := range localEnt {
		out = append(out, Entity{
			ID: e.ID, Kind: e.Kind,
			X:  math.Round(wrapNorm(e.X, worldW)*10) / 10,
			Y:  math.Round(wrapNorm(e.Y, worldH)*10) / 10,
			VX: math.Round(e.VX*10) / 10, VY: math.Round(e.VY*10) / 10,
		})
	}
	return out
}

// ================= 信令协议（与 webrtc-boids.html 对齐） =================
var (
	myId     = "go-" + randString(6)
	presTop  string
	sigTop   string
	mqttCli  mqtt.Client
	roomName string
)

func randString(n int) string {
	const cs = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = cs[rand.Intn(len(cs))]
	}
	return string(b)
}

func shouldInitiate(selfId, peerId string) bool { return selfId < peerId }

type presMsg struct {
	T  string `json:"t"`
	ID string `json:"id"`
}

type sdpJSON struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type sigMsg struct {
	From      string                   `json:"from"`
	Type      string                   `json:"type"`
	SDP       *sdpJSON                 `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit `json:"candidate,omitempty"`
}

// ================= WebRTC peer 管理 =================
type peerState struct {
	id          string
	pc          *webrtc.PeerConnection
	dc          *webrtc.DataChannel
	open        bool
	initiator   bool
	lastAttempt time.Time
}

var (
	peers   = map[string]*peerState{}
	peersMu sync.Mutex
)

func rtcConfig() webrtc.Configuration {
	return webrtc.Configuration{ICEServers: []webrtc.ICEServer{
		{URLs: []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"}},
		{URLs: []string{"stun:stun.miwifi.com:3478", "stun:stun.qq.com:3478"}},
	}}
}

func newPC(id string) (*webrtc.PeerConnection, error) {
	pc, err := webrtc.NewPeerConnection(rtcConfig())
	if err != nil {
		return nil, err
	}
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		j := c.ToJSON()
		publishSig(id, sigMsg{From: myId, Type: "candidate", Candidate: &j})
	})
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("🔄 PC[%s] %s", id, s)
		if s == webrtc.PeerConnectionStateFailed || s == webrtc.PeerConnectionStateClosed {
			teardown(id, "ice-"+s.String())
		}
	})
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Printf("📡 DataChannel [%s 接收]", id)
		setupDC(id, dc)
	})
	return pc, nil
}

func setupDC(id string, dc *webrtc.DataChannel) {
	peersMu.Lock()
	p := peers[id]
	peersMu.Unlock()
	if p == nil {
		return
	}
	dc.OnOpen(func() {
		peersMu.Lock()
		if pp := peers[id]; pp != nil {
			pp.open = true
		}
		peersMu.Unlock()
		log.Printf("✅✅✅ DataChannel 打开 → %s", id)
		sendData(id, map[string]any{"t": "hello", "role": *roleFlag})
	})
	dc.OnClose(func() {
		log.Printf("⚠ DataChannel 关闭 → %s", id)
		teardown(id, "dc-closed")
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		handleData(id, msg.Data)
	})
	peersMu.Lock()
	if pp := peers[id]; pp != nil {
		pp.dc = dc
	}
	peersMu.Unlock()
}

func teardown(id, reason string) {
	peersMu.Lock()
	p := peers[id]
	if p != nil {
		if p.dc != nil {
			_ = p.dc.Close()
		}
		if p.pc != nil {
			_ = p.pc.Close()
		}
		delete(peers, id)
	}
	peersMu.Unlock()
	mu.Lock()
	delete(remoteN, id)
	mu.Unlock()
	if p != nil {
		log.Printf("👋 拆除连接 %s (%s)", id, reason)
	}
}

// 发起方：建 PC + DataChannel + offer（与页面 initiateConn 对齐）
func initiate(id string, retry bool) {
	peersMu.Lock()
	if old := peers[id]; old != nil {
		peersMu.Unlock()
		teardown(id, "reinit")
		peersMu.Lock()
	}
	pc, err := newPC(id)
	if err != nil {
		peersMu.Unlock()
		log.Printf("❌ 建 PC 失败: %v", err)
		return
	}
	p := &peerState{id: id, pc: pc, initiator: true, lastAttempt: time.Now()}
	peers[id] = p
	peersMu.Unlock()

	dc, err := pc.CreateDataChannel("boids", nil) // nil = 默认 ordered/reliable，与页面一致
	if err != nil {
		log.Printf("❌ 建 DC 失败: %v", err)
		return
	}
	setupDC(id, dc)

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Printf("❌ CreateOffer: %v", err)
		return
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		log.Printf("❌ SetLocalDescription: %v", err)
		return
	}
	ld := pc.LocalDescription()
	publishSig(id, sigMsg{From: myId, Type: "offer", SDP: &sdpJSON{Type: "offer", SDP: ld.SDP}})
	if retry {
		log.Printf("📤 发送 Offer → %s（重试）", id)
	} else {
		log.Printf("📤 发送 Offer → %s", id)
	}
}

// 收到信令：offer 一律清场重答（与页面 onSigMessage 对齐）
func onSignal(m sigMsg) {
	from := m.From
	if from == "" || from == myId {
		return
	}
	switch m.Type {
	case "offer":
		if m.SDP == nil {
			return
		}
		log.Printf("📩 收到 Offer ← %s", from)
		teardown(from, "offer-reset")
		pc, err := newPC(from)
		if err != nil {
			return
		}
		peersMu.Lock()
		peers[from] = &peerState{id: from, pc: pc, initiator: false, lastAttempt: time.Now()}
		peersMu.Unlock()
		if err := pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: m.SDP.SDP}); err != nil {
			log.Printf("❌ SetRemoteDescription: %v", err)
			return
		}
		ans, err := pc.CreateAnswer(nil)
		if err != nil {
			log.Printf("❌ CreateAnswer: %v", err)
			return
		}
		if err := pc.SetLocalDescription(ans); err != nil {
			log.Printf("❌ SetLocalDescription(answer): %v", err)
			return
		}
		ld := pc.LocalDescription()
		publishSig(from, sigMsg{From: myId, Type: "answer", SDP: &sdpJSON{Type: "answer", SDP: ld.SDP}})
		log.Printf("📤 发送 Answer → %s", from)
	case "answer":
		if m.SDP == nil {
			return
		}
		log.Printf("📩 收到 Answer ← %s", from)
		peersMu.Lock()
		p := peers[from]
		peersMu.Unlock()
		if p != nil && p.pc != nil {
			if err := p.pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: m.SDP.SDP}); err != nil {
				log.Printf("⚠ answer 设置失败 %s: %v", from, err)
			}
		}
	case "candidate":
		peersMu.Lock()
		p := peers[from]
		peersMu.Unlock()
		if p != nil && p.pc != nil && m.Candidate != nil {
			if err := p.pc.AddICECandidate(*m.Candidate); err != nil {
				// 过期 candidate 忽略
			}
		}
	}
}

// ================= 数据通道消息（与页面 handleData 对齐） =================
func sendData(id string, obj any) {
	peersMu.Lock()
	p := peers[id]
	peersMu.Unlock()
	if p == nil || !p.open || p.dc == nil {
		return
	}
	b, _ := json.Marshal(obj)
	_ = p.dc.SendText(string(b))
}

func broadcastData(obj any) {
	b, _ := json.Marshal(obj)
	peersMu.Lock()
	list := make([]*peerState, 0, len(peers))
	for _, p := range peers {
		if p.open && p.dc != nil {
			list = append(list, p)
		}
	}
	peersMu.Unlock()
	for _, p := range list {
		_ = p.dc.SendText(string(b))
	}
}

func handleData(peerId string, raw []byte) {
	var m struct {
		T    string   `json:"t"`
		Role string   `json:"role"`
		E    []Entity `json:"e"`
		TS   float64  `json:"ts"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	switch m.T {
	case "hello":
		log.Printf("👋 %s 角色=%s", peerId, m.Role)
	case "state":
		mu.Lock()
		remoteN[peerId] = len(m.E)
		mu.Unlock()
	case "clear":
		mu.Lock()
		remoteN[peerId] = 0
		mu.Unlock()
		log.Printf("🗑 %s 清空了其实体", peerId)
	case "ping":
		sendData(peerId, map[string]any{"t": "pong", "ts": m.TS})
	case "exec":
		log.Printf("⚠ 收到远程执行请求 ← %s（Go 端不支持，已忽略）", peerId)
	}
}

// ================= MQTT 信令 =================
func publishPres(t string) {
	b, _ := json.Marshal(presMsg{T: t, ID: myId})
	mqttCli.Publish(presTop, 0, false, b)
}

func publishSig(to string, m sigMsg) {
	b, _ := json.Marshal(m)
	mqttCli.Publish(sigTop+to, 0, false, b)
}

func onPresence(id string) {
	if id == myId {
		return
	}
	peersMu.Lock()
	p := peers[id]
	peersMu.Unlock()
	if p != nil && (p.open || time.Since(p.lastAttempt) < 8*time.Second) {
		return // 已连接或协商中
	}
	if shouldInitiate(myId, id) {
		log.Printf("📞 发起连接 → %s", id)
		initiate(id, p != nil)
	}
	// 被动方：等 offer
}

func setupMQTT(broker string) error {
	presTop = "webrtc-boids-v1/" + roomName + "/pres"
	sigTop = "webrtc-boids-v1/" + roomName + "/sig/"

	bye, _ := json.Marshal(presMsg{T: "bye", ID: myId})
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("boids-" + myId + "-" + randString(4)).
		SetCleanSession(true).
		SetKeepAlive(30 * time.Second).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetWill(presTop, string(bye), 0, false)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("✅ MQTT 已连接 %s", broker)
		tok := c.SubscribeMultiple(map[string]byte{presTop: 0, sigTop + myId: 0}, func(_ mqtt.Client, msg mqtt.Message) {
			onMqttMessage(msg)
		})
		tok.Wait()
		if tok.Error() != nil {
			log.Printf("❌ 订阅失败: %v", tok.Error())
			return
		}
		publishPres("hi")
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("⚠ MQTT 断开: %v（自动重连中）", err)
	})

	mqttCli = mqtt.NewClient(opts)
	tok := mqttCli.Connect()
	tok.Wait()
	return tok.Error()
}

func onMqttMessage(msg mqtt.Message) {
	topic := msg.Topic()
	if topic == presTop {
		var pm presMsg
		if err := json.Unmarshal(msg.Payload(), &pm); err != nil {
			return
		}
		if pm.ID == myId {
			return
		}
		switch pm.T {
		case "hi", "hello":
			if pm.T == "hi" {
				publishPres("hello") // 回礼让新人发现自己
			}
			onPresence(pm.ID)
		case "bye":
			teardown(pm.ID, "bye")
		}
	} else if topic == sigTop+myId {
		var sm sigMsg
		if err := json.Unmarshal(msg.Payload(), &sm); err != nil {
			return
		}
		onSignal(sm)
	}
}

// ================= main =================
var (
	roleFlag   = flag.String("role", "flock", "角色: flock / predator")
	roomFlag   = flag.String("room", "boids-room-1", "房间名（需与 HTML 页面一致）")
	brokerFlag = flag.String("broker", "tcp://broker.emqx.io:1883", "MQTT broker 地址")
	nFlag      = flag.Int("n", 10, "启动生成的鸟数量")
	cxFlag     = flag.Float64("cx", 400, "生成中心 X")
	cyFlag     = flag.Float64("cy", 300, "生成中心 Y")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	roomName = *roomFlag
	log.SetFlags(log.Flags() &^ log.Ldate)
	log.Printf("📋 go-boids 启动 · myId=%s role=%s room=%s", myId, *roleFlag, roomName)

	if err := setupMQTT(*brokerFlag); err != nil {
		log.Fatalf("❌ MQTT 连接失败: %v", err)
	}

	spawnBoids(*nFlag, *cxFlag, *cyFlag)
	log.Printf("🐦 本地生成 %d 只鸟 @(%.0f,%.0f)", *nFlag, *cxFlag, *cyFlag)

	// 仿真循环 ~60fps
	go func() {
		tk := time.NewTicker(16 * time.Millisecond)
		defer tk.Stop()
		last := time.Now()
		for range tk.C {
			dt := time.Since(last).Seconds()
			last = time.Now()
			if dt > 0.05 {
				dt = 0.05
			}
			stepSim(dt)
		}
	}()

	// 状态广播 ~15Hz（与页面对齐）
	go func() {
		tk := time.NewTicker(66 * time.Millisecond)
		defer tk.Stop()
		for range tk.C {
			snap := snapshot()
			if len(snap) == 0 {
				continue
			}
			broadcastData(map[string]any{"t": "state", "e": snap})
		}
	}()

	// mesh 重试 tick 1s + presence 心跳 5s（与页面对齐）
	go func() {
		tk := time.NewTicker(1 * time.Second)
		defer tk.Stop()
		var lastAnnounce time.Time
		for now := range tk.C {
			peersMu.Lock()
			for id, p := range peers {
				if !p.open && p.initiator && now.Sub(p.lastAttempt) > 8*time.Second {
					log.Printf("🔁 重试连接 → %s", id)
					go initiate(id, true)
				}
			}
			peersMu.Unlock()
			if now.Sub(lastAnnounce) >= 5*time.Second {
				lastAnnounce = now
				publishPres("hi")
			}
		}
	}()

	// 远端实体统计日志 2s
	go func() {
		tk := time.NewTicker(2 * time.Second)
		defer tk.Stop()
		for range tk.C {
			peersMu.Lock()
			var open, total int
			ids := make([]string, 0)
			for id, p := range peers {
				total++
				if p.open {
					open++
					ids = append(ids, id)
				}
			}
			peersMu.Unlock()
			mu.Lock()
			remoteSum := 0
			parts := make([]string, 0)
			for id, n := range remoteN {
				remoteSum += n
				parts = append(parts, fmt.Sprintf("%s:%d", id, n))
			}
			localN := len(localEnt)
			mu.Unlock()
			log.Printf("📊 本地鸟=%d · 远端实体=%d [%s] · 连接 %d/%d {%s}",
				localN, remoteSum, strings.Join(parts, ","), open, total, strings.Join(ids, ","))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Printf("👋 退出，发送 bye")
	publishPres("bye")
	time.Sleep(300 * time.Millisecond)
	mqttCli.Disconnect(500)
}
