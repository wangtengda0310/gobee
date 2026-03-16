package llm

import "fmt"

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeInvalidRequest ErrorType = "invalid_request" // 请求无效
	ErrorTypeAuthentication ErrorType = "authentication"  // 认证失败
	ErrorTypePermission     ErrorType = "permission"      // 权限不足
	ErrorTypeNotFound       ErrorType = "not_found"       // 资源不存在
	ErrorTypeRateLimit      ErrorType = "rate_limit"      // 速率限制
	ErrorTypeServerError    ErrorType = "server_error"    // 服务器错误
	ErrorTypeTimeout        ErrorType = "timeout"         // 超时
	ErrorTypeOverloaded     ErrorType = "overloaded"      // 服务过载
	ErrorTypeContextLength  ErrorType = "context_length"  // 上下文长度超限
)

// LLMError LLM 相关错误
type LLMError struct {
	// Type 错误类型
	Type ErrorType `json:"type"`

	// Message 错误消息
	Message string `json:"message"`

	// Code 提供商特定的错误码
	Code string `json:"code,omitempty"`

	// StatusCode HTTP 状态码
	StatusCode int `json:"status_code,omitempty"`

	// Provider 提供商名称
	Provider string `json:"provider,omitempty"`

	// RetryAfter 重试等待时间 (秒)
	RetryAfter int `json:"retry_after,omitempty"`
}

func (e *LLMError) Error() string {
	if e.Provider != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Type, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 返回底层错误 (用于 errors.As)
func (e *LLMError) Unwrap() error {
	return nil
}

// NewLLMError 创建 LLM 错误
func NewLLMError(typ ErrorType, message string) *LLMError {
	return &LLMError{
		Type:    typ,
		Message: message,
	}
}

// WithCode 设置错误码
func (e *LLMError) WithCode(code string) *LLMError {
	e.Code = code
	return e
}

// WithStatusCode 设置 HTTP 状态码
func (e *LLMError) WithStatusCode(code int) *LLMError {
	e.StatusCode = code
	return e
}

// WithProvider 设置提供商
func (e *LLMError) WithProvider(provider string) *LLMError {
	e.Provider = provider
	return e
}

// WithRetryAfter 设置重试等待时间
func (e *LLMError) WithRetryAfter(seconds int) *LLMError {
	e.RetryAfter = seconds
	return e
}

// IsRetryable 检查错误是否可重试
func (e *LLMError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeRateLimit, ErrorTypeServerError, ErrorTypeTimeout, ErrorTypeOverloaded:
		return true
	default:
		return false
	}
}

// 错误构造函数

// ErrInvalidRequest 请求无效错误
func ErrInvalidRequest(message string) *LLMError {
	return NewLLMError(ErrorTypeInvalidRequest, message)
}

// ErrAuthentication 认证失败错误
func ErrAuthentication(message string) *LLMError {
	return NewLLMError(ErrorTypeAuthentication, message)
}

// ErrPermission 权限不足错误
func ErrPermission(message string) *LLMError {
	return NewLLMError(ErrorTypePermission, message)
}

// ErrNotFound 资源不存在错误
func ErrNotFound(message string) *LLMError {
	return NewLLMError(ErrorTypeNotFound, message)
}

// ErrRateLimit 速率限制错误
func ErrRateLimit(message string) *LLMError {
	return NewLLMError(ErrorTypeRateLimit, message)
}

// ErrServerError 服务器错误
func ErrServerError(message string) *LLMError {
	return NewLLMError(ErrorTypeServerError, message)
}

// ErrTimeout 超时错误
func ErrTimeout(message string) *LLMError {
	return NewLLMError(ErrorTypeTimeout, message)
}

// ErrOverloaded 服务过载错误
func ErrOverloaded(message string) *LLMError {
	return NewLLMError(ErrorTypeOverloaded, message)
}

// ErrContextLength 上下文长度超限错误
func ErrContextLength(message string) *LLMError {
	return NewLLMError(ErrorTypeContextLength, message)
}
