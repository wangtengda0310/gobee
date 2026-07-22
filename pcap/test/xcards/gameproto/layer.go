// Package gameproto 演示如何为 go-service 的 Unity 游戏协议实现自定义 Layer + 帧读取器。
//
// 消息格式（来自 go-service/framework/net/internal/internal/msg_reader.go）：
//
//	[4字节包头] [msgData]
//
//	包头: msgSize(3字节小端) + flag(1字节)
//	  - msgSize: msgData 的长度（不含包头本身的 4 字节）
//	  - flag bit0: 1=压缩(snappy), 0=未压缩
//	  - flag bit1: 1=加密, 0=未加密
//
//	msgData: [msgID(2字节小端)] [seqID(4字节小端)] [body...]
//	  - msgID: 消息类型（uint16，对应 proto 消息号）
//	  - seqID: 序列号（uint32）
//	  - body: 消息体（可能已压缩/加密，明文时是 protobuf 字节）
package gameproto

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"github.com/gopacket/gopacket"
)

const (
	// MsgHeadSize 包头大小：3字节msgSize + 1字节flag。
	MsgHeadSize = 4
	// MsgIDSize 消息ID大小（uint16）。
	MsgIDSize = 2
	// SeqSize 序列号大小（uint32）。
	SeqSize = 4
	// MaxMsgSize 外部消息长度上限。
	MaxMsgSize = 62 * 1024
)

// GameMessage 是一个完整的游戏消息（已从 TCP 流中切分出来）。
// 实现为 gopacket 自定义 Layer，可参与 gopacket 的 Layer 链。
type GameMessage struct {
	// MsgSize msgData 的长度（不含包头）。
	MsgSize uint32
	// Flag 包头第 4 字节：bit0=压缩, bit1=加密。
	Flag byte
	// MsgID 消息类型（uint16 小端）。
	MsgID uint16
	// SeqID 序列号（uint32 小端）。
	SeqID uint32
	// Body 消息体（如未加密未压缩，是原始 protobuf 字节）。
	Body []byte

	// layerContents / layerPayload 用于实现 gopacket.Layer 接口。
	layerContents []byte
	layerPayload  []byte
}

// LayerType 返回本自定义 Layer 的类型。
func (m *GameMessage) LayerType() gopacket.LayerType { return LayerTypeGameMessage }

// LayerContents 实现 gopacket.Layer 接口：返回本层的原始字节。
func (m *GameMessage) LayerContents() []byte { return m.layerContents }

// LayerPayload 实现 gopacket.Layer 接口：返回本层包含的 payload。
func (m *GameMessage) LayerPayload() []byte { return m.layerPayload }

// String 返回可读表示，用于日志。
func (m *GameMessage) String() string {
	return fmt.Sprintf("GameMsg{ID=%d, Seq=%d, BodyLen=%d, Encrypted=%v, Compressed=%v}",
		m.MsgID, m.SeqID, len(m.Body), m.IsEncrypted(), m.IsCompressed())
}

// IsEncrypted 是否加密。
func (m *GameMessage) IsEncrypted() bool { return m.Flag&0x2 != 0 }

// IsCompressed 是否压缩。
func (m *GameMessage) IsCompressed() bool { return m.Flag&0x1 != 0 }

// LayerTypeGameMessage 是自定义 LayerType，需在 init() 中注册。
var LayerTypeGameMessage gopacket.LayerType

func init() {
	// 注册自定义 LayerType。
	// 12345 是自定义类型编号（避开 gopacket 内置编号），Decoder 用 DecodeFunc。
	LayerTypeGameMessage = gopacket.RegisterLayerType(
		12345,
		gopacket.LayerTypeMetadata{
			Name:    "GameMessage",
			Decoder: gopacket.DecodeFunc(decodeGameMessage),
		},
	)
}

// decodeGameMessage 是 gopacket Decoder，从字节流解码出一个 GameMessage。
//
// 注意：这个 decoder 处理的是「单条完整消息」（已从 TCP 流中切分），
// 不是 TCP 流重组——流重组由 pcap.TCPStreamHandler 负责。
func decodeGameMessage(data []byte, p gopacket.PacketBuilder) error {
	if len(data) < MsgHeadSize {
		return fmt.Errorf("game message too short: %d bytes", len(data))
	}

	msg := &GameMessage{
		MsgSize: uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16,
		Flag:    data[3],
	}
	msg.layerContents = data[:MsgHeadSize]

	payload := data[MsgHeadSize:]
	if len(payload) < MsgIDSize+SeqSize {
		return fmt.Errorf("game message payload too short: need >= %d, got %d",
			MsgIDSize+SeqSize, len(payload))
	}

	msg.MsgID = binary.LittleEndian.Uint16(payload[0:2])
	msg.SeqID = binary.LittleEndian.Uint32(payload[2:6])
	msg.Body = payload[6:]
	msg.layerPayload = payload

	p.AddLayer(msg)
	// 不调用 NextDecoder——body 是应用层终点（protobuf 在上层解析）。
	return nil
}

// FrameReader 从 io.Reader 中按游戏协议格式逐条切分消息。
// 使用者把它放在 pcap.TCPStreamHandler 的回调里。
//
// fromClient 标识流的方向，决定解密时用哪个密钥：
//   - true: 客户端→服务端的流（客户端用 decryptKey 加密，抓包侧也用 decryptKey 解密）
//   - false: 服务端→客户端的流（服务端用 encryptKey 加密，抓包侧也用 encryptKey 解密）
type FrameReader struct {
	r          io.Reader
	fromClient bool
}

// NewFrameReader 创建一个帧读取器。
//   - r: 重组后的字节流（来自 pcap.TCPStreamHandler 的 io.Reader 回调）。
//   - fromClient: 流方向（客户端→服务端=true，服务端→客户端=false），决定解密密钥。
func NewFrameReader(r io.Reader, fromClient bool) *FrameReader {
	return &FrameReader{r: r, fromClient: fromClient}
}

// ReadMessage 读取一条完整消息。
// 返回 msgID, seqID, body, error。
// 对应 msg_reader.go 的 readARQMsg 逻辑，含解密（异或+循环移位）和解压（snappy）。
func (fr *FrameReader) ReadMessage() (msgID uint16, seqID uint32, body []byte, err error) {
	// 1. 读 4 字节包头。
	var head [MsgHeadSize]byte
	if _, err = io.ReadFull(fr.r, head[:]); err != nil {
		return 0, 0, nil, err
	}

	msgSize := uint32(head[0]) | uint32(head[1])<<8 | uint32(head[2])<<16
	flag := head[3]

	if msgSize > MaxMsgSize || msgSize < MsgIDSize+SeqSize {
		return 0, 0, nil, fmt.Errorf("msgSize 非法: %d (flag=%d)", msgSize, flag)
	}

	// 2. 读 msgSize 字节的 msgData。
	msgData := make([]byte, msgSize)
	if _, err = io.ReadFull(fr.r, msgData); err != nil {
		return 0, 0, nil, err
	}

	// 3. 解密（如 flag bit1 置位）。
	//    加密算法与 go-service crypt.DecryptData 相同：异或 + 循环右移。
	//    密钥方向：客户端→服务端用 decryptKey，服务端→客户端用 encryptKey。
	if flag&0x2 != 0 {
		key := encryptKey
		if fr.fromClient {
			key = decryptKey // 客户端发的，客户端用 decryptKey 加密
		}
		msgData = decryptData(msgData, key)
	}

	// 4. 解压（如 flag bit0 置位）。
	//    go-service 用 snappy 压缩。
	if flag&0x1 == 0 {
		// 未压缩，直接解析。
	} else {
		msgData, err = snappy.Decode(nil, msgData)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("snappy 解压失败: %w", err)
		}
	}

	// 5. 解析 msgID + seqID + body。
	if len(msgData) < MsgIDSize+SeqSize {
		return 0, 0, nil, fmt.Errorf("解密/解压后数据过短: %d bytes", len(msgData))
	}
	msgID = binary.LittleEndian.Uint16(msgData[0:2])
	seqID = binary.LittleEndian.Uint32(msgData[2:6])
	body = msgData[6:]

	return msgID, seqID, body, nil
}

// =============================================================================
// 加密/解密（移植自 go-service/framework/net/internal/internal/crypt）
//
// 算法：逐字节「异或 + 循环移位」，密钥 8 字节、硬编码。
//
// 加密（EncryptData）：左移 n 位 → 异或 key
// 解密（DecryptData）：异或 key → 右移 n 位（加密的逆操作）
//
// 密钥方向（与 go-service 一致）：
//   - encryptKey：服务端加密（发→客户端），客户端解密也用它
//   - decryptKey：客户端加密（发→服务端），服务端解密也用它
//
// 抓包侧解密：根据流方向选密钥即可（密钥和算法完全公开）。
// =============================================================================

var (
	// 服务端→客户端方向的密钥。
	encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}
	// 客户端→服务端方向的密钥。
	decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}
)

// decryptData 解密数据（对应 go-service crypt.DecryptData）。
// 逐字节：先异或 key，再循环右移 n 位（n = i%7+1）。
func decryptData(buf []byte, key []byte) []byte {
	keylen := len(key)
	for i := 0; i < len(buf); i++ {
		b := buf[i] ^ key[i%keylen]
		n := byte(i%7 + 1)
		buf[i] = (b >> n) | (b << (8 - n))
	}
	return buf
}

// encryptData 加密数据（对应 go-service crypt.EncryptData）。
// 逐字节：先循环左移 n 位，再异或 key（decryptData 的逆操作）。
func encryptData(buf []byte, key []byte) []byte {
	keylen := len(key)
	for i := 0; i < len(buf); i++ {
		n := byte(i%7 + 1)
		b := (buf[i] << n) | (buf[i] >> (8 - n))
		buf[i] = b ^ key[i%keylen]
	}
	return buf
}
