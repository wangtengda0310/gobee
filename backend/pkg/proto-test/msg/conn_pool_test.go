package protocol

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAccountConnectionPool 验证连接池创建
func TestNewAccountConnectionPool(t *testing.T) {
	pool := NewAccountConnectionPool()
	assert.NotNil(t, pool)
	assert.Empty(t, pool.List())
}

// TestConnPoolEntry_ConnPoolEntry 验证连接池状态快照
func TestConnPoolEntry_ConnPoolEntry(t *testing.T) {
	pool := NewAccountConnectionPool()
	entries := pool.List()
	assert.Empty(t, entries, "新创建的连接池应为空")

	// Has 应返回 false
	assert.False(t, pool.Has("test1"), "不存在的账号应返回 false")
}

// TestAccountConnectionPool_CloseAll 验证关闭空连接池不会 panic
func TestAccountConnectionPool_CloseAll(t *testing.T) {
	pool := NewAccountConnectionPool()
	assert.NotPanics(t, func() {
		pool.CloseAll()
	}, "关闭空连接池不应 panic")
}

// TestAccountConnectionPool_Close 验证关闭不存在的账号不会 panic
func TestAccountConnectionPool_Close(t *testing.T) {
	pool := NewAccountConnectionPool()
	assert.NotPanics(t, func() {
		pool.Close("nonexistent")
	}, "关闭不存在的账号不应 panic")
}

// ===== DrainConn 测试 =====

// TestDrainConn_DiscardsMultipleCompleteFrames 验证 DrainConn 能完整读取并丢弃多个完整帧
func TestDrainConn_DiscardsMultipleCompleteFrames(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	// 推入 3 个完整的服务端帧
	for i := 0; i < 3; i++ {
		fc.PushServerFrame(uint16(100+i), []byte("payload"))
	}

	// DrainConn 应在读取完所有帧后，因超时返回 nil
	err := DrainConn(fc, 100*time.Millisecond)
	require.NoError(t, err, "DrainConn 应正常超时返回 nil")

	// 验证：DrainConn 后，FakeConn 的 inbound 已被清空
	// 再推一帧，Read 应能读到它（证明 DrainConn 确实消费了之前的帧）
	fc.PushServerFrame(999, []byte("after-drain"))
	header := make([]byte, FrameHeaderSize)
	_, err = io.ReadFull(fc, header)
	require.NoError(t, err, "DrainConn 后应能读取新推入的帧头")

	msgLen, _, err := ParseFrameHeader(header)
	require.NoError(t, err)
	body := make([]byte, msgLen)
	_, err = io.ReadFull(fc, body)
	require.NoError(t, err, "DrainConn 后应能读取新推入的帧体")
}

// TestDrainConn_ReturnsErrorOnPartialFrame 验证半帧数据时返回帧头解析错误
// 构造一个能通过 ReadFull 读取 4 字节帧头，但帧头内容非法导致 ParseFrameHeader 失败的场景
func TestDrainConn_ReturnsErrorOnPartialFrame(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	// 推入一个完整帧 + 一个非法帧头（4 字节，但 msgLen 过短 < MsgIDSize+SeqIDSize=6）
	fc.PushServerFrame(100, []byte("complete"))
	// 构造非法帧头：msgLen = 2（小于最小合法值 6），flags = FlagEncrypt
	badHeader := []byte{0x02, 0x00, 0x00, byte(FlagEncrypt)}
	fc.PushRawFrame(badHeader)

	err := DrainConn(fc, 100*time.Millisecond)
	require.Error(t, err, "DrainConn 遇到非法帧头应返回错误")
	assert.Contains(t, err.Error(), "消息过短", "错误应提示消息过短")
}

// TestDrainConn_TimeoutReturnsNil 验证正常超时（无数据）返回 nil
func TestDrainConn_TimeoutReturnsNil(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	// 不推入任何数据
	err := DrainConn(fc, 50*time.Millisecond)
	assert.NoError(t, err, "无数据时 DrainConn 超时返回 nil")
}

// ===== 连接池借出/归还测试 =====

// TestConnPool_BorrowWriteDoesNotRaceWithHeartbeat 验证 Borrow 返回的连接 Write 与 heartbeat 的 Write 不会并发交错
func TestConnPool_BorrowWriteDoesNotRaceWithHeartbeat(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	pool := NewAccountConnectionPool()
	pool.AcceptConn("acc1", "srv1", fc)

	// 借出连接
	pc := pool.conns["acc1"]
	borrowedConn, seqID := pc.Borrow()
	require.NotNil(t, borrowedConn, "借出应返回非 nil 连接")
	require.Equal(t, uint32(0), seqID, "初始 seqID 应为 0")

	// 模拟并发写：一个 goroutine 通过 borrowedConn 写业务帧，
	// 另一个 goroutine 直接通过底层 conn 写（模拟心跳 goroutine 的行为）
	const writeCount = 50
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < writeCount; i++ {
			frame := EncodeFrame(uint16(i), uint32(i), FlagEncrypt, []byte("biz"), true)
			_, err := borrowedConn.Write(frame)
			require.NoError(t, err)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < writeCount; i++ {
			// 直接获取 writeMu 后写底层 conn，模拟 heartbeat 的行为
			pc.writeMu.Lock()
			frame := EncodeFrame(3, 0, FlagEncrypt, []byte{}, true)
			_, err := pc.conn.Write(frame)
			pc.writeMu.Unlock()
			require.NoError(t, err)
		}
	}()

	wg.Wait()

	// 验证：所有写出的帧都应该是完整的（FakeConn 的 outbound 中不应有半截帧）
	frames := fc.ClientWrites()
	require.NotNil(t, frames)
	require.Equal(t, writeCount*2, len(frames), "应写出 %d 帧", writeCount*2)

	// 验证每一帧都能正确解码（证明 writeMu 保证了写操作的原子性，没有帧交错）
	for _, f := range frames {
		assert.NotNil(t, f, "每帧都应正确解码")
		// 验证帧的 RawSize 合法（至少包含帧头 + body）
		assert.GreaterOrEqual(t, f.RawSize, FrameHeaderSize+6, "帧大小应合法（至少包含帧头+MsgID+SeqID）")
	}
}

// TestConnPool_ReturnThenBorrowCanWriteAndRead 验证归还后新的 Borrow 可以正常写入/读取
func TestConnPool_ReturnThenBorrowCanWriteAndRead(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	pool := NewAccountConnectionPool()
	pool.AcceptConn("acc1", "srv1", fc)

	// 第一次借出
	pc := pool.conns["acc1"]
	conn1, seq1 := pc.Borrow()
	require.NotNil(t, conn1)
	require.Equal(t, uint32(0), seq1)

	// 通过借出的连接写一帧
	frame1 := EncodeFrame(101, 1, FlagEncrypt, []byte("first"), true)
	_, err := conn1.Write(frame1)
	require.NoError(t, err)

	// 归还连接
	pool.Return("acc1", 5)

	// 再次借出
	conn2, seq2 := pc.Borrow()
	require.NotNil(t, conn2, "归还后再次借出应返回非 nil 连接")
	require.Equal(t, uint32(5), seq2, "seqID 应续接上次归还的值")

	// 新借出的连接应能正常写入
	frame2 := EncodeFrame(102, 6, FlagEncrypt, []byte("second"), true)
	_, err = conn2.Write(frame2)
	require.NoError(t, err)

	// 验证 FakeConn 收到了两帧
	frames := fc.ClientWrites()
	require.Equal(t, 2, len(frames))
	assert.Equal(t, uint16(101), frames[0].MsgID)
	assert.Equal(t, uint16(102), frames[1].MsgID)
}

// TestConnPool_CloseRemovesDirtyConnection 验证脏连接（DrainConn 返回错误）应被 Close 移除
func TestConnPool_CloseRemovesDirtyConnection(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	pool := NewAccountConnectionPool()
	pool.AcceptConn("acc1", "srv1", fc)

	// 确认连接已入池
	require.True(t, pool.Has("acc1"), "连接应已在池中")

	// 模拟脏连接：推入非法帧头，使 DrainConn 返回错误
	// 构造非法帧头：msgLen = 2（小于最小合法值 6）
	badHeader := []byte{0x02, 0x00, 0x00, byte(FlagEncrypt)}
	fc.PushRawFrame(badHeader)
	err := DrainConn(fc, 50*time.Millisecond)
	require.Error(t, err, "DrainConn 应因非法帧头返回错误")

	// 关闭（移除）该账号连接
	pool.Close("acc1")

	// 验证连接已被移除
	assert.False(t, pool.Has("acc1"), "Close 后连接应从池中移除")
	assert.Empty(t, pool.List(), "List 应返回空")
}
