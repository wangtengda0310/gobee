package protocol

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// ConnState 连接状态
type ConnState int

const (
	ConnStateIdle     ConnState = iota // 空闲，在池中可用
	ConnStateBorrowed                  // 已借出，正在使用
	ConnStateClosed                    // 已关闭
)

// PooledConn 池化连接
// 每个账号对应一个 PooledConn，管理 TCP 长连接的生命周期、心跳和并发安全。
//
// 并发安全设计：
//   - mu: 保护 state、seqID 等状态字段
//   - writeMu: 保护底层 conn 的并发写（心跳与业务写互斥）
//     Borrow() 返回 writeLockedConn，其 Write 方法自动获取 writeMu
//     保证同一时刻只有一个 goroutine 在写 conn
//   - stopHeart: 通知心跳 goroutine 停止（Borrow 时关闭，Return 后重建）
type PooledConn struct {
	mu         sync.Mutex
	writeMu    sync.Mutex // 保护底层 conn 的并发写（心跳与业务写互斥）
	accountID  string
	conn       net.Conn
	state      ConnState
	serverAddr string
	httpAddr   string
	stopHeart  chan struct{} // 停止心跳的信号
	seqID      uint32        // 服务端已知的最后 seqId（用于连接复用时续接序列号）
}

// writeLockedConn 包装 net.Conn，在 Write 时自动获取 PooledConn 的 writeMu。
// 借出期间心跳 goroutine 已停止，但 writeMu 仍保证同一时刻只有一个写者。
// 这是连接池并发安全的核心：防止心跳与 borrower 并发写同一 net.Conn。
type writeLockedConn struct {
	net.Conn
	pc *PooledConn
}

// Write 在写底层 conn 前获取 writeMu，保证写操作互斥
func (wlc *writeLockedConn) Write(p []byte) (n int, err error) {
	wlc.pc.writeMu.Lock()
	defer wlc.pc.writeMu.Unlock()
	return wlc.Conn.Write(p)
}

// ConnPoolEntry 连接池条目信息（只读快照，返回给前端）
type ConnPoolEntry struct {
	AccountID  string `json:"account_id"`
	ServerAddr string `json:"server_addr"`
	State      string `json:"state"`    // "idle" | "borrowed" | "closed"
	ConnAge    string `json:"conn_age"` // 连接存活时间
}

// AccountConnectionPool 账号连接池
// 按账号 ID 维护已登录的 TCP 长连接，支持借出/归还模式。
//
// 生命周期：
//   - GetOrCreate: 获取已有连接或新建（HTTP 登录 + TCP + LoginReq）
//   - Borrow: 标记为借出状态（阻止心跳写入干扰），返回 writeLockedConn
//   - Return: 归还到池中（恢复心跳）
//   - Close/CloseAll: 关闭连接并从池中移除
//
// 并发安全：
//   - p.mu 保护 conns map
//   - pc.mu 保护单个连接的状态（state、seqID）
//   - pc.writeMu 保护底层 conn 的写操作（心跳与业务互斥）
//   - 心跳 goroutine 在 Borrow 时通过关闭 stopHeart 停止，Return 后重建
//   - 连接存活检测由 TCP keepalive 负责，不再用 IsConnAlive 探测（避免丢弃帧首字节）
type AccountConnectionPool struct {
	mu    sync.Mutex
	conns map[string]*PooledConn // accountID -> PooledConn
}

// NewAccountConnectionPool 创建连接池
func NewAccountConnectionPool() *AccountConnectionPool {
	return &AccountConnectionPool{
		conns: make(map[string]*PooledConn),
	}
}

// GetOrCreate 获取已有连接或创建新连接
// 如果池中已有该账号的空闲连接，直接返回；
// 否则新建连接（HTTP 登录 + TCP 连接 + LoginReq）。
// 返回的连接处于 idle 状态，调用方需 Borrow 后再使用。
func (p *AccountConnectionPool) GetOrCreate(accountID, serverAddr, httpAddr string) (*PooledConn, error) {
	// 快速路径：检查已有连接（短持锁）
	p.mu.Lock()
	if pc, ok := p.conns[accountID]; ok {
		pc.mu.Lock()
		if pc.state == ConnStateIdle {
			pc.mu.Unlock()
			p.mu.Unlock()
			log.Printf("[ConnPool] 复用已有连接: account=%s", accountID)
			return pc, nil
		}
		pc.mu.Unlock()
		p.closeConnLocked(accountID)
		log.Printf("[ConnPool] 旧连接非空闲，重建: account=%s", accountID)
	}
	p.mu.Unlock()

	// 慢路径：在锁外拨号（避免阻塞整个连接池）
	conn, err := p.dialAndLogin(accountID, serverAddr, httpAddr)
	if err != nil {
		return nil, err
	}

	// 重新获取锁，插入新连接（处理并发竞争：另一 goroutine 可能已创建同账号连接）
	p.mu.Lock()
	defer p.mu.Unlock()
	if existing, ok := p.conns[accountID]; ok {
		existing.mu.Lock()
		if existing.state == ConnStateIdle {
			existing.mu.Unlock()
			_ = conn.Close() // 丢弃刚创建的连接，复用已有的
			log.Printf("[ConnPool] 并发创建冲突，复用已有连接: account=%s", accountID)
			return existing, nil
		}
		existing.mu.Unlock()
		p.closeConnLocked(accountID)
	}

	pc := &PooledConn{
		accountID:  accountID,
		conn:       conn,
		state:      ConnStateIdle,
		serverAddr: serverAddr,
		httpAddr:   httpAddr,
		stopHeart:  make(chan struct{}),
	}
	p.conns[accountID] = pc

	go p.heartbeat(pc)

	log.Printf("[ConnPool] 新建连接: account=%s, server=%s", accountID, serverAddr)
	return pc, nil
}

// Borrow 借出连接（标记为使用中，暂停心跳）
// 返回包装后的连接（Write 自动获取 writeMu）和当前服务端已知的 seqId。
// 调用方应从此值续接，而非从 1 开始。
// 返回的 writeLockedConn 保证写操作与心跳互斥，但读操作仍需调用方自行协调。
func (pc *PooledConn) Borrow() (net.Conn, uint32) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if pc.state != ConnStateIdle {
		return nil, 0
	}
	pc.state = ConnStateBorrowed
	// 通知心跳停止
	close(pc.stopHeart)
	pc.stopHeart = make(chan struct{})
	return &writeLockedConn{Conn: pc.conn, pc: pc}, pc.seqID
}

// UpdateSeqID 更新服务端已知的 seqId（发送消息后调用）
func (pc *PooledConn) UpdateSeqID(seqID uint32) {
	pc.mu.Lock()
	pc.seqID = seqID
	pc.mu.Unlock()
}

// Return 归还连接（恢复空闲状态，重启心跳）
// lastSeqID: 本次发送使用的最大 seqId（用于下次复用时续接）
// 归还前必须确保所有读取 goroutine（如 readDrainer）已完全退出，否则下一个 borrower 会与之并发读取同一 net.Conn。
// 调用方应通过 readerDone channel 同步等待 readDrainer 退出后再调用 Return。
func (p *AccountConnectionPool) Return(accountID string, lastSeqID uint32) {
	p.mu.Lock()
	pc, ok := p.conns[accountID]
	p.mu.Unlock()

	if !ok {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.state != ConnStateBorrowed {
		return
	}

	if pc.conn == nil {
		pc.state = ConnStateClosed
		p.mu.Lock()
		delete(p.conns, accountID)
		p.mu.Unlock()
		log.Printf("[ConnPool] 归还时发现连接为 nil，移除: account=%s", accountID)
		return
	}

	pc.state = ConnStateIdle
	pc.seqID = lastSeqID
	// 重启心跳
	go p.heartbeat(pc)
	log.Printf("[ConnPool] 连接已归还: account=%s, seqID=%d", accountID, lastSeqID)
}

// Close 关闭指定账号的连接
func (p *AccountConnectionPool) Close(accountID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeConnLocked(accountID)
}

// CloseAll 关闭所有连接
func (p *AccountConnectionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for accountID := range p.conns {
		p.closeConnLocked(accountID)
	}
	log.Printf("[ConnPool] 所有连接已关闭")
}

// List 获取连接池快照
func (p *AccountConnectionPool) List() []ConnPoolEntry {
	p.mu.Lock()
	defer p.mu.Unlock()

	entries := make([]ConnPoolEntry, 0, len(p.conns))
	for _, pc := range p.conns {
		pc.mu.Lock()
		state := "idle"
		switch pc.state {
		case ConnStateBorrowed:
			state = "borrowed"
		case ConnStateClosed:
			state = "closed"
		}
		entry := ConnPoolEntry{
			AccountID:  pc.accountID,
			ServerAddr: pc.serverAddr,
			State:      state,
		}
		pc.mu.Unlock()
		entries = append(entries, entry)
	}
	return entries
}

// Has 检查指定账号是否有可用连接
func (p *AccountConnectionPool) Has(accountID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	pc, ok := p.conns[accountID]
	if !ok {
		return false
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.state == ConnStateIdle
}

// dialAndLogin 建立 TCP 连接并完成登录
func (p *AccountConnectionPool) dialAndLogin(accountID, serverAddr, httpAddr string) (net.Conn, error) {
	// HTTP 登录获取 token
	token, _, authErr := AuthLogin(httpAddr, accountID)
	if authErr != nil {
		return nil, fmt.Errorf("HTTP 登录失败: %w", authErr)
	}
	log.Printf("[ConnPool] HTTP 登录成功: account=%s", accountID)

	// TCP 连接
	conn, dialErr := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if dialErr != nil {
		return nil, fmt.Errorf("TCP 连接失败: %v", dialErr)
	}

	// 设置 TCP keepalive，让操作系统检测死连接
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(3 * time.Minute)
	}

	// 发送 LoginReq
	if loginErr := sendLoginReq(conn, accountID, token, ""); loginErr != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("发送 LoginReq 失败: %v", loginErr)
	}

	// 等待 LoginResp
	if respErr := waitLoginResp(conn); respErr != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("等待 LoginResp 失败: %v", respErr)
	}

	// 等待服务端推送完成
	time.Sleep(2 * time.Second)

	return conn, nil
}

// heartbeat 心跳保活 goroutine
// 每 5 秒发送一次 Ping 帧，保持连接活跃。
// 并发安全：
//  1. 检查 state 前获取 pc.mu
//  2. 确认 state==Idle 后获取 writeMu
//  3. 获取 writeMu 后二次确认 state（防止间隙中被 Borrow）
//  4. 写完成后释放 writeMu
//
// 写失败时直接调用 p.Close(accountID) 关闭连接，避免锁顺序反转（原实现先锁 pc.mu 再锁 p.mu 导致死锁）。
func (p *AccountConnectionPool) heartbeat(pc *PooledConn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pc.stopHeart:
			return
		case <-ticker.C:
			pc.mu.Lock()
			if pc.state != ConnStateIdle {
				pc.mu.Unlock()
				return
			}
			pc.mu.Unlock()

			pc.writeMu.Lock()
			// 二次检查：获取 writeMu 后确认状态仍是 Idle
			pc.mu.Lock()
			if pc.state != ConnStateIdle {
				pc.mu.Unlock()
				pc.writeMu.Unlock()
				return
			}
			conn := pc.conn
			pc.mu.Unlock()

			pingFrame := EncodeFrame(3, 0, FlagEncrypt, []byte{}, true)
			_, err := conn.Write(pingFrame)
			pc.writeMu.Unlock()

			if err != nil {
				log.Printf("[ConnPool] 心跳失败，标记关闭: account=%s, err=%v", pc.accountID, err)
				p.Close(pc.accountID)
				return
			}
		}
	}
}

// closeConnLocked 关闭连接（调用方必须持有 p.mu 锁）
func (p *AccountConnectionPool) closeConnLocked(accountID string) {
	pc, ok := p.conns[accountID]
	if !ok {
		return
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.state == ConnStateClosed {
		return
	}

	// 停止心跳
	close(pc.stopHeart)
	pc.state = ConnStateClosed

	if pc.conn != nil {
		_ = pc.conn.Close()
	}
	delete(p.conns, accountID)
	log.Printf("[ConnPool] 连接已关闭: account=%s", accountID)
}

// AcceptConn 接受一个外部已建立的连接加入池
// 用于录制代理移交连接的场景
func (p *AccountConnectionPool) AcceptConn(accountID, serverAddr string, conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果已有该账号的连接，先关闭旧的
	if _, ok := p.conns[accountID]; ok {
		p.closeConnLocked(accountID)
	}

	// 设置 TCP keepalive，让操作系统检测死连接
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(3 * time.Minute)
	}

	pc := &PooledConn{
		accountID:  accountID,
		conn:       conn,
		state:      ConnStateIdle,
		serverAddr: serverAddr,
		stopHeart:  make(chan struct{}),
	}
	p.conns[accountID] = pc

	// 启动心跳
	go p.heartbeat(pc)
	log.Printf("[ConnPool] 接受外部连接: account=%s, server=%s", accountID, serverAddr)
}

// DrainConn 按帧读取并丢弃连接上的残留数据，直到超时或读取/解析错误。
// 与裸 Read 不同，本函数保证每次读取都在帧边界上，不会停在帧中间。
// 超时返回 nil（表示排空完成），其他错误返回具体错误。
// 适用于连接复用前清理积压帧，确保后续读帧从完整帧头开始。
func DrainConn(conn net.Conn, timeout time.Duration) error {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	for {
		header := make([]byte, FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil
			}
			return err
		}

		msgLen, _, err := ParseFrameHeader(header)
		if err != nil {
			return err
		}

		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil
			}
			return err
		}
	}
}
