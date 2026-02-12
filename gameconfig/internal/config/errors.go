package config

import (
	"fmt"
)

// 错误类型
var (
	// ErrFileNotFound 文件不存在
	ErrFileNotFound = fmt.Errorf("文件不存在")

	// ErrInvalidFormat 格式错误
	ErrInvalidFormat = fmt.Errorf("格式错误")

	// ErrSheetNotFound Sheet 不存在
	ErrSheetNotFound = fmt.Errorf("Sheet 不存在")

	// ErrTypeMismatch 类型不匹配
	ErrTypeMismatch = fmt.Errorf("类型不匹配")

	// ErrRequiredField 缺少必填字段
	ErrRequiredField = fmt.Errorf("缺少必填字段")

	// ErrInvalidVersion 无效的版本号
	ErrInvalidVersion = fmt.Errorf("无效的版本号")

	// ErrMigrationFailed 迁移失败
	ErrMigrationFailed = fmt.Errorf("迁移失败")
)

// ConfigError 配置错误（包含位置信息）
type ConfigError struct {
	File     string
	Row      int
	Col      int
	ColName  string
	Msg      string
	InnerErr error
}

func (e *ConfigError) Error() string {
	loc := e.File
	if e.Row > 0 {
		loc += fmt.Sprintf(" 行%d", e.Row)
	}
	if e.Col > 0 || e.ColName != "" {
		if e.ColName != "" {
			loc += fmt.Sprintf(" 列%d (%s)", e.Col, e.ColName)
		} else {
			loc += fmt.Sprintf(" 列%d", e.Col)
		}
	}

	msg := fmt.Sprintf("配置错误 [%s]", loc)
	if e.Msg != "" {
		msg += fmt.Sprintf("\n  %s", e.Msg)
	}
	if e.InnerErr != nil {
		msg += fmt.Sprintf("\n  %v", e.InnerErr)
	}
	return msg
}

func (e *ConfigError) Unwrap() error {
	return e.InnerErr
}

// NewConfigError 创建配置错误
func NewConfigError(file string, row, col int, colName, msg string, innerErr error) *ConfigError {
	return &ConfigError{
		File:     file,
		Row:      row,
		Col:      col,
		ColName:  colName,
		Msg:      msg,
		InnerErr: innerErr,
	}
}
