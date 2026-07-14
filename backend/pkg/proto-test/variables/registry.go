package variables

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	pb "git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_pb"
)

func init() {
	// 注册内置变量。
	// cityId 的 ExtractFunc 依赖 protocol.DecodedFrame，因此放在本包实现，
	// 由顶层包（如 wails.go）导入本包触发注册。
	params.SetBuiltinVariables([]params.VariableDef{
		{
			ShortName:   "cityId",
			DisplayName: "🏰 工会城战 - 城池ID",
			WatchMsgIDs: []uint16{
				uint16(pb.EGameMsgID_TransportRawNtf_id),
				uint16(pb.EGameMsgID_PveGuildCityDataNtf_id),
			},
			ExtractFunc:   ExtractGuildCityID,
			AvailableReqs: []string{"TeamSelectGuildCityReq"},
		},
		{
			ShortName:   "roomCreator",
			DisplayName: "🏠 房间列表 - 创建者ID",
			WatchMsgIDs: []uint16{
				uint16(pb.EGameMsgID_NewGetRoomListAck_id),
			},
			ExtractFunc:   ExtractRoomCreator,
			AvailableReqs: []string{"RoomLookOnReq"},
		},
		{
			ShortName:   "roomID",
			DisplayName: "🏠 房间列表 - 房间ID",
			WatchMsgIDs: []uint16{
				uint16(pb.EGameMsgID_NewGetRoomListAck_id),
			},
			ExtractFunc:   ExtractRoomID,
			AvailableReqs: []string{"RoomLookOnReq"},
		},
		{
			// openid 为账号级变量，值由发送循环根据当前 accountID 预置，
			// 不需要从服务端 Ntf 提取，因此 WatchMsgIDs 为空。
			// AvailableReqs 留空（nil），对所有 Req 可用。
			ShortName:   params.OpenIDShortName,
			DisplayName: "👤 当前账号",
			WatchMsgIDs: []uint16{},
		},
	})
}
