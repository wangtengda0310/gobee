// Package frame 提供简单的「4字节长度前缀 + protobuf」帧编解码。
//
// 这是 server / client / sniffer 共享的帧协议：
//
//	┌─── 4 字节包头 ───┐ ┌────── payload（变长）──────┐
//	│ msgType(2B LE)  │ │ msgLen(2B LE) │ protobuf... │
//	└─────────────────┘ └───────────────┴─────────────┘
//
// msgType: 消息类型枚举（见 MsgTypeEcho 等）。
// msgLen:  protobuf 数据的字节长度。
// protobuf: proto.Marshal 编码的消息体。
//
// 这个格式刻意简单（无加密/压缩），用于演示 Buf + pcap 的完整链路。
package frame

import (
	"encoding/binary"
	"fmt"
	"io"
)

// 消息类型枚举。
const (
	MsgTypeEchoRequest  uint16 = 1 // 客户端→服务端：回显请求
	MsgTypeEchoResponse uint16 = 2 // 服务端→客户端：回显响应
	MsgTypeSumRequest   uint16 = 3 // 客户端→服务端：求和请求
	MsgTypeSumResponse  uint16 = 4 // 服务端→客户端：求和响应
)

// MsgTypeName 返回消息类型的可读名称。
func MsgTypeName(t uint16) string {
	switch t {
	case MsgTypeEchoRequest:
		return "EchoRequest"
	case MsgTypeEchoResponse:
		return "EchoResponse"
	case MsgTypeSumRequest:
		return "SumRequest"
	case MsgTypeSumResponse:
		return "SumResponse"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// HeaderSize 包头固定大小：2字节 msgType + 2字节 msgLen。
const HeaderSize = 4

// MaxPayload 单条消息 payload 上限（uint16 最大值）。
const MaxPayload = 0xFFFF

// WriteFrame 向 io.Writer 写入一帧完整消息。
//
// 参数：
//   - w: 目标 writer（通常是 net.Conn）
//   - msgType: 消息类型枚举
//   - payload: protobuf 编码后的消息体
func WriteFrame(w io.Writer, msgType uint16, payload []byte) error {
	header := make([]byte, HeaderSize)
	binary.LittleEndian.PutUint16(header[0:2], msgType)
	binary.LittleEndian.PutUint16(header[2:4], uint16(len(payload)))
	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	return nil
}

// ReadFrame 从 io.Reader 读取一帧完整消息。
//
// 返回 msgType、payload（protobuf 字节）和 error。
// io.EOF / 连接关闭时返回 error。
func ReadFrame(r io.Reader) (msgType uint16, payload []byte, err error) {
	header := make([]byte, HeaderSize)
	if _, err = io.ReadFull(r, header); err != nil {
		return 0, nil, fmt.Errorf("read header: %w", err)
	}
	msgType = binary.LittleEndian.Uint16(header[0:2])
	msgLen := binary.LittleEndian.Uint16(header[2:4])
	if msgLen > MaxPayload {
		return 0, nil, fmt.Errorf("payload too large: %d", msgLen)
	}
	payload = make([]byte, msgLen)
	if _, err = io.ReadFull(r, payload); err != nil {
		return 0, nil, fmt.Errorf("read payload: %w", err)
	}
	return msgType, payload, nil
}
