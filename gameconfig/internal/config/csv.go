package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CSVReaderOptions CSV 读取器选项
type CSVReaderOptions struct {
	// SkipHeader 是否跳过表头行
	SkipHeader bool
	// Comment 注释行字符（0 表示无注释支持）
	Comment rune
	// TrimSpace 是否去除字段前后空格
	TrimSpace bool
}

// CSVReader CSV 读取器
type CSVReader struct {
	filePath string
	file     *os.File
	reader   *csv.Reader
	options  CSVReaderOptions
}

// NewCSVReader 创建 CSV 读取器（使用默认选项）
func NewCSVReader(filePath string) *CSVReader {
	return NewCSVReaderWithOptions(filePath, CSVReaderOptions{
		SkipHeader: false,
		Comment:    0,
		TrimSpace:  true,
	})
}

// NewCSVReaderWithOptions 创建 CSV 读取器（使用自定义选项）
func NewCSVReaderWithOptions(filePath string, options CSVReaderOptions) *CSVReader {
	return &CSVReader{
		filePath: filePath,
		options:  options,
	}
}

// Read 读取 CSV 文件的所有数据
func (r *CSVReader) Read() ([][]string, error) {
	// 打开文件
	file, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, r.filePath)
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)
	reader.Comment = r.options.Comment
	reader.TrimLeadingSpace = r.options.TrimSpace

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取 CSV 数据失败: %w", err)
	}

	// 跳过表头
	if r.options.SkipHeader && len(records) > 0 {
		records = records[1:]
	}

	return records, nil
}

// Close 关闭读取器
func (r *CSVReader) Close() error {
	if r.reader != nil {
		r.reader = nil
	}
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// inferType 推断值的类型
func inferType(value string) string {
	value = strings.TrimSpace(value)

	// 空值
	if value == "" {
		return "null"
	}

	// 布尔值
	lower := strings.ToLower(value)
	if lower == "true" || lower == "false" {
		return "bool"
	}

	// 整数
	if isInt(value) {
		return "int"
	}

	// 浮点数
	if isFloat(value) {
		return "float"
	}

	// 默认为字符串
	return "string"
}

// isInt 检查字符串是否是整数
func isInt(s string) bool {
	if s == "" {
		return false
	}
	// 允许负号
	start := 0
	if s[0] == '-' {
		start = 1
		if len(s) == 1 {
			return false
		}
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isFloat 检查字符串是否是浮点数
func isFloat(s string) bool {
	if s == "" {
		return false
	}
	dotFound := false
	start := 0
	if s[0] == '-' {
		start = 1
		if len(s) == 1 {
			return false
		}
	}
	// 必须至少有一个数字
	hasDigit := false
	for i := start; i < len(s); i++ {
		if s[i] == '.' {
			if dotFound {
				return false
			}
			dotFound = true
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		hasDigit = true
	}
	// 必须有小数点且至少有一个数字
	return dotFound && hasDigit
}

// parseValue 解析字符串值到合适的 Go 类型
func parseValue(value string) (interface{}, error) {
	value = strings.TrimSpace(value)

	// 空值
	if value == "" {
		return nil, nil
	}

	// 布尔值
	lower := strings.ToLower(value)
	if lower == "true" {
		return true, nil
	}
	if lower == "false" {
		return false, nil
	}

	// 整数
	if isInt(value) {
		if strings.HasPrefix(value, "-") {
			v, err := strconv.ParseInt(value, 10, 64)
			return v, err
		}
		v, err := strconv.ParseUint(value, 10, 64)
		return v, err
	}

	// 浮点数
	if isFloat(value) {
		v, err := strconv.ParseFloat(value, 64)
		return v, err
	}

	// 默认返回字符串
	return value, nil
}

// ConvertToType 将字符串值转换为目标类型
func ConvertToType(value string, targetType string) (interface{}, error) {
	switch targetType {
	case "int":
		return strconv.ParseInt(value, 10, 64)
	case "int64":
		return strconv.ParseInt(value, 10, 64)
	case "int32":
		v, err := strconv.ParseInt(value, 10, 32)
		return int32(v), err
	case "uint":
		return strconv.ParseUint(value, 10, 64)
	case "uint64":
		return strconv.ParseUint(value, 10, 64)
	case "uint32":
		v, err := strconv.ParseUint(value, 10, 32)
		return uint32(v), err
	case "float64":
		return strconv.ParseFloat(value, 64)
	case "float32":
		v, err := strconv.ParseFloat(value, 32)
		return float32(v), err
	case "bool":
		lower := strings.ToLower(value)
		if lower == "true" {
			return true, nil
		}
		if lower == "false" {
			return false, nil
		}
		return nil, fmt.Errorf("无法将 %q 转换为 bool", value)
	case "string":
		return value, nil
	default:
		return nil, fmt.Errorf("不支持的目标类型: %s", targetType)
	}
}
