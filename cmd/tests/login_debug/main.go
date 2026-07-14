package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// ============================================================
// 工具：对比 streamproto sendLoginReq 与真实 rain-robot LoginReq
// ============================================================
// 用法：
//   go run main.go <openID>
// 示例：
//   go run main.go test9
//
// 功能：
//   1. HTTP 认证获取 token
//   2. 使用 sendLoginReq 构造 LoginReq 并登录
//   3. 创建角色
//   4. 输出完整 LoginReq payload 的 hex dump 供对比

func main() {
	openID := "test9"
	if len(os.Args) > 1 {
		openID = os.Args[1]
	}

	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("=== LoginReq 对比工具 ===")
	log.Printf("openID: %s, httpAddr: %s", openID, httpAddr)

	// Step 1: HTTP 认证
	token, serverAddr, authErr := authLogin(httpAddr, openID)
	if authErr != nil {
		log.Fatalf("HTTP 认证失败: %v", authErr)
	}
	log.Printf("HTTP 认证成功: token=%s, serverAddr=%s", token[:16]+"...", serverAddr)

	// Step 2: 构造 LoginReq payload（使用 sendLoginReq 相同的逻辑）
	selfPayload := buildLoginReqPayload(openID, token)
	log.Printf("自构造 LoginReq payload: %d bytes", len(selfPayload))
	log.Printf("Payload hex: %x", selfPayload)

	// 打印字段级分析
	analyzeLoginReqPayload(selfPayload, "自构造")

	// Step 3: TCP 连接并发送 LoginReq
	conn, dialErr := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if dialErr != nil {
		log.Fatalf("TCP 连接失败: %v", dialErr)
	}
	defer conn.Close()
	log.Printf("TCP 已连接: %s", serverAddr)

	// 发送 LoginReq
	frame := EncodeFrame(1, 0, FlagEncrypt, selfPayload, true)
	if _, err := conn.Write(frame); err != nil {
		log.Fatalf("发送 LoginReq 失败: %v", err)
	}
	log.Printf("LoginReq 已发送 (%d bytes frame)", len(frame))

	// Step 4: 等待 LoginResp
	loginResult := waitLoginResp(conn)
	if loginResult != 0 {
		log.Printf("⚠ LoginResp Result=%d (非零表示登录可能有问题)", loginResult)
	} else {
		log.Printf("✓ LoginResp Result=0 登录成功")
	}

	// Step 5: 启动 reader 消费推送
	stopReader := make(chan struct{})
	go readDrainer(conn, stopReader)
	defer close(stopReader)
	time.Sleep(2 * time.Second)

	// Step 6: 发送 CreateRoleReq
	roleName := fmt.Sprintf("测试%s", openID)
	sendCreateRoleReq(conn, roleName)
	log.Printf("CreateRoleReq 已发送: roleName=%s", roleName)

	// 等待响应
	time.Sleep(5 * time.Second)
	log.Printf("完成，连接关闭")

	// 如果有录制的 LoginReq payload (login_payload_b64)，也输出对比
	log.Println("")
	log.Println("=== 对比说明 ===")
	log.Println("如果上面的自构造 LoginReq payload 能成功登录(Result=0)，")
	log.Println("说明 sendLoginReq 的结构是正确的。")
	log.Println("需要关注的是: 服务端推送了哪些消息，")
	log.Println("以及 CreateRoleAck 的具体内容。")
}

// ========== 辅助函数 ==========

func authLogin(httpAddr, openID string) (token, serverAddr string, err error) {
	url := fmt.Sprintf("http://%s/authlogin", httpAddr)
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return "", "", fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	var result struct {
		Code       int    `json:"code"`
		Token      string `json:"token"`
		ServerAddr string `json:"server_addr"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}
	if result.Code != 0 {
		return "", "", fmt.Errorf("登录失败: code=%d", result.Code)
	}
	return result.Token, result.ServerAddr, nil
}

// buildLoginReqPayload 使用与 sendLoginReq 完全相同的逻辑构造 LoginReq payload
func buildLoginReqPayload(openID, token string) []byte {
	var buf bytes.Buffer

	writeString(&buf, openID)                          // Account
	writeString(&buf, token)                           // Token
	binary.Write(&buf, binary.LittleEndian, uint64(0)) // UID = 0
	writeString(&buf, "0.8.3001.1.1")                  // Version
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // Metadata (map count=0)
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // ExtData ([]byte len=0)
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // ReqType = 0
	binary.Write(&buf, binary.LittleEndian, uint32(0)) // SeqID = 0
	writeString(&buf, "")                              // Entity = ""
	writeString(&buf, "")                              // Sign = ""

	return buf.Bytes()
}

// analyzeLoginReqPayload 分析 LoginReq payload 的字段级结构
func analyzeLoginReqPayload(payload []byte, label string) {
	br := bytes.NewReader(payload)
	readStr := func() string {
		var l uint16
		binary.Read(br, binary.LittleEndian, &l)
		s := make([]byte, l)
		br.Read(s)
		return string(s)
	}

	account := readStr()
	token := readStr()
	tokenDisplay := token
	if len(tokenDisplay) > 16 {
		tokenDisplay = tokenDisplay[:16] + "..."
	}
	var uid uint64
	binary.Read(br, binary.LittleEndian, &uid)
	version := readStr()
	var metaCount uint16
	binary.Read(br, binary.LittleEndian, &metaCount)
	var extLen uint16
	binary.Read(br, binary.LittleEndian, &extLen)
	var reqType uint16
	binary.Read(br, binary.LittleEndian, &reqType)
	var seqID uint32
	binary.Read(br, binary.LittleEndian, &seqID)
	entity := readStr()
	sign := readStr()

	log.Printf("[%s] Account=%q Token=%s UID=%d Version=%q MetadataCount=%d ExtLen=%d ReqType=%d SeqID=%d Entity=%q Sign=%q",
		label, account, tokenDisplay, uid, version, metaCount, extLen, reqType, seqID, entity, sign)
	log.Printf("[%s] 剩余未读取字节: %d (应为0)", label, br.Len())
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}

func sendCreateRoleReq(conn net.Conn, roleName string) {
	// CreateRoleReq 是 proto 消息 (MsgID=1006)
	// 字段: gender=2, isRandom=true, itemID=1020502, roleName
	// 使用 protowire 手动编码
	payload := marshalCreateRoleReq(roleName, 2, 1020502, true)

	// proto 消息需要 ByteStream 包装: [2B len][proto data]
	wrapped := make([]byte, 2+len(payload))
	binary.LittleEndian.PutUint16(wrapped, uint16(len(payload)))
	copy(wrapped[2:], payload)

	frame := EncodeFrame(1006, 1, FlagEncrypt, wrapped, true)
	conn.Write(frame)
}

// marshalCreateRoleReq protobuf 手动编码 CreateRoleReq
//
//	message CreateRoleReq {
//	  string roleName = 1;  // field 1, wire type 2
//	  int32 gender = 2;     // field 2, wire type 0
//	  bool isRandom = 3;    // field 3, wire type 0
//	  int32 itemID = 4;     // field 4, wire type 0
//	}
func marshalCreateRoleReq(roleName string, gender int32, itemID int32, isRandom bool) []byte {
	var buf bytes.Buffer

	// field 1: roleName (string, wire type 2)
	// tag = (1 << 3) | 2 = 0x0a
	buf.WriteByte(0x0a)
	// varint length
	writeVarint(&buf, uint64(len(roleName)))
	buf.WriteString(roleName)

	// field 2: gender (int32, wire type 0)
	// tag = (2 << 3) | 0 = 0x10
	buf.WriteByte(0x10)
	writeVarint(&buf, uint64(gender))

	// field 3: isRandom (bool, wire type 0)
	buf.WriteByte(0x18) // (3<<3)|0 = 0x18
	if isRandom {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	// field 4: itemID (int32, wire type 0)
	buf.WriteByte(0x20) // (4<<3)|0 = 0x20
	writeVarint(&buf, uint64(itemID))

	return buf.Bytes()
}

func writeVarint(buf *bytes.Buffer, v uint64) {
	for v >= 0x80 {
		buf.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	buf.WriteByte(byte(v))
}

// ========== 网络层（与 rain-robot 和 streamproto 一致）==========
// 密钥和算法直接复制自 streamproto/crypto.go 和 rain-robot encrypt.go

// Frame header constants
const (
	FrameHeaderSize = 4
	FlagEncrypt     = 0x02
)

var encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}
var decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}

func EncodeFrame(msgID uint16, seqID uint32, flags byte, payload []byte, isClientData bool) []byte {
	// msgLen = MsgID(2) + SeqID(4) + payload
	msgLen := 2 + 4 + len(payload)
	body := make([]byte, msgLen)
	binary.LittleEndian.PutUint16(body[0:2], msgID)
	binary.LittleEndian.PutUint32(body[2:6], seqID)
	copy(body[6:], payload)

	// XOR encrypt (same as streamproto EncodeFrame)
	if flags&FlagEncrypt != 0 {
		EncryptXOR(body, isClientData)
	}

	// 帧: [3B len][1B flags][body]
	frame := make([]byte, FrameHeaderSize+msgLen)
	frame[0] = byte(msgLen)
	frame[1] = byte(msgLen >> 8)
	frame[2] = byte(msgLen >> 16)
	frame[3] = flags
	copy(frame[4:], body)
	return frame
}

func EncryptXOR(data []byte, isClientData bool) {
	var key []byte
	if isClientData {
		key = decryptKey
	} else {
		key = encryptKey
	}
	keylen := len(key)
	for i := 0; i < len(data); i++ {
		n := byte(i%7 + 1)
		b := data[i]
		b = (b << n) | (b >> (8 - n)) // 循环左移 n 位
		data[i] = b ^ key[i%keylen]
	}
}

func DecryptXOR(data []byte, isClientData bool) {
	var key []byte
	if isClientData {
		key = decryptKey
	} else {
		key = encryptKey
	}
	keylen := len(key)
	for i := 0; i < len(data); i++ {
		b := data[i] ^ key[i%keylen]
		n := byte(i%7 + 1)
		data[i] = (b >> n) | (b << (8 - n)) // 循环右移 n 位
	}
}

func waitLoginResp(conn net.Conn) uint32 {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		header := make([]byte, FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("读取 LoginResp 帧头失败: %v", err)
			return 0xFFFFFFFF
		}

		msgLen := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Printf("读取 LoginResp body 失败: %v", err)
			return 0xFFFFFFFF
		}

		// 解密
		DecryptXOR(body, false) // false = 服务端发送的数据

		msgID := uint16(body[0]) | uint16(body[1])<<8
		_ = uint32(body[2]) | uint32(body[3])<<8 | uint32(body[4])<<8 | uint32(body[5])<<8 // seqID
		payload := body[6:]

		if msgID == 2 {
			// LoginResp: UID(8) + Result(4) + ErrStr + Metadata + ExtData + ZoneID(4) + HttpToken
			if len(payload) >= 12 {
				result := binary.LittleEndian.Uint32(payload[8:12])
				log.Printf("LoginResp: UID=%d, Result=%d, payloadLen=%d",
					binary.LittleEndian.Uint64(payload[0:8]), result, len(payload))
				return result
			}
		}

		msgName := fmt.Sprintf("Unknown(%d)", msgID)
		switch msgID {
		case 2:
			msgName = "LoginResp"
		case 4:
			msgName = "Pong"
		case 10:
			msgName = "KickOut"
		}
		log.Printf("← %s (MsgID=%d, %dB) [等待LoginResp中]", msgName, msgID, len(payload))
	}
}

func readDrainer(conn net.Conn, stop chan struct{}) {
	count := 0
	for {
		select {
		case <-stop:
			return
		default:
		}

		header := make([]byte, FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("[reader] 退出 (收到%d条): %v", count, err)
			return
		}
		count++

		msgLen := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Printf("[reader] 读取body失败: %v", err)
			return
		}

		// Decrypt
		DecryptXOR(body, false)

		msgID := uint16(body[0]) | uint16(body[1])<<8
		_ = uint32(body[2]) | uint32(body[3])<<8 | uint32(body[4])<<8 | uint32(body[5])<<8 // seqID
		payload := body[6:]

		msgName := fmt.Sprintf("Unknown(%d)", msgID)
		switch msgID {
		case 2:
			msgName = "LoginResp"
		case 4:
			msgName = "Pong"
		case 10:
			msgName = "KickOut"
		case 1007:
			msgName = "CreateRoleAck"
		case 1016:
			msgName = "RefreshUserDataNtf"
		case 1017:
			msgName = "LoginServerMsgOverNtf"
		}

		// 对于 proto 消息 (ID >= 1000)，解析 payload
		payloadPreview := ""
		if msgID >= 1000 && len(payload) >= 2 {
			protoLen := int(payload[0]) | int(payload[1])<<8
			if protoLen <= len(payload)-2 {
				protoData := payload[2 : 2+protoLen]
				payloadPreview = fmt.Sprintf(" protoLen=%d data=%x", protoLen, protoData[:min(len(protoData), 64)])
			}
		} else if msgID < 1000 {
			payloadPreview = fmt.Sprintf(" hex=%x", payload[:min(len(payload), 64)])
		}

		log.Printf("← %s (MsgID=%d, %dB)%s", msgName, msgID, len(payload), payloadPreview)
	}
}

// 需要 context 包
var _ = context.Background()
