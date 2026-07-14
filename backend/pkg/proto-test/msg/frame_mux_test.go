package protocol

import (
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// makeFrame 构造一个完整的协议帧（帧头 + 加密 body）
// msgID: 消息 ID, seqID: 序列号, payload: 消息负载
// 使用 FlagEncrypt 编码，模拟服务端→客户端方向
func makeFrame(msgID uint16, seqID uint32, payload []byte) []byte {
	return EncodeFrame(msgID, seqID, FlagEncrypt, payload, false)
}

// writeFramesToConn 将多个帧写入连接的写端
func writeFramesToConn(w net.Conn, frames ...[]byte) error {
	for _, f := range frames {
		if _, err := w.Write(f); err != nil {
			return err
		}
	}
	return nil
}

// TestDrainAndStart 验证 drain 阶段丢弃所有积压帧(不缓存 watchedID 帧)
// 新契约(2026-06-15):drain 只丢弃积压数据,不缓存。
// 原因:池复用连接的积压帧属于上次会话残留,缓存会让惰性变量提取读到过期数据。
func TestDrainAndStart(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	watchedIDs := []uint16{100, 200}
	mux := NewFrameMux(server, watchedIDs)

	// 构造 3 帧：MsgID=100(watched), MsgID=300(非 watched), MsgID=200(watched)
	frame1 := makeFrame(100, 1, []byte("payload-100"))
	frame2 := makeFrame(300, 2, []byte("payload-300"))
	frame3 := makeFrame(200, 3, []byte("payload-200"))

	// 在 goroutine 中写入数据（drain 阶段会设置 deadline）
	go func() {
		// 给 drain 一些时间准备好
		time.Sleep(10 * time.Millisecond)
		_ = writeFramesToConn(client, frame1, frame2, frame3)
		// 保持连接打开一段时间让 drain 消费
		time.Sleep(200 * time.Millisecond)
	}()

	// DrainAndStart 用较长超时确保能读到所有帧
	mux.DrainAndStart(500*time.Millisecond, nil, "test-account")

	// 新契约:drain 丢弃所有积压帧,cache 应为空(watchedID 帧也不缓存)
	_, ok := mux.GetCache(100)
	assert.False(t, ok, "drain 不应缓存 MsgID=100(积压帧属于上次会话残留)")

	_, ok = mux.GetCache(200)
	assert.False(t, ok, "drain 不应缓存 MsgID=200")

	_, ok = mux.GetCache(300)
	assert.False(t, ok, "MsgID=300 不在 watchedIDs 也不应缓存")

	mux.Stop()
}

// TestWaitMsgCacheHit 验证 cache 已有帧时 WaitMsg 立即返回
func TestWaitMsgCacheHit(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 手动写入 cache
	frame := &DecodedFrame{MsgID: 100, SeqID: 1, Payload: []byte("cached")}
	mux.mu.Lock()
	mux.cache[100] = frame
	mux.mu.Unlock()

	// WaitMsg 应立即返回（不阻塞）
	result, err := mux.WaitMsg(100, 2*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint16(100), result.MsgID)
	assert.Equal(t, uint32(1), result.SeqID)
}

// TestWaitMsgBlocking 验证 cache 空时 WaitMsg 阻塞等待 readLoop 写入帧后被唤醒
func TestWaitMsgBlocking(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	watchedIDs := []uint16{100}
	mux := NewFrameMux(server, watchedIDs)

	// 不调用 DrainAndStart，直接启动 readLoop（无积压帧）
	// readLoop 会阻塞在 io.ReadFull 等待数据
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 启动 WaitMsg goroutine
	waitDone := make(chan *DecodedFrame, 1)
	waitErr := make(chan error, 1)
	go func() {
		f, err := mux.WaitMsg(100, 5*time.Second)
		if err != nil {
			waitErr <- err
		} else {
			waitDone <- f
		}
	}()

	// 短暂等待确保 WaitMsg 已进入等待状态
	time.Sleep(50 * time.Millisecond)

	// 写入 watchedID 帧
	frame := makeFrame(100, 42, []byte("notify-payload"))
	_ = writeFramesToConn(client, frame)

	// 验证 WaitMsg 被唤醒并返回正确帧
	select {
	case f := <-waitDone:
		assert.NotNil(t, f)
		assert.Equal(t, uint16(100), f.MsgID)
		assert.Equal(t, uint32(42), f.SeqID)
	case err := <-waitErr:
		t.Fatalf("WaitMsg 不应返回错误: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("WaitMsg 超时未返回")
	}

	mux.Stop()
}

// TestWaitMsgTimeout 验证 cache 空且无帧写入时 WaitMsg 超时返回错误
func TestWaitMsgTimeout(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 启动 readLoop（会阻塞读取，但不写入任何帧）
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	start := time.Now()
	_, err := mux.WaitMsg(100, 500*time.Millisecond)
	elapsed := time.Since(start)

	assert.Error(t, err, "WaitMsg 应返回超时错误")
	assert.Contains(t, err.Error(), "超时")
	// 验证超时时间大致正确（允许 50% 误差）
	assert.True(t, elapsed >= 400*time.Millisecond, "实际等待时间 %v 应接近超时 500ms", elapsed)
	assert.True(t, elapsed < 1*time.Second, "实际等待时间 %v 不应远超超时", elapsed)

	mux.Stop()
}

// TestStopExitsReadLoop 验证 Stop() 后 readLoop 退出（wg.Wait 不超过 1 秒）
func TestStopExitsReadLoop(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 启动 readLoop（阻塞在 io.ReadFull）
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 短暂等待确保 readLoop 已启动
	time.Sleep(50 * time.Millisecond)

	// Stop 应在合理时间内退出
	done := make(chan struct{})
	go func() {
		mux.Stop()
		close(done)
	}()

	select {
	case <-done:
		// 正常退出
	case <-time.After(1 * time.Second):
		t.Fatal("Stop() 超过 1 秒未返回，readLoop 可能未退出")
	}
}

// TestConcurrentAccess 验证并发 readLoop 写 + WaitMsg 读的安全性
func TestConcurrentAccess(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	watchedIDs := []uint16{100, 200, 300}
	mux := NewFrameMux(server, watchedIDs)

	// 启动 readLoop
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 并发写入多个 watchedID 帧
	const numFrames = 50
	go func() {
		time.Sleep(20 * time.Millisecond)
		for i := 0; i < numFrames; i++ {
			// 只发送 watchedIDs 中的帧
			actualID := watchedIDs[i%3]
			frame := makeFrame(actualID, uint32(i), []byte("concurrent"))
			_, _ = client.Write(frame)
			time.Sleep(5 * time.Millisecond) // 控制写入速率
		}
	}()

	// 并发 WaitMsg
	var wg sync.WaitGroup
	var successCount int32
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			msgID := watchedIDs[idx]
			f, err := mux.WaitMsg(msgID, 3*time.Second)
			if err == nil && f != nil && f.MsgID == msgID {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}
	wg.Wait()

	// 至少部分 WaitMsg 应成功
	assert.True(t, atomic.LoadInt32(&successCount) > 0, "至少部分 WaitMsg 应成功获取帧")

	mux.Stop()
}

// TestChannelFullDrop 验证 frames channel 满时 readLoop 不阻塞
// 通过创建小缓冲 channel 的 FrameMux，写入超过缓冲的帧数，验证不阻塞
func TestChannelFullDrop(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{})

	// 手动替换 frames channel 为小缓冲（模拟满场景）
	// 先读取原始 channel 的引用
	mux.mu.Lock()
	// 创建一个只有 2 容量的新 channel 来替换
	smallFrames := make(chan *DecodedFrame, 2)
	// 填满 channel
	smallFrames <- &DecodedFrame{MsgID: 1, SeqID: 1, Payload: []byte("fill-1")}
	smallFrames <- &DecodedFrame{MsgID: 2, SeqID: 2, Payload: []byte("fill-2")}
	mux.frames = smallFrames
	mux.mu.Unlock()

	// 启动 readLoop
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 写入 5 帧，readLoop 不应阻塞
	frames := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		frames[i] = makeFrame(uint16(1000+i), uint32(i), []byte("overflow"))
	}

	// 用带超时的写入验证 readLoop 消费不阻塞
	done := make(chan error, 1)
	go func() {
		// 短暂延迟让 readLoop 启动
		time.Sleep(30 * time.Millisecond)
		for _, f := range frames {
			if _, err := client.Write(f); err != nil {
				done <- err
				return
			}
			time.Sleep(10 * time.Millisecond) // 帧间间隔
		}
		close(done) // 写入完成
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("写入帧失败: %v", err)
		}
		// 成功：readLoop 消费了帧并丢弃了多余的（channel 满）
	case <-time.After(3 * time.Second):
		t.Fatal("写入超时，readLoop 可能阻塞了")
	}

	mux.Stop()
}

// TestMultipleStop 验证多次调用 Stop 不 panic
func TestMultipleStop(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 不启动 readLoop，直接多次 Stop
	assert.NotPanics(t, func() {
		mux.Stop()
		mux.Stop()
		mux.Stop()
	}, "多次 Stop 不应 panic")
}

// TestStopWakesWaitMsg 验证 Stop 后 WaitMsg 不永久阻塞
// WaitMsg 使用较长超时（5s），Stop 在 100ms 后调用
// WaitMsg 应因 done channel 关闭而立即返回错误
func TestStopWakesWaitMsg(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 启动 readLoop
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// WaitMsg 在另一个 goroutine 中等待
	waitDone := make(chan error, 1)
	go func() {
		_, err := mux.WaitMsg(100, 5*time.Second) // 长超时
		waitDone <- err
	}()

	// 确保 WaitMsg 已进入等待
	time.Sleep(50 * time.Millisecond)

	// Stop 应关闭 done channel → WaitMsg 的 select 收到 done 信号
	go func() {
		time.Sleep(100 * time.Millisecond)
		mux.Stop()
	}()

	select {
	case err := <-waitDone:
		// WaitMsg 应返回 "已停止" 错误
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "停止")
	case <-time.After(3 * time.Second):
		t.Fatal("WaitMsg 在 Stop 后仍未返回")
	}
}

// TestDrainDecodingError 验证 drain 阶段解码错误不会终止 drain
// 新契约(2026-06-15):drain 丢弃所有积压帧(含 valid 帧),不缓存
func TestDrainDecodingError(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 构造一个有效的帧后跟一个无效帧（body 太短）
	validFrame := makeFrame(100, 1, []byte("valid"))
	// 构造无效帧：帧头声明 body 长度 10，但只包含 6 字节 body（MsgID+SeqID 无 payload 是合法的）
	// 更好的方式：构造一个 body 小于 MsgIDSize+SeqIDSize 的帧
	invalidBody := []byte{0x01, 0x00} // 只有 2 字节，不足 MsgIDSize+SeqIDSize=6
	invalidFrame := make([]byte, FrameHeaderSize+len(invalidBody))
	invalidFrame[0] = byte(len(invalidBody))
	invalidFrame[1] = 0
	invalidFrame[2] = 0
	invalidFrame[3] = byte(FlagEncrypt)
	copy(invalidFrame[FrameHeaderSize:], invalidBody)

	// 在 valid 后面加一个 valid 来证明 drain 没有因 invalid 退出
	validFrame2 := makeFrame(100, 2, []byte("valid2"))

	go func() {
		time.Sleep(10 * time.Millisecond)
		// 写入 valid → invalid → valid，证明 invalid 不终止 drain
		// 注意：由于 XOR 加密是在 body 上操作的，invalidBody 也会被加密
		// 所以 invalid 帧的 DecodeFrame 会因 body 太短而失败
		_ = writeFramesToConn(client, validFrame, invalidFrame, validFrame2)
		time.Sleep(200 * time.Millisecond)
	}()

	mux.DrainAndStart(500*time.Millisecond, nil, "test")

	// 新契约:drain 丢弃所有积压帧,valid 帧也不应缓存
	_, ok := mux.GetCache(100)
	assert.False(t, ok, "drain 不应缓存 valid 帧(积压帧属于上次会话残留)")

	mux.Stop()
}

// TestFramesChannel 验证 frames channel 接收实时帧
func TestFramesChannel(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{})

	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 写入帧
	frame := makeFrame(1000, 1, []byte("stream"))
	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = client.Write(frame)
	}()

	// 从 frames channel 读取
	select {
	case f := <-mux.Frames():
		assert.NotNil(t, f)
		assert.Equal(t, uint16(1000), f.MsgID)
	case <-time.After(2 * time.Second):
		t.Fatal("frames channel 未收到帧")
	}

	mux.Stop()
}

// TestReadLoopContinuesOnError 验证 DecodeFrame 错误时 readLoop 继续读下一帧
// 构造一个通过 ParseFrameHeader 但 DecodeFrame 失败的帧（压缩标记但 body 是垃圾数据）
func TestReadLoopContinuesOnError(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 构造坏帧：帧头正确，msgLen=10 >= 6 (MsgIDSize+SeqIDSize)，flags 带压缩标记
	// body 是 10 字节垃圾数据 → Snappy 解压失败 → DecodeFrame 返回 error
	// readLoop continue 跳过（body 已完整读取，下一帧 header 位置正确）
	const badMsgLen = 10
	badBody := make([]byte, badMsgLen)
	for i := range badBody {
		badBody[i] = 0xFF // 无效的 Snappy 数据
	}
	badFrame := make([]byte, FrameHeaderSize+badMsgLen)
	badFrame[0] = byte(badMsgLen)
	badFrame[1] = 0
	badFrame[2] = 0
	badFrame[3] = byte(FlagCompress) // 标记为压缩，触发 Snappy 解压失败
	copy(badFrame[FrameHeaderSize:], badBody)

	// 构造有效帧（watched）
	goodFrame := makeFrame(100, 1, []byte("after-error"))

	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = client.Write(badFrame)
		time.Sleep(30 * time.Millisecond) // 确保 bad 帧被 readLoop 完整消费
		_, _ = client.Write(goodFrame)
	}()

	// 验证 bad 帧后 readLoop 继续读到了 good 帧
	f, err := mux.WaitMsg(100, 3*time.Second)
	assert.NoError(t, err, "bad 帧后应继续读到 good 帧")
	assert.NotNil(t, f)
	assert.Equal(t, uint16(100), f.MsgID)

	mux.Stop()
}

// TestEncodeDecodeRoundTrip 验证 EncodeFrame/DecodeFrame 往返一致性
// 辅助测试：确保 makeFrame 构造的帧能被正确解码
func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		msgID   uint16
		seqID   uint32
		payload []byte
	}{
		{"空 payload", 100, 1, []byte{}},
		{"短 payload", 200, 42, []byte("hello")},
		{"长 payload", 300, 999, make([]byte, 256)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := makeFrame(tt.msgID, tt.seqID, tt.payload)
			decoded, err := DecodeFrame(encoded, false) // 服务端→客户端
			assert.NoError(t, err)
			assert.Equal(t, tt.msgID, decoded.MsgID)
			assert.Equal(t, tt.seqID, decoded.SeqID)
			assert.Equal(t, tt.payload, decoded.Payload)
		})
	}
}

// TestDrainAndStartWithPipeEndClose 验证 pipe 写端关闭后 drain 正常退出
// 新契约(2026-06-15):drain 丢弃积压帧,不缓存
func TestDrainAndStartWithPipeEndClose(t *testing.T) {
	client, server := net.Pipe()

	mux := NewFrameMux(server, []uint16{100})

	// 在 goroutine 中写入帧后关闭
	go func() {
		frame := makeFrame(100, 1, []byte("before-close"))
		_, _ = client.Write(frame)
		time.Sleep(50 * time.Millisecond) // 等 drain 消费
		_ = client.Close()
	}()

	// drain 应能读到帧并正常退出(不 panic)
	mux.DrainAndStart(500*time.Millisecond, nil, "test")

	// 新契约:drain 丢弃积压帧,不应缓存
	_, ok := mux.GetCache(100)
	assert.False(t, ok, "drain 不应缓存积压帧")

	mux.Stop()
	_ = server.Close()
}

// TestWaitMsgReturnsLatestCache 验证 WaitMsg 返回 cache 中最新的帧（覆盖式缓存）
func TestWaitMsgReturnsLatestCache(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	watchedIDs := []uint16{100}
	mux := NewFrameMux(server, watchedIDs)

	// 启动 readLoop
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 连续写入两个同 MsgID 的帧
	frame1 := makeFrame(100, 1, []byte("first"))
	frame2 := makeFrame(100, 2, []byte("second"))
	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = client.Write(frame1)
		time.Sleep(20 * time.Millisecond)
		_, _ = client.Write(frame2)
	}()

	// WaitMsg 应返回第二个（最新）帧
	f, err := mux.WaitMsg(100, 3*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, f)
	// 注意：由于 WaitMsg 可能读到第一个帧就返回，
	// 但 cache 会被第二个帧覆盖，所以验证最终 cache 状态
	time.Sleep(50 * time.Millisecond)
	cached, ok := mux.GetCache(100)
	assert.True(t, ok)
	assert.Equal(t, uint32(2), cached.SeqID, "cache 应为最新帧")

	mux.Stop()
}

// TestLargePayload 验证较大 payload 的帧正常处理
func TestLargePayload(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 构造 1KB payload
	largePayload := make([]byte, 1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}
	frame := makeFrame(100, 1, largePayload)

	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = client.Write(frame)
	}()

	f, err := mux.WaitMsg(100, 3*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, largePayload, f.Payload)

	mux.Stop()
}

// TestParseFrameHeaderInDrain 验证 drain 阶段帧头解析错误能正常退出 drain
func TestParseFrameHeaderInDrain(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	// 写入一个帧头声明超大长度的帧
	badHeader := make([]byte, FrameHeaderSize)
	// 设置 msgLen = MaxPacketSize + 1
	binary.LittleEndian.PutUint16(badHeader[0:2], 0) // 低 2 字节
	badHeader[2] = 1                                 // 高字节 = 1，所以 msgLen = 65536 > MaxPacketSize
	badHeader[3] = byte(FlagEncrypt)

	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = client.Write(badHeader)
	}()

	// drain 应正常退出（不会 panic）
	mux.DrainAndStart(500*time.Millisecond, nil, "test")
	mux.Stop()
}

// TestReadLoopExitOnConnClose 验证连接关闭后 readLoop 正常退出
func TestReadLoopExitOnConnClose(t *testing.T) {
	client, server := net.Pipe()

	mux := NewFrameMux(server, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	time.Sleep(30 * time.Millisecond)

	// 关闭客户端写端 → server 端 ReadFull 会返回 io.EOF
	_ = client.Close()

	// readLoop 应在合理时间内退出
	exited := make(chan struct{})
	go func() {
		mux.wg.Wait()
		close(exited)
	}()

	select {
	case <-exited:
		// 正常
	case <-time.After(2 * time.Second):
		t.Fatal("readLoop 在连接关闭后未退出")
	}

	_ = server.Close()
}

// TestFramesChannelOrdering 验证 frames channel 保持帧的顺序
func TestFramesChannelOrdering(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	const numFrames = 5
	frames := make([][]byte, numFrames)
	for i := 0; i < numFrames; i++ {
		frames[i] = makeFrame(uint16(1000+i), uint32(i), []byte{byte(i)})
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		for _, f := range frames {
			_, _ = client.Write(f)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// 按顺序读取
	for i := 0; i < numFrames; i++ {
		select {
		case f := <-mux.Frames():
			assert.Equal(t, uint16(1000+i), f.MsgID, "帧顺序应保持一致")
			assert.Equal(t, uint32(i), f.SeqID)
		case <-time.After(2 * time.Second):
			t.Fatalf("等待第 %d 帧超时", i)
		}
	}

	mux.Stop()
}

// TestGetCacheEmpty 验证空 cache 返回 false
func TestGetCacheEmpty(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})

	_, ok := mux.GetCache(100)
	assert.False(t, ok, "空 cache 应返回 false")

	_, ok = mux.GetCache(999)
	assert.False(t, ok, "未注册的 MsgID 应返回 false")
}

// TestReadLoopBodyReadError 验证 body 读取失败时 readLoop 退出
func TestReadLoopBodyReadError(t *testing.T) {
	client, server := net.Pipe()
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	mux := NewFrameMux(server, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test-account")

	// 写一个声称 body 长度为 100 的帧头，但只写 10 字节就关闭
	header := make([]byte, FrameHeaderSize)
	header[0] = 100 // msgLen = 100
	header[1] = 0
	header[2] = 0
	header[3] = byte(FlagEncrypt)

	go func() {
		time.Sleep(20 * time.Millisecond)
		_, _ = client.Write(header)
		// 只写 10 字节 body，不够 100 字节
		partial := make([]byte, 10)
		_, _ = client.Write(partial)
		// 关闭连接，触发 io.ReadFull 返回 unexpected EOF
		_ = client.Close()
	}()

	// readLoop 应退出
	done := make(chan struct{})
	go func() {
		mux.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常退出
	case <-time.After(2 * time.Second):
		t.Fatal("readLoop 在 body 读取失败后未退出")
	}
}
