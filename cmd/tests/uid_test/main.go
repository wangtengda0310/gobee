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
	"time"
)

// 最小诊断：同一个 openID 登录两次，对比 UID
// 不做任何假设，只看事实

func main() {
	openID := "diag_uid"
	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// 第一次：HTTP 认证 + TCP 登录
	token1, addr1 := auth(httpAddr, openID)
	log.Printf("第1次: token=%s... server=%s", token1[:8], addr1)

	uid1 := login(addr1, openID, token1)
	log.Printf("第1次: LoginResp.UID=%d", uid1)

	time.Sleep(2 * time.Second)

	// 第二次：同样操作
	token2, addr2 := auth(httpAddr, openID)
	log.Printf("第2次: token=%s... server=%s", token2[:8], addr2)

	uid2 := login(addr2, openID, token2)
	log.Printf("第2次: LoginResp.UID=%d", uid2)

	if uid1 == uid2 {
		log.Printf("✅ 两次 UID 一致: %d", uid1)
	} else {
		log.Printf("❌ UID 不一致: %d vs %d", uid1, uid2)
	}
}

func auth(httpAddr, openID string) (token, serverAddr string) {
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, _ := http.Post("http://"+httpAddr+"/authlogin", "application/json",
		bytes.NewBufferString(body))
	if resp != nil {
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		var r struct {
			Code       int    `json:"code"`
			Token      string `json:"token"`
			ServerAddr string `json:"server_addr"`
		}
		json.Unmarshal(data, &r)
		if r.Code == 0 {
			return r.Token, r.ServerAddr
		}
		log.Fatalf("auth 失败: code=%d body=%s", r.Code, string(data))
	}
	log.Fatal("auth HTTP 失败")
	return
}

// login TCP 登录，返回 LoginResp 中的 UID
func login(serverAddr, account, token string) uint64 {
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		log.Fatalf("TCP 连接失败: %v", err)
	}
	defer conn.Close()

	// 构造 LoginReq payload
	payload := buildLoginReq(account, token)

	// 加密 + 组帧
	frame := encodeMsg(1, 0, payload, true)
	conn.Write(frame)

	// 读 LoginResp
	uid := readLoginResp(conn)
	return uid
}

func buildLoginReq(account, token string) []byte {
	var b bytes.Buffer
	wstr(&b, account)                                // Account
	wstr(&b, token)                                  // Token
	binary.Write(&b, binary.LittleEndian, uint64(0)) // UID=0
	wstr(&b, "0.8.3001.1.1")                         // Version
	binary.Write(&b, binary.LittleEndian, uint16(0)) // Metadata=空map
	binary.Write(&b, binary.LittleEndian, uint16(0)) // ExtData=空
	binary.Write(&b, binary.LittleEndian, uint16(0)) // ReqType=0
	binary.Write(&b, binary.LittleEndian, uint32(0)) // SeqID=0
	wstr(&b, "")                                     // Entity
	wstr(&b, "")                                     // Sign
	return b.Bytes()
}

func wstr(b *bytes.Buffer, s string) {
	binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}

func encodeMsg(msgID uint16, seqID uint32, payload []byte, encrypt bool) []byte {
	body := make([]byte, 6+len(payload)) // 2B MsgID + 4B SeqID + payload
	binary.LittleEndian.PutUint16(body[0:2], msgID)
	binary.LittleEndian.PutUint32(body[2:6], seqID)
	copy(body[6:], payload)

	if encrypt {
		encryptXOR(body) // 客户端发送 isClient=true → key=decryptKey
	}

	// 帧: [3B len LE][1B flags][body]
	flags := byte(0)
	if encrypt {
		flags |= 0x02
	}
	f := make([]byte, 4+len(body))
	f[0] = byte(len(body))
	f[1] = byte(len(body) >> 8)
	f[2] = byte(len(body) >> 16)
	f[3] = flags
	copy(f[4:], body)
	return f
}

var encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}
var decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}

func encryptXOR(data []byte) {
	key := decryptKey // isClient=true
	for i := range data {
		n := byte(i%7 + 1)
		b := data[i]
		b = (b << n) | (b >> (8 - n)) // 循环左移
		data[i] = b ^ key[i%len(key)]
	}
}

func decryptXOR(data []byte) {
	key := encryptKey // isClient=true, 服务端发送时用 encryptKey
	for i := range data {
		b := data[i] ^ key[i%len(key)]
		n := byte(i%7 + 1)
		data[i] = (b >> n) | (b << (8 - n))
	}
}

func readLoginResp(conn net.Conn) uint64 {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	for {
		hdr := make([]byte, 4)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			log.Fatalf("read hdr: %v", err)
		}
		mlen := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
		body := make([]byte, mlen)
		if _, err := io.ReadFull(conn, body); err != nil {
			log.Fatalf("read body: %v", err)
		}

		// 解密
		if hdr[3]&0x02 != 0 {
			decryptXOR(body)
		}

		msgID := binary.LittleEndian.Uint16(body[0:2])
		payload := body[6:]

		if msgID == 2 && len(payload) >= 12 {
			uid := binary.LittleEndian.Uint64(payload[0:8])
			result := binary.LittleEndian.Uint32(payload[8:12])
			log.Printf("  LoginResp: UID=%d Result=%d payloadLen=%d", uid, result, len(payload))
			return uid
		}
		log.Printf("  skip MsgID=%d", msgID)
	}
}
