package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	proto "github.com/gogo/protobuf/proto"
)

// 框架内部消息名称映射（MsgID < 1000）
var frameworkMsgNames = map[uint16]string{
	1: "LoginReq",
	2: "LoginResp",
	3: "Ping",
	4: "Pong",
}

// FormatFrame 将解码后的帧格式化为终端可读字符串
func FormatFrame(frame *DecodedFrame, connID uint64, dir string) string {
	ts := time.Now().Format("15:04:05.000")
	msgName := getMsgName(frame.MsgID)

	// 标志位信息
	flags := ""
	if frame.Flags&FlagCompress != 0 {
		flags += "C"
	}
	if frame.Flags&FlagEncrypt != 0 {
		flags += "E"
	}
	if flags == "" {
		flags = "-"
	}

	line := fmt.Sprintf("[%s] #%d %s MsgID=%d(%s) SeqID=%d Flags=[%s] Payload=%dB",
		ts, connID, dir, frame.MsgID, msgName, frame.SeqID, flags, len(frame.Payload))

	// 尝试 Proto 反序列化（仅 MsgID >= 1000）
	if frame.MsgID >= 1000 && len(frame.Payload) > 0 {
		if jsonStr := tryProtoToJSON(frame.MsgID, frame.Payload); jsonStr != "" {
			line += "\n  " + jsonStr
		}
	}

	return line
}

// getMsgName 获取消息名称
func getMsgName(msgID uint16) string {
	if msgID < 1000 {
		if name, ok := frameworkMsgNames[msgID]; ok {
			return name
		}
		return fmt.Sprintf("Framework_%d", msgID)
	}
	return GetMsgName(msgID)
}

// tryProtoToJSON 尝试将 payload 反序列化为 Proto 消息并输出 JSON
func tryProtoToJSON(msgID uint16, payload []byte) string {
	msg, ok := NewMessage(msgID)
	if !ok {
		return ""
	}

	// Proto 消息的 payload 格式：[2B 长度 LE][proto 二进制数据]
	var protoData []byte
	if len(payload) >= 2 {
		dataLen := int(payload[0]) | int(payload[1])<<8
		if dataLen <= len(payload)-2 {
			protoData = payload[2 : 2+dataLen]
		}
	}

	if len(protoData) == 0 {
		return ""
	}

	if err := proto.Unmarshal(protoData, msg); err != nil {
		return fmt.Sprintf("(反序列化失败: %v)", err)
	}

	// 转为 JSON
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return ""
	}

	// 格式化 JSON，限制长度
	jsonStr := string(jsonBytes)
	if len(jsonStr) > 500 {
		jsonStr = jsonStr[:500] + "..."
	}

	// 缩进：每行前加 "  "
	jsonStr = strings.ReplaceAll(jsonStr, "\\n", "\n  ")
	return jsonStr
}
