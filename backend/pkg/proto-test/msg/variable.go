package protocol

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
	proto "github.com/gogo/protobuf/proto"
)

func init() {
	// 向 params 包注入 proto 消息工厂，供 variables 包等城市Id提取使用
	params.MessageFactory = NewMessage
}

// StripByteStreamPrefix 是 params.StripByteStreamPrefix 的别名，保留以保持向后兼容。
func StripByteStreamPrefix(payload []byte) []byte {
	return params.StripByteStreamPrefix(payload)
}

// UnmarshalProtoPayload 是 params.UnmarshalProtoPayload 的别名，保留以保持向后兼容。
func UnmarshalProtoPayload(msgID uint16, payload []byte) (proto.Message, error) {
	return params.UnmarshalProtoPayload(msgID, payload)
}
