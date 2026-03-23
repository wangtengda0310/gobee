package tool

import "errors"

var (
	// ErrToolNotFound 工具未找到
	ErrToolNotFound = errors.New("工具未找到")

	// ErrToolAlreadyExists 工具已存在
	ErrToolAlreadyExists = errors.New("工具已存在")

	// ErrNoHandler 未设置处理函数
	ErrNoHandler = errors.New("未设置处理函数")

	// ErrInvalidArguments 无效参数
	ErrInvalidArguments = errors.New("无效参数")

	// ErrExecutionFailed 执行失败
	ErrExecutionFailed = errors.New("执行失败")

	// ErrTimeout 执行超时
	ErrTimeout = errors.New("执行超时")
)

// ToolError 工具执行错误
// 包含工具名称和原始错误
type ToolError struct {
	Name    string
	Message string
	Cause   error
}

// Error 实现 error 接口
func (e *ToolError) Error() string {
	if e.Cause != nil {
		return e.Name + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Name + ": " + e.Message
}

// Unwrap 返回原始错误
func (e *ToolError) Unwrap() error {
	return e.Cause
}

// NewToolError 创建工具错误
func NewToolError(name, message string, cause error) *ToolError {
	return &ToolError{
		Name:    name,
		Message: message,
		Cause:   cause,
	}
}
