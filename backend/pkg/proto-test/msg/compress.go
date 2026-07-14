package protocol

import "github.com/golang/snappy"

// Decompress 解压 Snappy 压缩的消息体数据
func Decompress(data []byte) ([]byte, error) {
	return snappy.Decode(nil, data)
}

// Compress 使用 Snappy 压缩数据
func Compress(data []byte) []byte {
	return snappy.Encode(nil, data)
}
