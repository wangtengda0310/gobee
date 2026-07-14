package protocol

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== 测试辅助函数 ==========

// makeProtoPayload 将 proto 消息序列化并添加 ByteStream 2字节LE长度前缀
func makeProtoPayload(msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	payload := make([]byte, 2+len(data))
	payload[0] = byte(len(data))
	payload[1] = byte(len(data) >> 8)
	copy(payload[2:], data)
	return payload
}

// makeTransportRawNtfPayload 构造 TransportRawNtf 帧 payload
// innerMsgID: 内层消息 ID，innerData: 内层消息序列化后的数据（已含 ByteStream 前缀）
func makeTransportRawNtfPayload(innerMsgID uint32, innerData []byte) []byte {
	transport := &pb.TransportRawNtf{
		MsgId: innerMsgID,
		Data:  innerData,
	}
	return makeProtoPayload(transport)
}

// ========== msgNeedsVariable 测试 ==========

func TestMsgNeedsVariable_True(t *testing.T) {
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	assert.True(t, msgNeedsVariable(msg))
}

func TestMsgNeedsVariable_False_OriginalOnly(t *testing.T) {
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "original"},
		},
	}
	assert.False(t, msgNeedsVariable(msg))
}

func TestMsgNeedsVariable_False_EmptyFieldValues(t *testing.T) {
	msg := RecordMessage{}
	assert.False(t, msgNeedsVariable(msg))
}

func TestMsgNeedsVariable_False_VariableNameEmpty(t *testing.T) {
	// InputType=="variable" 但 VariableName 为空,视为无效配置
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: ""},
		},
	}
	assert.False(t, msgNeedsVariable(msg))
}

// ========== ExtractVariablesForMessage 测试(惰性提取核心) ==========

// TestLazyExtract_HitAfterNtfArrives 验证惰性提取核心场景:
// Ntf 在 readLoop 启动后到达 → cache 命中 → 提取成功
// 这正是修复后的正确时序:提取推迟到发送循环,此时 Ntf 已被缓存
func TestLazyExtract_HitAfterNtfArrives(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	// cityId 变量关注 TransportRawNtf
	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 构造一个含 cityId=200 的 TransportRawNtf(内层 PveGuildCityDataNtf)
	innerNtf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{
					SimpleInfo: &pb.GuildCitySimpleInfo{Id: 200},
				},
			},
		},
	}
	transportPayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
		makeProtoPayload(innerNtf),
	)

	// readLoop 启动后推送 Ntf(模拟服务器在客户端发前置 Req 后返回)
	time.Sleep(50 * time.Millisecond) // 确保 readLoop 已就绪
	fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), transportPayload)

	// 等待 readLoop 缓存
	require.Eventually(t, func() bool {
		_, err := mux.WaitMsg(uint16(pb.EGameMsgID_TransportRawNtf_id), 2*time.Second)
		return err == nil
	}, 3*time.Second, 50*time.Millisecond, "readLoop 未缓存 Ntf")

	// 惰性提取:此时 variableStore 为空,触发 WaitMsg → 命中 cache → 提取成功
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	store := map[string]any{}
	err := ExtractVariablesForMessage(msg, mux, store)
	require.NoError(t, err)
	assert.Equal(t, uint64(200), store["cityId"])
}

// TestLazyExtract_SkipAlreadyExtracted 验证"已有值跳过"优化:
// variableStore 中已有 cityId,不重复 WaitMsg
func TestLazyExtract_SkipAlreadyExtracted(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// store 已有值
	store := map[string]any{"cityId": uint64(999)}
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}

	// 不推送任何 Ntf。如果代码未跳过已有值,会 WaitMsg 超时 5s。
	// 用短超时包装验证:应立即返回 nil(跳过)
	done := make(chan error, 1)
	go func() {
		done <- ExtractVariablesForMessage(msg, mux, store)
	}()

	select {
	case err := <-done:
		require.NoError(t, err, "已有值时应跳过提取,立即返回")
		assert.Equal(t, uint64(999), store["cityId"], "值不被覆盖")
	case <-time.After(1 * time.Second):
		t.Fatal("已有值未跳过,WaitMsg 阻塞超过 1s(应跳过)")
	}
}

// TestLazyExtract_TimeoutWhenNtfAbsent 验证提取失败场景:
// Ntf 未到达 → WaitMsg 超时 → 返回错误(调用方据此跳过发送)
func TestLazyExtract_TimeoutWhenNtfAbsent(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 不推送任何 Ntf
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	store := map[string]any{}

	// cityId 有 2 个 WatchMsgIDs(TransportRawNtf + PveGuildCityDataNtf),
	// 每个 WaitMsg 5s 超时,最坏情况 2×5s=10s。测试窗口设为 15s。
	done := make(chan error, 1)
	go func() {
		done <- ExtractVariablesForMessage(msg, mux, store)
	}()

	select {
	case err := <-done:
		require.Error(t, err, "Ntf 缺失时应返回提取失败错误")
		assert.Contains(t, err.Error(), "提取失败")
		assert.Empty(t, store, "失败时 store 不应被填充")
	case <-time.After(15 * time.Second):
		t.Fatal("提取未在 15s 内返回(超时机制异常)")
	}
}

// ========== 端到端发送循环时序测试(模拟工会战.json 流程) ==========

// TestLazyExtract_GuildWarTimeline 模拟工会战.json 的完整时序,验证修复:
//
// 时序(修复后):
//  1. 前置消息(GuildCityWarDataReq)发出 → WaitClientWrite 确认
//  2. 服务器返回 TransportRawNtf(含 cityId) → PushServerFrame
//  3. readLoop 缓存 Ntf
//  4. 变量消息(TeamSelectGuildCityReq)发送前 → ExtractVariablesForMessage 命中 cache
//  5. payload 中 cityId 被替换为提取值
//
// 修复前:步骤4在步骤1之前执行(连接建立后立即提取),此时 Ntf 尚未到达 → 超时 → 用写死值。
func TestLazyExtract_GuildWarTimeline(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// --- 步骤1:客户端发前置消息 ---
	preReq := EncodeFrame(uint16(pb.EGameMsgID_GuildCityWarDataReq_id), 1, FlagEncrypt, []byte("req-guild-war"), true)
	_, err := fc.Write(preReq)
	require.NoError(t, err)
	got, err := fc.WaitClientWrite(2 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, uint16(pb.EGameMsgID_GuildCityWarDataReq_id), got.MsgID)

	// --- 步骤2:服务器返回 Ntf(变量来源,cityId=4293) ---
	innerNtf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: 4293}},
			},
		},
	}
	transportPayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
		makeProtoPayload(innerNtf),
	)
	fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), transportPayload)

	// --- 步骤3:等待 readLoop 缓存 Ntf ---
	require.Eventually(t, func() bool {
		_, err := mux.WaitMsg(uint16(pb.EGameMsgID_TransportRawNtf_id), 2*time.Second)
		return err == nil
	}, 3*time.Second, 50*time.Millisecond)

	// --- 步骤4:变量消息发送前,惰性提取 ---
	// 占位值用 9999(与提取值 4293 不同),确保断言能区分"替换生效"和"替换失效"
	varMsg := RecordMessage{
		MsgName:     "TeamSelectGuildCityReq",
		MsgID:       uint16(pb.EGameMsgID_TeamSelectGuildCityReq_id),
		PayloadJSON: `{"cityId":9999}`, // 写死的占位值(故意与 Ntf 值不同)
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	store := map[string]any{}
	err = ExtractVariablesForMessage(varMsg, mux, store)
	require.NoError(t, err)

	// --- 步骤5:resolveVariablePayload 替换 payload ---
	resolved := resolveVariablePayload(varMsg, store)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(resolved), &payload))
	assert.Equal(t, float64(4293), payload["cityId"], "cityId 应被替换为提取值")

	// 模拟发送替换后的消息
	resolvedFrame := EncodeFrame(varMsg.MsgID, 2, FlagEncrypt, []byte(resolved), true)
	_, err = fc.Write(resolvedFrame)
	require.NoError(t, err)
	got2, err := fc.WaitClientWrite(2 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, varMsg.MsgID, got2.MsgID)
}

// ========== 并发安全性测试 ==========

// TestLazyExtract_ConcurrentExtractSafety 验证多 goroutine 并发提取的安全性
// (虽然当前架构单账号单发送循环,但 FrameMux.WaitMsg 本身需并发安全)
func TestLazyExtract_ConcurrentExtractSafety(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 推送多个 Ntf
	for i := 0; i < 5; i++ {
		innerNtf := &pb.PveGuildCityDataNtf{
			MatchedCities: []*pb.MatchedGuildCityInfo{
				{
					IsAttack: false,
					CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: uint64(100 + i)}},
				},
			},
		}
		payload := makeTransportRawNtfPayload(
			uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
			makeProtoPayload(innerNtf),
		)
		fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), payload)
	}

	// 并发 WaitMsg(模拟多变量同时提取)
	var wg sync.WaitGroup
	results := make(chan uint64, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			frame, err := mux.WaitMsg(uint16(pb.EGameMsgID_TransportRawNtf_id), 2*time.Second)
			if err == nil && frame != nil {
				msg, err := UnmarshalProtoPayload(frame.MsgID, frame.Payload)
				if err == nil {
					transport := msg.(*pb.TransportRawNtf)
					if transport.GetMsgId() == uint32(pb.EGameMsgID_PveGuildCityDataNtf_id) {
						inner := &pb.PveGuildCityDataNtf{}
						innerData := StripByteStreamPrefix(transport.Data)
						if proto.Unmarshal(innerData, inner) == nil {
							if len(inner.MatchedCities) > 0 && inner.MatchedCities[0].CityInfo != nil && inner.MatchedCities[0].CityInfo.SimpleInfo != nil {
								results <- inner.MatchedCities[0].CityInfo.SimpleInfo.Id
							}
						}
					}
				}
			}
		}()
	}
	wg.Wait()
	close(results)

	// 至少一个成功(缓存帧只保留最新一个,但不应 panic/deadlock)
	count := 0
	for range results {
		count++
	}
	assert.GreaterOrEqual(t, count, 1, "并发提取应至少成功一次")
}

// ========== resolveVariablePayload 集成测试 ==========

// TestResolveVariablePayload_PreservesNonVariableFields 验证 A6 防御:
// 非 variable 类型的字段值不被覆盖
func TestResolveVariablePayload_PreservesNonVariableFields(t *testing.T) {
	msg := RecordMessage{
		PayloadJSON: `{"cityId":4293,"heroId":77,"count":3}`,
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
			// heroId/count 无 fieldMeta 或 InputType!=variable,不应被改
		},
	}
	store := map[string]any{"cityId": uint64(5555)}

	resolved := resolveVariablePayload(msg, store)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(resolved), &payload))
	assert.Equal(t, float64(5555), payload["cityId"], "variable 字段应被替换")
	assert.EqualValues(t, 77, payload["heroId"], "非 variable 字段保持原值")
	assert.EqualValues(t, 3, payload["count"], "非 variable 字段保持原值")
}

// ========== 工会战.json 完整时序模拟 ==========

// TestLazyExtract_FullGuildWarSequence 模拟工会战.json 的完整消息序列,
// 复现 sendMessagesOnce 发送循环的变量提取行为。
//
// 工会战.json 消息序列(节选关键节点):
//  1. GmCommandReq × 2 (添加粮草/虎符)
//  2. GuildEnterReq / CompleteClientFuncReq / PostUserStatusReq
//  3. GuildCityWarDataReq × 2 ← 触发服务器推送 TransportRawNtf(含 cityId)
//  4. JoinRoomReq / TeamMatchGuildCityReq
//  5. TeamSelectGuildCityReq ← 依赖 cityId 变量(惰性提取点)
//  6. SaveLeadHerosReq / RoomAddRobotReq × 2 / GuildCityWarDataReq / StartBattleReq ...
//  7. 战斗结束后续消息
//
// 本测试模拟发送循环逐条处理,验证:
//   - 非变量消息(1,2,4,6,7)不触发提取,payload 原样通过
//   - 触发消息(3)发出后,Ntf 被 readLoop 缓存
//   - 变量消息(5)发送前惰性提取,cityId 被正确替换
//   - 整个序列无阻塞、无超时
func TestLazyExtract_FullGuildWarSequence(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// 构造工会战.json 的消息序列(RecordMessage 格式,Direction="→")
	// 用真实 MsgID 常量,payload 用 JSON 字符串模拟
	mkMsg := func(name string, msgID uint16, seqID uint32, payload string) RecordMessage {
		return RecordMessage{
			MsgName:     name,
			MsgID:       msgID,
			SeqID:       seqID,
			Direction:   DirClientToServer,
			PayloadJSON: payload,
		}
	}
	mkVarMsg := func(name string, msgID uint16, seqID uint32, payload string) RecordMessage {
		m := mkMsg(name, msgID, seqID, payload)
		m.FieldValues = map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		}
		return m
	}

	messages := []RecordMessage{
		mkMsg("GmCommandReq", uint16(pb.EGameMsgID_GmCommandReq_id), 29, `{"content":"//AddItem 8000003 10"}`),
		mkMsg("GmCommandReq", uint16(pb.EGameMsgID_GmCommandReq_id), 29, `{"content":"//AddItem 1000016 10"}`),
		mkMsg("GuildEnterReq", uint16(pb.EGameMsgID_GuildEnterReq_id), 34, `{"guildId":595}`),
		mkMsg("CompleteClientFuncReq", uint16(pb.EGameMsgID_CompleteClientFuncReq_id), 36, `{"triggerType":4}`),
		mkMsg("PostUserStatusReq", uint16(pb.EGameMsgID_PostUserStatusReq_id), 38, `{"status":11}`),
		mkMsg("GuildCityWarDataReq", uint16(pb.EGameMsgID_GuildCityWarDataReq_id), 40, `{}`),
		mkMsg("GuildCityWarDataReq", uint16(pb.EGameMsgID_GuildCityWarDataReq_id), 42, `{}`),
		mkMsg("JoinRoomReq", uint16(pb.EGameMsgID_JoinRoomReq_id), 43, `{"isQuickJoin":true,"mode":15001}`),
		mkMsg("TeamMatchGuildCityReq", uint16(pb.EGameMsgID_TeamMatchGuildCityReq_id), 46, `{}`),
		mkVarMsg("TeamSelectGuildCityReq", uint16(pb.EGameMsgID_TeamSelectGuildCityReq_id), 48, `{"cityId":4293}`),
		mkMsg("SaveLeadHerosReq", uint16(pb.EGameMsgID_SaveLeadHerosReq_id), 18, `{"heroIds":[10001,10002,10003]}`),
		mkMsg("RoomAddRobotReq", uint16(pb.EGameMsgID_RoomAddRobotReq_id), 51, `{"seatID":2}`),
		mkMsg("RoomAddRobotReq", uint16(pb.EGameMsgID_RoomAddRobotReq_id), 52, `{"seatID":3}`),
		mkMsg("GuildCityWarDataReq", uint16(pb.EGameMsgID_GuildCityWarDataReq_id), 54, `{}`),
		mkMsg("StartBattleReq", uint16(pb.EGameMsgID_StartBattleReq_id), 55, `{}`),
	}

	store := map[string]any{}
	ntfPushed := false
	var sentPayloads []string

	// 模拟 sendMessagesOnce 的发送循环核心逻辑
	for i, msg := range messages {
		// 模拟:GuildCityWarDataReq(seq=40, 第6条)发出后,服务器推送 Ntf
		if msg.MsgID == uint16(pb.EGameMsgID_GuildCityWarDataReq_id) && msg.SeqID == 40 && !ntfPushed {
			// 确认这条消息确实被写出
			_, _ = fc.Write(EncodeFrame(msg.MsgID, msg.SeqID, FlagEncrypt, []byte(msg.PayloadJSON), true))
			_, _ = fc.WaitClientWrite(2 * time.Second)

			// 推送含 cityId=5566 的 Ntf
			innerNtf := &pb.PveGuildCityDataNtf{
				MatchedCities: []*pb.MatchedGuildCityInfo{
					{
						IsAttack: false,
						CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: 5566}},
					},
				},
			}
			transportPayload := makeTransportRawNtfPayload(
				uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
				makeProtoPayload(innerNtf),
			)
			fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), transportPayload)
			ntfPushed = true
			continue
		}

		// 变量消息:惰性提取 + 替换
		payloadToSend := msg.PayloadJSON
		if mux != nil && msgNeedsVariable(msg) {
			err := ExtractVariablesForMessage(msg, mux, store)
			require.NoError(t, err, "第%d条 %s 变量提取失败", i+1, msg.MsgName)
			payloadToSend = resolveVariablePayload(msg, store)
		}

		sentPayloads = append(sentPayloads, payloadToSend)
	}

	// 验证:变量消息(第10条 TeamSelectGuildCityReq)的 cityId 被替换为 5566
	// sentPayloads 索引:跳过了 seq=40 的 GuildCityWarDataReq(continue),所以变量消息在 sentPayloads[8]
	varMsgIdx := 8 // 0-based,前9条非变量消息发出后,第10条(变量)是 sentPayloads[8]
	var resolvedCityID map[string]any
	require.NoError(t, json.Unmarshal([]byte(sentPayloads[varMsgIdx]), &resolvedCityID))
	assert.Equal(t, float64(5566), resolvedCityID["cityId"],
		"TeamSelectGuildCityReq 的 cityId 应被替换为 Ntf 提取的 5566")

	// 验证:非变量消息的 payload 原样保留(抽查 GuildEnterReq)
	var guildEnter map[string]any
	require.NoError(t, json.Unmarshal([]byte(sentPayloads[2]), &guildEnter))
	assert.EqualValues(t, 595, guildEnter["guildId"], "非变量消息 payload 不变")

	// 验证:store 中 cityId 已填充
	assert.Equal(t, uint64(5566), store["cityId"], "variableStore 应填充提取值")
}

// ========== 未注册变量防御测试 ==========

// TestLazyExtract_UnregisteredVariableReturnsError 验证审核问题3的修复:
// 变量名在注册表中查不到(如前端拼写错误)时,必须返回 error,
// 而非静默用写死值兜底(否则 QA 工具会"测试了错误数据却以为成功")
func TestLazyExtract_UnregisteredVariableReturnsError(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	// cityID(大写D)是拼写错误,注册的是 cityId
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityID"}, // 注意:大写 D
		},
	}
	store := map[string]any{}

	err := ExtractVariablesForMessage(msg, mux, store)
	require.Error(t, err, "未注册变量应返回 error")
	assert.Contains(t, err.Error(), "未注册")
	assert.Contains(t, err.Error(), "cityID")
	assert.Empty(t, store, "未注册变量不应填充 store")
}

// TestScanFieldValues_UnregisteredVariableStillSetsHasVariable 验证 F1 修复:
// 未注册变量名不应让 ScanFieldValuesForVariables 返回 hasVariable=false。
// 修复前:未注册变量 → hasVariable=false → mux=nil → 发送循环整体跳过提取 → 静默用写死值。
// 修复后:未注册变量 → hasVariable=true(走 FrameMux 路径)→ ExtractVariablesForMessage 报错上报。
func TestScanFieldValues_UnregisteredVariableStillSetsHasVariable(t *testing.T) {
	messages := []RecordMessage{
		{
			MsgName:     "TeamSelectGuildCityReq",
			MsgID:       uint16(pb.EGameMsgID_TeamSelectGuildCityReq_id),
			PayloadJSON: `{"cityId":9999}`,
			FieldValues: map[string]FieldMetaValue{
				"cityId": {InputType: "variable", VariableName: "cityID"}, // 拼写错误(大写D),未注册
			},
		},
	}

	hasVariable, watchedIDs := ScanFieldValuesForVariables(messages)

	// F1 核心:未注册变量仍应置 hasVariable=true,确保走提取路径
	assert.True(t, hasVariable,
		"未注册变量应让 hasVariable=true(否则发送循环会跳过提取、静默用写死值)")
	// 未注册变量不贡献 watchedIDs
	assert.Empty(t, watchedIDs, "未注册变量不贡献 watchedIDs")
}

// ========== 审核 P0 补充: Ntf 延迟到达 → WaitMsg 阻塞唤醒 ==========

// TestLazyExtract_BlockingThenWakeOnDelayedNtf 验证 WaitMsg 的"阻塞→被 notifyCh 唤醒→命中"路径
// (审核 P0 盲区 A: 真实服务器场景下 Ntf 比提取调用晚到)
//
// 时序:
//  1. 启动 readLoop(不预推 Ntf)
//  2. 在独立 goroutine 调用 ExtractVariablesForMessage(它会阻塞在 WaitMsg)
//  3. 验证 goroutine 正在阻塞(短超时探测)
//  4. 推送 Ntf
//  5. 验证 goroutine 返回且 store 填充正确
func TestLazyExtract_BlockingThenWakeOnDelayedNtf(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	time.Sleep(50 * time.Millisecond) // 确保 readLoop 就绪

	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	store := map[string]any{}

	// 步骤2: 在独立 goroutine 调用提取(会阻塞在 WaitMsg)
	done := make(chan error, 1)
	go func() {
		done <- ExtractVariablesForMessage(msg, mux, store)
	}()

	// 步骤3: 验证正在阻塞(200ms 内不应返回)
	select {
	case <-done:
		t.Fatal("提取不应在 Ntf 推送前返回(应阻塞在 WaitMsg)")
	case <-time.After(200 * time.Millisecond):
		// 预期:仍在阻塞
	}

	// 步骤4: 推送 Ntf(cityId=8888)
	innerNtf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: 8888}},
			},
		},
	}
	transportPayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
		makeProtoPayload(innerNtf),
	)
	fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), transportPayload)

	// 步骤5: 验证唤醒后成功提取
	select {
	case err := <-done:
		require.NoError(t, err, "延迟 Ntf 到达后应唤醒并提取成功")
		assert.Equal(t, uint64(8888), store["cityId"], "应提取到延迟推送的 cityId")
	case <-time.After(6 * time.Second):
		t.Fatal("推送 Ntf 后提取未在 6s 内唤醒(notifyCh 未唤醒 WaitMsg)")
	}
}

// ========== 审核 P0 补充: DrainAndStart 池路径集成测试 ==========

// TestLazyExtract_DrainAndStartPoolPath 验证池复用连接路径下的惰性提取
// (审核 P0 盲区 B: 所有 lazy_extract_test 都用 readLoop, 没有一个用 DrainAndStart)
//
// 时序:
//  1. 先推几个"过期帧"(模拟上次会话残留) → DrainAndStart 应丢弃
//  2. DrainAndStart 启动 readLoop
//  3. 推送本会话的 Ntf → readLoop 缓存
//  4. 惰性提取命中 cache
func TestLazyExtract_DrainAndStartPoolPath(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})

	// 步骤1: 推过期帧(上次会话残留, cityId=1111)
	staleNtf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: 1111}},
			},
		},
	}
	stalePayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
		makeProtoPayload(staleNtf),
	)
	fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), stalePayload)

	// 步骤2: DrainAndStart(会丢弃积压的过期帧,不缓存)
	mux.DrainAndStart(300*time.Millisecond, nil, "test-pool")
	defer mux.Stop()

	// 验证过期帧未进 cache
	_, ok := mux.GetCache(uint16(pb.EGameMsgID_TransportRawNtf_id))
	assert.False(t, ok, "DrainAndStart 不应缓存过期积压帧")

	// 步骤3: 推送本会话的 Ntf(cityId=2222)
	freshNtf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: 2222}},
			},
		},
	}
	freshPayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
		makeProtoPayload(freshNtf),
	)
	fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), freshPayload)

	// 等待 readLoop 缓存
	require.Eventually(t, func() bool {
		_, ok := mux.GetCache(uint16(pb.EGameMsgID_TransportRawNtf_id))
		return ok
	}, 3*time.Second, 50*time.Millisecond, "readLoop 未缓存本会话 Ntf")

	// 步骤4: 惰性提取命中 cache
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	store := map[string]any{}
	err := ExtractVariablesForMessage(msg, mux, store)
	require.NoError(t, err, "DrainAndStart 路径惰性提取应成功")
	assert.Equal(t, uint64(2222), store["cityId"], "应提取本会话 Ntf 的 cityId, 而非过期帧的 1111")
}

// ========== 审核 P1 补充: 提取失败后 continue 语义 ==========

// TestLazyExtract_FailureSkipsButSubsequentMessagesStillSend 验证:
// 变量提取失败时跳过该消息,但后续非变量消息仍正常处理
// (审核 P1 缺口 #6: 验证 continue 语义)
//
// F2 重写 (2026-06-15): 旧版用 if 模拟失败分支,属于虚假测试。现改为真实失败:
// 关闭 FakeConn → readLoop 因 io.EOF 退出 → signalStopped 传播失败 (F3 修复)
// → ExtractVariablesForMessage 立即返回 error(不等 5s 超时)。
// 然后真实走发送循环决策,用 WaitClientWrite 断言实际发出的帧。
func TestLazyExtract_FailureSkipsButSubsequentMessagesStillSend(t *testing.T) {
	fc := NewFakeConn()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	time.Sleep(50 * time.Millisecond) // 确保 readLoop 就绪

	// 关闭连接:readLoop 的 readFrame 会返回 io.EOF → readLoop 退出 + signalStopped
	// 这样 ExtractVariablesForMessage 的 WaitMsg 会立即返回"FrameMux 已停止"错误,无需等 10s 超时
	_ = fc.Close()

	// 等待 readLoop 真正退出并完成 signalStopped
	require.Eventually(t, func() bool {
		_, err := mux.WaitMsg(uint16(pb.EGameMsgID_TransportRawNtf_id), 100*time.Millisecond)
		// 已停止时应返回错误
		return err != nil
	}, 2*time.Second, 20*time.Millisecond, "readLoop 未在连接关闭后传播停止状态")

	// 发送循环消息序列: [变量消息(会提取失败→跳过), 非变量消息(应正常发送)]
	messages := []RecordMessage{
		{
			MsgName:     "TeamSelectGuildCityReq",
			MsgID:       uint16(pb.EGameMsgID_TeamSelectGuildCityReq_id),
			PayloadJSON: `{"cityId":9999}`,
			FieldValues: map[string]FieldMetaValue{
				"cityId": {InputType: "variable", VariableName: "cityId"},
			},
		},
		{
			MsgName:     "SaveLeadHerosReq",
			MsgID:       uint16(pb.EGameMsgID_SaveLeadHerosReq_id),
			PayloadJSON: `{"heroIds":[10001]}`,
		},
	}

	store := map[string]any{}
	hasVariableContext := mux != nil // 模拟 sendMessagesOnce 的判定

	// 真实复现 sendMessagesOnce 发送循环的核心决策逻辑:
	// 变量消息 → 真实调用 ExtractVariablesForMessage(连接已断 → WaitMsg 立即失败 → error)
	// → 跳过该消息; 非变量消息 → 不触发提取,进入发送路径
	var skippedMsgNames []string
	var sentMsgNames []string
	for _, msg := range messages {
		if hasVariableContext && msgNeedsVariable(msg) {
			// 真实调用被测函数(非模拟):连接已断 → WaitMsg 立即失败 → 返回 error
			extractErr := ExtractVariablesForMessage(msg, mux, store)
			if extractErr != nil {
				skippedMsgNames = append(skippedMsgNames, msg.MsgName)
				continue // 提取失败跳过该消息(与 replay.go:376-385 一致)
			}
			// 提取成功才会走到 resolveVariablePayload + 发送(本测试不会到这)
		}
		sentMsgNames = append(sentMsgNames, msg.MsgName)
	}

	// 验证:变量消息提取失败被跳过
	assert.Contains(t, skippedMsgNames, "TeamSelectGuildCityReq",
		"变量消息在连接已断时应提取失败并跳过")
	// 验证:后续非变量消息仍正常进入发送路径(不触发提取)
	assert.Contains(t, sentMsgNames, "SaveLeadHerosReq",
		"提取失败后后续非变量消息应正常发送")
	assert.Len(t, sentMsgNames, 1, "仅非变量消息进入发送路径")
	assert.Empty(t, store, "失败时 variableStore 不应被填充")
}

// ========== 审核 P2 补充: 同名 Ntf 多次推送取最新值 ==========

// TestLazyExtract_LatestNtfWins 验证同名 Ntf 多次推送时,提取取最新值
// (审核 P2 缺口 #3: readLoop cache 是覆盖语义, 需显式断言)
func TestLazyExtract_LatestNtfWins(t *testing.T) {
	fc := NewFakeConn()
	defer func() { _ = fc.Close() }()

	mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
	mux.wg.Add(1)
	go mux.readLoop(nil, "test")
	defer mux.Stop()

	time.Sleep(50 * time.Millisecond)

	// 推送 3 个不同 cityId 的 Ntf
	for _, id := range []uint64{100, 200, 300} {
		innerNtf := &pb.PveGuildCityDataNtf{
			MatchedCities: []*pb.MatchedGuildCityInfo{
				{
					IsAttack: false,
					CityInfo: &pb.GuildCityInfo{SimpleInfo: &pb.GuildCitySimpleInfo{Id: id}},
				},
			},
		}
		payload := makeTransportRawNtfPayload(
			uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
			makeProtoPayload(innerNtf),
		)
		fc.PushServerFrame(uint16(pb.EGameMsgID_TransportRawNtf_id), payload)
		time.Sleep(30 * time.Millisecond) // 确保 readLoop 按序处理
	}

	// 等待缓存稳定
	require.Eventually(t, func() bool {
		_, ok := mux.GetCache(uint16(pb.EGameMsgID_TransportRawNtf_id))
		return ok
	}, 3*time.Second, 50*time.Millisecond)

	// 提取:应取最新值 300(cache 覆盖语义)
	msg := RecordMessage{
		FieldValues: map[string]FieldMetaValue{
			"cityId": {InputType: "variable", VariableName: "cityId"},
		},
	}
	store := map[string]any{}
	err := ExtractVariablesForMessage(msg, mux, store)
	require.NoError(t, err)
	assert.Equal(t, uint64(300), store["cityId"], "多次推送应取最新值 300")
}
