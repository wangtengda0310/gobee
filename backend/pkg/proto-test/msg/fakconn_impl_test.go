package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// fakeNetError 实现 net.Error 接口，供 FakeConn 返回超时错误。
// DrainConn 通过 netErr.Timeout() 判断超时，因此 FakeConn 必须返回 net.Error 而非普通 error。
type fakeNetError struct{ msg string }

func (e *fakeNetError) Error() string   { return e.msg }
func (e *fakeNetError) Timeout() bool   { return true }
func (e *fakeNetError) Temporary() bool { return true }

var fakeTimeoutErr net.Error = &fakeNetError{msg: "i/o timeout"}

// FakeConn 是一个实现 net.Conn 的内存连接,用于单元测试中脚本化编排协议交互。
//
// 设计参考 Netty EmbeddedChannel(单对象脚本式编排)而非纯 net.Pipe(双端同步管道)。
// 核心优势:测试代码可以精确控制"服务器在第几条客户端消息后推送什么帧",
// 这对验证 FrameMux 的 readLoop/WaitMsg/变量提取时序至关重要——
// 真实游戏服务器的 Ntf 推送时机不确定,无法稳定复现时序相关 bug。
//
// 数据流向(与真实 TCP 一致):
//
//	客户端写(conn.Write)  → outboundBuf → 测试读取(ClientWrites/WaitClientWrite)
//	测试推入(PushServerFrame) → inboundCh  → 客户端读(conn.Read, readLoop 消费)
//
// 并发模型:
//   - PushServerFrame 非阻塞写入 inboundCh(满时 panic,避免测试静默丢帧)
//   - Read 阻塞等待 inboundCh(支持 SetReadDeadline)
//   - Write 非阻塞写入 outboundBuf(测试侧主动消费)
//
// 帧编解码:PushServerFrame 接收 (msgID, payload),内部用 EncodeFrame 编码为完整协议帧
// (含加密/帧头),使 readLoop 的 DecodeFrame 能正确解码。WaitClientWrite 返回 DecodedFrame。
type FakeConn struct {
	mu           sync.Mutex
	inboundCh    chan []byte // 服务端 → 客户端(测试 Push → readLoop Read)
	outboundMu   sync.Mutex
	outboundBuf  [][]byte // 客户端 → 服务端(发送循环 Write → 测试 ClientWrites)
	outboundCond *sync.Cond

	writeDeadline time.Time
	closed        bool
	closeCh       chan struct{}

	// readDeadlineCh: SetReadDeadline 通过关闭此 channel 唤醒阻塞的 Read。
	// 每次设置新的非零 deadline 时创建新 channel,过期或被设为 now 时关闭它。
	// FrameMux.Stop 依赖 SetReadDeadline(now) 强制中断 io.ReadFull,必须能唤醒已阻塞的 Read。
	readDeadlineMu   sync.Mutex
	readDeadlineCh   chan struct{}
	readDeadlineWhen time.Time

	// 读缓冲:inboundCh 传递整帧,Read 被调用方(如 io.ReadFull)按任意大小切片读取。
	// inboundBuf 保存当前帧尚未被 Read 消费的剩余字节。
	inboundMu  sync.Mutex
	inboundBuf []byte
}

// NewFakeConn 创建一个内存连接实例
func NewFakeConn() *FakeConn {
	fc := &FakeConn{
		inboundCh:      make(chan []byte, 256),
		outboundBuf:    make([][]byte, 0),
		closeCh:        make(chan struct{}),
		readDeadlineCh: make(chan struct{}), // 初始:永远打开(无 deadline 时 Read 不被它唤醒)
	}
	fc.outboundCond = sync.NewCond(&fc.outboundMu)
	return fc
}

// PushServerFrame 编码并推入一个服务端→客户端方向的协议帧(非阻塞)。
// msgID: 消息 ID; payload: proto 序列化后的消息体(不含 ByteStream 长度前缀,本函数会补上)。
// 使用 FlagEncrypt 编码,与真实服务端帧一致,readLoop 的 DecodeFrame 能正确解密。
// 如果 inbound channel 已满则 panic(测试应控制推送节奏,满意味着测试逻辑错误)。
func (fc *FakeConn) PushServerFrame(msgID uint16, payload []byte) {
	if fc.IsClosed() {
		return
	}
	frame := fc.encodeServerFrame(msgID, 0, payload)
	select {
	case fc.inboundCh <- frame:
	default:
		panic("FakeConn: inbound channel 已满,测试推送节奏过快(请增大 channel 缓冲或控制推送频率)")
	}
}

// PushRawFrame 推入预编码的原始帧字节(与 PushServerFrame 不同,不做任何编码)。
// 供需要精确控制 flags/seqID 的测试使用。
func (fc *FakeConn) PushRawFrame(rawFrame []byte) {
	if fc.IsClosed() {
		return
	}
	cp := make([]byte, len(rawFrame))
	copy(cp, rawFrame)
	select {
	case fc.inboundCh <- cp:
	default:
		panic("FakeConn: inbound channel 已满")
	}
}

// encodeServerFrame 编码服务端→客户端帧(isClientData=false)
func (fc *FakeConn) encodeServerFrame(msgID uint16, seqID uint32, payload []byte) []byte {
	return EncodeFrame(msgID, seqID, FlagEncrypt, payload, false)
}

// WaitClientWrite 阻塞等待客户端(发送循环)写出下一帧,返回解码后的 DecodedFrame。
// 超时返回错误,避免测试死锁。测试据此断言客户端发出了预期的消息。
func (fc *FakeConn) WaitClientWrite(timeout time.Duration) (*DecodedFrame, error) {
	frame, ok := fc.popOutbound(timeout)
	if !ok {
		return nil, fmt.Errorf("等待客户端写出超时 (%v)", timeout)
	}
	decoded, err := DecodeFrame(frame, true) // 客户端→服务端方向
	if err != nil {
		return nil, fmt.Errorf("解码客户端帧失败: %w", err)
	}
	return decoded, nil
}

// ClientWrites 非阻塞返回当前所有客户端写出的帧(已解码),无数据返回 nil。
func (fc *FakeConn) ClientWrites() []*DecodedFrame {
	frames := fc.drainOutbound()
	if len(frames) == 0 {
		return nil
	}
	result := make([]*DecodedFrame, 0, len(frames))
	for _, raw := range frames {
		if decoded, err := DecodeFrame(raw, true); err == nil {
			result = append(result, decoded)
		}
	}
	return result
}

// popOutbound 取出一条客户端写出的帧,带超时
func (fc *FakeConn) popOutbound(timeout time.Duration) ([]byte, bool) {
	fc.outboundMu.Lock()
	defer fc.outboundMu.Unlock()

	if len(fc.outboundBuf) > 0 {
		frame := fc.outboundBuf[0]
		fc.outboundBuf = fc.outboundBuf[1:]
		return frame, true
	}

	// 超时触发器:到点后设置标志并 Broadcast 唤醒 Wait。
	// Cond 自身不支持超时,用 time.AfterFunc 在独立 goroutine 中触发唤醒。
	timedOut := false
	timer := time.AfterFunc(timeout, func() {
		fc.outboundMu.Lock()
		timedOut = true
		fc.outboundCond.Broadcast()
		fc.outboundMu.Unlock()
	})
	defer timer.Stop()

	for len(fc.outboundBuf) == 0 && !timedOut && !fc.closed {
		fc.outboundCond.Wait()
	}

	if len(fc.outboundBuf) > 0 {
		frame := fc.outboundBuf[0]
		fc.outboundBuf = fc.outboundBuf[1:]
		return frame, true
	}
	return nil, false
}

// drainOutbound 取出所有客户端写出的帧(非阻塞)
func (fc *FakeConn) drainOutbound() [][]byte {
	fc.outboundMu.Lock()
	defer fc.outboundMu.Unlock()
	frames := fc.outboundBuf
	fc.outboundBuf = make([][]byte, 0)
	return frames
}

// IsClosed 返回连接是否已关闭
func (fc *FakeConn) IsClosed() bool {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.closed
}

// ========== net.Conn 接口实现 ==========

// Read 实现 io.Reader(被 FrameMux.readLoop 的 io.ReadFull 调用)。
//
// 关键:调用方(io.ReadFull)会分次读取——先读 4 字节帧头,再按帧头声明的长度读 body。
// 因此 Read 不能"一次返回一整帧",必须维护字节流语义:从 inboundBuf 取出调用方所需字节数,
// inboundBuf 耗尽时从 inboundCh 取下一帧拼接到 inboundBuf 尾部。
// 这样多次 Read 能跨帧边界连续读取,与真实 TCP 流行为一致。
//
// 超时机制:SetReadDeadline 通过 readDeadlineCh 唤醒阻塞的 Read。
// FrameMux.Stop 依赖 SetReadDeadline(now) 强制中断 io.ReadFull。
func (fc *FakeConn) Read(b []byte) (int, error) {
	fc.mu.Lock()
	closed := fc.closed
	fc.mu.Unlock()
	if closed {
		return 0, io.EOF
	}

	fc.inboundMu.Lock()
	defer fc.inboundMu.Unlock()

retry:
	// inboundBuf 有剩余字节:直接拷贝,无需等待
	if len(fc.inboundBuf) > 0 {
		n := copy(b, fc.inboundBuf)
		fc.inboundBuf = fc.inboundBuf[n:]
		return n, nil
	}

	// inboundBuf 空:获取当前 deadline 信号 channel 和 deadline 时刻
	fc.readDeadlineMu.Lock()
	dlCh := fc.readDeadlineCh
	dlWhen := fc.readDeadlineWhen
	fc.readDeadlineMu.Unlock()

	// 构造超时 timer(若 deadline 已设置)
	var timerC <-chan time.Time
	if !dlWhen.IsZero() {
		if d := time.Until(dlWhen); d > 0 {
			timer := time.NewTimer(d)
			defer timer.Stop()
			timerC = timer.C
		} else {
			return 0, fakeTimeoutErr
		}
	}

	// dlCh 为 nil 时(SetReadDeadline 已过期并清理),select 中 case <-nil 永远阻塞。
	// 若此时无其他可触发 case,Read 会永远阻塞。直接返回 i/o timeout 避免死锁。
	if dlCh == nil {
		return 0, fakeTimeoutErr
	}

	select {
	case data, ok := <-fc.inboundCh:
		if !ok {
			return 0, io.EOF
		}
		n := copy(b, data)
		if n < len(data) {
			fc.inboundBuf = append(fc.inboundBuf, data[n:]...)
		}
		return n, nil
	case <-timerC:
		return 0, fakeTimeoutErr
	case <-dlCh:
		// SetReadDeadline 触发的唤醒。判断 deadline 是否已过期。
		if !dlWhen.IsZero() && !time.Now().Before(dlWhen) {
			return 0, fakeTimeoutErr
		}
		// deadline 被重置到未来或清零:重新读取 deadline 状态并再次等待。
		// 仍持有 inboundMu,无需重新加锁(goto retry 在同一锁作用域内)。
		goto retry
	case <-fc.closeCh:
		return 0, io.EOF
	}
}

// Write 实现 io.Writer(被 sendRawMessage 的 conn.Write 调用)
// 将客户端写出的帧存入 outboundBuf,供测试侧 WaitClientWrite/ClientWrites 读取。
func (fc *FakeConn) Write(b []byte) (int, error) {
	fc.mu.Lock()
	closed := fc.closed
	fc.mu.Unlock()
	if closed {
		return 0, io.ErrClosedPipe
	}

	cp := make([]byte, len(b))
	copy(cp, b)

	fc.outboundMu.Lock()
	fc.outboundBuf = append(fc.outboundBuf, cp)
	fc.outboundCond.Signal()
	fc.outboundMu.Unlock()

	return len(b), nil
}

// Close 关闭连接,唤醒所有阻塞的 Read/Write
func (fc *FakeConn) Close() error {
	fc.mu.Lock()
	if fc.closed {
		fc.mu.Unlock()
		return nil
	}
	fc.closed = true
	fc.mu.Unlock()
	close(fc.closeCh)

	fc.outboundMu.Lock()
	fc.outboundCond.Broadcast()
	fc.outboundMu.Unlock()
	return nil
}

// LocalAddr 返回本地地址(测试占位)
func (fc *FakeConn) LocalAddr() net.Addr { return fakeAddr{} }

// RemoteAddr 返回远程地址(测试占位)
func (fc *FakeConn) RemoteAddr() net.Addr { return fakeAddr{} }

// SetDeadline 设置读写 deadline
func (fc *FakeConn) SetDeadline(t time.Time) error {
	fc.mu.Lock()
	fc.writeDeadline = t
	fc.mu.Unlock()
	return fc.SetReadDeadline(t)
}

// SetReadDeadline 设置读 deadline。
// 关键:通过关闭 readDeadlineCh 唤醒正在阻塞的 Read,使 FrameMux.Stop 的
// SetReadDeadline(now) 能强制中断 io.ReadFull(readLoop 依赖此机制退出)。
func (fc *FakeConn) SetReadDeadline(t time.Time) error {
	fc.readDeadlineMu.Lock()
	// 关闭旧信号 channel,唤醒阻塞在它上面的 Read
	if fc.readDeadlineCh != nil {
		close(fc.readDeadlineCh)
	}
	// 创建新信号 channel 并记录 deadline 时刻
	fc.readDeadlineCh = make(chan struct{})
	fc.readDeadlineWhen = t
	ch := fc.readDeadlineCh
	fc.readDeadlineMu.Unlock()

	// 若 deadline 在未来,启动定时器到点后关闭 channel 再次唤醒 Read
	if !t.IsZero() {
		if d := time.Until(t); d > 0 {
			time.AfterFunc(d, func() {
				fc.readDeadlineMu.Lock()
				defer fc.readDeadlineMu.Unlock()
				// 只关闭当前 channel(避免关闭已被替换的旧 channel)
				if fc.readDeadlineCh == ch {
					close(fc.readDeadlineCh)
					fc.readDeadlineCh = nil
				}
			})
		} else {
			// deadline 已过期(now 或过去):立即关闭新 channel,
			// 让 Read 在 goto retry 后重新获取 dlCh 时能立即触发并检测到过期
			fc.readDeadlineMu.Lock()
			if fc.readDeadlineCh == ch {
				close(fc.readDeadlineCh)
				fc.readDeadlineCh = nil
			}
			fc.readDeadlineMu.Unlock()
		}
	}
	// t 为零值:readDeadlineCh 已创建,Read 下一轮会因 dlWhen 为零而不等待 deadline
	return nil
}

// SetWriteDeadline 设置写 deadline
func (fc *FakeConn) SetWriteDeadline(t time.Time) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.writeDeadline = t
	return nil
}

// fakeAddr 测试用占位地址
type fakeAddr struct{}

func (fakeAddr) Network() string { return "fake" }
func (fakeAddr) String() string  { return "fake-conn" }

// MakeServerPayload 构造服务端帧的 payload(2字节LE长度前缀 + proto 数据)。
// 许多游戏消息(如 TransportRawNtf 内层的 PveGuildCityDataNtf)使用 ByteStream 格式,
// payload 前两个字节是小端长度前缀。此 helper 供测试构造合法 payload。
func MakeServerPayload(protoData []byte) []byte {
	payload := make([]byte, 2+len(protoData))
	binary.LittleEndian.PutUint16(payload[0:2], uint16(len(protoData)))
	copy(payload[2:], protoData)
	return payload
}
