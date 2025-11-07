package csvunmarshaler

// Unmarshaler 定义了将CSV数据转换为结构体的接口
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

// UnmarshalerFunc 函数类型，实现Unmarshaler接口
type UnmarshalerFunc func(data []byte, v interface{}) error

// Unmarshal 实现Unmarshaler接口
func (f UnmarshalerFunc) Unmarshal(data []byte, v interface{}) error {
	return f(data, v)
}