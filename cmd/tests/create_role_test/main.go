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

func main() {
	openID := fmt.Sprintf("ut_new_%d", time.Now().Unix())
	if len(os.Args) > 1 {
		openID = os.Args[1]
	}
	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	token, serverAddr := httpAuth(httpAddr, openID)
	log.Printf("AuthLogin: openID=%s token=%s... server=%s", openID, token[:8], serverAddr)

	conn := tcpConnect(serverAddr)

	// reader first
	msgCh := make(chan msgEntry, 10)
	stopRead := make(chan struct{})
	go readAll(conn, stopRead, msgCh)

	// LoginReq
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

	// Read LoginResp
	var uid uint64
	for e := range msgCh {
		if e.msgID == 2 && len(e.payload) >= 12 {
			uid = binary.LittleEndian.Uint64(e.payload[0:8])
			log.Printf("LoginResp: UID=%d Result=%d", uid, binary.LittleEndian.Uint32(e.payload[8:12]))
			break
		}
	}

	// Drain initial pushes
	time.Sleep(3 * time.Second)

	// Send CreateRoleReq with real proto
	roleName := fmt.Sprintf("test%d", time.Now().UnixNano()%10000000)
	sendCreateRoleReq(conn, roleName)

	// Read CreateRoleAck (with 15s timeout for sensitive word check)
	timeout := time.After(20 * time.Second) // 敏感字 HTTP 检测可能耗时长
	select {
	case e := <-msgCh:
		if e.msgID == 1007 {
			plen := int(e.payload[0]) | int(e.payload[1])<<8
			pdata := e.payload[2 : 2+plen]
			errCode := parseErrCode(pdata)
			log.Printf("CreateRoleAck: ErrCode=%d", errCode)
		}
	case <-timeout:
		log.Printf("CreateRoleAck timeout")
	}
	close(stopRead)
	conn.Close()

	log.Printf("UID=%d", uid)
}

type msgEntry struct {
	msgID   uint16
	payload []byte
}

func readAll(conn net.Conn, stop <-chan struct{}, out chan<- msgEntry) {
	defer close(out)
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
		if _, err := io.ReadFull(conn, bd); err != nil {
			return
		}
		raw := make([]byte, 4+len(bd))
		copy(raw, hdr)
		copy(raw[4:], bd)
		f, err := streamproto.DecodeFrame(raw, false)
		if err != nil {
			continue
		}
		out <- msgEntry{f.MsgID, f.Payload}
	}
}

func sendCreateRoleReq(conn net.Conn, roleName string) {
	msg, ok := streamproto.NewMessage(1006)
	if !ok {
		log.Fatal("CreateRoleReq not registered")
	}
	jsonPayload := fmt.Sprintf(`{"roleName":"%s","gender":1,"isRandom":true,"itemID":1020303}`, roleName)
	json.Unmarshal([]byte(jsonPayload), msg)
	pd, _ := proto.Marshal(msg)
	payload := make([]byte, 2+len(pd))
	payload[0] = byte(len(pd))
	payload[1] = byte(len(pd) >> 8)
	copy(payload[2:], pd)
	frame := streamproto.EncodeFrame(1006, 1, streamproto.FlagEncrypt, payload, true)
	conn.Write(frame)
	log.Printf("CreateRoleReq sent: %d bytes proto, roleName=%s", len(pd), roleName)
}

func parseErrCode(data []byte) uint32 {
	if len(data) > 0 && data[0]>>3 == 1 {
		var v uint64
		for i := 1; i < len(data); i++ {
			v |= uint64(data[i]&0x7f) << (7 * (i - 1))
			if data[i]&0x80 == 0 {
				return uint32(v)
			}
		}
	}
	return 0
}

func httpAuth(httpAddr, openID string) (token, serverAddr string) {
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, _ := http.Post("http://"+httpAddr+"/authlogin", "application/json", bytes.NewBufferString(body))
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
		log.Fatalf("auth failed: %s", string(data))
	}
	log.Fatal("auth HTTP failed")
	return
}

func tcpConnect(addr string) net.Conn {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Fatalf("TCP failed: %v", err)
	}
	return conn
}

func wstr(b *bytes.Buffer, s string) {
	binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}
