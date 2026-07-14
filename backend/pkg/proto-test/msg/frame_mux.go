// FrameMux — 帧多路复用器，替代 readDrainer 提供变量提取所需的帧缓存能力
//
// 核心职责：
//   - 持续从 TCP 连接读取帧（readLoop 生产者）
//   - 按 watchedIDs 缓存最新帧（map 覆盖式）
//   - 提供阻塞等待特定 MsgID 的能力（WaitMsg 消费者，channel 通知）
//   - 实时帧流（frames channel）
//   - 安全退出（stop + SetReadDeadline 强制中断 io.ReadFull）
//
// 并发模型：
//
//	readLoop (写 cache)  ←→  WaitMsg (读 cache)  ←→  外部查询
//	       mu (RWMutex) 保护 cache，notifyCh (channel) 唤醒等待者
//
// 生命周期：
//
//	NewFrameMux → DrainAndStart → [WaitMsg 调用] → Stop
package protocol

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// FrameMux 帧多路复用器
// 从 TCP 连接持续读取帧，按 watchedIDs 缓存，支持阻塞等待和实时帧流
type FrameMux struct {
	conn       net.Conn
	mu         sync.RWMutex             // 保护 cache 和 stopped
	cache      map[uint16]*DecodedFrame // 按 MsgID 缓存最新帧（仅 watchedIDs）
	watchedIDs map[uint16]bool          // 需要缓存的 MsgID 集合
	notifyCh   chan struct{}            // readLoop 写 cache 后发信号唤醒 WaitMsg
	frames     chan *DecodedFrame       // 实时帧流（缓冲 1024）
	done       chan struct{}            // 停止信号
	stopped    bool                     // 标记已停止（mu 保护）
	once       sync.Once                // 确保 done 只关闭一次
	wg         sync.WaitGroup           // 等待 readLoop 退出
}

// NewFrameMux 创建帧多路复用器
// conn: TCP 连接（已建立，调用方负责 Borrow 后传入）
// watchedIDs: 需要缓存到 cache 的 MsgID 集合（变量提取关注的消息）
func NewFrameMux(conn net.Conn, watchedIDs []uint16) *FrameMux {
	m := &FrameMux{
		conn:       conn,
		cache:      make(map[uint16]*DecodedFrame),
		watchedIDs: make(map[uint16]bool),
		notifyCh:   make(chan struct{}, 1), // 缓冲 1，允许一次信号积压
		frames:     make(chan *DecodedFrame, 1024),
		done:       make(chan struct{}),
	}
	for _, id := range watchedIDs {
		m.watchedIDs[id] = true
	}
	return m
}

// decodeError 标记帧解码阶段的错误（readLoop 应 continue 而非 return）
type decodeError struct{ err error }

func (e *decodeError) Error() string { return e.err.Error() }
func (e *decodeError) Unwrap() error { return e.err }

func isDecodeError(err error) bool {
	_, ok := err.(*decodeError)
	return ok
}

// readFrame 从连接读取一帧并解码（内部方法，供 DrainAndStart 和 readLoop 复用）
func (m *FrameMux) readFrame() (*DecodedFrame, error) {
	header := make([]byte, FrameHeaderSize)
	if _, err := io.ReadFull(m.conn, header); err != nil {
		return nil, err
	}

	msgLen, _, err := ParseFrameHeader(header)
	if err != nil {
		return nil, &decodeError{err: fmt.Errorf("帧头解析失败: %w", err)}
	}

	body := make([]byte, msgLen)
	if _, err := io.ReadFull(m.conn, body); err != nil {
		return nil, &decodeError{err: fmt.Errorf("读取消息体失败: %w", err)}
	}

	raw := make([]byte, FrameHeaderSize+msgLen)
	copy(raw, header)
	copy(raw[FrameHeaderSize:], body)
	frame, err := DecodeFrame(raw, false)
	if err != nil {
		return nil, &decodeError{err: err}
	}
	return frame, nil
}

// DrainAndStart 先消费积压帧（替代 DrainConn），然后启动 readLoop
// drainTimeout: 排空阶段超时（通常 100ms，与 DrainConn 一致）
// onMessage: 前端推送回调（可为 nil）
// accountID: 当前账号标识
func (m *FrameMux) DrainAndStart(drainTimeout time.Duration, onMessage ReplayMessageCallback, accountID string) {
	// drain 阶段：设置短 deadline 逐帧读取并丢弃积压帧。
	// 重要:drain 只丢弃积压数据,不缓存 watchedMsgID 帧。
	// 原因:池复用连接的积压帧属于"上一次会话"的残留,若缓存会让后续惰性变量提取
	// (WaitMsg 命中 cache)读到过期数据。变量提取的数据源必须是 readLoop 启动后
	// 当前会话收到的新帧。
	_ = m.conn.SetReadDeadline(time.Now().Add(drainTimeout))
	for {
		frame, err := m.readFrame()
		if err != nil {
			// 超时或连接错误，drain 结束
			break
		}

		log.Printf("[FrameMux] drain 丢弃积压: MsgID=%d, SeqID=%d, %dB", frame.MsgID, frame.SeqID, len(frame.Payload))
	}

	// drain 结束：清除 deadline
	_ = m.conn.SetReadDeadline(time.Time{})

	// 启动 readLoop goroutine
	m.wg.Add(1)
	go m.readLoop(onMessage, accountID)
}

// readLoop 持续从连接读取帧并分发
//
// 流程：
//  1. 检查 done 信号 → 非阻塞退出
//  2. 读取 4 字节帧头 → ParseFrameHeader → 读取 body → DecodeFrame
//  3. watchedMsgID 帧 → 持写锁写入 cache → 发送通知信号
//  4. 所有帧 → 非阻塞写入 frames channel（满时丢弃）
//  5. DecodeFrame 错误 → continue（不 return）
//  6. io.ReadFull 错误（连接关闭/stop 强制中断）→ return
func (m *FrameMux) readLoop(onMessage ReplayMessageCallback, accountID string) {
	defer m.wg.Done()
	defer log.Printf("[FrameMux] readLoop 已退出")
	// F3 修复 (2026-06-15): readLoop 因连接异常退出时，必须传播失败状态
	// (stopped=true + close(done))，否则 WaitMsg 不知道 readLoop 已死，
	// 后续每条变量消息都要各自等满 5s 超时，多账号并发场景延迟放大。
	defer m.signalStopped()

	for {
		// 检查停止信号（非阻塞）
		select {
		case <-m.done:
			return
		default:
		}

		frame, err := m.readFrame()
		if err != nil {
			select {
			case <-m.done:
				return
			default:
			}
			// 解码错误（如 snappy 解压失败）→ 跳过此帧继续读
			if isDecodeError(err) {
				log.Printf("[FrameMux] readLoop 帧解码失败，跳过: %v", err)
				continue
			}
			// 连接错误（io.EOF、deadline 等）→ 退出
			log.Printf("[FrameMux] readLoop 读取帧失败: %v", err)
			return
		}

		msgName := GetMsgName(frame.MsgID)
		if msgName == "" {
			switch frame.MsgID {
			case 2:
				msgName = "LoginResp"
			case 4:
				msgName = "Pong"
			default:
				msgName = fmt.Sprintf("Unknown(%d)", frame.MsgID)
			}
		}

		// watchedMsgID → 缓存到 cache + 通知等待者
		if m.watchedIDs[frame.MsgID] {
			m.mu.Lock()
			m.cache[frame.MsgID] = frame
			m.mu.Unlock()
			// 非阻塞发送通知（channel 缓冲 1，如果已有未消费的信号则跳过）
			select {
			case m.notifyCh <- struct{}{}:
			default:
			}
		}

		// 非阻塞写入 frames channel（满时丢弃）
		select {
		case m.frames <- frame:
		default:
			// channel 满了，丢弃当前帧（避免 readLoop 阻塞）
			log.Printf("[FrameMux] frames channel 已满，丢弃帧: MsgID=%d", frame.MsgID)
		}

		// 推送前端回调
		if onMessage != nil && frame.MsgID >= 1000 {
			if jsonPayload, err := payloadToJSON(frame.MsgID, frame.Payload); err == nil {
				onMessage(msgName, frame.MsgID, frame.SeqID, jsonPayload, 0, DirServerToClient, accountID)
			}
		}

		log.Printf("[FrameMux] ← %s (MsgID=%d, SeqID=%d, %dB)", msgName, frame.MsgID, frame.SeqID, len(frame.Payload))
	}
}

// WaitMsg 等待特定 MsgID 的帧（阻塞，带超时）
//
// 流程：
//  1. 持锁查 cache → 有值立即返回
//  2. 无值 → 进入等待循环：
//     a. 释放锁，等待 notifyCh 信号或超时
//     b. 收到信号后重新获取锁查 cache
//  3. 超时 → 返回 error
//  4. Stop 后 → 返回 error（连接已关闭）
//
// 使用 channel（notifyCh）替代 sync.Cond 实现通知，避免 Broadcast 丢失问题。
// WaitMsg 阻塞等待指定 MsgID 的缓存帧，直到命中、超时或 FrameMux 停止。
//
// 契约（F5 文档化）：同一 FrameMux 同时仅允许一个 WaitMsg 调用者。
// 原因：notifyCh 缓冲为 1，readLoop 用非阻塞 send 合并通知；若多个 WaitMsg 并发等待，
// 某个等待者可能错过属于自己的通知而误等超时。当前架构下 sendMessagesOnce 单线程
// 串行调用 ExtractVariablesForMessage → WaitMsg，满足单等待者前提。
func (m *FrameMux) WaitMsg(msgID uint16, timeout time.Duration) (*DecodedFrame, error) {
	deadline := time.Now().Add(timeout)
	for {
		// 检查 stopped 状态
		m.mu.RLock()
		stopped := m.stopped
		m.mu.RUnlock()
		if stopped {
			return nil, fmt.Errorf("FrameMux 已停止，等待 MsgID=%d 失败", msgID)
		}

		// 持读锁查 cache
		m.mu.RLock()
		frame, ok := m.cache[msgID]
		m.mu.RUnlock()
		if ok {
			return frame, nil
		}

		// 计算剩余超时
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, fmt.Errorf("等待 MsgID=%d 超时 (%v)", msgID, timeout)
		}

		// 等待通知信号或超时
		select {
		case <-m.notifyCh:
			// 收到通知，循环回到顶部重新查 cache
		case <-time.After(remaining):
			return nil, fmt.Errorf("等待 MsgID=%d 超时 (%v)", msgID, timeout)
		case <-m.done:
			return nil, fmt.Errorf("FrameMux 已停止，等待 MsgID=%d 失败", msgID)
		}
	}
}

// GetCache 获取指定 MsgID 的缓存帧（非阻塞，用于外部查询）
func (m *FrameMux) GetCache(msgID uint16) (*DecodedFrame, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	frame, ok := m.cache[msgID]
	return frame, ok
}

// Frames 返回实时帧流 channel（只读）
func (m *FrameMux) Frames() <-chan *DecodedFrame {
	return m.frames
}

// signalStopped 标记 FrameMux 已停止并关闭 done 信号（幂等，由 once 保护）。
// readLoop 连接异常退出和 Stop() 都通过它传播失败状态，
// 让 WaitMsg 通过 done/stopped 分支立即返回，而非等满超时。
func (m *FrameMux) signalStopped() {
	m.once.Do(func() {
		m.mu.Lock()
		m.stopped = true
		m.mu.Unlock()
		close(m.done)
		// 强制 io.ReadFull 返回（设置已过期的 deadline）
		_ = m.conn.SetReadDeadline(time.Now())
	})
}

// Stop 安全停止 FrameMux
//
// 退出流程：
//  1. close(done) → 信号 readLoop 退出
//  2. conn.SetReadDeadline(time.Now()) → 强制 io.ReadFull 返回 timeout 错误
//  3. 标记 stopped → WaitMsg 立即返回错误
//  4. wg.Wait() → 等待 readLoop goroutine 退出
func (m *FrameMux) Stop() {
	m.signalStopped()
	m.wg.Wait()
}
