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

	sp "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// 诊断：连续登录 test20 和 test21，对比 AuthLogin 分配地址 vs TCP 连接地址
func main() {
	httpAddr := "10.254.114.204:20144"
	frontendServerAddr := "10.254.114.204:18000" // 前端配置的固定地址

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	for _, uid := range []string{"test20", "test21"} {
		log.Printf("=== %s ===", uid)

		// 1. AuthLogin
		body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, uid)
		resp, _ := http.Post("http://"+httpAddr+"/authlogin",
			"application/json", bytes.NewBufferString(body))
		if resp != nil {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			var r struct {
				Code       int    `json:"code"`
				Token      string `json:"token"`
				ServerAddr string `json:"server_addr"`
				OpenID     string `json:"open_id"`
			}
			json.Unmarshal(data, &r)
			if r.Code != 0 {
				log.Printf("  AuthLogin FAILED: code=%d", r.Code)
				continue
			}
			log.Printf("  AuthLogin: open_id=%s, server_addr=%s, token=%s...",
				r.OpenID, r.ServerAddr, r.Token[:8])

			// 2. TCP 连接到前端配置地址（模拟 wails3 行为）
			conn, err := net.DialTimeout("tcp", frontendServerAddr, 5*time.Second)
			if err != nil {
				log.Printf("  TCP 连接 %s 失败: %v", frontendServerAddr, err)
				continue
			}

			// 3. LoginReq
			loginPayload := buildLoginReq(r.OpenID, r.Token)
			frame := sp.EncodeFrame(1, 0, sp.FlagEncrypt, loginPayload, true)
			conn.Write(frame)

			// 4. 读取 LoginResp（从帧头开始读）
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			for {
				hdr := make([]byte, 4)
				if _, err := io.ReadFull(conn, hdr); err != nil {
					log.Printf("  read hdr: %v", err)
					break
				}
				ml := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
				bd := make([]byte, ml)
				if _, err := io.ReadFull(conn, bd); err != nil {
					log.Printf("  read body: %v", err)
					break
				}
				raw := make([]byte, 4+len(bd))
				copy(raw, hdr)
				copy(raw[4:], bd)
				f, err := sp.DecodeFrame(raw, false)
				if err != nil {
					continue
				}
				if f.MsgID == 2 && len(f.Payload) >= 12 {
					lrUID := binary.LittleEndian.Uint64(f.Payload[0:8])
					lrResult := binary.LittleEndian.Uint32(f.Payload[8:12])
					log.Printf("  LoginResp: UID=%d, Result=%d (分配的Lobby=%s, 连接的Lobby=%s)",
						lrUID, lrResult, r.ServerAddr, frontendServerAddr)
					break
				}
			}
			conn.Close()
		}
		time.Sleep(1 * time.Second) // 间隔1秒，让服务端有时间清理
	}
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
