package variables

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
)

// ExtractRoomCreator 从 NewGetRoomListAck 帧中提取第一个房间的创建者 ID
func ExtractRoomCreator(frame any) (any, error) {
	ack, err := parseNewGetRoomListAck(frame)
	if err != nil {
		return nil, err
	}
	if ack == nil {
		return nil, nil
	}
	if len(ack.Infos) == 0 || ack.Infos[0] == nil {
		return nil, fmt.Errorf("NewGetRoomListAck 无有效房间信息")
	}
	return ack.Infos[0].Creator, nil
}

// ExtractRoomID 从 NewGetRoomListAck 帧中提取第一个房间的房间 ID
func ExtractRoomID(frame any) (any, error) {
	ack, err := parseNewGetRoomListAck(frame)
	if err != nil {
		return nil, err
	}
	if ack == nil {
		return nil, nil
	}
	if len(ack.Infos) == 0 || ack.Infos[0] == nil {
		return nil, fmt.Errorf("NewGetRoomListAck 无有效房间信息")
	}
	return ack.Infos[0].RoomID, nil
}

// parseNewGetRoomListAck 从解码帧中解析出 NewGetRoomListAck
func parseNewGetRoomListAck(frame any) (*pb.NewGetRoomListAck, error) {
	f, ok := frame.(*params.DecodedFrame)
	if !ok || f == nil {
		return nil, fmt.Errorf("frame 类型断言失败")
	}
	msgID := uint16(pb.EGameMsgID_NewGetRoomListAck_id)
	if f.MsgID != msgID {
		return nil, nil
	}
	msg, err := params.UnmarshalProtoPayload(f.MsgID, f.Payload)
	if err != nil {
		return nil, fmt.Errorf("解析 NewGetRoomListAck 失败: %w", err)
	}
	return msg.(*pb.NewGetRoomListAck), nil
}
