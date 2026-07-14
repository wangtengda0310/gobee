package prototest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	streamproto "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	proto "github.com/gogo/protobuf/proto"
)

// pendingFrame 拦截待放行帧
type pendingFrame struct {
	seqID       uint32
	msgID       uint16
	header      []byte
	body        []byte
	payloadJSON json.RawMessage
}

// connInterceptState 单连接拦截状态
type connInterceptState struct {
	connID     uint64
	serverConn net.Conn
	pending    []*pendingFrame
	mu         sync.Mutex
	writeMu    sync.Mutex // 保护 serverConn 写入（防止并发写入交错）
	proxySeqId uint32     // 代理级别的 seqId 计数器（重写后发往服务端的 seqId）
}

// nextProxySeqID 返回下一个 proxySeqId 并递增
func (s *connInterceptState) nextProxySeqID() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxySeqId++
	return s.proxySeqId
}

// peekProxySeqID 返回当前 proxySeqId（不递增）
func (s *connInterceptState) peekProxySeqID() uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.proxySeqId
}

// RecordProgress 录制进度（返回给前端）
type RecordProgress struct {
	Status       string `json:"status"`        // "idle" | "running" | "stopped" | "error"
	MessageCount int    `json:"message_count"` // 已录制消息数
	ServerAddr   string `json:"server_addr"`   // 目标服务器地址
	ErrorMessage string `json:"error_message"` // 错误信息
}

// RecordWorker 异步录制工作器
type RecordWorker struct {
	mu       sync.Mutex
	emitter  EventEmitter
	cancel   context.CancelFunc
	progress RecordProgress
	running  bool
	// 代理相关
	tcpListener  net.Listener
	httpListener net.Listener
	recorder     *streamproto.Recorder
	// 地址映射：远程地址 -> 本地地址（供 HTTP 代理替换 server_addr）
	tcpRedirectMap sync.Map
	// 拦截模式：true 时客户端 Req 消息（MsgID >= 1000）将被拦截并推送到前端
	filterMode bool
	// 每连接拦截状态（filterMode 时使用）
	connStates sync.Map // uint64 -> *connInterceptState
	// 连接池引用（停止录制时移交连接）
	connPool *streamproto.AccountConnectionPool
	// 每连接的账号映射（connID -> accountID，由 LoginReq 解析得到）
	connAccounts sync.Map // uint64 -> string
}

// NewRecordWorker 创建录制工作器
// emitter 可以是 Wails application.App.Event，也可以是任何实现了 EventEmitter 接口的对象。
func NewRecordWorker(emitter EventEmitter) *RecordWorker {
	return &RecordWorker{
		emitter: emitter,
		progress: RecordProgress{
			Status: "idle",
		},
	}
}

// IsRunning 是否正在录制
func (w *RecordWorker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// SetConnPool 设置连接池引用（停止录制时移交连接）
func (w *RecordWorker) SetConnPool(pool *streamproto.AccountConnectionPool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.connPool = pool
}

// GetProgress 获取当前进度
func (w *RecordWorker) GetProgress() RecordProgress {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.progress
}

// GetRecording 导出当前内存中的录制数据（供"保存为用例"使用）
// 若无录制数据（未录制或已停止清空），返回 nil
func (w *RecordWorker) GetRecording() *streamproto.Recording {
	w.mu.Lock()
	recorder := w.recorder
	w.mu.Unlock()
	if recorder == nil {
		return nil
	}
	return recorder.ToRecording()
}

// StartListen 启动本地 TCP/HTTP 监听（不开始录制）
// serverAddr: 目标 TCP 服务器地址（如 "10.254.114.204:18000"）
// httpAddr: 目标 HTTP 地址（如 "10.254.114.204:20144"）
// tcpListenPort: 本地 TCP 监听端口
// httpListenPort: 本地 HTTP 监听端口
// 应用启动时调用，监听端口长期保持，直到调用 StopListen 或 Stop
func (w *RecordWorker) StartListen(serverAddr string, httpAddr string, tcpListenPort int, httpListenPort int) error {
	w.mu.Lock()
	// 如果已经在监听，先停止旧监听再重新启动（支持修改端口后重载）
	// 关闭旧监听器在锁内完成，避免释放锁期间发生竞态
	if w.tcpListener != nil || w.httpListener != nil {
		log.Printf("[RecordWorker] 监听已存在，先停止后重新监听")
		oldTCP := w.tcpListener
		oldHTTP := w.httpListener
		w.tcpListener = nil
		w.httpListener = nil
		w.progress = RecordProgress{Status: "idle"}
		if oldTCP != nil {
			_ = oldTCP.Close()
		}
		if oldHTTP != nil {
			_ = oldHTTP.Close()
		}
	}

	w.progress = RecordProgress{
		Status:     "listening",
		ServerAddr: serverAddr,
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	// 本地监听地址由配置参数决定
	localTCPAddr := fmt.Sprintf(":%d", tcpListenPort)
	localHTTPAddr := fmt.Sprintf(":%d", httpListenPort)

	// 注册 TCP 地址映射：客户端连接本地 tcpListenPort，实际转发到 serverAddr
	localAddr := fmt.Sprintf("127.0.0.1:%d", tcpListenPort)
	w.tcpRedirectMap.Store(serverAddr, localAddr)

	// 启动 TCP 代理
	tcpLn, err := net.Listen("tcp", localTCPAddr)
	if err != nil {
		w.finish("error", fmt.Sprintf("TCP监听失败: %v", err))
		return fmt.Errorf("TCP监听失败: %v", err)
	}
	w.tcpListener = tcpLn

	go w.tcpProxyLoop(ctx, tcpLn, serverAddr)

	// 启动 HTTP 代理
	httpLn, err := net.Listen("tcp", localHTTPAddr)
	if err != nil {
		_ = w.tcpListener.Close()
		w.finish("error", fmt.Sprintf("HTTP监听失败: %v", err))
		return fmt.Errorf("HTTP监听失败: %v", err)
	}
	w.httpListener = httpLn

	go w.httpProxyLoop(ctx, httpLn, httpAddr, serverAddr)

	w.emitProgress()
	log.Printf("[RecordWorker] 监听已启动: TCP=%s->%s, HTTP=%s->%s",
		localTCPAddr, serverAddr, localHTTPAddr, httpAddr)
	log.Printf("[RecordWorker] 客户端应连接: TCP=%s, HTTP=%s", localAddr, fmt.Sprintf("127.0.0.1:%d", httpListenPort))

	return nil
}

// StartRecord 开始录制（必须先调用 StartListen 成功）
// filterMode: 是否启用实时修改模式（强制客户端修改报文后再发送到服务器）
func (w *RecordWorker) StartRecord(filterMode bool) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("录制已经在进行中")
	}
	if w.tcpListener == nil || w.httpListener == nil {
		w.mu.Unlock()
		return fmt.Errorf("监听未启动，无法开始录制")
	}
	serverAddr := w.progress.ServerAddr
	w.running = true
	w.filterMode = filterMode
	w.progress.Status = "running"
	w.mu.Unlock()

	// 初始化录制器（设置回调：每录制一条消息就推送进度）
	w.recorder = streamproto.NewRecorder(serverAddr)
	w.recorder.SetOnRecord(func(count int, latestMsg *streamproto.RecordEntry) {
		w.mu.Lock()
		w.progress.MessageCount = count
		w.mu.Unlock()
		w.emitProgressWithLatestMsg(latestMsg)
	})
	w.recorder.Start()

	w.emitProgress()
	log.Printf("[RecordWorker] 录制已开始，filterMode=%v", filterMode)

	return nil
}

// Start 启动录制（兼容旧接口：同时监听并开始录制）
// 新代码建议分别调用 StartListen + StartRecord
// serverAddr: 目标 TCP 服务器地址（如 "10.254.114.204:18000"）
// httpAddr: 目标 HTTP 地址（如 "10.254.114.204:20144"）
// filterMode: 是否启用实时修改模式
func (w *RecordWorker) Start(serverAddr string, httpAddr string, filterMode bool) error {
	colonIdx := strings.LastIndex(serverAddr, ":")
	if colonIdx < 0 {
		return fmt.Errorf("服务器地址格式错误: %s", serverAddr)
	}
	tcpPort, err := strconv.Atoi(serverAddr[colonIdx+1:])
	if err != nil {
		return fmt.Errorf("服务器地址端口格式错误: %s", serverAddr)
	}
	if err := w.StartListen(serverAddr, httpAddr, tcpPort, 20144); err != nil {
		return err
	}
	return w.StartRecord(filterMode)
}

// IsFilterMode 是否处于实时拦截模式
func (w *RecordWorker) IsFilterMode() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.filterMode
}

// SetFilterMode 设置实时拦截模式（关闭后后续消息透传）
func (w *RecordWorker) SetFilterMode(enabled bool) {
	w.mu.Lock()
	w.filterMode = enabled
	w.mu.Unlock()
	log.Printf("[RecordWorker] filterMode=%v", enabled)
}

// StopRecord 停止录制（保留监听）
func (w *RecordWorker) StopRecord() {
	log.Printf("[RecordWorker] StopRecord() 被调用")

	// 停录前 best-effort 放行所有 pending（前端应已带 edits 调用过 ReleaseAllPending）
	if w.IsFilterMode() {
		if err := w.ReleaseAllPending(""); err != nil {
			log.Printf("[RecordWorker] 停录放行 pending 失败: %v", err)
		}
		w.SetFilterMode(false)
	}

	w.mu.Lock()
	w.running = false
	w.recorder = nil
	w.progress.Status = "listening"
	w.progress.MessageCount = 0
	w.mu.Unlock()
	w.emitProgress()
	log.Printf("[RecordWorker] 录制已停止，监听仍保持")
}

// StopListen 停止监听（关闭 TCP/HTTP 监听器）
func (w *RecordWorker) StopListen() {
	log.Printf("[RecordWorker] StopListen() 被调用")

	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	tcpLn := w.tcpListener
	httpLn := w.httpListener
	w.tcpListener = nil
	w.httpListener = nil
	w.tcpRedirectMap.Clear()
	w.progress = RecordProgress{Status: "idle"}
	w.mu.Unlock()

	if tcpLn != nil {
		if err := tcpLn.Close(); err != nil {
			log.Printf("[RecordWorker] TCP 监听器关闭失败: %v", err)
		}
	}
	if httpLn != nil {
		if err := httpLn.Close(); err != nil {
			log.Printf("[RecordWorker] HTTP 监听器关闭失败: %v", err)
		}
	}

	w.emitProgress()
	log.Printf("[RecordWorker] 监听已停止")
}

// Stop 停止录制和监听（完整停止）
func (w *RecordWorker) Stop() {
	log.Printf("[RecordWorker] Stop() 被调用")
	w.StopRecord()
	w.StopListen()
}

// emitProgress 推送进度到前端
func (w *RecordWorker) emitProgress() {
	w.emitProgressWithLatestMsg(nil)
}

// emitProgressWithLatestMsg 推送进度到前端（附带最新消息）
// 使用 RecordEntryView 合约类型，确保字段名与前端 bindings 一致
func (w *RecordWorker) emitProgressWithLatestMsg(latestMsg *streamproto.RecordEntry) {
	w.mu.Lock()
	p := w.progress
	count := p.MessageCount
	w.mu.Unlock()

	data := map[string]any{
		"status":        p.Status,
		"message_count": count,
		"server_addr":   p.ServerAddr,
		"error_message": p.ErrorMessage,
	}
	if latestMsg != nil {
		data["latest_msg"] = singleEntryToView(latestMsg, count-1)
	}
	log.Printf("[RecordWorker] emit event: status=%v msg_count=%v", data["status"], data["message_count"])
	w.emitter.Emit("record:progress", data)
}

// finish 清理运行状态（不清空监听器，停止录制时监听应保持）
func (w *RecordWorker) finish(status string, errMsg string) {
	w.mu.Lock()
	w.running = false
	w.progress.Status = status
	w.progress.ErrorMessage = errMsg
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	w.mu.Unlock()
	w.emitProgress()
}

// updateMessageCount 更新消息计数并推送
func (w *RecordWorker) updateMessageCount() {
	w.mu.Lock()
	if w.recorder != nil {
		w.progress.MessageCount = w.recorder.GetMessageCount()
	}
	w.mu.Unlock()
	w.emitProgress()
}

// tcpProxyLoop TCP 代理循环
func (w *RecordWorker) tcpProxyLoop(ctx context.Context, ln net.Listener, remoteAddr string) {
	log.Printf("[RecordWorker] TCP 代理监听 %s, 转发到 %s", ln.Addr(), remoteAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("[RecordWorker] TCP Accept 失败: %v", err)
				continue
			}
		}

		go w.handleTCPConn(ctx, conn, remoteAddr)
	}
}

// handleTCPConn 处理单个 TCP 连接
func (w *RecordWorker) handleTCPConn(ctx context.Context, clientConn net.Conn, remoteAddr string) {
	id := streamproto.ConnIDCounter.Add(1)
	startTime := time.Now()

	log.Printf("[RecordWorker] TCP #%d 新连接 %s -> %s", id, clientConn.RemoteAddr(), remoteAddr)

	remoteConn, err := net.DialTimeout("tcp", remoteAddr, 5*time.Second)
	if err != nil {
		log.Printf("[RecordWorker] TCP #%d 连接目标失败: %v", id, err)
		_ = clientConn.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// 提前创建连接状态（确保 writeMu 可用）
	connState := w.getOrCreateConnState(id, remoteConn)

	// 记录哪个方向先完成：1=C→S先完成（客户端断开，服务端连接可移交），2=S→C先完成
	var firstDone int32

	// 客户端 -> 服务端（录制方向）
	go func() {
		defer wg.Done()
		var n int64
		if w.filterMode {
			n = w.interceptAndParse(clientConn, remoteConn, id)
		} else {
			n = streamproto.RelayAndParse(clientConn, remoteConn, id, streamproto.DirClientToServer, true, w.recorder, w.onLoginReqParsed, &connState.writeMu, connState.nextProxySeqID)
		}
		log.Printf("[RecordWorker] TCP #%d 客户端->服务端 完成, %d 字节, 耗时 %s", id, n, time.Since(startTime).Truncate(time.Millisecond))
		atomic.StoreInt32(&firstDone, 1)
	}()

	// 服务端 -> 客户端（seqId 不重写，服务端回复原样转发）
	go func() {
		defer wg.Done()
		n := streamproto.RelayAndParse(remoteConn, clientConn, id, streamproto.DirServerToClient, false, w.recorder, nil, nil, nil)
		log.Printf("[RecordWorker] TCP #%d 服务端->客户端 完成, %d 字节, 耗时 %s", id, n, time.Since(startTime).Truncate(time.Millisecond))
		atomic.CompareAndSwapInt32(&firstDone, 0, 2) // 只有 C→S 还没结束时才设为 2
		_ = clientConn.Close()
	}()

	wg.Wait()

	w.removeConnState(id)

	// 尝试将服务端连接移交给连接池
	// 条件：C→S 先结束（客户端断开）且 S→C 后结束（服务端连接仍存活）
	// 如果 S→C 先结束，说明服务端断开了连接，无法移交
	handedOff := false
	if atomic.LoadInt32(&firstDone) == 1 {
		handedOff = w.tryHandoffConnection(id, remoteAddr, remoteConn)
	}

	if !handedOff {
		_ = remoteConn.Close()
	}

	// 推送进度更新（录制数据驻留内存，不再自动落盘）
	if w.recorder != nil {
		w.updateMessageCount()
	}

	log.Printf("[RecordWorker] TCP #%d 连接关闭", id)
}

// httpProxyLoop HTTP 代理循环
func (w *RecordWorker) httpProxyLoop(ctx context.Context, ln net.Listener, remoteAddr string, remoteIP string) {
	handler := http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		ts := time.Now().Format("15:04:05.000")
		qs := ""
		if req.URL.RawQuery != "" {
			qs = "?" + req.URL.RawQuery
		}
		log.Printf("[RecordWorker] HTTP [%s] -> %s %s%s", ts, req.Method, req.URL.Path, qs)

		// 构建到目标服务器的请求
		outURL := fmt.Sprintf("http://%s%s%s", remoteAddr, req.URL.Path, qs)
		if req.URL.RawQuery != "" {
			outURL = fmt.Sprintf("http://%s%s?%s", remoteAddr, req.URL.Path, req.URL.RawQuery)
		}

		outReq, err := http.NewRequest(req.Method, outURL, req.Body)
		if err != nil {
			log.Printf("[RecordWorker] HTTP [%s] <- 构建请求失败: %v", ts, err)
			http.Error(wr, "代理请求失败", http.StatusBadGateway)
			return
		}

		// 复制请求头
		for k, vs := range req.Header {
			for _, v := range vs {
				outReq.Header.Add(k, v)
			}
		}
		outReq.Host = remoteAddr

		resp, err := http.DefaultClient.Do(outReq)
		if err != nil {
			log.Printf("[RecordWorker] HTTP [%s] <- 请求目标失败: %v", ts, err)
			http.Error(wr, "目标不可达", http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		// 读取响应体到内存（需要完整读取才能做地址替换）
		bodyBytes, err := readAllWithTimeout(resp.Body, 5*time.Second)
		if err != nil {
			log.Printf("[RecordWorker] HTTP [%s] <- 读取响应体失败: %v", ts, err)
			http.Error(wr, "读取响应失败", http.StatusBadGateway)
			return
		}

		// 替换 server_addr / server_list 中指向远程服务器的地址为本地代理地址
		bodyStr := string(bodyBytes)
		replaced := false
		w.tcpRedirectMap.Range(func(key, value any) bool {
			remoteAddr := key.(string)
			localAddr := value.(string)
			colonIdx := strings.LastIndex(remoteAddr, ":")
			if colonIdx < 0 {
				return true
			}
			rip := remoteAddr[:colonIdx]
			// 替换所有该 IP 出现的地方（无论端口是什么）
			if strings.Contains(bodyStr, rip) {
				prefix := rip + ":"
				for {
					idx := strings.Index(bodyStr, prefix)
					if idx < 0 {
						break
					}
					portStart := idx + len(prefix)
					portEnd := portStart
					for portEnd < len(bodyStr) && bodyStr[portEnd] >= '0' && bodyStr[portEnd] <= '9' {
						portEnd++
					}
					oldAddr := bodyStr[idx:portEnd]
					bodyStr = bodyStr[:idx] + localAddr + bodyStr[portEnd:]
					replaced = true
					log.Printf("[RecordWorker] HTTP [MAP] 地址替换: %s -> %s", oldAddr, localAddr)
				}
			}
			return true
		})
		if replaced {
			bodyBytes = []byte(bodyStr)
		}

		// 复制响应头（Content-Length 由 Write 自动设置，跳过原始值）
		for k, vs := range resp.Header {
			if strings.EqualFold(k, "Content-Length") {
				continue
			}
			for _, v := range vs {
				wr.Header().Add(k, v)
			}
		}
		wr.WriteHeader(resp.StatusCode)

		written, _ := wr.Write(bodyBytes)

		ts2 := time.Now().Format("15:04:05.000")
		extra := ""
		if replaced {
			extra = " [已重定向]"
		}
		log.Printf("[RecordWorker] HTTP [%s] <- %d %d字节%s", ts2, resp.StatusCode, written, extra)
	})

	log.Printf("[RecordWorker] HTTP 代理监听 %s, 转发到 http://%s", ln.Addr(), remoteAddr)
	server := &http.Server{Handler: handler}
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()
	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Printf("[RecordWorker] HTTP 代理退出: %v", err)
	}
}

func (w *RecordWorker) getOrCreateConnState(connID uint64, serverConn net.Conn) *connInterceptState {
	if v, ok := w.connStates.Load(connID); ok {
		return v.(*connInterceptState)
	}
	state := &connInterceptState{
		connID:     connID,
		serverConn: serverConn,
		pending:    make([]*pendingFrame, 0),
	}
	actual, _ := w.connStates.LoadOrStore(connID, state)
	return actual.(*connInterceptState)
}

func (w *RecordWorker) removeConnState(connID uint64) {
	w.connStates.Delete(connID)
}

// onLoginReqParsed 录制模式下解析到 LoginReq 时的回调
// 从 payload 中提取账号名，记录到 connAccounts 映射中
func (w *RecordWorker) onLoginReqParsed(connID uint64, payload []byte) {
	accountID, err := streamproto.ExtractAccountFromLoginPayload(payload)
	if err != nil {
		log.Printf("[RecordWorker] TCP #%d 解析 LoginReq 账号失败: %v", connID, err)
		return
	}
	w.connAccounts.Store(connID, accountID)
	log.Printf("[RecordWorker] TCP #%d 检测到账号: %s", connID, accountID)
}

// tryHandoffConnection 尝试将代理的服务端连接移交给连接池
// 当客户端断开（C→S 先结束）时调用，此时 remoteConn 可能仍存活
// 返回 true 表示移交成功（调用方不要再 Close），false 表示移交失败（调用方负责 Close）
func (w *RecordWorker) tryHandoffConnection(connID uint64, serverAddr string, remoteConn net.Conn) bool {
	w.mu.Lock()
	pool := w.connPool
	w.mu.Unlock()

	if pool == nil {
		return false
	}

	accountIDVal, ok := w.connAccounts.Load(connID)
	if !ok {
		return false
	}
	accountID := accountIDVal.(string)

	// 移交给连接池
	log.Printf("[RecordWorker] TCP #%d 移交连接到连接池: account=%s", connID, accountID)
	pool.AcceptConn(accountID, serverAddr, remoteConn)
	w.connAccounts.Delete(connID)
	return true
}

func (w *RecordWorker) forwardFrame(serverConn net.Conn, connID uint64, dir string, header, body []byte) bool {
	state := w.getOrCreateConnState(connID, serverConn)

	// client→server 方向：重写 seqId 为代理级别的递增 seqId
	if dir == streamproto.DirClientToServer {
		flags := header[3]
		newSeqID := state.nextProxySeqID()
		if _, err := streamproto.RewriteSeqID(body, flags, newSeqID); err != nil {
			log.Printf("[TCP] #%d seqId 重写失败: %v", connID, err)
		}
	}

	state.writeMu.Lock()
	defer state.writeMu.Unlock()

	if _, werr := serverConn.Write(header); werr != nil {
		log.Printf("[TCP] #%d %s 写入帧头失败: %v", connID, dir, werr)
		return false
	}
	if _, werr := serverConn.Write(body); werr != nil {
		log.Printf("[TCP] #%d %s 写入消息体失败: %v", connID, dir, werr)
		return false
	}
	return true
}

func (w *RecordWorker) framePayloadJSON(frame *streamproto.DecodedFrame) json.RawMessage {
	entry := &streamproto.RecordEntry{
		MsgID:     frame.MsgID,
		MsgName:   streamproto.GetMsgName(frame.MsgID),
		SeqID:     frame.SeqID,
		Direction: streamproto.DirClientToServer,
	}
	msg, ok := streamproto.NewMessage(frame.MsgID)
	if ok {
		var protoData []byte
		if len(frame.Payload) >= 2 {
			dataLen := int(frame.Payload[0]) | int(frame.Payload[1])<<8
			if dataLen <= len(frame.Payload)-2 {
				protoData = frame.Payload[2 : 2+dataLen]
			}
		}
		if len(protoData) == 0 {
			protoData = []byte{}
		}
		if err := proto.Unmarshal(protoData, msg); err == nil {
			if jsonBytes, err := json.Marshal(msg); err == nil {
				entry.PayloadJSON = jsonBytes
			}
		}
	}
	return entry.PayloadJSON
}

// ReleasePendingMessages 快照并放行指定连接的 pending 消息（按 seq_id 升序）
// editsJSON: map[seq_id string]payload JSON；未提供编辑的使用原始 payload
func (w *RecordWorker) ReleasePendingMessages(connID uint64, editsJSON string) error {
	v, ok := w.connStates.Load(connID)
	if !ok {
		return nil
	}
	state := v.(*connInterceptState)

	var edits map[string]json.RawMessage
	if editsJSON != "" {
		if err := json.Unmarshal([]byte(editsJSON), &edits); err != nil {
			return fmt.Errorf("解析 edits 失败: %w", err)
		}
	}

	state.mu.Lock()
	snapshot := state.pending
	state.pending = make([]*pendingFrame, 0)
	serverConn := state.serverConn
	state.mu.Unlock()

	if len(snapshot) == 0 {
		return nil
	}

	sort.Slice(snapshot, func(i, j int) bool {
		return snapshot[i].seqID < snapshot[j].seqID
	})

	for i, pf := range snapshot {
		payloadJSON := string(pf.payloadJSON)
		if edits != nil {
			if edit, ok := edits[strconv.FormatUint(uint64(pf.seqID), 10)]; ok && len(edit) > 0 {
				payloadJSON = string(edit)
			}
		}

		// 使用代理级别的 seqId（而非客户端原始 seqId），确保服务端看到连续递增
		proxySeqID := state.nextProxySeqID()
		frame, err := streamproto.EncodeClientMessage(pf.msgID, proxySeqID, payloadJSON)
		if err != nil {
			state.mu.Lock()
			state.pending = append(snapshot[i:], state.pending...)
			state.mu.Unlock()
			return fmt.Errorf("编码失败 clientSeqID=%d proxySeqID=%d: %w", pf.seqID, proxySeqID, err)
		}

		state.writeMu.Lock()
		_, werr := serverConn.Write(frame)
		state.writeMu.Unlock()
		if werr != nil {
			state.mu.Lock()
			state.pending = append(snapshot[i:], state.pending...)
			state.mu.Unlock()
			return fmt.Errorf("写入失败 clientSeqID=%d proxySeqID=%d: %w", pf.seqID, proxySeqID, werr)
		}

		if w.recorder != nil {
			w.recorder.UpdatePayloadBySeqID(pf.seqID, streamproto.DirClientToServer, json.RawMessage(payloadJSON))
		}
		log.Printf("[RecordWorker] TCP #%d 放行 clientSeqID=%d → proxySeqID=%d MsgID=%d", connID, pf.seqID, proxySeqID, pf.msgID)
	}

	return nil
}

// ReleaseAllPending 放行所有连接的 pending 消息
func (w *RecordWorker) ReleaseAllPending(editsJSON string) error {
	var firstErr error
	w.connStates.Range(func(key, _ any) bool {
		connID := key.(uint64)
		if err := w.ReleasePendingMessages(connID, editsJSON); err != nil && firstErr == nil {
			firstErr = err
		}
		return true
	})
	return firstErr
}

// IsRecording 录制是否活跃
func (w *RecordWorker) IsRecording() bool {
	return w.IsRunning()
}

// HasActiveConnections 是否有活跃的代理连接
func (w *RecordWorker) HasActiveConnections() bool {
	hasActive := false
	w.connStates.Range(func(key, value any) bool {
		s := value.(*connInterceptState)
		s.mu.Lock()
		if s.serverConn != nil {
			hasActive = true
		}
		s.mu.Unlock()
		return !hasActive
	})
	return hasActive
}

// GetActiveAccounts 返回录制代理中所有活跃连接对应的账号列表
func (w *RecordWorker) GetActiveAccounts() []string {
	var accounts []string
	w.connStates.Range(func(key, value any) bool {
		s := value.(*connInterceptState)
		s.mu.Lock()
		active := s.serverConn != nil
		s.mu.Unlock()
		if active {
			connID := key.(uint64)
			if v, ok := w.connAccounts.Load(connID); ok {
				accounts = append(accounts, v.(string))
			}
		}
		return true
	})
	return accounts
}

// HasAccountConnection 检查指定账号是否在录制代理中有活跃连接
func (w *RecordWorker) HasAccountConnection(accountID string) bool {
	found := false
	w.connStates.Range(func(key, value any) bool {
		s := value.(*connInterceptState)
		s.mu.Lock()
		active := s.serverConn != nil
		s.mu.Unlock()
		if active {
			connID := key.(uint64)
			if v, ok := w.connAccounts.Load(connID); ok && v.(string) == accountID {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// InjectMessages 通过录制代理的活动连接注入消息（不新建连接）
// 使用代理的 proxySeqId 保证服务端看到连续递增的 seqId
// 返回发送的消息数量和错误
func (w *RecordWorker) InjectMessages(messagesJSON string, repeatCount int, onProgress func(total, sent int, currentMsg string) bool, onMessage streamproto.ReplayMessageCallback) (int, error) {
	var messages []streamproto.RecordMessage
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		return 0, fmt.Errorf("解析消息列表失败: %v", err)
	}

	// 按账号精确查找活跃连接（利用 connAccounts 反查 connID）
	var state *connInterceptState
	var actualAccount string
	w.connAccounts.Range(func(key, value any) bool {
		accountID := value.(string)
		connID := key.(uint64)
		if s, ok := w.connStates.Load(connID); ok {
			cs := s.(*connInterceptState)
			cs.mu.Lock()
			hasConn := cs.serverConn != nil
			cs.mu.Unlock()
			if hasConn {
				state = cs
				actualAccount = accountID
				return false
			}
		}
		return true
	})

	if state == nil {
		return 0, fmt.Errorf("录制代理无活跃连接，无法注入消息")
	}

	state.mu.Lock()
	serverConn := state.serverConn
	state.mu.Unlock()

	log.Printf("[RecordWorker] InjectMessages: 通过代理连接 #%d (账号: %s) 注入 %d 条消息, 重复 %d 次",
		state.connID, actualAccount, len(messages), repeatCount)

	total := len(messages) * repeatCount
	sent := 0

	for r := 0; r < repeatCount; r++ {
		for _, msg := range messages {
			if msg.Direction != "" && msg.Direction != streamproto.DirClientToServer {
				continue
			}

			// 使用代理的 proxySeqId
			proxySeqID := state.nextProxySeqID()
			frame, err := streamproto.EncodeClientMessage(msg.MsgID, proxySeqID, msg.PayloadJSON)
			if err != nil {
				log.Printf("[RecordWorker] 注入消息编码失败: %v", err)
				continue
			}

			state.writeMu.Lock()
			_, werr := serverConn.Write(frame)
			state.writeMu.Unlock()
			if werr != nil {
				return sent, fmt.Errorf("注入消息写入失败: %v", werr)
			}

			if onMessage != nil {
				onMessage(msg.MsgName, msg.MsgID, proxySeqID, msg.PayloadJSON, 0, streamproto.DirClientToServer, actualAccount)
			}

			sent++

			if onProgress != nil {
				if !onProgress(total, sent, msg.MsgName) {
					return sent, fmt.Errorf("注入已取消")
				}
			}

			if streamproto.SendIntervalMs > 0 {
				time.Sleep(time.Duration(streamproto.SendIntervalMs) * time.Millisecond)
			}
		}
	}

	// 等待服务端 Ack
	if streamproto.AckWaitMs > 0 {
		time.Sleep(time.Duration(streamproto.AckWaitMs) * time.Millisecond)
	}

	log.Printf("[RecordWorker] InjectMessages 完成: 发送 %d 条 (proxySeqId 至 %d)", sent, state.peekProxySeqID())
	return sent, nil
}

// interceptAndParse 拦截模式：解析客户端->服务端方向的帧
// LoginReq/Ping 立即透传；MsgID>=1000 入 pending 并推送前端；关闭 filterMode 后透传
func (w *RecordWorker) interceptAndParse(clientConn net.Conn, serverConn net.Conn, connID uint64) int64 {
	var total int64
	dir := streamproto.DirClientToServer
	state := w.getOrCreateConnState(connID, serverConn)

	// safeWrite 通过 writeMu 保护写入 serverConn
	safeWrite := func(data []byte) bool {
		state.writeMu.Lock()
		defer state.writeMu.Unlock()
		_, err := serverConn.Write(data)
		return err == nil
	}

	for {
		header := make([]byte, streamproto.FrameHeaderSize)
		n, err := io.ReadFull(clientConn, header)
		if err != nil {
			if err != io.EOF && total > 0 {
				log.Printf("[TCP] #%d %s 读取帧头结束: %v", connID, dir, err)
			}
			return total
		}
		total += int64(n)

		msgLen, _, herr := streamproto.ParseFrameHeader(header)
		if herr != nil {
			log.Printf("[TCP] #%d %s 帧头解析失败: %v", connID, dir, herr)
			safeWrite(header)
			return total
		}

		body := make([]byte, msgLen)
		n, err = io.ReadFull(clientConn, body)
		if err != nil {
			log.Printf("[TCP] #%d %s 读取消息体失败: %v", connID, dir, err)
			safeWrite(header)
			return total
		}
		total += int64(n)

		raw := make([]byte, len(header)+len(body))
		copy(raw, header)
		copy(raw[len(header):], body)

		frame, derr := streamproto.DecodeFrame(raw, true)
		if derr != nil {
			log.Printf("[TCP] #%d %s 解码失败: %v", connID, dir, derr)
			safeWrite(header)
			safeWrite(body)
			continue
		}

		// filterMode 关闭后：透传并录制 Proto 消息
		if !w.IsFilterMode() {
			if w.recorder != nil && frame.MsgID >= 1000 {
				if err := w.recorder.RecordFrame(frame, dir); err != nil {
					log.Printf("[录制] 错误: %v", err)
				}
			}
			if !w.forwardFrame(serverConn, connID, dir, header, body) {
				return total
			}
			continue
		}

		// LoginReq：立即透传 + 录制 login payload + 提取账号
		if frame.MsgID == 1 {
			if w.recorder != nil {
				w.recorder.RecordLoginPayload(frame.Payload)
				log.Printf("[录制] LoginReq payload(%dB)", len(frame.Payload))
			}
			w.onLoginReqParsed(connID, frame.Payload)
			if !w.forwardFrame(serverConn, connID, dir, header, body) {
				return total
			}
			continue
		}

		// Ping：立即透传，不入 pending
		if frame.MsgID == 3 {
			if !w.forwardFrame(serverConn, connID, dir, header, body) {
				return total
			}
			continue
		}

		// Proto 消息：入 pending，不转发
		if frame.MsgID >= 1000 {
			payloadJSON := w.framePayloadJSON(frame)
			if w.recorder != nil {
				// 静默写入录制缓冲；前端仅通过 record:intercepted 追加，避免与 progress 重复
				if err := w.recorder.RecordFrameSilent(frame, dir); err != nil {
					log.Printf("[录制] 错误: %v", err)
				}
			}

			hdrCopy := make([]byte, len(header))
			copy(hdrCopy, header)
			bodyCopy := make([]byte, len(body))
			copy(bodyCopy, body)

			state.mu.Lock()
			state.pending = append(state.pending, &pendingFrame{
				seqID:       frame.SeqID,
				msgID:       frame.MsgID,
				header:      hdrCopy,
				body:        bodyCopy,
				payloadJSON: payloadJSON,
			})
			state.mu.Unlock()

			log.Printf("[RecordWorker] TCP #%d 拦截消息 MsgID=%d (%s), SeqID=%d", connID, frame.MsgID, streamproto.GetMsgName(frame.MsgID), frame.SeqID)
			w.emitInterceptedMessage(frame, connID)
			continue
		}

		// 其他框架消息：透传 + 日志
		log.Printf("[RecordWorker] TCP #%d 透传框架消息 MsgID=%d", connID, frame.MsgID)
		if !w.forwardFrame(serverConn, connID, dir, header, body) {
			return total
		}
	}
}

// emitInterceptedMessage 推送被拦截的消息到前端
// 使用 singleEntryToView 确保数据契约，数据格式为 map[string]any{"latest_msg": entryView, "conn_id": connID}
func (w *RecordWorker) emitInterceptedMessage(frame *streamproto.DecodedFrame, connID uint64) {
	// 构建 RecordEntry 用于 singleEntryToView 转换
	entry := &streamproto.RecordEntry{
		OffsetMs:    0, // 拦截消息不记录时间偏移
		MsgID:       frame.MsgID,
		MsgName:     streamproto.GetMsgName(frame.MsgID),
		SeqID:       frame.SeqID,
		Direction:   streamproto.DirClientToServer,
		PayloadJSON: nil,
	}

	// 将 payload 序列化为 JSON（复用 RecordFrame 中的逻辑）
	msg, ok := streamproto.NewMessage(frame.MsgID)
	if ok {
		var protoData []byte
		if len(frame.Payload) >= 2 {
			dataLen := int(frame.Payload[0]) | int(frame.Payload[1])<<8
			if dataLen <= len(frame.Payload)-2 {
				protoData = frame.Payload[2 : 2+dataLen]
			}
		}
		if len(protoData) == 0 {
			protoData = []byte{}
		}
		if err := proto.Unmarshal(protoData, msg); err == nil {
			if jsonBytes, err := json.Marshal(msg); err == nil {
				entry.PayloadJSON = jsonBytes
			}
		}
	}

	// 使用 singleEntryToView 转换为 RecordEntryView（唯一转换点，确保数据契约）
	entryView := singleEntryToView(entry, 0)

	data := map[string]any{
		"latest_msg": entryView,
		"conn_id":    connID,
	}
	log.Printf("[RecordWorker] emit intercepted: MsgID=%d, connID=%d", frame.MsgID, connID)
	w.emitter.Emit("record:intercepted", data)
}

// readAllWithTimeout 带超时的读取全部响应体
func readAllWithTimeout(body io.Reader, timeout time.Duration) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := io.ReadAll(body)
		done <- result{data, err}
	}()
	select {
	case r := <-done:
		return r.data, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("读取响应体超时")
	}
}
