package protocol

import (
	"encoding/json"
	"testing"

	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helloReqMsgID 取 HelloReq 的 MsgID（EGameMsgID_HelloReq_id = 1004，>= 1000 满足 payloadToJSON 的协议消息条件）
const helloReqMsgID = uint16(pb.EGameMsgID_HelloReq_id)

// TestPayloadToJSON_FrameworkMessage 验证 MsgID < 1000 的框架消息不被序列化
func TestPayloadToJSON_FrameworkMessage(t *testing.T) {
	_, err := payloadToJSON(2, nil) // MsgID=2 是 LoginResp，属于框架消息
	assert.Error(t, err, "框架消息（MsgID < 1000）应返回错误")
	assert.Contains(t, err.Error(), "框架消息", "错误信息应说明是框架消息")
}

// TestPayloadToJSON_UnregisteredMsg 验证未注册的协议消息返回错误
func TestPayloadToJSON_UnregisteredMsg(t *testing.T) {
	// 9999 不在 msgRegistry 中
	_, err := payloadToJSON(9999, []byte{0, 0})
	assert.Error(t, err, "未注册的 MsgID 应返回错误")
	assert.Contains(t, err.Error(), "消息未注册", "错误信息应说明消息未注册")
}

// TestPayloadToJSON_EmptyPayload 验证空 proto 数据时仍能序列化出零值 JSON
func TestPayloadToJSON_EmptyPayload(t *testing.T) {
	// payload 只有 2 字节 dataLen=0，无实际 proto 数据
	jsonStr, err := payloadToJSON(helloReqMsgID, []byte{0, 0})
	require.NoError(t, err, "空 proto 数据应成功序列化为零值对象")
	assert.NotEmpty(t, jsonStr, "零值对象 JSON 不应为空")

	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &m), "输出应为合法 JSON")
}

// TestEncodeClientMessage_UnregisteredMsg 验证未注册消息编码失败
func TestEncodeClientMessage_UnregisteredMsg(t *testing.T) {
	_, err := EncodeClientMessage(9999, 1, `{}`)
	assert.Error(t, err, "未注册的 MsgID 应编码失败")
	assert.Contains(t, err.Error(), "消息未注册", "错误信息应说明消息未注册")
}

// TestEncodeClientMessage_InvalidJSON 验证非法 JSON payload 编码失败
func TestEncodeClientMessage_InvalidJSON(t *testing.T) {
	_, err := EncodeClientMessage(helloReqMsgID, 1, `{invalid`)
	assert.Error(t, err, "非法 JSON 应编码失败")
	assert.Contains(t, err.Error(), "JSON 反序列化失败", "错误信息应说明是 JSON 反序列化失败")
}

// TestEncodeClientMessage_HelloReq_RoundTrip 验证完整编解码往返：
// JSON payload → EncodeClientMessage → DecodeFrame → payloadToJSON → 还原 JSON
// 这是 transport.go 两个纯函数协同工作的核心契约，帧编码错误会导致静默数据损坏
func TestEncodeClientMessage_HelloReq_RoundTrip(t *testing.T) {
	const seqID = uint32(42)
	original := map[string]any{
		"Name": "tester",
		"Id":   float64(12345),
	}
	originalJSON, err := json.Marshal(original)
	require.NoError(t, err)

	// 编码：JSON → 加密协议帧
	frame, err := EncodeClientMessage(helloReqMsgID, seqID, string(originalJSON))
	require.NoError(t, err, "HelloReq 应编码成功")
	assert.NotEmpty(t, frame, "编码后的帧不应为空")

	// 解码：加密协议帧 → DecodedFrame（isClientData=true 对称解密）
	decoded, err := DecodeFrame(frame, true)
	require.NoError(t, err, "对称解码应成功")
	assert.Equal(t, helloReqMsgID, decoded.MsgID, "MsgID 往返应一致")
	assert.Equal(t, seqID, decoded.SeqID, "SeqID 往返应一致")
	assert.NotEmpty(t, decoded.Payload, "解码后的 payload 不应为空")

	// 反序列化：payload → JSON（payloadToJSON 的核心职责）
	roundTripJSON, err := payloadToJSON(decoded.MsgID, decoded.Payload)
	require.NoError(t, err, "payloadToJSON 应成功还原 JSON")

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal([]byte(roundTripJSON), &roundTrip), "往返 JSON 应可解析")
	assert.Equal(t, "tester", roundTrip["Name"], "Name 字段往返应一致")
	assert.EqualValues(t, 12345, roundTrip["Id"], "Id 字段往返应一致")
}
