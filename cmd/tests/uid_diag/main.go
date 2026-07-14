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
)

// 诊断目标：两次用同一种方式登录同一个账号，看 UID 是否一致

func main() {
	openID := "diag_test9"
	if len(os.Args) > 1 {
		openID = os.Args[1]
	}
	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// 第一次登录
	log.Println("===== 第1次登录 =====")
	uid1 := doLogin(openID, httpAddr)
	if uid1 == 0 {
		log.Fatal("第一次登录失败")
	}
	log.Printf("第1次登录 UID=%d", uid1)

	// 第二次登录同一个账号
	time.Sleep(2 * time.Second)
	log.Println("===== 第2次登录 =====")
	uid2 := doLogin(openID, httpAddr)
	if uid2 == 0 {
		log.Fatal("第二次登录失败")
	}
	log.Printf("第2次登录 UID=%d", uid2)

	// 比较
	if uid1 == uid2 {
		log.Printf("✅ UID 一致 (%d) — accountid 映射正常", uid1)
	} else {
		log.Printf("❌ UID 不一致! 第1次=%d, 第2次=%d — accountid 映射有问题!", uid1, uid2)
	}
}

func doLogin(openID, httpAddr string) uint64 {
	// HTTP 认证
	token, serverAddr, err := authLogin(httpAddr, openID)
	if err != nil {
		log.Printf("HTTP认证失败: %v", err)
		return 0
	}

	// TCP 连接
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Printf("TCP连接失败: %v", err)
		return 0
	}
	defer conn.Close()

	// 发送 LoginReq (与 sendLoginReq 完全一致的逻辑)
	payload := buildLoginReqPayload(openID, token)
	frame := EncodeFrame(1, 0, FlagEncrypt, payload, true)
	if _, err := conn.Write(frame); err != nil {
		log.Printf("发送LoginReq失败: %v", err)
		return 0
	}

	// 读取 LoginResp
	uid, result := readLoginResp(conn)
	log.Printf("LoginResp: UID=%d, Result=%d", uid, result)
	return uid
}

// ========== 网络层（直接复制 streamproto）==========

const (
	FrameHeaderSize = 4
	FlagEncrypt     = 0x02
)

var encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}
var decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}

func EncodeFrame(msgID uint16, seqID uint32, flags byte, payload []byte, isClientData bool) []byte {
	msgLen := 2 + 4 + len(payload)
	body := make([]byte, msgLen)
	binary.LittleEndian.PutUint16(body[0:2], msgID)
	binary.LittleEndian.PutUint32(body[2:6], seqID)
	copy(body[6:], payload)

	if flags&FlagEncrypt != 0 {
		EncryptXOR(body, isClientData)
	}

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
	for i := range data {
		n := byte(i%7 + 1)
		b := data[i]
		b = (b << n) | (b >> (8 - n))
		data[i] = b ^ key[i%len(key)]
	}
}

func DecryptXOR(data []byte, isClientData bool) {
	var key []byte
	if isClientData {
		key = decryptKey
	} else {
		key = encryptKey
	}
	for i := range data {
		b := data[i] ^ key[i%len(key)]
		n := byte(i%7 + 1)
		data[i] = (b >> n) | (b << (8 - n))
	}
}

// ========== LoginReq ==========

func buildLoginReqPayload(openID, token string) []byte {
	var buf bytes.Buffer
	writeString(&buf, openID)
	writeString(&buf, token)
	binary.Write(&buf, binary.LittleEndian, uint64(0))
	writeString(&buf, "0.8.3001.1.1")
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // Metadata
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // ExtData
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // ReqType
	binary.Write(&buf, binary.LittleEndian, uint32(0)) // SeqID
	writeString(&buf, "")                              // Entity
	writeString(&buf, "")                              // Sign
	return buf.Bytes()
}

func writeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}

// ========== HTTP 认证 ==========

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
		return "", "", fmt.Errorf("code=%d", result.Code)
	}
	return result.Token, result.ServerAddr, nil
}

// ========== 读取 LoginResp ==========

func readLoginResp(conn net.Conn) (uid uint64, result uint32) {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		header := make([]byte, FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			log.Printf("读帧头失败: %v", err)
			return 0, 0xFFFFFFFF
		}
		msgLen := int(header[0]) | int(header[1])<<8 | int(header[2])<<16
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Printf("读body失败: %v", err)
			return 0, 0xFFFFFFFF
		}

		DecryptXOR(body, false)

		msgID := uint16(body[0]) | uint16(body[1])<<8
		payload := body[6:]

		if msgID == 2 && len(payload) >= 12 {
			uid = binary.LittleEndian.Uint64(payload[0:8])
			result = binary.LittleEndian.Uint32(payload[8:12])
			return
		}
		// skip non-LoginResp messages
	}
}
