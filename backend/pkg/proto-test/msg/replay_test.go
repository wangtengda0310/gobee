package protocol

import (
	"encoding/json"
	"testing"
	"time"

	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

// TestResolveVariablePayload_OpenID 验证 openid 账号级变量被替换为当前账号
func TestResolveVariablePayload_OpenID(t *testing.T) {
	msg := RecordMessage{
		MsgID:       1,
		MsgName:     "TestReq",
		PayloadJSON: `{"account":"placeholder","other":"keep"}`,
		FieldValues: map[string]FieldMetaValue{
			"account": {InputType: "variable", VariableName: "openid"},
		},
	}
	store := map[string]any{"openid": "test5"}

	resolved := resolveVariablePayload(msg, store)

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(resolved), &result))
	assert.Equal(t, "test5", result["account"])
	assert.Equal(t, "keep", result["other"])
}

// TestPrepareVariableContext_OpenIDOnly 验证仅 openid 变量时不创建 FrameMux，
// 但仍创建 variableStore 并预置当前账号
func TestPrepareVariableContext_OpenIDOnly(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	messages := []RecordMessage{
		{
			MsgID:       1,
			MsgName:     "TestReq",
			PayloadJSON: `{"account":"placeholder"}`,
			FieldValues: map[string]FieldMetaValue{
				"account": {InputType: "variable", VariableName: "openid"},
			},
		},
	}

	mux, stopReader, readerDone, store := prepareVariableContext(fc, messages, false, nil, "test3")

	assert.Nil(t, mux, "仅账号级变量不应创建 FrameMux")
	assert.NotNil(t, stopReader, "应启动 readDrainer")
	assert.NotNil(t, readerDone, "应返回 readerDone 通道")
	assert.NotNil(t, store, "应创建 variableStore")
	assert.Equal(t, "test3", store["openid"])

	// 给 readDrainer 一点时间启动并进入 io.ReadFull 阻塞
	time.Sleep(50 * time.Millisecond)

	// 模拟 cleanup() 行为：先发 stop 信号，再强制中断 io.ReadFull
	close(stopReader)
	_ = fc.SetReadDeadline(time.Now())

	// 等待 readDrainer 同步退出
	select {
	case <-readerDone:
		// 预期：readDrainer 已退出
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readDrainer 未在预期时间内退出")
	}

	// 清理：恢复 deadline 避免影响后续 Close
	_ = fc.SetReadDeadline(time.Time{})
	defer func() { _ = fc.Close() }()
}

// TestSendMessages_RangeInvalid 验证账号范围参数校验：rangeStart > rangeEnd 时立即返回错误
func TestSendMessages_RangeInvalid(t *testing.T) {
	// rangeStart=3, rangeEnd=1 是无效参数，应立即返回错误（无需实际网络连接）
	err := SendMessages("127.0.0.1:18000", "127.0.0.1:20144", "test", nil, 1, 3, 1, nil, nil, nil, nil)
	assert.Error(t, err, "rangeStart > rangeEnd 应返回错误")
	assert.Contains(t, err.Error(), "账号范围无效", "错误消息应包含'账号范围无效'")
	assert.Contains(t, err.Error(), "3", "错误消息应包含 rangeStart 值")
	assert.Contains(t, err.Error(), "1", "错误消息应包含 rangeEnd 值")
}

// TestSendMessages_RangeValid 验证合法的范围参数不会在参数校验阶段报错
// 注意：此测试会尝试实际连接服务器并超时，错误消息应来自网络层而非参数校验层
func TestSendMessages_RangeValid(t *testing.T) {
	// rangeStart=1, rangeEnd=1 是默认值，应通过参数校验
	err := SendMessages("127.0.0.1:18000", "127.0.0.1:20144", "test", nil, 1, 1, 1, nil, nil, nil, nil)
	// 参数校验通过，应进入网络连接阶段（连接被拒绝或超时）
	if err != nil {
		assert.NotContains(t, err.Error(), "账号范围无效", "rangeStart=1, rangeEnd=1 不应触发范围校验错误")
	}
}

// TestSendMessages_RangeEdge 验证边界情况
func TestSendMessages_RangeEdge(t *testing.T) {
	tests := []struct {
		name       string
		rangeStart int
		rangeEnd   int
		expectErr  bool
	}{
		{
			name:       "相等（单账号）",
			rangeStart: 1,
			rangeEnd:   1,
			expectErr:  false,
		},
		{
			name:       "rangeStart 小于 rangeEnd（多账号）",
			rangeStart: 1,
			rangeEnd:   5,
			expectErr:  false,
		},
		{
			name:       "rangeStart 大于 rangeEnd（无效）",
			rangeStart: 5,
			rangeEnd:   1,
			expectErr:  true,
		},
		{
			name:       "rangeStart 大于 rangeEnd（零与负数）",
			rangeStart: 0,
			rangeEnd:   -1,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendMessages("127.0.0.1:18000", "127.0.0.1:20144", "test", nil, 1, tt.rangeStart, tt.rangeEnd, nil, nil, nil, nil)
			if tt.expectErr {
				if err == nil {
					// 如果没报错，检查是否意外通过了参数校验
					t.Error("期望参数校验错误，但未返回错误")
				} else {
					assert.Contains(t, err.Error(), "账号范围无效", "错误应来自参数校验")
				}
			} else {
				// 参数校验应通过，错误来自网络层
				if err != nil {
					assert.NotContains(t, err.Error(), "账号范围无效", "参数校验应通过")
				}
			}
		})
	}
}

// TestReadDrainer_StopAndDeadlineExit 验证 close(stop) + SetReadDeadline(time.Now()) 后
// readDrainer 能在合理时间内退出并关闭 done channel。
// 这是 cleanup() 同步等待 readDrainer 退出的核心机制。
func TestReadDrainer_StopAndDeadlineExit(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	stop := make(chan struct{})
	done := make(chan struct{})

	go readDrainer(fc, stop, done, nil, "test1")

	// 给 readDrainer 一点时间启动并进入 io.ReadFull 阻塞
	time.Sleep(50 * time.Millisecond)

	// 模拟 cleanup() 的标准退出流程：先发 stop 信号，再强制中断 Read
	close(stop)
	_ = fc.SetReadDeadline(time.Now())

	// 验证 done 在合理时间内被关闭（500ms 是 cleanup() 的超时阈值）
	select {
	case <-done:
		// 预期：readDrainer 已同步退出
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readDrainer 未在预期时间内退出，cleanup 同步等待机制可能失效")
	}

	// 清理：恢复 deadline 避免影响后续 Close
	_ = fc.SetReadDeadline(time.Time{})
}

// TestReadDrainer_NoResidualAfterRestart 验证旧 reader 退出后，
// 新 reader 不会读到旧 reader 的残留字节。
// 这是连接池复用场景的关键：归还连接后下一个 borrower 必须看到干净的流。
func TestReadDrainer_NoResidualAfterRestart(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	// 构造一个合法的服务端帧（使用已注册的 MsgID=1004 HelloReq，proto 数据合法确保 payloadToJSON 成功）
	helloReq := &pb.HelloReq{Name: "test"}
	protoData, _ := proto.Marshal(helloReq)
	payload := MakeServerPayload(protoData)
	frame := fc.encodeServerFrame(helloReqMsgID, 0, payload)

	// 预推帧到 FakeConn（模拟连接池中积压的旧帧）
	fc.PushRawFrame(frame)

	// 启动第一个 readDrainer
	stop1 := make(chan struct{})
	done1 := make(chan struct{})
	go readDrainer(fc, stop1, done1, nil, "test1")

	// 给第一个 reader 一点时间读取预推的帧
	time.Sleep(100 * time.Millisecond)

	// 停止第一个 reader（模拟 cleanup）
	close(stop1)
	_ = fc.SetReadDeadline(time.Now())
	select {
	case <-done1:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("第一个 readDrainer 未退出")
	}
	_ = fc.SetReadDeadline(time.Time{})

	// 现在启动第二个 readDrainer（模拟下一个 borrower 从连接池复用）
	stop2 := make(chan struct{})
	done2 := make(chan struct{})
	var msgCount2 int
	go readDrainer(fc, stop2, done2, func(msgName string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) {
		msgCount2++
	}, "test2")

	// 给第二个 reader 启动时间
	time.Sleep(100 * time.Millisecond)

	// 第二个 reader 不应读到任何消息（因为第一个 reader 已消费完所有积压帧）
	assert.Equal(t, 0, msgCount2, "新 reader 不应读到旧 reader 的残留字节")

	// 清理第二个 reader
	close(stop2)
	_ = fc.SetReadDeadline(time.Now())
	select {
	case <-done2:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("第二个 readDrainer 未退出")
	}
	_ = fc.SetReadDeadline(time.Time{})
}

// TestPrepareVariableContext_OpenIDOnly_BorrowedFromPool 验证只有 openid 变量 +
// borrowedFromPool=true 时，prepareVariableContext 会先 DrainConn 再启动 readDrainer。
// 预推的帧应被 DrainConn 排空，readDrainer 不会读到它们。
func TestPrepareVariableContext_OpenIDOnly_BorrowedFromPool(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	// 构造合法的服务端帧（使用已注册的 MsgID=1004 HelloReq，proto 数据合法）
	helloReq := &pb.HelloReq{Name: "test"}
	protoData, _ := proto.Marshal(helloReq)
	payload := MakeServerPayload(protoData)
	frame := fc.encodeServerFrame(helloReqMsgID, 0, payload)

	// 在调用 prepareVariableContext 前预推帧（模拟连接池中积压的旧帧）
	fc.PushRawFrame(frame)
	fc.PushRawFrame(frame)

	messages := []RecordMessage{
		{
			MsgID:       1,
			MsgName:     "TestReq",
			PayloadJSON: `{"account":"placeholder"}`,
			FieldValues: map[string]FieldMetaValue{
				"account": {InputType: "variable", VariableName: "openid"},
			},
		},
	}

	var msgCount int
	onMessage := ReplayMessageCallback(func(msgName string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) {
		msgCount++
	})

	// borrowedFromPool=true：应先 DrainConn 排空积压帧，再启动 readDrainer
	mux, stopReader, readerDone, store := prepareVariableContext(fc, messages, true, onMessage, "test4")

	assert.Nil(t, mux, "仅 openid 变量不应创建 FrameMux")
	assert.NotNil(t, stopReader, "应启动 readDrainer")
	assert.NotNil(t, readerDone, "应返回 readerDone 通道")
	assert.NotNil(t, store, "应创建 variableStore")
	assert.Equal(t, "test4", store["openid"])

	// 给 readDrainer 启动时间
	time.Sleep(100 * time.Millisecond)

	// 关键断言：readDrainer 不应读到任何预推的帧（因为 DrainConn 已排空）
	assert.Equal(t, 0, msgCount, "readDrainer 不应读到 DrainConn 已排空的旧帧")

	// 模拟 cleanup 退出
	close(stopReader)
	_ = fc.SetReadDeadline(time.Now())
	select {
	case <-readerDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readDrainer 未退出")
	}
	_ = fc.SetReadDeadline(time.Time{})
}

// TestPrepareVariableContext_OpenIDOnly_NotBorrowed 验证只有 openid 变量 +
// borrowedFromPool=false 时，不会 DrainConn，直接启动 readDrainer。
// 预推的帧会被 readDrainer 消费（从 inbound 中读取并丢弃）。
func TestPrepareVariableContext_OpenIDOnly_NotBorrowed(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	// 构造合法的服务端帧（使用已注册的 MsgID=1004 HelloReq，proto 数据合法）
	helloReq := &pb.HelloReq{Name: "test"}
	protoData, _ := proto.Marshal(helloReq)
	payload := MakeServerPayload(protoData)
	frame := fc.encodeServerFrame(helloReqMsgID, 0, payload)

	// 预推帧
	fc.PushRawFrame(frame)

	messages := []RecordMessage{
		{
			MsgID:       1,
			MsgName:     "TestReq",
			PayloadJSON: `{"account":"placeholder"}`,
			FieldValues: map[string]FieldMetaValue{
				"account": {InputType: "variable", VariableName: "openid"},
			},
		},
	}

	// borrowedFromPool=false：新连接，不应 DrainConn，直接启动 readDrainer
	mux, stopReader, readerDone, store := prepareVariableContext(fc, messages, false, nil, "test5")

	assert.Nil(t, mux, "仅 openid 变量不应创建 FrameMux")
	assert.NotNil(t, stopReader, "应启动 readDrainer")
	assert.NotNil(t, readerDone, "应返回 readerDone 通道")
	assert.NotNil(t, store, "应创建 variableStore")
	assert.Equal(t, "test5", store["openid"])

	// 给 readDrainer 启动和消费时间
	time.Sleep(100 * time.Millisecond)

	// 关键验证：停止旧 reader
	close(stopReader)
	_ = fc.SetReadDeadline(time.Now())
	select {
	case <-readerDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("readDrainer 未退出")
	}
	_ = fc.SetReadDeadline(time.Time{})

	// 再推一帧，验证连接仍然可用（旧 reader 已退出，不会竞争读取）
	fc.PushRawFrame(frame)

	// 启动新 reader，验证它能读到新帧（证明旧 reader 已消费了预推帧且已退出）
	stop2 := make(chan struct{})
	done2 := make(chan struct{})
	var msgCount2 int
	go readDrainer(fc, stop2, done2, func(msgName string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) {
		msgCount2++
	}, "test5-new")

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, msgCount2, "新 reader 应能读到新推入的帧，证明旧 reader 已退出且连接可用")

	close(stop2)
	_ = fc.SetReadDeadline(time.Now())
	select {
	case <-done2:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("第二个 readDrainer 未退出")
	}
	_ = fc.SetReadDeadline(time.Time{})
}
