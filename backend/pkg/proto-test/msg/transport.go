package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	proto "github.com/gogo/protobuf/proto"
)

// EncodeClientMessage 将 JSON payload 编码为客户端→服务端的加密协议帧
func EncodeClientMessage(msgID uint16, seqID uint32, payloadJSON string) ([]byte, error) {
	msg, ok := NewMessage(msgID)
	if !ok {
		return nil, fmt.Errorf("消息未注册: MsgID=%d", msgID)
	}
	if err := json.Unmarshal([]byte(payloadJSON), msg); err != nil {
		return nil, fmt.Errorf("JSON 反序列化失败: %v", err)
	}
	protoData, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("proto 编码失败: %v", err)
	}
	payload := make([]byte, 2+len(protoData))
	payload[0] = byte(len(protoData))
	payload[1] = byte(len(protoData) >> 8)
	copy(payload[2:], protoData)
	return EncodeFrame(msgID, seqID, FlagEncrypt, payload, true), nil
}

// sendRawMessage 发送单条消息（从 JSON payload 直接编码发送）
func sendRawMessage(conn net.Conn, msgID uint16, payloadJSON string, seqID uint32) error {
	frame, err := EncodeClientMessage(msgID, seqID, payloadJSON)
	if err != nil {
		return err
	}
	if _, err := conn.Write(frame); err != nil {
		return fmt.Errorf("TCP 发送失败: %v", err)
	}
	return nil
}

// payloadToJSON 将原始 payload 反序列化为 JSON 字符串（MsgID >= 1000 的 Proto 消息）
func payloadToJSON(msgID uint16, payload []byte) (string, error) {
	if msgID < 1000 {
		return "", fmt.Errorf("框架消息不序列化")
	}
	msg, ok := NewMessage(msgID)
	if !ok {
		return "", fmt.Errorf("消息未注册: MsgID=%d", msgID)
	}
	var protoData []byte
	if len(payload) >= 2 {
		dataLen := int(payload[0]) | int(payload[1])<<8
		if dataLen <= len(payload)-2 {
			protoData = payload[2 : 2+dataLen]
		}
	}
	if len(protoData) > 0 {
		if err := proto.Unmarshal(protoData, msg); err != nil {
			return "", fmt.Errorf("反序列化失败: %v", err)
		}
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %v", err)
	}
	return string(jsonBytes), nil
}

// readDrainer 读取并消费服务端推送的消息，防止服务端写缓冲区满导致断连
// onMessage 非空时，同步将解码后的消息推送到前端
// accountID: 当前连接的账号标识
// done: 函数退出时 close(done) 通知调用方同步等待完成
//
// 并发安全与生命周期：
//   - 调用方通过 close(stop) 通知退出
//   - 若 readDrainer 阻塞在 io.ReadFull，调用方应设置 conn.SetReadDeadline(time.Now()) 强制中断
//   - 函数退出时 close(done) 通知 cleanup() 同步完成
//   - 必须在 done 收到信号后再归还连接，否则下一个 borrower 会与之并发读取同一 net.Conn
//
// 帧读取：按完整帧读取（header → ParseFrameHeader → body），保证帧边界正确。
func readDrainer(conn net.Conn, stop chan struct{}, done chan struct{}, onMessage ReplayMessageCallback, accountID string) {
	defer close(done)
	count := 0
	start := time.Now()
	for {
		select {
		case <-stop:
			return
		default:
		}

		// 按帧读取，不使用 SetReadDeadline（避免干扰共享连接）
		header := make([]byte, FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("[重放] 读取协程退出 (收到%d条, 耗时%v): %v", count, time.Since(start), err)
			return
		}
		count++

		msgLen, _, err := ParseFrameHeader(header)
		if err != nil {
			log.Printf("[重放] 帧头解析失败 (header hex: %x): %v", header, err)
			return
		}

		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Printf("[重放] 读取消息体失败 (msgLen=%d): %v", msgLen, err)
			return
		}

		// 解码帧以获取 MsgID
		raw := make([]byte, FrameHeaderSize+msgLen)
		copy(raw, header)
		copy(raw[FrameHeaderSize:], body)
		frame, err := DecodeFrame(raw, false)
		if err != nil {
			log.Printf("[重放] 解码失败 (raw hex: %x): %v", raw[:min(len(raw), 40)], err)
			continue
		}

		msgName := GetMsgName(frame.MsgID)
		if msgName == "" {
			switch frame.MsgID {
			case 2:
				msgName = "LoginResp"
			case 4:
				msgName = "Pong"
			case 10:
				msgName = "KickOut"
			default:
				msgName = fmt.Sprintf("Unknown(%d)", frame.MsgID)
			}
		}
		// 将服务端推送的消息推送到前端（Proto 消息才推送）
		if onMessage != nil && frame.MsgID >= 1000 {
			if jsonPayload, err := payloadToJSON(frame.MsgID, frame.Payload); err == nil {
				onMessage(msgName, frame.MsgID, frame.SeqID, jsonPayload, 0, DirServerToClient, accountID)
			}
		}
		log.Printf("[重放] ← %s (MsgID=%d, SeqID=%d, %dB)", msgName, frame.MsgID, frame.SeqID, len(frame.Payload))
	}
}
