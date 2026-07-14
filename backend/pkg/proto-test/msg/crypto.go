// 帧格式常量和 XOR 加解密算法
//
// 帧头布局（4字节）:
//
//	Byte 0-2: 消息长度（3B LE，值 = 总帧长 - 4）
//	Byte 3:   标志位（bit0=压缩, bit1=加密）
//
// 消息体布局（紧跟帧头）:
//
//	Byte 0-1: MsgID（2B LE）
//	Byte 2-5: SeqID（4B LE）
//	Byte 6+:  Payload
package protocol

// 帧头大小（3B长度 + 1B标志位）
const FrameHeaderSize = 4

// MsgID 大小
const MsgIDSize = 2

// SeqID 大小
const SeqIDSize = 4

// 标志位
const (
	FlagCompress byte = 0x01 // Snappy 压缩
	FlagEncrypt  byte = 0x02 // XOR 加密
)

// 最大包大小
const MaxPacketSize = 62 * 1024

// XOR 加密密钥（来源：rain-robot/project/xcard/xcard_net_lib/key.go）
var (
	// 服务端发送数据使用的加密密钥（代理解密服务端→客户端方向时使用）
	encryptKey = []byte{253, 1, 56, 52, 62, 176, 42, 138}
	// 客户端发送数据使用的加密密钥（代理解密客户端→服务端方向时使用）
	decryptKey = []byte{41, 247, 6, 255, 138, 78, 197, 129}
)

// DecryptXOR 解密数据
// isClientData=true 表示这是客户端发送的数据（使用 decryptKey 加密）
// isClientData=false 表示这是服务端发送的数据（使用 encryptKey 加密）
func DecryptXOR(data []byte, isClientData bool) {
	var key []byte
	if isClientData {
		// 客户端发送时用 decryptKey 加密，解密用同一密钥
		key = decryptKey
	} else {
		// 服务端发送时用 encryptKey 加密，解密用同一密钥
		key = encryptKey
	}
	keylen := len(key)

	for i := 0; i < len(data); i++ {
		b := data[i] ^ key[i%keylen]
		n := byte(i%7 + 1)
		data[i] = (b >> n) | (b << (8 - n))
	}
}

// EncryptXOR 加密数据（DecryptXOR 的逆操作）
// isClientData=true 表示这是客户端发送的数据（使用 decryptKey 加密）
// isClientData=false 表示这是服务端发送的数据（使用 encryptKey 加密）
func EncryptXOR(data []byte, isClientData bool) {
	var key []byte
	if isClientData {
		key = decryptKey
	} else {
		key = encryptKey
	}
	keylen := len(key)

	for i := 0; i < len(data); i++ {
		n := byte(i%7 + 1)
		b := data[i]
		// 循环左移 n 位
		b = (b << n) | (b >> (8 - n))
		data[i] = b ^ key[i%keylen]
	}
}
