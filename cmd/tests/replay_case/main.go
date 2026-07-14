package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	proto "github.com/gogo/protobuf/proto"

	sp "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// 从 test.json 读取消息，使用 streamproto 完全相同的逻辑发送
// 用法: go run . <openID>
// 示例: go run . test_cmd_001

func main() {
	httpAddr := "10.254.114.204:20144"
	caseFile := "cases/proto_cases/test.json"

	openID := fmt.Sprintf("test_cmd_%d", time.Now().Unix()%100000)
	if len(os.Args) > 1 {
		openID = os.Args[1]
	}

	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("=== 从 %s 重放消息到账号 %s ===", caseFile, openID)

	// 1. 读取 test.json
	messages := loadCaseFile(caseFile)
	log.Printf("加载 %d 条消息", len(messages))

	// 2. HTTP 认证
	token, serverAddr, err := authLogin(httpAddr, openID)
	if err != nil {
		log.Fatalf("AuthLogin 失败: %v", err)
	}
	log.Printf("AuthLogin: token=%s..., server=%s", token[:8], serverAddr)

	// 3. TCP 连接
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Fatalf("TCP 连接失败: %v", err)
	}
	defer conn.Close()
	log.Printf("TCP 已连接: %s", serverAddr)

	// 4. 发送 LoginReq（与 streamproto.sendLoginReq 完全一致）
	err = sendLoginReq(conn, openID, token, "")
	if err != nil {
		log.Fatalf("LoginReq 失败: %v", err)
	}
	log.Printf("LoginReq 已发送")

	// 5. 等待 LoginResp（与 streamproto.waitLoginResp 完全一致）
	err = waitLoginResp(conn)
	if err != nil {
		log.Fatalf("等待 LoginResp 失败: %v", err)
	}
	log.Printf("LoginResp 成功")

	// 6. 启动 readDrainer（与 streamproto 完全一致）
	stopReader := make(chan struct{})
	go readDrainer(conn, stopReader, nil, openID)

	// 7. 等待推送完成
	time.Sleep(2 * time.Second)

	// 8. 逐条发送消息（跳过 ← 方向的 Ack/Ntf）
	seqID := uint32(1)
	for i, msg := range messages {
		if msg.Direction != "" && msg.Direction != "→" && msg.Direction != sp.DirClientToServer {
			log.Printf("[%d/%d] 跳过 %s (direction=%q)", i+1, len(messages), msg.MsgName, msg.Direction)
			continue
		}

		// 使用 streamproto.sendRawMessage 完全相同的逻辑
		err = sendRawMessage(conn, msg.MsgID, msg.PayloadJSON, seqID)
		if err != nil {
			log.Printf("[%d/%d] ❌ %s 发送失败: %v", i+1, len(messages), msg.MsgName, err)
		} else {
			log.Printf("[%d/%d] → %s (MsgID=%d, SeqID=%d)", i+1, len(messages), msg.MsgName, msg.MsgID, seqID)
		}
		seqID++

		if sp.SendIntervalMs > 0 {
			time.Sleep(time.Duration(sp.SendIntervalMs) * time.Millisecond)
		}
	}

	// 9. 等待最后一批 Ack
	if sp.AckWaitMs > 0 {
		time.Sleep(time.Duration(sp.AckWaitMs) * time.Millisecond)
	}

	close(stopReader)
	log.Printf("=== 完成 ===")
}

// ========== 从 JSON 加载用例 ==========

// RecordMessageRaw 还原 test.json 中的消息（payload_json 是 object 而非 string 时用此格式）
type RecordMessageRaw struct {
	MsgID      uint16          `json:"msg_id"`
	MsgName    string          `json:"msg_name"`
	PayloadRaw json.RawMessage `json:"payload_json"`
	OffsetMs   int             `json:"offset_ms"`
	Direction  string          `json:"direction"`
	SeqID      uint32          `json:"seq_id"`
}

func loadCaseFile(path string) []sp.RecordMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}
	var rec struct {
		Messages []RecordMessageRaw `json:"messages"`
	}
	if err := json.Unmarshal(data, &rec); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
	}

	result := make([]sp.RecordMessage, len(rec.Messages))
	for i, raw := range rec.Messages {
		payloadStr := string(raw.PayloadRaw)
		// 如果 payload_json 是 object，重新 marshal 为 string
		if len(raw.PayloadRaw) > 0 && raw.PayloadRaw[0] == '{' {
			payloadStr = string(raw.PayloadRaw)
		}
		result[i] = sp.RecordMessage{
			MsgID:       raw.MsgID,
			MsgName:     raw.MsgName,
			PayloadJSON: payloadStr,
			OffsetMs:    raw.OffsetMs,
			Direction:   raw.Direction,
			SeqID:       raw.SeqID,
		}
	}
	return result
}

// ========== HTTP 认证（与 streamproto.AuthLogin 完全一致）==========

func authLogin(httpAddr, openID string) (token, serverAddr string, err error) {
	url := fmt.Sprintf("http://%s/authlogin", httpAddr)
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return "", "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code       int    `json:"code"`
		Token      string `json:"token"`
		ServerAddr string `json:"server_addr"`
	}
	json.Unmarshal(respBody, &result)
	if result.Code != 0 {
		return "", "", fmt.Errorf("auth失败 code=%d", result.Code)
	}
	return result.Token, result.ServerAddr, nil
}

// ========== LoginReq（与 streamproto.sendLoginReq 完全一致）==========

func sendLoginReq(conn net.Conn, openID, token, loginPayloadB64 string) error {
	var payload []byte

	if loginPayloadB64 == "" {
		var buf bytes.Buffer
		writeString(&buf, openID)
		writeString(&buf, token)
		binary.Write(&buf, binary.LittleEndian, uint64(0))
		writeString(&buf, "0.8.3001.1.1")
		binary.Write(&buf, binary.LittleEndian, uint16(0))
		binary.Write(&buf, binary.LittleEndian, uint16(0))
		binary.Write(&buf, binary.LittleEndian, uint16(0))
		binary.Write(&buf, binary.LittleEndian, uint32(0))
		writeString(&buf, "")
		writeString(&buf, "")
		payload = buf.Bytes()
		log.Printf("  LoginReq payload: %d bytes", len(payload))
	}

	frame := sp.EncodeFrame(1, 0, sp.FlagEncrypt, payload, true)
	_, err := conn.Write(frame)
	return err
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}

// ========== 等待 LoginResp（与 streamproto.waitLoginResp 完全一致）==========

func waitLoginResp(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		header := make([]byte, sp.FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			return fmt.Errorf("读取帧头失败: %v", err)
		}
		msgLen, _, err := sp.ParseFrameHeader(header)
		if err != nil {
			return fmt.Errorf("解析帧头失败: %v", err)
		}
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			return fmt.Errorf("读取消息体失败: %v", err)
		}
		raw := make([]byte, sp.FrameHeaderSize+msgLen)
		copy(raw, header)
		copy(raw[sp.FrameHeaderSize:], body)

		frame, err := sp.DecodeFrame(raw, false)
		if err != nil {
			return fmt.Errorf("解码失败: %v", err)
		}

		if frame.MsgID == 2 {
			if len(frame.Payload) >= 12 {
				result := binary.LittleEndian.Uint32(frame.Payload[8:12])
				if result != 0 {
					return fmt.Errorf("登录失败: Result=%d", result)
				}
			}
			return nil
		}
		// 跳过非 LoginResp 消息
	}
}

// ========== 发送消息（与 streamproto.sendRawMessage 完全一致）==========

func sendRawMessage(conn net.Conn, msgID uint16, payloadJSON string, seqID uint32) error {
	msg, ok := sp.NewMessage(msgID)
	if !ok {
		return fmt.Errorf("消息未注册: MsgID=%d", msgID)
	}
	if err := json.Unmarshal([]byte(payloadJSON), msg); err != nil {
		return fmt.Errorf("JSON反序列化失败: %v", err)
	}
	protoData, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Proto编码失败: %v", err)
	}
	payload := make([]byte, 2+len(protoData))
	payload[0] = byte(len(protoData))
	payload[1] = byte(len(protoData) >> 8)
	copy(payload[2:], protoData)
	frame := sp.EncodeFrame(msgID, seqID, sp.FlagEncrypt, payload, true)
	_, err = conn.Write(frame)
	return err
}

// ========== readDrainer（与 streamproto.readDrainer 完全一致）==========

func readDrainer(conn net.Conn, stop chan struct{}, onMessage sp.ReplayMessageCallback, accountID string) {
	count := 0
	for {
		select {
		case <-stop:
			return
		default:
		}

		header := make([]byte, sp.FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("[reader] 退出 (收到%d条): %v", count, err)
			return
		}
		count++
		msgLen, _, err := sp.ParseFrameHeader(header)
		if err != nil {
			return
		}
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			return
		}
		raw := make([]byte, sp.FrameHeaderSize+msgLen)
		copy(raw, header)
		copy(raw[sp.FrameHeaderSize:], body)
		frame, err := sp.DecodeFrame(raw, false)
		if err != nil {
			continue
		}

		msgName := sp.GetMsgName(frame.MsgID)
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
		log.Printf("  ← %s (MsgID=%d, %dB)", msgName, frame.MsgID, len(frame.Payload))
	}
}
