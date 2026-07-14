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

// 验证脚本：对新账号先重放创号.json，再重放添加黄金.json
// 用法: go run . [openID]
// 默认自动生成新账号名

func main() {
	httpAddr := "10.254.114.204:20144"
	caseDir := "cases/proto_cases"

	openID := fmt.Sprintf("test_verify_%d", time.Now().Unix()%100000)
	if len(os.Args) > 1 {
		openID = os.Args[1]
	}

	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("=== 验证脚本: 账号=%s ===", openID)

	// ====== 阶段 1: 创号 ======
	log.Println("===== 阶段 1: 重放 创号.json =====")
	uid1 := replayFile(caseDir+"/创号.json", httpAddr, openID)
	log.Printf("阶段1完成: UID=%d", uid1)

	time.Sleep(3 * time.Second) // 等待跨实例同步

	// ====== 阶段 2: 添加黄金 ======
	log.Println("===== 阶段 2: 重放 添加黄金.json =====")
	uid2 := replayFile(caseDir+"/添加黄金.json", httpAddr, openID)
	log.Printf("阶段2完成: UID=%d", uid2)

	if uid1 == uid2 {
		log.Printf("✅ UID 一致 (%d) — 两次登录复用同一账号，GM 命令应已生效", uid1)
	} else {
		log.Printf("❌ UID 不一致! (%d vs %d) — 账号映射有问题", uid1, uid2)
	}

	log.Printf("=== 请在游戏客户端登录 %s 验证创角+黄金是否生效 ===", openID)
}

// ========== 重放核心（完整复用 streamproto 逻辑）==========

func replayFile(filePath, httpAddr, openID string) uint64 {
	// 1. 加载 JSON
	messages := loadCaseFile(filePath)
	log.Printf("  加载 %d 条消息", len(messages))

	// 2. HTTP 认证
	token, serverAddr, err := authLogin(httpAddr, openID)
	if err != nil {
		log.Fatalf("  AuthLogin 失败: %v", err)
	}
	log.Printf("  AuthLogin: server=%s token=%s...", serverAddr, token[:8])

	// 3. TCP 连接
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Fatalf("  TCP 连接失败: %v", err)
	}
	defer conn.Close()
	log.Printf("  TCP 已连接: %s", serverAddr)

	// 4. LoginReq
	loginPayload := buildLoginReqPayload(openID, token)
	frame := sp.EncodeFrame(1, 0, sp.FlagEncrypt, loginPayload, true)
	conn.Write(frame)
	log.Printf("  LoginReq 已发送 (%d bytes)", len(loginPayload))

	// 5. 读取 LoginResp（跳过非 LoginResp 的推送）
	uid := readLoginResp(conn)
	log.Printf("  LoginResp: UID=%d", uid)

	// 6. 启动 readDrainer
	stopReader := make(chan struct{})
	go readDrainer(conn, stopReader)
	defer close(stopReader)

	// 7. 等待推送完成
	time.Sleep(2 * time.Second)

	// 8. 逐条发送消息（跳过 ← 方向）
	seqID := uint32(1)
	for i, msg := range messages {
		if msg.Direction != "" && msg.Direction != "→" && msg.Direction != sp.DirClientToServer {
			log.Printf("  [%d/%d] 跳过 %s (←)", i+1, len(messages), msg.MsgName)
			continue
		}
		if err := sendRawMessage(conn, msg.MsgID, msg.PayloadJSON, seqID); err != nil {
			log.Printf("  [%d/%d] ❌ %s: %v", i+1, len(messages), msg.MsgName, err)
		} else {
			log.Printf("  [%d/%d] → %s (SeqID=%d)", i+1, len(messages), msg.MsgName, seqID)
		}
		seqID++
		if sp.SendIntervalMs > 0 {
			time.Sleep(time.Duration(sp.SendIntervalMs) * time.Millisecond)
		}
	}

	// 9. 等待最后 Ack
	if sp.AckWaitMs > 0 {
		time.Sleep(time.Duration(sp.AckWaitMs) * time.Millisecond)
	}

	return uid
}

// ========== 读取用例文件 ==========

type rawMsg struct {
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
		Messages []rawMsg `json:"messages"`
	}
	if err := json.Unmarshal(data, &rec); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
	}
	result := make([]sp.RecordMessage, len(rec.Messages))
	for i, raw := range rec.Messages {
		payloadStr := string(raw.PayloadRaw)
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
		return "", "", err
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
		return "", "", fmt.Errorf("code=%d body=%s", result.Code, string(respBody))
	}
	return result.Token, result.ServerAddr, nil
}

// ========== LoginReq ==========

func buildLoginReqPayload(openID, token string) []byte {
	var buf bytes.Buffer
	writeStr(&buf, openID)
	writeStr(&buf, token)
	binary.Write(&buf, binary.LittleEndian, uint64(0))
	writeStr(&buf, "0.8.3001.1.1")
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	writeStr(&buf, "")
	writeStr(&buf, "")
	return buf.Bytes()
}

func writeStr(b *bytes.Buffer, s string) {
	binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}

// ========== 读取 LoginResp ==========

func readLoginResp(conn net.Conn) uint64 {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		header := make([]byte, 4)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Fatalf("  读取 LoginResp 帧头失败: %v", err)
		}
		msgLen := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Fatalf("  读取 LoginResp body 失败: %v", err)
		}
		raw := make([]byte, 4+len(body))
		copy(raw, header)
		copy(raw[4:], body)
		frame, err := sp.DecodeFrame(raw, false)
		if err != nil {
			continue
		}
		if frame.MsgID == 2 && len(frame.Payload) >= 12 {
			uid := binary.LittleEndian.Uint64(frame.Payload[0:8])
			result := binary.LittleEndian.Uint32(frame.Payload[8:12])
			if result != 0 {
				log.Fatalf("  LoginResp Result=%d (登录失败)", result)
			}
			return uid
		}
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

// ========== readDrainer ==========

func readDrainer(conn net.Conn, stop chan struct{}) {
	count := 0
	for {
		select {
		case <-stop:
			return
		default:
		}
		header := make([]byte, 4)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("  [reader] 退出 (收到%d条): %v", count, err)
			return
		}
		count++
		ml := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
		body := make([]byte, ml)
		io.ReadFull(conn, body)
		raw := make([]byte, 4+len(body))
		copy(raw, header)
		copy(raw[4:], body)
		frame, _ := sp.DecodeFrame(raw, false)

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
		// 只打印关键 Ack
		switch frame.MsgID {
		case 1007, 1002, 1026, 1017:
			log.Printf("  ← %s (MsgID=%d, %dB)", msgName, frame.MsgID, len(frame.Payload))
		default:
			if frame.MsgID == 1016 && len(frame.Payload) <= 3 {
				log.Printf("  ← RefreshUserDataNtf (新账号, isHaveRole=false)")
			}
		}
	}
}
