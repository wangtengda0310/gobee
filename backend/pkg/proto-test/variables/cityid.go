package variables

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	proto "github.com/gogo/protobuf/proto"
)

// ExtractGuildCityID 从 PveGuildCityDataNtf 或 TransportRawNtf 帧中提取可选中的城池 ID
func ExtractGuildCityID(frame any) (any, error) {
	f, ok := frame.(*params.DecodedFrame)
	if !ok || f == nil {
		return nil, fmt.Errorf("frame 类型断言失败")
	}
	ntf, err := parsePveGuildCityDataFromFrame(f)
	if err != nil {
		return nil, err
	}
	if ntf == nil {
		return nil, nil
	}
	return pickSelectableCityID(ntf)
}

// parsePveGuildCityDataFromFrame 从解码帧中解析出 PveGuildCityDataNtf
func parsePveGuildCityDataFromFrame(frame *params.DecodedFrame) (*pb.PveGuildCityDataNtf, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame 为空")
	}

	msgIDPveGuildCityDataNtf := uint16(pb.EGameMsgID_PveGuildCityDataNtf_id)
	msgIDTransportRawNtf := uint16(pb.EGameMsgID_TransportRawNtf_id)

	if frame.MsgID == msgIDPveGuildCityDataNtf {
		msg, err := params.UnmarshalProtoPayload(frame.MsgID, frame.Payload)
		if err != nil {
			return nil, err
		}
		return msg.(*pb.PveGuildCityDataNtf), nil
	}

	if frame.MsgID != msgIDTransportRawNtf {
		return nil, nil
	}

	transport, err := params.UnmarshalProtoPayload(frame.MsgID, frame.Payload)
	if err != nil {
		return nil, fmt.Errorf("解析 TransportRawNtf 失败: %w", err)
	}
	rawNtf := transport.(*pb.TransportRawNtf)

	if rawNtf.GetMsgId() != uint32(msgIDPveGuildCityDataNtf) {
		return nil, nil
	}

	inner := &pb.PveGuildCityDataNtf{}
	innerData := params.StripByteStreamPrefix(rawNtf.Data)
	if err := proto.Unmarshal(innerData, inner); err != nil {
		if err2 := proto.Unmarshal(rawNtf.Data, inner); err2 != nil {
			return nil, fmt.Errorf("解析 TransportRawNtf 内层 PveGuildCityDataNtf 失败: %w", err)
		}
	}
	return inner, nil
}

// pickSelectableCityID 从 PveGuildCityDataNtf 的 matchedCities 中选择一个城池 ID
func pickSelectableCityID(ntf *pb.PveGuildCityDataNtf) (uint64, error) {
	if ntf == nil || len(ntf.MatchedCities) == 0 {
		return 0, fmt.Errorf("matchedCities 为空")
	}
	for _, city := range ntf.MatchedCities {
		if city == nil || city.CityInfo == nil || city.CityInfo.SimpleInfo == nil {
			continue
		}
		if !city.IsAttack {
			return city.CityInfo.SimpleInfo.Id, nil
		}
	}
	for _, city := range ntf.MatchedCities {
		if city != nil && city.CityInfo != nil && city.CityInfo.SimpleInfo != nil {
			return city.CityInfo.SimpleInfo.Id, nil
		}
	}
	return 0, fmt.Errorf("matchedCities 中无有效城池")
}
