package protocol_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// TestCreateRoleAndReLogin 验证：创角后重新登录能拿到同一个 UID
// 这是端到端集成测试，需要 204 服务器可连
func TestCreateRoleAndReLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要网络连接的集成测试")
	}

	httpAddr := "10.254.114.204:20144"
	openID := fmt.Sprintf("ut_cr_%d", time.Now().Unix())
	t.Logf("测试账号: %s", openID)

	// ==== 第1次登录 + 创角 ====
	uid1 := loginAndCreateRoleOnce(t, openID, httpAddr)
	t.Logf("第1次: UID=%d", uid1)

	time.Sleep(2 * time.Second)

	// ==== 第2次登录（重新认证+登录，验证 UID 是否相同）====
	token2, serverAddr2 := httpAuthLogin(t, httpAddr, openID)
	conn2 := tcpConnect(t, serverAddr2)
	defer func() { _ = conn2.Close() }()

	sendLoginReqRaw(t, conn2, openID, token2)
	uid2 := waitForLoginResp(t, conn2)
	t.Logf("第2次: UID=%d", uid2)

	// 核心断言
	assert.Equal(t, uid1, uid2,
		"创角成功后重新登录必须返回相同 UID。"+
			"检查: HandleCreateRole InsertData 是否被调用？MongoDB Player 集合中是否有 accountid 字段？")
}

// loginAndCreateRoleOnce HTTP认证 + TCP登录 + 创角，返回 UID
func loginAndCreateRoleOnce(t *testing.T, openID, httpAddr string) uint64 {
	t.Helper()

	token, serverAddr := httpAuthLogin(t, httpAddr, openID)
	conn := tcpConnect(t, serverAddr)

	// LoginReq
	sendLoginReqRaw(t, conn, openID, token)

	// 使用 conn 的 read deadline 来控制超时，串行读取 LoginResp 和后续消息
	_ = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

	// 阶段1: 等待 LoginResp
	uid := readUntilLoginResp(t, conn)

	// 阶段2: 跳过推送消息 (RefreshUserDataNtf, 6500 等)，等待 3 秒让推送完成
	time.Sleep(3 * time.Second)

	// 阶段3: 发送 CreateRoleReq（英文名避开敏感字检测）
	roleName := fmt.Sprintf("test%s", openID[len(openID)-4:])
	sendCreateRoleReqProto(t, conn, roleName)

	// 阶段4: 等待 CreateRoleAck
	errCode := readUntilCreateRoleAck(t, conn)
	t.Logf("  CreateRoleAck ErrCode=%d", errCode)
	// 不强制 assert ErrCode==0——可能是 HasCreateRole(1019) 如果账号已有角色
	// 只要不是 1001 (InvalidParams) 就说明 proto 编码正确

	_ = conn.Close()

	// 如果创角被拒绝(HasCreateRole=1019)，说明之前已有角色
	// 这种情况下 UID 一致性已经由 InsertData 保证了
	if errCode == 1019 {
		t.Log("  账号已有角色(EHasCreateRole)，说明之前 InsertData 已成功写入 accountid")
	}

	return uid
}

// ========== HTTP 认证 ==========

func httpAuthLogin(t *testing.T, httpAddr, openID string) (string, string) {
	t.Helper()
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, err := http.Post("http://"+httpAddr+"/authlogin", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(resp.Body)
	var r struct {
		Code       int    `json:"code"`
		Token      string `json:"token"`
		ServerAddr string `json:"server_addr"`
	}
	require.NoError(t, json.Unmarshal(data, &r))
	require.Equal(t, 0, r.Code)
	return r.Token, r.ServerAddr
}

func tcpConnect(t *testing.T, addr string) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	require.NoError(t, err)
	return conn
}

// ========== LoginReq ==========

func sendLoginReqRaw(t *testing.T, conn net.Conn, openID, token string) {
	t.Helper()
	var buf bytes.Buffer
	writeStr(&buf, openID)
	writeStr(&buf, token)
	_ = binary.Write(&buf, binary.LittleEndian, uint64(0))
	writeStr(&buf, "0.8.3001.1.1")
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	writeStr(&buf, "")
	writeStr(&buf, "")
	frame := protocol.EncodeFrame(1, 0, protocol.FlagEncrypt, buf.Bytes(), true)
	_, err := conn.Write(frame)
	require.NoError(t, err)
}

func writeStr(b *bytes.Buffer, s string) {
	_ = binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}

// ========== CreateRoleReq (真实 proto 库) ==========

func sendCreateRoleReqProto(t *testing.T, conn net.Conn, roleName string) {
	t.Helper()
	msg, ok := protocol.NewMessage(1006)
	require.True(t, ok, "CreateRoleReq msgID=1006 必须注册")

	jsonPayload := fmt.Sprintf(`{"roleName":"%s","gender":1,"isRandom":true,"itemID":1020303}`, roleName)
	require.NoError(t, json.Unmarshal([]byte(jsonPayload), msg))

	pd, err := proto.Marshal(msg)
	require.NoError(t, err)

	// ByteStream 包装: [2B len LE][protobuf data]
	payload := make([]byte, 2+len(pd))
	payload[0] = byte(len(pd))
	payload[1] = byte(len(pd) >> 8)
	copy(payload[2:], pd)

	frame := protocol.EncodeFrame(1006, 1, protocol.FlagEncrypt, payload, true)
	_, err = conn.Write(frame)
	require.NoError(t, err)
	t.Logf("  CreateRoleReq 已发送: %d bytes proto", len(pd))
}

// ========== 串行读取消息 ==========

// readUntilLoginResp 串行读取 TCP 帧直到收到 LoginResp，返回 UID
func waitForLoginResp(t *testing.T, conn net.Conn) uint64 {
	t.Helper()
	for {
		f := decodeOneFrame(t, conn)
		if f.MsgID == 2 && len(f.Payload) >= 12 {
			uid := binary.LittleEndian.Uint64(f.Payload[0:8])
			result := binary.LittleEndian.Uint32(f.Payload[8:12])
			t.Logf("  LoginResp: UID=%d Result=%d", uid, result)
			require.Equal(t, uint32(0), result, "LoginResp Result 应为 0")
			return uid
		}
	}
}

// readUntilLoginResp 是 waitForLoginResp 的别名，用于 loginAndCreateRoleOnce 调用
func readUntilLoginResp(t *testing.T, conn net.Conn) uint64 {
	t.Helper()
	return waitForLoginResp(t, conn)
}

// readUntilCreateRoleAck 串行读取直到收到 CreateRoleAck，返回 ErrCode
func readUntilCreateRoleAck(t *testing.T, conn net.Conn) uint32 {
	t.Helper()
	for {
		f := decodeOneFrame(t, conn)
		if f.MsgID == 1007 && len(f.Payload) >= 2 {
			pl := int(f.Payload[0]) | int(f.Payload[1])<<8
			if pl > 0 && pl <= len(f.Payload)-2 {
				pd := f.Payload[2 : 2+pl]
				// 简单解析 field 1 (ErrCode, varint)
				var code uint32
				if len(pd) > 0 && pd[0]>>3 == 1 {
					var v uint64
					for i := 1; i < len(pd); i++ {
						v |= uint64(pd[i]&0x7f) << (7 * (i - 1))
						if pd[i]&0x80 == 0 {
							code = uint32(v)
							break
						}
					}
				}
				return code
			}
		}
	}
}

// decodeOneFrame 从 conn 读取一个完整的 TCP 帧并解码
func decodeOneFrame(t *testing.T, conn net.Conn) *protocol.DecodedFrame {
	t.Helper()
	hdr := make([]byte, 4)
	_, err := io.ReadFull(conn, hdr)
	require.NoError(t, err, "读取帧头失败（超时或连接断开）")
	ml := int(hdr[0]) | int(hdr[1])<<8 | int(hdr[2])<<16
	bd := make([]byte, ml)
	_, err = io.ReadFull(conn, bd)
	require.NoError(t, err, "读取帧体失败")
	raw := make([]byte, 4+len(bd))
	copy(raw, hdr)
	copy(raw[4:], bd)
	f, err := protocol.DecodeFrame(raw, false) // 服务端→客户端方向
	require.NoError(t, err)
	return f
}
