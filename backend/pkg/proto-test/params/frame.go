package params

// DecodedFrame 解码后的协议帧
// 放在 params 包是为了让 variables 等上层包无需依赖 msg 包即可引用帧类型。
type DecodedFrame struct {
	MsgID   uint16 // 消息 ID
	SeqID   uint32 // 序列号
	Flags   byte   // 标志位
	Payload []byte // 解密解压后的 payload
	RawSize int    // 原始帧总字节数（含帧头）
}
