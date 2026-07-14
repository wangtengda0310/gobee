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

// 完全模拟 wails3 的行为：
// - 使用前端配置的固定 serverAddr（不是 AuthLogin 分配的）
// - 连续多个账号，每个独立 HTTP 登录 + TCP 连接 + 发送消息 + 断开
// - sendMessagesOnce 的完整流程

func main() {
	frontendServer := "10.254.114.204:18000" // 模拟前端配置
	httpAddr := "10.254.114.204:20144"
	openID := "simtest"
	rangeStart := 20
	rangeEnd := 30

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// ====== 阶段 1: 批量创号 ======
	log.Println("===== 阶段 1: 批量创号 =====")
	caseMessages := loadCaseFile("cases/proto_cases/创号.json")
	for i := rangeStart; i <= rangeEnd; i++ {
		accountID := fmt.Sprintf("%s%d", openID, i)
		log.Printf("--- [%d/%d] account=%s ---", i-rangeStart+1, rangeEnd-rangeStart+1, accountID)
		uid, err := replayOnce(frontendServer, httpAddr, accountID, caseMessages)
		if err != nil {
			log.Printf("❌ %s: %v", accountID, err)
		} else {
			log.Printf("✓ %s: UID=%d", accountID, uid)
		}
	}

	time.Sleep(3 * time.Second)

	// ====== 阶段 2: 批量加黄金 ======
	log.Println("===== 阶段 2: 批量添加黄金 =====")
	goldMessages := loadCaseFile("cases/proto_cases/添加黄金.json")
	for i := rangeStart; i <= rangeEnd; i++ {
		accountID := fmt.Sprintf("%s%d", openID, i)
		log.Printf("--- [%d/%d] account=%s ---", i-rangeStart+1, rangeEnd-rangeStart+1, accountID)
		uid, err := replayOnce(frontendServer, httpAddr, accountID, goldMessages)
		if err != nil {
			log.Printf("❌ %s: %v", accountID, err)
		} else {
			log.Printf("✓ %s: UID=%d", accountID, uid)
		}
	}
}

func replayOnce(serverAddr, httpAddr, accountID string, messages []sp.RecordMessage) (uint64, error) {
	// HTTP 认证
	token, err := authLogin(httpAddr, accountID)
	if err != nil {
		return 0, fmt.Errorf("auth: %v", err)
	}

	// TCP 连接（用固定地址，和 wails3 一样）
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		return 0, fmt.Errorf("tcp: %v", err)
	}
	defer conn.Close()

	// LoginReq
	loginPayload := buildLoginReq(accountID, token)
	frame := sp.EncodeFrame(1, 0, sp.FlagEncrypt, loginPayload, true)
	conn.Write(frame)

	// LoginResp
	uid := readLoginResp(conn)
	if uid == 0 {
		// 有错误码
		return 0, fmt.Errorf("LoginResp Result=%d", lastLoginResult)
	}

	// reader
	stop := make(chan struct{})
	go readDrainer(conn, stop)
	defer close(stop)

	time.Sleep(2 * time.Second)

	// 发送消息
	seqID := uint32(1)
	for _, msg := range messages {
		if msg.Direction != "" && msg.Direction != "→" && msg.Direction != sp.DirClientToServer {
			continue
		}
		sendRawMessage(conn, msg.MsgID, msg.PayloadJSON, seqID)
		seqID++
		if sp.SendIntervalMs > 0 {
			time.Sleep(time.Duration(sp.SendIntervalMs) * time.Millisecond)
		}
	}

	if sp.AckWaitMs > 0 {
		time.Sleep(time.Duration(sp.AckWaitMs) * time.Millisecond)
	}

	return uid, nil
}

var lastLoginResult uint32

func readLoginResp(conn net.Conn) uint64 {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})
	for {
		hdr := make([]byte, 4)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			return 0
		}
		ml := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
		bd := make([]byte, ml)
		io.ReadFull(conn, bd)
		raw := make([]byte, 4+len(bd))
		copy(raw, hdr)
		copy(raw[4:], bd)
		f, err := sp.DecodeFrame(raw, false)
		if err != nil {
			continue
		}
		if f.MsgID == 2 && len(f.Payload) >= 12 {
			uid := binary.LittleEndian.Uint64(f.Payload[0:8])
			lastLoginResult = binary.LittleEndian.Uint32(f.Payload[8:12])
			return uid
		}
	}
}

// ========== helpers ==========

type rawMsg struct {
	MsgID      uint16          `json:"msg_id"`
	MsgName    string          `json:"msg_name"`
	PayloadRaw json.RawMessage `json:"payload_json"`
	OffsetMs   int             `json:"offset_ms"`
	Direction  string          `json:"direction"`
	SeqID      uint32          `json:"seq_id"`
}

func loadCaseFile(path string) []sp.RecordMessage {
	data, _ := os.ReadFile(path)
	var rec struct {
		Messages []rawMsg `json:"messages"`
	}
	json.Unmarshal(data, &rec)
	result := make([]sp.RecordMessage, len(rec.Messages))
	for i, raw := range rec.Messages {
		result[i] = sp.RecordMessage{
			MsgID: raw.MsgID, MsgName: raw.MsgName, PayloadJSON: string(raw.PayloadRaw),
			OffsetMs: raw.OffsetMs, Direction: raw.Direction, SeqID: raw.SeqID,
		}
	}
	return result
}

func authLogin(httpAddr, openID string) (string, error) {
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, err := http.Post("http://"+httpAddr+"/authlogin", "application/json", bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var r struct {
		Code  int    `json:"code"`
		Token string `json:"token"`
	}
	json.Unmarshal(data, &r)
	if r.Code != 0 {
		return "", fmt.Errorf("code=%d", r.Code)
	}
	return r.Token, nil
}

func buildLoginReq(openID, token string) []byte {
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

func sendRawMessage(conn net.Conn, msgID uint16, payloadJSON string, seqID uint32) error {
	msg, ok := sp.NewMessage(msgID)
	if !ok {
		return fmt.Errorf("msg not found: %d", msgID)
	}
	json.Unmarshal([]byte(payloadJSON), msg)
	pd, _ := proto.Marshal(msg)
	payload := make([]byte, 2+len(pd))
	payload[0] = byte(len(pd))
	payload[1] = byte(len(pd) >> 8)
	copy(payload[2:], pd)
	frame := sp.EncodeFrame(msgID, seqID, sp.FlagEncrypt, payload, true)
	conn.Write(frame)
	return nil
}

func readDrainer(conn net.Conn, stop chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		hdr := make([]byte, 4)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			return
		}
		ml := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
		bd := make([]byte, ml)
		io.ReadFull(conn, bd)
	}
}
