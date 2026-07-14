package params

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"
)

// MessageFactory 根据 MsgID 创建 proto.Message 实例的工厂函数。
// 由 msg 包在 init 时注入，避免 params 包反向依赖 msg 包。
var MessageFactory func(msgID uint16) (proto.Message, bool)

// StripByteStreamPrefix 剥离 ByteStream 的 2字节LE长度前缀
func StripByteStreamPrefix(payload []byte) []byte {
	if len(payload) >= 2 {
		dataLen := int(payload[0]) | int(payload[1])<<8
		if dataLen > 0 && dataLen <= len(payload)-2 {
			return payload[2 : 2+dataLen]
		}
	}
	return payload
}

// UnmarshalProtoPayload 将帧 payload 反序列化为 proto 消息实例
func UnmarshalProtoPayload(msgID uint16, payload []byte) (proto.Message, error) {
	if MessageFactory == nil {
		return nil, fmt.Errorf("MessageFactory 未注入")
	}
	msg, ok := MessageFactory(msgID)
	if !ok {
		return nil, fmt.Errorf("消息未注册: MsgID=%d", msgID)
	}
	protoData := StripByteStreamPrefix(payload)
	if err := proto.Unmarshal(protoData, msg); err != nil {
		return nil, fmt.Errorf("反序列化 MsgID=%d 失败: %w", msgID, err)
	}
	return msg, nil
}
