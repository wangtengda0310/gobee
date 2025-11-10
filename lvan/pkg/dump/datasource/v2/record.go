package v2

import (
	"fmt"
	"reflect"
	"time"
)

// FieldType 字段类型枚举
type FieldType int

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeString
	FieldTypeInt
	FieldTypeInt64
	FieldTypeFloat64
	FieldTypeBool
	FieldTypeTime
	FieldTypeBytes
	FieldTypeNull
)

// String 返回字段类型的字符串表示
func (ft FieldType) String() string {
	switch ft {
	case FieldTypeString:
		return "string"
	case FieldTypeInt:
		return "int"
	case FieldTypeInt64:
		return "int64"
	case FieldTypeFloat64:
		return "float64"
	case FieldTypeBool:
		return "bool"
	case FieldTypeTime:
		return "time"
	case FieldTypeBytes:
		return "bytes"
	case FieldTypeNull:
		return "null"
	default:
		return "unknown"
	}
}

// Record 统一数据记录接口
// 定义数据记录的基本操作和行为
type Record interface {
	// 字段操作
	SetField(name string, value interface{})
	GetField(name string) (interface{}, error)
	GetString(name string) (string, error)
	GetInt(name string) (int, error)
	GetInt64(name string) (int64, error)
	GetFloat64(name string) (float64, error)
	GetBool(name string) (bool, error)
	GetTime(name string) (time.Time, error)
	GetBytes(name string) ([]byte, error)
	HasField(name string) bool
	GetFields() []string
	GetFieldType(name string) FieldType
	RemoveField(name string)

	// 元数据操作
	SetMetadata(key, value string)
	GetMetadata(key string) string
	GetAllMetadata() map[string]string
	ClearMetadata()

	// 序列化操作
	ToMap() map[string]interface{}
	FromMap(data map[string]interface{}) error
	Clone() Record

	// 验证操作
	Validate() error
}

// recordImpl Record的默认实现
type recordImpl struct {
	fields    map[string]interface{}
	fieldTypes map[string]FieldType
	metadata  map[string]string
}

// NewRecord 创建新的Record实例
func NewRecord() Record {
	return &recordImpl{
		fields:     make(map[string]interface{}),
		fieldTypes: make(map[string]FieldType),
		metadata:   make(map[string]string),
	}
}

// NewRecordWithData 从map创建Record实例
func NewRecordWithData(data map[string]interface{}) Record {
	record := NewRecord()
	record.FromMap(data)
	return record
}

// inferType 推断值的类型
func inferType(value interface{}) FieldType {
	if value == nil {
		return FieldTypeNull
	}

	switch value.(type) {
	case string:
		return FieldTypeString
	case int:
		return FieldTypeInt
	case int64:
		return FieldTypeInt64
	case float64:
		return FieldTypeFloat64
	case bool:
		return FieldTypeBool
	case time.Time:
		return FieldTypeTime
	case []byte:
		return FieldTypeBytes
	default:
		// 尝试反射推断
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.String:
			return FieldTypeString
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			return FieldTypeInt
		case reflect.Int64:
			return FieldTypeInt64
		case reflect.Float32, reflect.Float64:
			return FieldTypeFloat64
		case reflect.Bool:
			return FieldTypeBool
		case reflect.Slice:
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				return FieldTypeBytes
			}
		}
		return FieldTypeUnknown
	}
}

// SetField 设置字段值
func (r *recordImpl) SetField(name string, value interface{}) {
	if r.fields == nil {
		r.fields = make(map[string]interface{})
		r.fieldTypes = make(map[string]FieldType)
	}
	r.fields[name] = value
	r.fieldTypes[name] = inferType(value)
}

// GetField 获取字段值
func (r *recordImpl) GetField(name string) (interface{}, error) {
	if r.fields == nil {
		return nil, fmt.Errorf("字段 '%s' 不存在", name)
	}

	value, exists := r.fields[name]
	if !exists {
		return nil, fmt.Errorf("字段 '%s' 不存在", name)
	}

	return value, nil
}

// GetString 获取字符串字段值
func (r *recordImpl) GetString(name string) (string, error) {
	value, err := r.GetField(name)
	if err != nil {
		return "", err
	}

	switch value.(type) {
	case string:
		return value.(string), nil
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", value), nil
	default:
		return "", fmt.Errorf("字段 '%s' 不是字符串类型，实际类型: %T", name, value)
	}
}

// GetInt 获取整数字段值
func (r *recordImpl) GetInt(name string) (int, error) {
	value, err := r.GetField(name)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		if v > int64(^uint(0)>>1) || v < int64(^int(0)>>1) {
			return 0, fmt.Errorf("字段 '%s' 的int64值 %d 超出int范围", name, v)
		}
		return int(v), nil
	case float64:
		if v != float64(int(v)) {
			return 0, fmt.Errorf("字段 '%s' 的float64值 %f 不是整数", name, v)
		}
		return int(v), nil
	case string:
		var result int
		_, parseErr := fmt.Sscanf(v, "%d", &result)
		if parseErr != nil {
			return 0, fmt.Errorf("字段 '%s' 的字符串值 '%s' 无法转换为int", name, v)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("字段 '%s' 不是整数类型，实际类型: %T", name, value)
	}
}

// GetInt64 获取int64字段值
func (r *recordImpl) GetInt64(name string) (int64, error) {
	value, err := r.GetField(name)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		if v != float64(int64(v)) {
			return 0, fmt.Errorf("字段 '%s' 的float64值 %f 不是整数", name, v)
		}
		return int64(v), nil
	case string:
		var result int64
		_, parseErr := fmt.Sscanf(v, "%d", &result)
		if parseErr != nil {
			return 0, fmt.Errorf("字段 '%s' 的字符串值 '%s' 无法转换为int64", name, v)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("字段 '%s' 不是整数类型，实际类型: %T", name, value)
	}
}

// GetFloat64 获取float64字段值
func (r *recordImpl) GetFloat64(name string) (float64, error) {
	value, err := r.GetField(name)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		var result float64
		_, parseErr := fmt.Sscanf(v, "%f", &result)
		if parseErr != nil {
			return 0, fmt.Errorf("字段 '%s' 的字符串值 '%s' 无法转换为float64", name, v)
		}
		return result, nil
	default:
		return 0, fmt.Errorf("字段 '%s' 不是数字类型，实际类型: %T", name, value)
	}
}

// GetBool 获取布尔字段值
func (r *recordImpl) GetBool(name string) (bool, error) {
	value, err := r.GetField(name)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int64:
		return reflect.ValueOf(value).Int() != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return v == "true" || v == "1" || v == "yes", nil
	default:
		return false, fmt.Errorf("字段 '%s' 不是布尔类型，实际类型: %T", name, value)
	}
}

// GetTime 获取时间字段值
func (r *recordImpl) GetTime(name string) (time.Time, error) {
	value, err := r.GetField(name)
	if err != nil {
		return time.Time{}, err
	}

	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// 尝试解析多种时间格式
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
			"2006/01/02 15:04:05",
			"2006/01/02",
		}

		for _, format := range formats {
			if t, parseErr := time.Parse(format, v); parseErr == nil {
				return t, nil
			}
		}

		return time.Time{}, fmt.Errorf("字段 '%s' 的字符串值 '%s' 无法解析为时间", name, v)
	default:
		return time.Time{}, fmt.Errorf("字段 '%s' 不是时间类型，实际类型: %T", name, value)
	}
}

// GetBytes 获取字节数组字段值
func (r *recordImpl) GetBytes(name string) ([]byte, error) {
	value, err := r.GetField(name)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("字段 '%s' 不是字节数组类型，实际类型: %T", name, value)
	}
}

// HasField 检查字段是否存在
func (r *recordImpl) HasField(name string) bool {
	if r.fields == nil {
		return false
	}
	_, exists := r.fields[name]
	return exists
}

// GetFields 获取所有字段名
func (r *recordImpl) GetFields() []string {
	if r.fields == nil {
		return []string{}
	}

	fields := make([]string, 0, len(r.fields))
	for name := range r.fields {
		fields = append(fields, name)
	}
	return fields
}

// GetFieldType 获取字段类型
func (r *recordImpl) GetFieldType(name string) FieldType {
	if r.fieldTypes == nil {
		return FieldTypeUnknown
	}

	if fieldType, exists := r.fieldTypes[name]; exists {
		return fieldType
	}

	return FieldTypeUnknown
}

// RemoveField 移除字段
func (r *recordImpl) RemoveField(name string) {
	if r.fields != nil {
		delete(r.fields, name)
	}
	if r.fieldTypes != nil {
		delete(r.fieldTypes, name)
	}
}

// SetMetadata 设置元数据
func (r *recordImpl) SetMetadata(key, value string) {
	if r.metadata == nil {
		r.metadata = make(map[string]string)
	}
	r.metadata[key] = value
}

// GetMetadata 获取元数据
func (r *recordImpl) GetMetadata(key string) string {
	if r.metadata == nil {
		return ""
	}
	return r.metadata[key]
}

// GetAllMetadata 获取所有元数据
func (r *recordImpl) GetAllMetadata() map[string]string {
	if r.metadata == nil {
		return make(map[string]string)
	}

	// 返回副本以避免外部修改
	metadata := make(map[string]string, len(r.metadata))
	for k, v := range r.metadata {
		metadata[k] = v
	}
	return metadata
}

// ClearMetadata 清空元数据
func (r *recordImpl) ClearMetadata() {
	r.metadata = make(map[string]string)
}

// ToMap 转换为map
func (r *recordImpl) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	// 复制字段
	if r.fields != nil {
		for k, v := range r.fields {
			result[k] = v
		}
	}

	// 添加元数据（使用特殊前缀）
	if r.metadata != nil {
		for k, v := range r.metadata {
			result["_metadata_"+k] = v
		}
	}

	return result
}

// FromMap 从map创建Record
func (r *recordImpl) FromMap(data map[string]interface{}) error {
	for k, v := range data {
		// 跳过元数据字段
		if len(k) > 10 && k[:10] == "_metadata_" {
			metaKey := k[10:]
			if strValue, ok := v.(string); ok {
				r.SetMetadata(metaKey, strValue)
			}
			continue
		}

		r.SetField(k, v)
	}
	return nil
}

// Clone 克隆Record
func (r *recordImpl) Clone() Record {
	clone := NewRecord()

	// 克隆字段
	if r.fields != nil {
		for k, v := range r.fields {
			clone.SetField(k, v)
		}
	}

	// 克隆元数据
	if r.metadata != nil {
		for k, v := range r.metadata {
			clone.SetMetadata(k, v)
		}
	}

	return clone
}

// Validate 验证Record
func (r *recordImpl) Validate() error {
	// 基础验证：确保有字段
	if len(r.GetFields()) == 0 {
		return fmt.Errorf("Record不能为空")
	}

	// 验证字段名不为空
	for _, fieldName := range r.GetFields() {
		if fieldName == "" {
			return fmt.Errorf("字段名不能为空")
		}
	}

	return nil
}