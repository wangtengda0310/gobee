// Package gameproto 演示如何为 go-service 的 Unity 游戏协议实现自定义 Layer + 帧读取器。
//
// 这是一个完整的示例，展示 pcap 包的「协议无关基础设施」设计哲学：
// pcap 包负责抓包和 TCP 流重组，本包负责具体的游戏协议解析。
//
// # 消息格式（来自 go-service/framework/net/internal/internal/msg_reader.go）
//
// 每条游戏消息的在线格式如下（所有多字节字段均为小端）：
//
//	┌─────────────── 4 字节包头 ───────────────┐ ┌─────── msgData（msgSize 字节）───────┐
//	│ msgSize (3B LE)  │  flag (1B)             │ │ msgID(2B) │ seqID(4B) │ body(...)  │
//	└──────────────────┴────────────────────────┘ └───────────┴───────────┴────────────┘
//
//	包头：
//	  - msgSize: msgData 的长度（不含包头本身的 4 字节），3 字节小端无符号整数。
//	  - flag 第 4 字节：
//	      bit0 = 1 表示 body 经 snappy 压缩，0 表示未压缩。
//	      bit1 = 1 表示 msgData 已加密，0 表示未加密。
//
//	msgData：
//	  - msgID: 消息类型（uint16 小端），< 1000 为框架消息（LoginReq/Ping/Pong 等），
//	    >= 1000 为游戏消息（对应 proto 消息号）。
//	  - seqID: 序列号（uint32 小端），客户端递增，用于请求-响应配对。
//	  - body: 消息体，格式因消息类型而异：
//	      LoginReq(1):  自定义 ByteStream [2B len][Account][2B len][Token][8B][version]...
//	      LoginResp(2): 自定义 ByteStream [8B UID][4B Result][2B len][Version]...
//	      Ping(3):      空（body 长度为 0）。
//	      Pong(4):      原始时间戳（8 字节 uint64，非 protobuf）。
//	      游戏消息(>=1000): [2B len][protobuf 数据]（子信封格式）。
//
// # 加密
//
// 当 flag bit1=1 时，msgData 整体被加密（不含包头）。
// 加密算法是自定义的轻量对称加密（非 AES），逐字节「异或 + 循环移位」：
//
//	加密：byte[i] = rotate_left(byte[i], n) ^ key[i % 8]，其中 n = i%7 + 1
//	解密：byte[i] = rotate_right(byte[i] ^ key[i % 8], n)
//
// 密钥只有 8 字节且硬编码在 go-service 源码中（见下方 encryptKey/decryptKey）。
// 两个方向的密钥不同：
//   - 客户端→服务端：用 decryptKey 加密（服务端也用 decryptKey 解密）
//   - 服务端→客户端：用 encryptKey 加密（客户端也用 encryptKey 解密）
//
// 密钥命名看起来反直觉，这是因为它们是以「服务端视角」命名的：
// encryptKey = 服务端用来加密（发给客户端）的密钥。
// decryptKey = 服务端用来解密（收自客户端）的密钥。
//
// 抓包侧解密时，根据 TCP 流方向选密钥即可。
package gameproto

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"github.com/gopacket/gopacket"
	"google.golang.org/protobuf/encoding/protowire"
)

// 协议常量（与 go-service/framework/net/internal/internal/consts/const.go 一致）。
const (
	// MsgHeadSize 包头大小：3 字节 msgSize + 1 字节 flag = 4 字节。
	MsgHeadSize = 4
	// MsgIDSize 消息 ID 大小（uint16 = 2 字节）。
	MsgIDSize = 2
	// SeqSize 序列号大小（uint32 = 4 字节）。
	SeqSize = 4
	// MaxMsgSize 外部消息 msgData 长度上限（go-service 的 maxPacket 约束）。
	MaxMsgSize = 62 * 1024
)

// -----------------------------------------------------------------------------
// 自定义 gopacket Layer：GameMessage
// -----------------------------------------------------------------------------
// GameMessage 把一条完整的游戏消息实现为 gopacket 的自定义 Layer。
// 注册后可以用 gopacket 标准方式访问：
//
//	pkt := gopacket.NewPacket(rawBytes, LayerTypeGameMessage, gopacket.Default)
//	if layer := pkt.Layer(LayerTypeGameMessage); layer != nil {
//	    msg := layer.(*GameMessage)
//	    fmt.Println(msg.MsgID, msg.Body)
//	}
//
// 但实际使用中更推荐用 FrameReader（下方）配合 pcap.TCPStreamHandler，
// 因为 gopacket 的 Layer 机制处理的是「单条完整消息」，
// 而 TCP 流上的消息边界需要流重组后才能确定。
// -----------------------------------------------------------------------------

// GameMessage 是一个完整的游戏消息（已从 TCP 流中切分出来）。
// 实现 gopacket.Layer 接口，可参与 gopacket 的 Layer 链。
type GameMessage struct {
	// MsgSize msgData 的长度（不含包头），从包头前 3 字节解码。
	MsgSize uint32
	// Flag 包头第 4 字节：bit0=压缩(snappy), bit1=加密。
	// 用 IsEncrypted()/IsCompressed() 方法判断，不要直接位操作。
	Flag byte
	// MsgID 消息类型（uint16 小端），决定 body 的解析方式。
	// < 1000 = 框架消息（LoginReq/Ping/Pong 等，非 protobuf）。
	// >= 1000 = 游戏消息（body 是 [2B len][protobuf] 子信封格式）。
	MsgID uint16
	// SeqID 序列号（uint32 小端），客户端递增，用于请求-响应配对和重放。
	SeqID uint32
	// Body 消息体（已解密、已解压），格式因 MsgID 而异。
	Body []byte

	// 以下字段实现 gopacket.Layer 接口（LayerContents/LayerPayload），
	// 不对外暴露（小写），仅供 gopacket 内部使用。
	layerContents []byte // 本层（包头）的原始字节
	layerPayload  []byte // 本层 payload（msgData）的原始字节
}

// LayerType 返回本自定义 Layer 的类型标识。
// 返回 LayerTypeGameMessage（在 init() 中通过 RegisterLayerType 注册）。
func (m *GameMessage) LayerType() gopacket.LayerType { return LayerTypeGameMessage }

// LayerContents 实现 gopacket.Layer 接口：返回本层（包头）的原始字节。
func (m *GameMessage) LayerContents() []byte { return m.layerContents }

// LayerPayload 实现 gopacket.Layer 接口：返回本层包含的 payload（msgData）。
func (m *GameMessage) LayerPayload() []byte { return m.layerPayload }

// String 返回可读表示，用于日志输出。
func (m *GameMessage) String() string {
	return fmt.Sprintf("GameMsg{ID=%d, Seq=%d, BodyLen=%d, Encrypted=%v, Compressed=%v}",
		m.MsgID, m.SeqID, len(m.Body), m.IsEncrypted(), m.IsCompressed())
}

// IsEncrypted 返回该消息是否被加密（flag bit1）。
func (m *GameMessage) IsEncrypted() bool { return m.Flag&0x2 != 0 }

// IsCompressed 返回该消息是否被压缩（flag bit0）。
func (m *GameMessage) IsCompressed() bool { return m.Flag&0x1 != 0 }

// LayerTypeGameMessage 是本自定义 Layer 的类型标识。
// 在 init() 中通过 RegisterLayerType 注册到 gopacket 全局注册表，
// 注册后 gopacket 才能识别并解码这个 Layer 类型。
var LayerTypeGameMessage gopacket.LayerType

func init() {
	// 注册自定义 LayerType 到 gopacket 全局注册表。
	// 参数 12345 是自定义类型编号——必须避开 gopacket 内置编号（1-1000 左右）。
	// Decoder 指定了如何从原始字节解码出 GameMessage（调用 decodeGameMessage）。
	LayerTypeGameMessage = gopacket.RegisterLayerType(
		12345,
		gopacket.LayerTypeMetadata{
			Name:    "GameMessage",
			Decoder: gopacket.DecodeFunc(decodeGameMessage),
		},
	)
}

// decodeGameMessage 是 gopacket Decoder 函数。
// gopacket 在 NewPacket 时调用它，从原始字节解码出一个 GameMessage Layer。
//
// 注意：这个 decoder 处理的是「单条完整消息」（已从 TCP 流中切分）。
// 如果字节流包含多条消息，只解码第一条——
// TCP 流上的消息切分由 FrameReader.ReadMessage 负责。
//
// 参数：
//   - data: 完整的消息字节（含包头）
//   - p: gopacket 的 PacketBuilder，用于 AddLayer 和 NextDecoder
func decodeGameMessage(data []byte, p gopacket.PacketBuilder) error {
	if len(data) < MsgHeadSize {
		return fmt.Errorf("game message too short: %d bytes", len(data))
	}

	// 解析包头：msgSize(3字节小端) + flag(1字节)。
	msg := &GameMessage{
		MsgSize: uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16,
		Flag:    data[3],
	}
	msg.layerContents = data[:MsgHeadSize]

	// 解析 msgData：msgID(2字节) + seqID(4字节) + body(剩余)。
	payload := data[MsgHeadSize:]
	if len(payload) < MsgIDSize+SeqSize {
		return fmt.Errorf("game message payload too short: need >= %d, got %d",
			MsgIDSize+SeqSize, len(payload))
	}

	msg.MsgID = binary.LittleEndian.Uint16(payload[0:2])
	msg.SeqID = binary.LittleEndian.Uint32(payload[2:6])
	msg.Body = payload[6:]
	msg.layerPayload = payload

	// 将本 Layer 添加到 gopacket 的 Layer 链。
	p.AddLayer(msg)
	// 不调用 p.NextDecoder()——body 是应用层终点，
	// protobuf 解析在上层（使用者自己的代码或 DumpProtobufRaw）处理。
	return nil
}

// -----------------------------------------------------------------------------
// FrameReader：从重组后的 TCP 字节流逐条切分游戏消息
// -----------------------------------------------------------------------------
// FrameReader 是实际推荐使用的消息切分器。
// 配合 pcap.TCPStreamHandler 的 io.Reader 回调使用：
//
//	handler := pcap.NewTCPStreamHandler("game", func(flow pcap.FlowKey, r io.Reader) error {
//	    reader := gameproto.NewFrameReader(r, fromClient)
//	    for {
//	        msgID, seqID, body, err := reader.ReadMessage()
//	        if err != nil { return err }
//	        // 处理消息...
//	    }
//	})
//
// 它的工作原理对应 go-service 的 msg_reader.go::readARQMsg：
// 循环读取「4字节包头 → msgSize字节msgData → 解密/解压 → 提取 msgID/seqID/body」。
// -----------------------------------------------------------------------------

// FrameReader 从 io.Reader 中按游戏协议格式逐条切分消息。
type FrameReader struct {
	r          io.Reader // 重组后的字节流（来自 pcap.TCPStreamHandler 的回调）
	fromClient bool      // 流方向，决定解密时用哪个密钥
}

// NewFrameReader 创建一个帧读取器。
//
// 参数：
//   - r: 重组后的字节流（来自 pcap.TCPStreamHandler 的 io.Reader 回调）。
//   - fromClient: 流方向。
//     true = 客户端→服务端的流（客户端用 decryptKey 加密，解密也用 decryptKey）。
//     false = 服务端→客户端的流（服务端用 encryptKey 加密，解密也用 encryptKey）。
//
// 方向判断方法：比较源端口和目标端口。
// 客户端用临时高端口（如 54315），服务端用固定端口（如 18000）。
// srcPort > dstPort → 客户端→服务端（fromClient=true）。
func NewFrameReader(r io.Reader, fromClient bool) *FrameReader {
	return &FrameReader{r: r, fromClient: fromClient}
}

// ReadMessage 从字节流中读取一条完整的游戏消息。
//
// 返回值：
//   - msgID: 消息类型（< 1000 框架消息，>= 1000 游戏消息）
//   - seqID: 序列号
//   - body: 消息体（已解密、已解压），格式因 msgID 而异
//   - err: io.EOF 表示流结束；其它 error 表示消息格式错误
//
// 内部步骤（对应 msg_reader.go::readARQMsg）：
//  1. 读 4 字节包头（msgSize + flag）
//  2. 读 msgSize 字节的 msgData
//  3. 如 flag bit1=1，解密（异或+循环移位）
//  4. 如 flag bit0=1，解压（snappy）
//  5. 提取 msgID(2B) + seqID(4B) + body(剩余)
func (fr *FrameReader) ReadMessage() (msgID uint16, seqID uint32, body []byte, err error) {
	// 步骤 1：读 4 字节包头。
	var head [MsgHeadSize]byte
	if _, err = io.ReadFull(fr.r, head[:]); err != nil {
		return 0, 0, nil, err
	}

	// 解析 msgSize（3 字节小端）和 flag（1 字节）。
	msgSize := uint32(head[0]) | uint32(head[1])<<8 | uint32(head[2])<<16
	flag := head[3]

	// 校验 msgSize 合法性（防止畸形包导致分配过大内存）。
	if msgSize > MaxMsgSize || msgSize < MsgIDSize+SeqSize {
		return 0, 0, nil, fmt.Errorf("msgSize 非法: %d (flag=%d)", msgSize, flag)
	}

	// 步骤 2：读 msgSize 字节的 msgData。
	msgData := make([]byte, msgSize)
	if _, err = io.ReadFull(fr.r, msgData); err != nil {
		return 0, 0, nil, err
	}

	// 步骤 3：解密（如 flag bit1 置位）。
	// 密钥方向：客户端→服务端用 decryptKey，服务端→客户端用 encryptKey。
	if flag&0x2 != 0 {
		key := encryptKey
		if fr.fromClient {
			key = decryptKey // 客户端发的，客户端用 decryptKey 加密
		}
		msgData = decryptData(msgData, key)
	}

	// 步骤 4：解压（如 flag bit0 置位）。
	// go-service 使用 snappy 压缩算法。
	if flag&0x1 != 0 {
		msgData, err = snappy.Decode(nil, msgData)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("snappy 解压失败: %w", err)
		}
	}

	// 步骤 5：提取 msgID + seqID + body。
	// msgData 布局：[msgID(2B)][seqID(4B)][body(剩余)]
	if len(msgData) < MsgIDSize+SeqSize {
		return 0, 0, nil, fmt.Errorf("解密/解压后数据过短: %d bytes", len(msgData))
	}
	msgID = binary.LittleEndian.Uint16(msgData[0:2])
	seqID = binary.LittleEndian.Uint32(msgData[2:6])
	body = msgData[6:]

	return msgID, seqID, body, nil
}

// -----------------------------------------------------------------------------
// 加密/解密（移植自 go-service/framework/net/internal/internal/crypt）
// -----------------------------------------------------------------------------
// 算法：逐字节「异或 + 循环移位」，密钥 8 字节、硬编码。
//
// 加密（encryptData）：左移 n 位 → 异或 key
// 解密（decryptData）：异或 key → 右移 n 位（加密的逆操作）
//
// 这不是标准的加密算法（如 AES），安全性很低，但在游戏协议中常见——
// 目的是防止简单的抓包工具直接读出明文，而非抵抗密码学攻击。
//
// 密钥方向（与 go-service 一致）：
//   - encryptKey：服务端加密（发→客户端），客户端解密也用它
//   - decryptKey：客户端加密（发→服务端），服务端解密也用它
//
// 抓包侧解密：根据流方向选密钥即可（密钥和算法完全公开）。
// -----------------------------------------------------------------------------

var (
	// encryptKey 服务端→客户端方向的密钥（服务端用它加密，客户端用它解密）。
	// 来源：go-service/framework/net/internal/internal/crypt/key.go
	encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}

	// decryptKey 客户端→服务端方向的密钥（客户端用它加密，服务端用它解密）。
	// 来源：go-service/framework/net/internal/internal/crypt/key.go
	decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}
)

// decryptData 解密数据（对应 go-service crypt.DecryptData）。
//
// 逐字节操作，对第 i 个字节：
//  1. 先与 key[i % 8] 异或（撤销加密时的异或）
//  2. 再循环右移 n 位（撤销加密时的左移），n = i%7 + 1
//
// 注意：n 随字节索引变化（i%7+1，范围 1-7），因此这不是简单的重复异或，
// 而是类似流密码的结构（每个位置的变换不同）。
//
// 参数：
//   - buf: 待解密数据（就地修改，返回同一 slice）
//   - key: 8 字节密钥（encryptKey 或 decryptKey）
func decryptData(buf []byte, key []byte) []byte {
	keylen := len(key)
	for i := 0; i < len(buf); i++ {
		// 步骤 1：异或。
		b := buf[i] ^ key[i%keylen]
		// 步骤 2：循环右移 n 位（n = i%7+1，范围 1-7）。
		n := byte(i%7 + 1)
		buf[i] = (b >> n) | (b << (8 - n))
	}
	return buf
}

// encryptData 加密数据（对应 go-service crypt.EncryptData）。
// 是 decryptData 的逆操作：先循环左移 n 位，再异或 key。
// 在抓包监控场景中通常不需要加密（只需要解密），但保留以供完整性。
func encryptData(buf []byte, key []byte) []byte {
	keylen := len(key)
	for i := 0; i < len(buf); i++ {
		// 步骤 1：循环左移 n 位（n = i%7+1，范围 1-7）。
		n := byte(i%7 + 1)
		b := (buf[i] << n) | (buf[i] >> (8 - n))
		// 步骤 2：异或。
		buf[i] = b ^ key[i%keylen]
	}
	return buf
}

// -----------------------------------------------------------------------------
// LoginReq / LoginResp 的 ByteStream 格式解析
// -----------------------------------------------------------------------------
// LoginReq 和 LoginResp 是框架消息（MsgID=1/2），不走 protobuf，
// 使用 go-service 自定义的 ByteStream 格式（[2B len][value] 序列）。
// -----------------------------------------------------------------------------

// readByteStreamString 从 ByteStream 中读取一个 [2B len LE][string] 字段。
// 返回读取到的字符串和剩余的字节（便于链式调用）。
//
// ByteStream 格式中字符串字段的编码方式：
//
//	[2 字节长度（小端）][字符串内容（UTF-8，无终止符）]
//
// 例如 "test1" 编码为：05 00 74 65 73 74 31
func readByteStreamString(data []byte) (string, []byte) {
	if len(data) < 2 {
		return "", data
	}
	strLen := int(data[0]) | int(data[1])<<8
	if strLen > len(data)-2 {
		return fmt.Sprintf("(len=%d, but only %d bytes left)", strLen, len(data)-2), data[2:]
	}
	return string(data[2 : 2+strLen]), data[2+strLen:]
}

// DumpLoginReq 解析 LoginReq（MsgID=1）的 ByteStream 格式并返回可读文本。
//
// LoginReq 的 body 布局（来自 proto-test msg/auth.go::sendLoginReq）：
//
//	[2B len][Account 字符串]      ← 用户账号（如 "test1"）
//	[2B len][Token 字符串]        ← HTTP 登录获取的令牌
//	[8B uint64]                   ← 平台标识/标记位（通常为 0）
//	[2B len][Version 字符串]      ← 客户端版本号（如 "0.5.3013.1.0"）
//	[...]                         ← 后续字段（协议版本、预留等）
//
// 输出示例：
//
//	Account: "test1"
//	Token: "ddd54050afea97ac..."
//	Platform/Flags: 0
//	Version: "0.5.3013.1.0"
func DumpLoginReq(data []byte) string {
	var result string
	remaining := data

	account, remaining := readByteStreamString(remaining)
	result += fmt.Sprintf("  Account: %q\n", account)

	token, remaining := readByteStreamString(remaining)
	if len(token) > 16 {
		token = token[:16] + "..." // 截断长 token，避免日志过长
	}
	result += fmt.Sprintf("  Token: %q\n", token)

	if len(remaining) >= 8 {
		platform := binary.LittleEndian.Uint64(remaining[:8])
		result += fmt.Sprintf("  Platform/Flags: %d\n", platform)
		remaining = remaining[8:]
	}

	version, remaining := readByteStreamString(remaining)
	result += fmt.Sprintf("  Version: %q\n", version)

	return result
}

// DumpLoginResp 解析 LoginResp（MsgID=2）的 ByteStream 格式并返回可读文本。
//
// LoginResp 的 body 布局（来自 proto-test msg/auth.go::waitLoginResp）：
//
//	[8B uint64 UID]               ← 用户唯一 ID
//	[4B uint32 Result]            ← 登录结果码（0=成功，非 0=失败原因）
//	[2B len][Version 字符串]      ← 服务端版本号
//	[...]                         ← 后续字段
//
// 输出示例：
//
//	UID: 4294967952
//	Result: 0 (成功)
//	Version: "0.8.3001.1.1"
func DumpLoginResp(data []byte) string {
	var result string
	if len(data) >= 12 {
		uid := binary.LittleEndian.Uint64(data[:8])
		resultCode := binary.LittleEndian.Uint32(data[8:12])
		result += fmt.Sprintf("  UID: %d\n", uid)
		result += fmt.Sprintf("  Result: %d (%s)\n", resultCode, loginResultText(resultCode))
		data = data[12:]
	}
	version, data := readByteStreamString(data)
	if version != "" {
		result += fmt.Sprintf("  Version: %q\n", version)
	}
	return result
}

// loginResultText 返回登录结果码的可读文本。
func loginResultText(code uint32) string {
	if code == 0 {
		return "成功"
	}
	return fmt.Sprintf("失败(%d)", code)
}

// -----------------------------------------------------------------------------
// 消息名称映射
// -----------------------------------------------------------------------------

// MsgName 返回消息 ID 的可读名称。
//
// 消息 ID 的分段规则（来自 go-service msgId.pb.go）：
//   - < 1000: 框架消息（LoginReq/Ping/Pong/KickOut 等），非 protobuf。
//   - 1001-1999: 大厅消息（创建角色、修改名称等）。
//   - >= 2000: 游戏逻辑消息（战斗、匹配等）。
//
// 框架消息（< 1000）的名称在此硬编码。
// 游戏消息（>= 1000）的名称需要使用者的 proto 注册表（msg_registry.go）映射，
// 此处返回 "GameMsg(ID)" 占位。
func MsgName(msgID uint16) string {
	switch msgID {
	case 1:
		return "LoginReq"
	case 2:
		return "LoginResp"
	case 3:
		return "Ping"
	case 4:
		return "Pong"
	case 10:
		return "KickOut"
	default:
		if msgID >= 1000 {
			return fmt.Sprintf("GameMsg(%d)", msgID)
		}
		return fmt.Sprintf("Unknown(%d)", msgID)
	}
}

// -----------------------------------------------------------------------------
// Protobuf wire format 原始解析（不依赖 .pb.go 文件）
// -----------------------------------------------------------------------------
// DumpProtobufRaw 将 protobuf 字节流解析为可读的字段列表。
// 类似 `protoc --decode_raw`——不需要任何 .proto 或 .pb.go 文件，
// 直接解析 wire format，适用于调试/监控场景。
//
// 支持的 wire type：
//   - 0 (varint): 变长整数，直接输出数值。
//   - 1 (64bit): 固定 8 字节，输出十六进制。
//   - 2 (bytes): 变长字节串，自动判断是 string 还是嵌套 message。
//   - 5 (32bit): 固定 4 字节，输出十六进制。
//
// 对于 wire type 2 (bytes)，自动尝试递归解析为嵌套 message，
// 如果内容全是可打印 ASCII 则当作 string 输出。
// -----------------------------------------------------------------------------

// DumpProtobufRaw 将 protobuf 字节流解析为可读的多行文本。
// 每行一个字段：fieldNumber(typeName): value
// 嵌套 message 缩进显示。
func DumpProtobufRaw(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}
	var result string
	for len(data) > 0 {
		// protowire.ConsumeField 解析一个完整的 field（tag + value）。
		// 返回 fieldNumber, wireType, 消费的字节数 n。
		num, wireType, n := protowire.ConsumeField(data)
		if n < 0 {
			// 解析失败——可能遇到 gogo protobuf 的自定义扩展或数据不完整。
			result += fmt.Sprintf("  (decode error at %d bytes)\n", len(data))
			break
		}
		fieldData := data[:n] // 本 field 的完整字节（含 tag）
		data = data[n:]       // 剩余字节

		typeName := wireTypeName(wireType)
		value := decodeFieldValue(fieldData, wireType, num)
		result += fmt.Sprintf("  field %d (%s): %s\n", num, typeName, value)
	}
	return result
}

// wireTypeName 返回 protobuf wire type 的可读名称。
func wireTypeName(wt protowire.Type) string {
	switch wt {
	case 0:
		return "varint"
	case 1:
		return "64bit"
	case 2:
		return "bytes"
	case 5:
		return "32bit"
	default:
		return fmt.Sprintf("wire%d", wt)
	}
}

// decodeFieldValue 尝试将一个 protobuf field 的值解码为可读形式。
//
// 对于 wire type 2 (bytes)，尝试智能判断内容类型：
//   - 如果全是可打印 ASCII（0x20-0x7e）→ 当作 string 输出。
//   - 否则尝试递归解析为嵌套 message。
//   - 如果嵌套解析也失败 → 输出十六进制。
func decodeFieldValue(data []byte, wt protowire.Type, num protowire.Number) string {
	switch wt {
	case 0: // varint：变长整数（int32/int64/bool/enum 等）。
		v, _ := protowire.ConsumeVarint(data)
		return fmt.Sprintf("%d", v)
	case 1: // 64bit：固定 8 字节（double/fixed64）。
		v, _ := protowire.ConsumeFixed64(data)
		return fmt.Sprintf("0x%016x", v)
	case 2: // bytes：变长字节串（string/嵌套message/repeated bytes）。
		b, _ := protowire.ConsumeBytes(data)
		if isPrintable(b) {
			// 全是可打印 ASCII → 当作 string。
			return fmt.Sprintf("%q", string(b))
		}
		// 尝试当作嵌套 message 递归解析。
		nested := DumpProtobufRaw(b)
		return fmt.Sprintf("{\n%s  }", indent(nested))
	case 5: // 32bit：固定 4 字节（float/fixed32）。
		v, _ := protowire.ConsumeFixed32(data)
		return fmt.Sprintf("0x%08x", v)
	default:
		return fmt.Sprintf("%x", data)
	}
}

// isPrintable 判断字节是否全是可打印 ASCII 字符（0x20-0x7e）。
// 用于区分 protobuf bytes 字段是 string 还是二进制/嵌套 message。
func isPrintable(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

// indent 给多行文本的每一行加两个空格缩进（用于嵌套 message 的层级显示）。
func indent(s string) string {
	var result string
	for _, line := range splitLines(s) {
		result += "  " + line + "\n"
	}
	return result
}

// splitLines 按换行符分割字符串（兼容 \n 和 \r\n）。
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
