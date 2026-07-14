package protocol

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// AuthError HTTP 登录返回的业务错误，包含服务端错误码。
// 用于上层识别特定错误（如限流 -600）并决定是否重试。
type AuthError struct {
	Code    int
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: code=%d", e.Message, e.Code)
}

// IsRateLimitError 判断错误是否为登录服限流（RATE_LIMIT = -600）。
func IsRateLimitError(err error) bool {
	authErr, ok := errors.AsType[*AuthError](err)
	return ok && authErr.Code == -600
}

// Authenticator 负责为指定账号获取并释放 TCP 连接。
// 实现可自主选择新建连接、复用连接池连接或其他方式；调用方只关心最终得到的连接和起始 seqID。
type Authenticator interface {
	// Authenticate 返回已准备好发送消息的 TCP 连接、是否来自连接池以及当前起始 seqID。
	// skipDrain 提示实现：当连接来自连接池且调用方将自行管理帧缓存/排空时，可跳过 DrainConn。
	// 新建立连接的起始 seqID 应为 0，且 borrowedFromPool=false；连接池复用场景的 seqID 由连接池管理。
	Authenticate(ctx context.Context, accountID string, skipDrain bool) (conn net.Conn, borrowedFromPool bool, startSeqID uint32, err error)

	// Return 归还或关闭连接。
	// 实现根据连接来源决定：连接池实现应 Return 到池中，新建连接实现应 Close。
	Return(accountID string, conn net.Conn, lastSeqID uint32)
}

// HTTPAuthenticator 通过 HTTP 登录 + TCP LoginReq/Resp 建立游戏连接。
type HTTPAuthenticator struct {
	ServerAddr string // TCP 游戏服务器地址
	HTTPAddr   string // HTTP 登录服务地址
}

// Authenticate 实现 Authenticator 接口。
// HTTP 登录总是新建 TCP 连接，因此 skipDrain 参数被忽略。
func (a *HTTPAuthenticator) Authenticate(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
	token, _, authErr := AuthLogin(a.HTTPAddr, accountID)
	if authErr != nil {
		return nil, false, 0, fmt.Errorf("HTTP 登录失败: %w", authErr)
	}
	log.Printf("[重放] [account=%s] HTTP 登录成功 (httpAddr=%s, token=%s)", accountID, a.HTTPAddr, token[:8])

	conn, dialErr := net.DialTimeout("tcp", a.ServerAddr, 5*time.Second)
	if dialErr != nil {
		return nil, false, 0, fmt.Errorf("连接服务端失败: %v", dialErr)
	}
	log.Printf("[重放] [account=%s] 已连接 %s", accountID, a.ServerAddr)

	if loginErr := sendLoginReq(conn, accountID, token, ""); loginErr != nil {
		_ = conn.Close()
		return nil, false, 0, fmt.Errorf("发送 LoginReq 失败: %v", loginErr)
	}
	log.Printf("[重放] [account=%s] LoginReq 已发送", accountID)

	if respErr := waitLoginResp(conn); respErr != nil {
		_ = conn.Close()
		return nil, false, 0, fmt.Errorf("等待 LoginResp 失败: %v", respErr)
	}
	log.Printf("[重放] [account=%s] 登录成功", accountID)

	// 等待服务端推送完成（登录后常见有一批初始化 Ntf）
	time.Sleep(2 * time.Second)
	return conn, false, 0, nil
}

// Return 关闭新建连接。
func (a *HTTPAuthenticator) Return(accountID string, conn net.Conn, lastSeqID uint32) {
	if conn != nil {
		_ = conn.Close()
		log.Printf("[重放] [account=%s] 连接已关闭", accountID)
	}
}

// AuthLogin HTTP 登录获取 token（可选，用于需要登录的场景）
func AuthLogin(httpAddr string, openID string) (token string, serverAddr string, err error) {
	url := fmt.Sprintf("http://%s/authlogin", httpAddr)
	body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, openID)
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return "", "", fmt.Errorf("HTTP 登录请求失败: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
		return "", "", &AuthError{Code: result.Code, Message: "登录失败"}
	}

	return result.Token, result.ServerAddr, nil
}

// sendLoginReq 发送 LoginReq 框架消息
// loginPayloadB64 非空时使用录制的解密 payload（替换 token），否则自行序列化
func sendLoginReq(conn net.Conn, openID string, token string, loginPayloadB64 string) error {
	var payload []byte

	if loginPayloadB64 != "" {
		// 使用录制的解密 payload，替换其中的 Token 字段
		savedPayload, err := base64.StdEncoding.DecodeString(loginPayloadB64)
		if err != nil {
			return fmt.Errorf("解码 LoginReq payload 失败: %v", err)
		}

		// LoginReq ByteStream 格式: [2B len][Account][2B len][Token]...
		// 跳过 Account (2B 长度 + 内容)，然后替换 Token
		if len(savedPayload) < 2 {
			return fmt.Errorf("LoginReq payload 太短")
		}
		accountLen := int(savedPayload[0]) | int(savedPayload[1])<<8
		tokenOffset := 2 + accountLen

		if tokenOffset+2 > len(savedPayload) {
			return fmt.Errorf("LoginReq payload 格式错误")
		}
		oldTokenLen := int(savedPayload[tokenOffset]) | int(savedPayload[tokenOffset+1])<<8

		// 构建新 payload: Account + 新 Token + 剩余部分
		var buf bytes.Buffer
		buf.Write(savedPayload[:tokenOffset])               // Account 部分
		writeString(&buf, token)                            // 新 Token
		buf.Write(savedPayload[tokenOffset+2+oldTokenLen:]) // Token 后的部分

		payload = buf.Bytes()
		log.Printf("[重放] 使用录制的 LoginReq payload (%d 字节)，已替换 token", len(payload))
	} else {
		// 自行序列化（兼容旧版录制文件）
		var buf bytes.Buffer

		writeString(&buf, openID)
		writeString(&buf, token)
		_ = binary.Write(&buf, binary.LittleEndian, uint64(0))
		writeString(&buf, "0.8.3001.1.1")
		_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
		_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
		writeString(&buf, "")
		writeString(&buf, "")

		payload = buf.Bytes()
		log.Printf("[重放] 使用自行序列化的 LoginReq (%d 字节)", len(payload))
	}

	// 编码帧: MsgID=1, SeqID=0, flags=FlagEncrypt
	frame := EncodeFrame(1, 0, FlagEncrypt, payload, true)

	if _, err := conn.Write(frame); err != nil {
		return fmt.Errorf("TCP 发送失败: %v", err)
	}
	return nil
}

// ExtractAccountFromLoginPayload 从 LoginReq 解密后的 payload 中提取账号名
// LoginReq ByteStream 格式: [2B len][Account string][2B len][Token string]...
func ExtractAccountFromLoginPayload(payload []byte) (string, error) {
	if len(payload) < 2 {
		return "", fmt.Errorf("payload 太短")
	}
	accountLen := int(payload[0]) | int(payload[1])<<8
	if 2+accountLen > len(payload) {
		return "", fmt.Errorf("account 字段不完整: 声明 %d 字节，剩余 %d 字节", accountLen, len(payload)-2)
	}
	return string(payload[2 : 2+accountLen]), nil
}

// writeString 写入 ByteStream 格式的字符串（2B长度LE + 内容）
func writeString(buf *bytes.Buffer, s string) {
	_ = binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}

// waitLoginResp 等待 LoginResp（跳过其他消息）
func waitLoginResp(conn net.Conn) error {
	// 设置读取超时，避免无限等待
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }() // 清除超时

	for {
		// 读取帧头
		header := make([]byte, FrameHeaderSize)
		if _, err := io.ReadFull(conn, header); err != nil {
			return fmt.Errorf("读取帧头失败: %v", err)
		}

		msgLen, _, err := ParseFrameHeader(header)
		if err != nil {
			return fmt.Errorf("解析帧头失败: %v", err)
		}

		// 读取消息体
		body := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, body); err != nil {
			return fmt.Errorf("读取消息体失败: %v", err)
		}

		// 组合完整帧并解码
		raw := make([]byte, FrameHeaderSize+msgLen)
		copy(raw, header)
		copy(raw[FrameHeaderSize:], body)

		frame, err := DecodeFrame(raw, false) // 服务端→客户端
		if err != nil {
			return fmt.Errorf("解码失败: %v", err)
		}

		// 找到 LoginResp 则跳出循环
		if frame.MsgID == 2 {
			// 检查 LoginResp 结果
			// LoginResp ByteStream 格式: UID(uint64) + Result(uint32) + ...
			if len(frame.Payload) >= 12 {
				// 跳过 UID(8字节), 读取 Result(4字节)
				result := binary.LittleEndian.Uint32(frame.Payload[8:12])
				if result != 0 {
					return fmt.Errorf("登录失败: Result=%d", result)
				}
			}
			return nil
		}

		// 跳过其他消息（如 Ntf 等），继续等待 LoginResp
		msgName := GetMsgName(frame.MsgID)
		if msgName == "" {
			switch frame.MsgID {
			case 4:
				msgName = "Pong"
			case 10:
				msgName = "KickOut"
			default:
				msgName = fmt.Sprintf("Unknown(%d)", frame.MsgID)
			}
		}
		log.Printf("[重放] 跳过消息，等待 LoginResp: %s (MsgID=%d)", msgName, frame.MsgID)
	}
}
