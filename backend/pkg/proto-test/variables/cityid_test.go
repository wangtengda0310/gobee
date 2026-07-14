package variables

import (
	"testing"

	_ "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

// ========== 辅助函数 ==========

// makeDecodedFrame 构造一个测试用 DecodedFrame
func makeDecodedFrame(msgID uint16, payload []byte) *params.DecodedFrame {
	return &params.DecodedFrame{
		MsgID:   msgID,
		SeqID:   1,
		Flags:   0,
		Payload: payload,
	}
}

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

// ========== 测试用例 ==========

// TestExtractGuildCityID 构造 PveGuildCityDataNtf 帧（含 2 个 matchedCities），
// 验证提取第一个未攻打的 cityId
func TestExtractGuildCityID(t *testing.T) {
	ntf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: true,
				CityInfo: &pb.GuildCityInfo{
					SimpleInfo: &pb.GuildCitySimpleInfo{Id: 100},
				},
			},
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{
					SimpleInfo: &pb.GuildCitySimpleInfo{Id: 200},
				},
			},
		},
	}

	msgID := uint16(pb.EGameMsgID_PveGuildCityDataNtf_id)
	frame := makeDecodedFrame(msgID, makeProtoPayload(ntf))

	val, err := ExtractGuildCityID(frame)
	assert.NoError(t, err)
	assert.Equal(t, uint64(200), val)
}

// TestExtractFromTransportRawNtf 构造 TransportRawNtf 帧（内层 PveGuildCityDataNtf），
// 验证解信封后提取
func TestExtractFromTransportRawNtf(t *testing.T) {
	innerNtf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: false,
				CityInfo: &pb.GuildCityInfo{
					SimpleInfo: &pb.GuildCitySimpleInfo{Id: 42},
				},
			},
		},
	}
	innerPayload := makeProtoPayload(innerNtf)

	transportPayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_PveGuildCityDataNtf_id),
		innerPayload,
	)

	msgID := uint16(pb.EGameMsgID_TransportRawNtf_id)
	frame := makeDecodedFrame(msgID, transportPayload)

	val, err := ExtractGuildCityID(frame)
	assert.NoError(t, err)
	assert.Equal(t, uint64(42), val)
}

// TestExtractNonTargetInnerMsg TransportRawNtf 内层为其他消息，返回 nil
func TestExtractNonTargetInnerMsg(t *testing.T) {
	// 内层为 HelloReq（非 PveGuildCityDataNtf）
	innerMsg := &pb.HelloReq{
		Name: "test",
	}
	innerPayload := makeProtoPayload(innerMsg)

	transportPayload := makeTransportRawNtfPayload(
		uint32(pb.EGameMsgID_HelloReq_id),
		innerPayload,
	)

	msgID := uint16(pb.EGameMsgID_TransportRawNtf_id)
	frame := makeDecodedFrame(msgID, transportPayload)

	val, err := ExtractGuildCityID(frame)
	assert.NoError(t, err)
	assert.Nil(t, val, "非目标内层消息应返回 nil")
}

// TestExtractNoAttackableCity 所有城池 isAttack=true，fallback 到第一个有效城池
func TestExtractNoAttackableCity(t *testing.T) {
	ntf := &pb.PveGuildCityDataNtf{
		MatchedCities: []*pb.MatchedGuildCityInfo{
			{
				IsAttack: true,
				CityInfo: &pb.GuildCityInfo{
					SimpleInfo: &pb.GuildCitySimpleInfo{Id: 300},
				},
			},
			{
				IsAttack: true,
				CityInfo: &pb.GuildCityInfo{
					SimpleInfo: &pb.GuildCitySimpleInfo{Id: 400},
				},
			},
		},
	}

	msgID := uint16(pb.EGameMsgID_PveGuildCityDataNtf_id)
	frame := makeDecodedFrame(msgID, makeProtoPayload(ntf))

	val, err := ExtractGuildCityID(frame)
	assert.NoError(t, err)
	assert.Equal(t, uint64(300), val, "全部已攻占时应回退到第一个有效城池")
}
