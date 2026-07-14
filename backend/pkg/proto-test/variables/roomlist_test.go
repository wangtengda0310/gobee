package variables

import (
	"testing"

	_ "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	"github.com/stretchr/testify/assert"
)

// TestExtractRoomCreator 从 NewGetRoomListAck 中提取第一个房间的创建者 ID
func TestExtractRoomCreator(t *testing.T) {
	ack := &pb.NewGetRoomListAck{
		Infos: []*pb.RoomInfo{
			{Creator: 12345, RoomID: 67890},
			{Creator: 11111, RoomID: 22222},
		},
	}
	frame := makeDecodedFrame(uint16(pb.EGameMsgID_NewGetRoomListAck_id), makeProtoPayload(ack))

	val, err := ExtractRoomCreator(frame)
	assert.NoError(t, err)
	assert.Equal(t, uint64(12345), val)
}

// TestExtractRoomID 从 NewGetRoomListAck 中提取第一个房间的房间 ID
func TestExtractRoomID(t *testing.T) {
	ack := &pb.NewGetRoomListAck{
		Infos: []*pb.RoomInfo{
			{Creator: 12345, RoomID: 67890},
		},
	}
	frame := makeDecodedFrame(uint16(pb.EGameMsgID_NewGetRoomListAck_id), makeProtoPayload(ack))

	val, err := ExtractRoomID(frame)
	assert.NoError(t, err)
	assert.Equal(t, uint64(67890), val)
}

// TestExtractRoomFromEmptyInfos Infos 为空时返回错误
func TestExtractRoomFromEmptyInfos(t *testing.T) {
	ack := &pb.NewGetRoomListAck{Infos: []*pb.RoomInfo{}}
	frame := makeDecodedFrame(uint16(pb.EGameMsgID_NewGetRoomListAck_id), makeProtoPayload(ack))

	_, err := ExtractRoomCreator(frame)
	assert.Error(t, err)

	_, err = ExtractRoomID(frame)
	assert.Error(t, err)
}

// TestExtractRoomNonTargetMsgID 非 NewGetRoomListAck 帧返回 nil
func TestExtractRoomNonTargetMsgID(t *testing.T) {
	frame := makeDecodedFrame(uint16(pb.EGameMsgID_HelloReq_id), makeProtoPayload(&pb.HelloReq{Name: "test"}))

	val, err := ExtractRoomCreator(frame)
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = ExtractRoomID(frame)
	assert.NoError(t, err)
	assert.Nil(t, val)
}

// TestExtractRoomInvalidFrameType frame 类型错误时返回错误
func TestExtractRoomInvalidFrameType(t *testing.T) {
	_, err := ExtractRoomCreator("not a frame")
	assert.Error(t, err)

	_, err = ExtractRoomID(123)
	assert.Error(t, err)
}

// TestFindRoomVariablesByShortName 验证注册表中能找到新变量
func TestFindRoomVariablesByShortName(t *testing.T) {
	creatorDef := params.FindVariableByShortName("roomCreator")
	assert.NotNil(t, creatorDef)
	assert.Equal(t, "roomCreator", creatorDef.ShortName)
	assert.NotNil(t, creatorDef.ExtractFunc)
	assert.Equal(t, []uint16{uint16(pb.EGameMsgID_NewGetRoomListAck_id)}, creatorDef.WatchMsgIDs)

	roomIDDef := params.FindVariableByShortName("roomID")
	assert.NotNil(t, roomIDDef)
	assert.Equal(t, "roomID", roomIDDef.ShortName)
	assert.NotNil(t, roomIDDef.ExtractFunc)
	assert.Equal(t, []uint16{uint16(pb.EGameMsgID_NewGetRoomListAck_id)}, roomIDDef.WatchMsgIDs)
}
