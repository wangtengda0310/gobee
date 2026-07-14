package protocol

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFakeConn_PushAndWait 验证基本数据流:
// 测试侧 PushServerFrame → FakeConn → FrameMux.readLoop 解码 → cache
// FrameMux.write(conn) → FakeConn outboundBuf → 测试侧 WaitClientWrite 读取
func TestFakeConn_PushAndWait(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")

	// 模拟服务器推送一帧
	fc.PushServerFrame(100, []byte("server-data"))

	// FrameMux readLoop 应读到并缓存
	frame, err := mux.WaitMsg(100, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, uint16(100), frame.MsgID)
	assert.Equal(t, []byte("server-data"), frame.Payload)

	mux.Stop()
}

// TestFakeConn_ClientWriteRoundTrip 验证客户端写出帧的回读
// 模拟 sendRawMessage:conn.Write(EncodeClientMessage) → 测试 WaitClientWrite 解码
func TestFakeConn_ClientWriteRoundTrip(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	// 模拟发送循环写出一条客户端帧(加密方向 isClientData=true)
	clientFrame := EncodeFrame(200, 5, FlagEncrypt, []byte("client-req"), true)
	n, err := fc.Write(clientFrame)
	require.NoError(t, err)
	assert.Equal(t, len(clientFrame), n)

	// 测试侧读回
	decoded, err := fc.WaitClientWrite(2 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, uint16(200), decoded.MsgID)
	assert.Equal(t, uint32(5), decoded.SeqID)
	assert.Equal(t, []byte("client-req"), decoded.Payload)
}

// TestFakeConn_InterleavedPushAndWrite 验证双向同时工作:
// 后台 FrameMux 读服务端推送,主 goroutine 写客户端消息并回读
func TestFakeConn_InterleavedPushAndWrite(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 场景:客户端先发一条 Req,服务器回一个 Ntf,客户端再发第二条 Req
	// 这正是工会战变量提取的典型时序

	// 1. 客户端发第一条
	req1 := EncodeFrame(200, 1, FlagEncrypt, []byte("req-1"), true)
	_, _ = fc.Write(req1)
	got1, err := fc.WaitClientWrite(1 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, []byte("req-1"), got1.Payload)

	// 2. 服务器推送 Ntf(变量来源)
	fc.PushServerFrame(100, []byte("ntf-data"))

	// 3. readLoop 缓存 Ntf
	ntf, err := mux.WaitMsg(100, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, []byte("ntf-data"), ntf.Payload)

	// 4. 客户端发第二条(可能携带从 Ntf 提取的变量)
	req2 := EncodeFrame(201, 2, FlagEncrypt, []byte("req-2"), true)
	_, _ = fc.Write(req2)
	got2, err := fc.WaitClientWrite(1 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, []byte("req-2"), got2.Payload)
}

// TestFakeConn_ReadDeadlineInterrupt 验证 SetReadDeadline 能中断阻塞的 Read
// FrameMux.Stop 依赖此机制强制中断 io.ReadFull
func TestFakeConn_ReadDeadlineInterrupt(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	// 不推送任何帧,Read 会阻塞
	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 1024)
		_, err := fc.Read(buf)
		done <- err
	}()

	// 设置短 deadline,Read 应在 deadline 后返回 timeout 错误
	_ = fc.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	select {
	case err := <-done:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	case <-time.After(2 * time.Second):
		t.Fatal("Read 未被 deadline 中断")
	}
}

// TestFakeConn_StopInterruptsRead 验证 FrameMux.Stop 能通过 SetReadDeadline(now) 中断 readLoop
// 这是 cleanup() 的关键路径:Stop → SetReadDeadline(now) → Read 返回 → readLoop 退出
func TestFakeConn_StopInterruptsRead(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")

	// 短暂等待 readLoop 进入阻塞 Read
	time.Sleep(50 * time.Millisecond)

	// Stop 设置 SetReadDeadline(now),readLoop 的 Read 应返回错误并退出
	stopDone := make(chan struct{})
	go func() {
		mux.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		// 正常:Stop 在合理时间内完成(Stop 内部 wg.Wait 等待 readLoop 退出)
	case <-time.After(3 * time.Second):
		t.Fatal("FrameMux.Stop 超时,readLoop 可能未被 SetReadDeadline 中断")
	}
}

// TestFakeConn_ConcurrentWriteRead 验证并发写(发送循环)+ 读(readLoop)的安全性
func TestFakeConn_ConcurrentWriteRead(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 并发:一个 goroutine 持续推服务端帧,另一个持续写客户端帧
	const numMessages = 20
	var wg sync.WaitGroup

	// 推服务端帧
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numMessages; i++ {
			fc.PushServerFrame(100, []byte{byte(i)})
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// 写客户端帧
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numMessages; i++ {
			_, _ = fc.Write(EncodeFrame(200, uint32(i), FlagEncrypt, []byte{byte(i)}, true))
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// 读回客户端帧,验证数量
	readCount := 0
	for i := 0; i < numMessages; i++ {
		if _, err := fc.WaitClientWrite(2 * time.Second); err != nil {
			t.Fatalf("第 %d 条客户端帧读取失败: %v", i, err)
		}
		readCount++
	}
	assert.Equal(t, numMessages, readCount)

	wg.Wait()
}

// TestFakeConn_CloseWakesBlockedRead 验证 Close 唤醒阻塞的 Read(返回 io.EOF)
func TestFakeConn_CloseWakesBlockedRead(t *testing.T) {
	fc := NewFakeConn()

	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 1024)
		_, err := fc.Read(buf)
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)
	_ = fc.Close()

	select {
	case err := <-done:
		assert.Equal(t, io.EOF, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Close 未唤醒阻塞的 Read")
	}
}

// TestFakeConn_PushRawFrame 验证 PushRawFrame 能推入预编码帧
func TestFakeConn_PushRawFrame(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{100})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 用 makeFrame(现有测试 helper)构造帧,再 PushRawFrame
	raw := makeFrame(100, 42, []byte("raw-payload"))
	fc.PushRawFrame(raw)

	frame, err := mux.WaitMsg(100, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, uint16(100), frame.MsgID)
	assert.Equal(t, uint32(42), frame.SeqID)
	assert.Equal(t, []byte("raw-payload"), frame.Payload)
}

// TestFakeConn_MakeServerPayload 验证 ByteStream 长度前缀构造
func TestFakeConn_MakeServerPayload(t *testing.T) {
	protoData := []byte{0x01, 0x02, 0x03, 0x04}
	payload := MakeServerPayload(protoData)
	assert.Equal(t, 6, len(payload))        // 2字节长度 + 4字节数据
	assert.Equal(t, uint8(4), payload[0])   // 低字节: 长度=4
	assert.Equal(t, uint8(0), payload[1])   // 高字节: 0
	assert.Equal(t, protoData, payload[2:]) // 原始数据
}
