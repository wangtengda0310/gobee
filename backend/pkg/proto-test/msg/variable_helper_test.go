package protocol

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
	proto "github.com/gogo/protobuf/proto"
)

func init() {
	// 为 msg 包测试补充注册 cityId 变量。
	// variables 包在运行时由 wails.go 导入触发注册，但 msg 包测试不依赖 variables 包，
	// 因此需要自行注册以保证 lazy_extract 等端到端测试能真实提取 cityId。
	params.AppendBuiltinVariables([]params.VariableDef{
		{
			ShortName:   "cityId",
			DisplayName: "测试用 cityId",
			WatchMsgIDs: []uint16{
				uint16(pb.EGameMsgID_TransportRawNtf_id),
				uint16(pb.EGameMsgID_PveGuildCityDataNtf_id),
			},
			ExtractFunc: testExtractCityID,
		},
	})
}

// testExtractCityID 测试用 cityId 提取函数，逻辑与 variables.ExtractGuildCityID 一致
func testExtractCityID(frame any) (any, error) {
	f, ok := frame.(*DecodedFrame)
	if !ok || f == nil {
		return nil, fmt.Errorf("frame 类型断言失败")
	}
	ntf, err := testParsePveGuildCityDataFromFrame(f)
	if err != nil {
		return nil, err
	}
	if ntf == nil {
		return nil, nil
	}
	return testPickSelectableCityID(ntf)
}

func testParsePveGuildCityDataFromFrame(frame *DecodedFrame) (*pb.PveGuildCityDataNtf, error) {
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

func testPickSelectableCityID(ntf *pb.PveGuildCityDataNtf) (uint64, error) {
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
