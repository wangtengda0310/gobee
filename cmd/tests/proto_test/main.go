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

	streamproto "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// 核心验证：用真实 proto 库构造 CreateRoleReq + 正确 ByteStream 包装
// 验证创角成功后重新登录是否获得相同 UID
func main() {
	openID := fmt.Sprintf("ut_new_%d", time.Now().Unix())

	if len(os.Args) > 1 {
		openID = os.Args[1]
	}
	if len(openID) == 0 {
		openID = "proto_test95"
	}
	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// ==== 第1次登录 + 创角 ====
	log.Println("===== 第1次: 登录 + 创角 (使用真实 proto 库) =====")
	uid1, conn1 := loginAndCreateRole(openID, httpAddr)
	conn1.Close()

	time.Sleep(2 * time.Second)

	// ==== 第2次登录 ====
	log.Println("===== 第2次: 仅登录查看 UID =====")
	uid2 := loginOnly(openID, httpAddr)

	if uid1 == uid2 {
		log.Printf("✅ UID 一致 (%d) — HandleCreateRole InsertData 成功", uid1)
	} else {
		log.Printf("❌ UID 不一致! (%d vs %d)", uid1, uid2)
	}
}

func loginAndCreateRole(openID, httpAddr string) (uint64, net.Conn) {
	token, serverAddr := httpAuth(httpAddr, openID)
	conn := tcpConnect(serverAddr)

	// LoginReq
	uid := doLogin(conn, openID, token)
	log.Printf("  LoginResp: UID=%d", uid)

	// reader
	stop := make(chan struct{})
	ackCh := make(chan uint32, 1)
	go readAck(conn, stop, ackCh)

	time.Sleep(2 * time.Second)

	// CreateRoleReq（用真实 proto 库）
	roleName := fmt.Sprintf("test%s%d", openID[len(openID)-4:], time.Now().UnixNano()%10000)
	sendCreateRoleReqProto(conn, roleName)

	select {
	case code := <-ackCh:
		log.Printf("  CreateRoleAck: ErrCode=%d", code)
	case <-time.After(10 * time.Second):
		log.Printf("  CreateRoleAck 超时")
	}
	close(stop)
	return uid, conn
}

func loginOnly(openID, httpAddr string) uint64 {
	token, serverAddr := httpAuth(httpAddr, openID)
	conn := tcpConnect(serverAddr)
	defer conn.Close()
	return doLogin(conn, openID, token)
}

func httpAuth(httpAddr, openID string) (token, serverAddr string) {
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, _ := http.Post("http://"+httpAddr+"/authlogin",
		"application/json", bytes.NewBufferString(body))
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
		log.Fatalf("auth 失败: %s", string(data))
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

func doLogin(conn net.Conn, openID, token string) uint64 {
	var buf bytes.Buffer
	wstr(&buf, openID)
	wstr(&buf, token)
	binary.Write(&buf, binary.LittleEndian, uint64(0))
	wstr(&buf, "0.8.3001.1.1")
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	wstr(&buf, "")
	wstr(&buf, "")

	frame := streamproto.EncodeFrame(1, 0, streamproto.FlagEncrypt, buf.Bytes(), true)
	conn.Write(frame)
	return readLoginResp(conn)
}

func wstr(b *bytes.Buffer, s string) {
	binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}

// ========== CreateRoleReq（真实 proto 库）==========

func sendCreateRoleReqProto(conn net.Conn, roleName string) {
	// 获取注册的 CreateRoleReq 消息类型
	msg, ok := streamproto.NewMessage(1006)
	if !ok {
		log.Fatal("CreateRoleReq 未注册")
	}

	// JSON → proto
	payloadJSON := fmt.Sprintf(`{"roleName":"%s","gender":1,"isRandom":true,"itemID":1020303}`, roleName)
	if err := json.Unmarshal([]byte(payloadJSON), msg); err != nil {
		log.Fatalf("JSON→proto 失败: %v", err)
	}

	// proto.Marshal
	protoData, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalf("proto.Marshal 失败: %v", err)
	}
	log.Printf("  CreateRoleReq proto: %d bytes hex=%x", len(protoData), protoData)

	// ByteStream 包装: [2B len LE][proto data]
	payload := make([]byte, 2+len(protoData))
	payload[0] = byte(len(protoData))
	payload[1] = byte(len(protoData) >> 8)
	copy(payload[2:], protoData)

	frame := streamproto.EncodeFrame(1006, 1, streamproto.FlagEncrypt, payload, true)
	conn.Write(frame)
	log.Printf("  CreateRoleReq 已发送: roleName=%s", roleName)
}

// ========== 读取 LoginResp ==========

func readLoginResp(conn net.Conn) uint64 {
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	defer conn.SetReadDeadline(time.Time{})
	for {
		hdr := make([]byte, 4)
		io.ReadFull(conn, hdr)
		mlen := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
		body := make([]byte, mlen)
		io.ReadFull(conn, body)
		raw := make([]byte, 4+len(body))
		copy(raw, hdr)
		copy(raw[4:], body)
		frame, _ := streamproto.DecodeFrame(raw, false)
		if frame.MsgID == 2 && len(frame.Payload) >= 12 {
			uid := binary.LittleEndian.Uint64(frame.Payload[0:8])
			result := binary.LittleEndian.Uint32(frame.Payload[8:12])
			log.Printf("  LoginResp: UID=%d Result=%d", uid, result)
			return uid
		}
	}
}

func readAck(conn net.Conn, stop <-chan struct{}, out chan<- uint32) {
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
		raw := make([]byte, 4+len(body))
		copy(raw, hdr)
		copy(raw[4:], body)
		frame, _ := streamproto.DecodeFrame(raw, false)

		if frame.MsgID == 1007 && len(frame.Payload) >= 2 {
			protoLen := int(frame.Payload[0]) | int(frame.Payload[1])<<8
			if protoLen > 0 {
				protoData := frame.Payload[2 : 2+protoLen]
				// 简单解析 ErrCode: field 1, wire 0 (varint)
				var errCode uint32
				if len(protoData) > 0 && protoData[0]>>3 == 1 {
					var v uint64
					for i := 1; i < len(protoData); i++ {
						v |= uint64(protoData[i]&0x7f) << (7 * (i - 1))
						if protoData[i]&0x80 == 0 {
							errCode = uint32(v)
							break
						}
					}
				}
				out <- errCode
				return
			}
		}
	}
}
