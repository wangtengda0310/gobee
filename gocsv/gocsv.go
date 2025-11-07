package gocsv

import (
	"github.com/gocarina/gocsv"
)

// Unmarshaler 定义了将CSV数据转换为结构体的接口
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

// GoCSVUnmarshaler 使用gocarina/gocsv库实现Unmarshaler接口
type GoCSVUnmarshaler struct{}

// NewGoCSVUnmarshaler 创建新的GoCSVUnmarshaler实例
func NewGoCSVUnmarshaler() *GoCSVUnmarshaler {
	return &GoCSVUnmarshaler{}
}

// Unmarshal 将CSV数据解析到指定的结构体切片中
func (g *GoCSVUnmarshaler) Unmarshal(data []byte, v interface{}) error {
	return gocsv.UnmarshalBytes(data, v)
}