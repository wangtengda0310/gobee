package protocol

import (
	"encoding/binary"
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
)

// DecodedFrame 解码后的协议帧（类型别名，实际定义在 params 包）
type DecodedFrame = params.DecodedFrame

// DecodeFrame 从原始 TCP 数据中解码一个协议帧
// raw 必须包含完整的帧（帧头+消息体），至少 FrameHeaderSize 字节
// isClientData=true 表示这是客户端→服务端方向的数据
func DecodeFrame(raw []byte, isClientData bool) (*DecodedFrame, error) {
	if len(raw) < FrameHeaderSize {
		return nil, fmt.Errorf("数据太短: %d 字节", len(raw))
	}

	// 解析帧头
	msgLen := int(raw[0]) | int(raw[1])<<8 | int(raw[2])<<16
	flags := raw[3]

	// 消息体
	bodyStart := FrameHeaderSize
	bodyEnd := bodyStart + msgLen
	if bodyEnd > len(raw) {
		return nil, fmt.Errorf("消息不完整: 声明 %d 字节，实际 %d 字节", msgLen, len(raw)-bodyStart)
	}

	body := make([]byte, msgLen)
	copy(body, raw[bodyStart:bodyEnd])

	// 解密
	if flags&FlagEncrypt != 0 {
		DecryptXOR(body, isClientData)
	}

	// 解压
	if flags&FlagCompress != 0 {
		decompressed, err := Decompress(body)
		if err != nil {
			return nil, fmt.Errorf("snappy 解压失败: %w", err)
		}
		body = decompressed
	}

	// 解析 MsgID + SeqID
	if len(body) < MsgIDSize+SeqIDSize {
		return nil, fmt.Errorf("消息体太短: %d 字节（至少需要 %d）", len(body), MsgIDSize+SeqIDSize)
	}

	msgID := binary.LittleEndian.Uint16(body[0:MsgIDSize])
	seqID := binary.LittleEndian.Uint32(body[MsgIDSize : MsgIDSize+SeqIDSize])
	payload := body[MsgIDSize+SeqIDSize:]

	return &DecodedFrame{
		MsgID:   msgID,
		SeqID:   seqID,
		Flags:   flags,
		Payload: payload,
		RawSize: bodyEnd,
	}, nil
}

// ParseFrameHeader 从 TCP 流的开头解析帧头，返回消息体长度和标志位
// 返回 (消息体长度, 标志位, error)
func ParseFrameHeader(header []byte) (msgLen int, flags byte, err error) {
	if len(header) < FrameHeaderSize {
		return 0, 0, fmt.Errorf("帧头不完整: %d 字节", len(header))
	}
	msgLen = int(header[0]) | int(header[1])<<8 | int(header[2])<<16
	flags = header[3]

	if msgLen > MaxPacketSize {
		return 0, 0, fmt.Errorf("消息过长: %d 字节（最大 %d）", msgLen, MaxPacketSize)
	}
	if msgLen < MsgIDSize+SeqIDSize {
		return 0, 0, fmt.Errorf("消息过短: %d 字节", msgLen)
	}
	return msgLen, flags, nil
}

// EncodeFrame 将消息编码为协议帧
// isClientData=true 表示客户端→服务端方向（使用正确的密钥加密）
func EncodeFrame(msgID uint16, seqID uint32, flags byte, payload []byte, isClientData bool) []byte {
	// 组装消息体: [MsgID LE][SeqID LE][payload]
	bodyLen := MsgIDSize + SeqIDSize + len(payload)
	body := make([]byte, bodyLen)
	binary.LittleEndian.PutUint16(body[0:MsgIDSize], msgID)
	binary.LittleEndian.PutUint32(body[MsgIDSize:MsgIDSize+SeqIDSize], seqID)
	copy(body[MsgIDSize+SeqIDSize:], payload)

	// 压缩（如果启用且消息体足够大）
	if flags&FlagCompress != 0 && len(body) > 100 {
		body = Compress(body)
	}

	// 加密（如果启用）
	if flags&FlagEncrypt != 0 {
		EncryptXOR(body, isClientData)
	}

	// 组装完整帧: [3B bodyLen LE][1B flags][body]
	frame := make([]byte, FrameHeaderSize+len(body))
	// 3字节小端长度
	frame[0] = byte(len(body))
	frame[1] = byte(len(body) >> 8)
	frame[2] = byte(len(body) >> 16)
	frame[3] = flags
	copy(frame[FrameHeaderSize:], body)

	return frame
}

// RewriteSeqID 重写加密 body 中的 seqId（就地修改），同时返回 msgID
// body 是原始加密状态的 TCP 消息体，flags 是帧标志位
// 仅用于 client→server 方向
func RewriteSeqID(body []byte, flags byte, newSeqID uint32) (uint16, error) {
	if flags&FlagEncrypt != 0 {
		DecryptXOR(body, true)
	}
	if len(body) < MsgIDSize+SeqIDSize {
		return 0, fmt.Errorf("body too short for seqId rewrite: %d bytes", len(body))
	}
	msgID := binary.LittleEndian.Uint16(body[0:MsgIDSize])
	binary.LittleEndian.PutUint32(body[MsgIDSize:MsgIDSize+SeqIDSize], newSeqID)
	if flags&FlagEncrypt != 0 {
		EncryptXOR(body, true)
	}
	return msgID, nil
}
