package csvutil

import (
	"encoding/csv"
	"github.com/jszwec/csvutil"
	"io"
)

// Unmarshaler 定义了将CSV数据转换为结构体的接口
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

// CSVUtilUnmarshaler 使用jszwec/csvutil库实现Unmarshaler接口
type CSVUtilUnmarshaler struct{}

// NewCSVUtilUnmarshaler 创建新的CSVUtilUnmarshaler实例
func NewCSVUtilUnmarshaler() *CSVUtilUnmarshaler {
	return &CSVUtilUnmarshaler{}
}

// Unmarshal 将CSV数据解析到指定的结构体切片中
func (c *CSVUtilUnmarshaler) Unmarshal(data []byte, v interface{}) error {
	// 使用csvutil.Unmarshal直接解析数据
	return csvutil.Unmarshal(data, v)
}

// CSVUtilDecoder csvutil的decoder包装器
type CSVUtilDecoder struct {
	decoder *csvutil.Decoder
}

// NewCSVUtilDecoder 创建新的CSVUtilDecoder实例
func NewCSVUtilDecoder(r io.Reader) (*CSVUtilDecoder, error) {
	decoder, err := csvutil.NewDecoder(csv.NewReader(r))
	if err != nil {
		return nil, err
	}
	return &CSVUtilDecoder{decoder: decoder}, nil
}

// Decode 解码下一条记录
func (d *CSVUtilDecoder) Decode(v interface{}) error {
	return d.decoder.Decode(v)
}