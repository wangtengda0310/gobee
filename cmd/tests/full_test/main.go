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

// 完整验证：
// 1. 登录 test95 → 查 UID
// 2. 发送 CreateRoleReq → 查 Ack
// 3. 断开连接
// 4. 重新登录 test95 → 查 UID（是否同一个）
// 5. 发送 CreateRoleReq → 查 Ack 的 ErrCode（预期 1019 = 已有角色）

func main() {
	openID := "test95"
	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// ==== 第1次登录 + 创角 ====
	log.Println("===== 第1次: 登录 + 创角 =====")
	uid1, conn1 := loginAndCreateRole(openID, httpAddr, "测试角色名1")
	conn1.Close()
	log.Printf("第1次: UID=%d, 已断开连接", uid1)

	time.Sleep(2 * time.Second)

	// ==== 第2次登录 (不创角，只看 UID) ====
	log.Println("===== 第2次: 仅登录查看 UID =====")
	uid2 := loginOnly(openID, httpAddr)
	log.Printf("第2次: UID=%d", uid2)

	if uid1 == uid2 {
		log.Printf("✅ UID 一致 (%d) — HandleCreateRole 中的 InsertData 成功写入了 accountid", uid1)
	} else {
		log.Printf("❌ UID 不一致! 第1次=%d, 第2次=%d — 说明 InsertData 没生效或 Player 文档有问题", uid1, uid2)
	}

	// ==== 第3次登录 + 尝试重复创角 ====
	log.Println("===== 第3次: 登录 + 尝试重复创角 =====")
	uid3, conn3 := loginAndCreateRole(openID, httpAddr, "测试角色名2")
	conn3.Close()
	log.Printf("第3次: UID=%d", uid3)
}

func loginAndCreateRole(openID, httpAddr, roleName string) (uint64, net.Conn) {
	token, serverAddr := auth(httpAddr, openID)
	conn := tcpConnect(serverAddr)

	// LoginReq
	sendLoginReq(conn, openID, token)
	uid := readLoginResp(conn)
	log.Printf("  LoginResp: UID=%d", uid)

	// 启动 reader 消费推送
	stop := make(chan struct{})
	go readDrainer(conn, stop)

	time.Sleep(2 * time.Second) // 等推送完成

	// CreateRoleReq
	sendCreateRoleReq(conn, roleName)
	time.Sleep(3 * time.Second) // 等 Ack
	close(stop)

	return uid, conn
}

func loginOnly(openID, httpAddr string) uint64 {
	token, serverAddr := auth(httpAddr, openID)
	conn := tcpConnect(serverAddr)
	defer conn.Close()

	sendLoginReq(conn, openID, token)
	uid := readLoginResp(conn)
	return uid
}

// ========== Helpers ==========

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

func tcpConnect(addr string) net.Conn {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Fatalf("TCP 失败: %v", err)
	}
	return conn
}

// ========== LoginReq ==========

func sendLoginReq(conn net.Conn, account, token string) {
	payload := buildLoginReqPayload(account, token)
	frame := encodeFrame(1, 0, payload, true)
	conn.Write(frame)
}

func buildLoginReqPayload(account, token string) []byte {
	var b bytes.Buffer
	wstr(&b, account)
	wstr(&b, token)
	binary.Write(&b, binary.LittleEndian, uint64(0))
	wstr(&b, "0.8.3001.1.1")
	binary.Write(&b, binary.LittleEndian, uint16(0))
	binary.Write(&b, binary.LittleEndian, uint16(0))
	binary.Write(&b, binary.LittleEndian, uint16(0))
	binary.Write(&b, binary.LittleEndian, uint32(0))
	wstr(&b, "")
	wstr(&b, "")
	return b.Bytes()
}

func wstr(b *bytes.Buffer, s string) {
	binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}

// ========== CreateRoleReq ==========

func sendCreateRoleReq(conn net.Conn, roleName string) {
	// protobuf 手动编码 CreateRoleReq
	var b bytes.Buffer
	// field 1: roleName (string, wire=2)
	b.WriteByte(0x0a)
	writeVarint(&b, uint64(len(roleName)))
	b.WriteString(roleName)
	// field 2: gender=2 (int32, wire=0)
	b.WriteByte(0x10)
	writeVarint(&b, 2)
	// field 3: isRandom=true (bool, wire=0)
	b.WriteByte(0x18)
	b.WriteByte(1)
	// field 4: itemID=1020502 (int32, wire=0)
	b.WriteByte(0x20)
	writeVarint(&b, 1020502)

	// ByteStream 包装
	payload := b.Bytes()
	wrapped := make([]byte, 2+len(payload))
	binary.LittleEndian.PutUint16(wrapped, uint16(len(payload)))
	copy(wrapped[2:], payload)

	frame := encodeFrame(1006, 1, wrapped, true) // ✅ 修复：使用 wrapped 而非 payload
	conn.Write(frame)
	log.Printf("  CreateRoleReq 已发送: roleName=%s", roleName)
}

func writeVarint(b *bytes.Buffer, v uint64) {
	for v >= 0x80 {
		b.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	b.WriteByte(byte(v))
}

// ========== Network ==========

var encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}
var decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}

func encodeFrame(msgID uint16, seqID uint32, payload []byte, encrypt bool) []byte {
	body := make([]byte, 6+len(payload))
	binary.LittleEndian.PutUint16(body[0:2], msgID)
	binary.LittleEndian.PutUint32(body[2:6], seqID)
	copy(body[6:], payload)

	if encrypt {
		encryptXOR(body)
	}

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

func encryptXOR(data []byte) {
	key := decryptKey
	for i := range data {
		n := byte(i%7 + 1)
		b := data[i]
		b = (b << n) | (b >> (8 - n))
		data[i] = b ^ key[i%len(key)]
	}
}

func decryptXOR(data []byte) {
	key := encryptKey
	for i := range data {
		b := data[i] ^ key[i%len(key)]
		n := byte(i%7 + 1)
		data[i] = (b >> n) | (b << (8 - n))
	}
}

// ========== 读取 LoginResp ==========

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
		if hdr[3]&0x02 != 0 {
			decryptXOR(body)
		}
		msgID := binary.LittleEndian.Uint16(body[0:2])
		payload := body[6:]
		if msgID == 2 && len(payload) >= 12 {
			uid := binary.LittleEndian.Uint64(payload[0:8])
			result := binary.LittleEndian.Uint32(payload[8:12])
			log.Printf("  LoginResp: UID=%d Result=%d", uid, result)
			return uid
		}
	}
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
		mlen := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
		body := make([]byte, mlen)
		if _, err := io.ReadFull(conn, body); err != nil {
			return
		}
		if hdr[3]&0x02 != 0 {
			decryptXOR(body)
		}
		msgID := binary.LittleEndian.Uint16(body[0:2])
		payload := body[6:]

		switch msgID {
		case 1007: // CreateRoleAck
			if len(payload) >= 2 {
				protoLen := int(payload[0]) | int(payload[1])<<8
				protoData := payload[2 : 2+protoLen]
				errCode := readCreateRoleAckErrCode(protoData)
				log.Printf("  ← CreateRoleAck: ErrCode=%d proto=%x", errCode, protoData)
			}
		case 1016:
			log.Printf("  ← RefreshUserDataNtf (%d bytes)", len(payload))
		case 1017:
			log.Printf("  ← LoginServerMsgOverNtf")
		default:
			if msgID >= 1000 {
				log.Printf("  ← MsgID=%d (%d bytes)", msgID, len(payload))
			}
		}
	}
}

func readCreateRoleAckErrCode(data []byte) uint32 {
	// 简单解析 protobuf field 1 (ErrCode, varint)
	if len(data) > 0 && data[0]>>3 == 1 { // field 1, wire=0
		var v uint64
		for i := 1; i < len(data); i++ {
			v |= uint64(data[i]&0x7f) << (7 * (i - 1))
			if data[i]&0x80 == 0 {
				break
			}
		}
		return uint32(v)
	}
	return 0
}
