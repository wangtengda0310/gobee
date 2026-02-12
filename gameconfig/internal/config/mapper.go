package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// FieldInfo 字段信息
type FieldInfo struct {
	Name     string       // 字段名
	Index    int          // 结构体字段索引
	Type     reflect.Type  // 字段类型
	Options  map[string]string // tag 选项
}

// StructMapper 结构体映射器
type StructMapper[T any] struct {
	typ       reflect.Type
	fields    map[string]*FieldInfo // excel tag -> FieldInfo
	fieldName map[string]*FieldInfo // 结构体字段名 -> FieldInfo
}

// NewStructMapper 创建结构体映射器
func NewStructMapper[T any]() *StructMapper[T] {
	var zero T
	typ := reflect.TypeOf(zero)

	// 如果是指针类型，获取元素类型
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	mapper := &StructMapper[T]{
		typ:       typ,
		fields:    make(map[string]*FieldInfo),
		fieldName: make(map[string]*FieldInfo),
	}

	// 解析结构体字段
	mapper.parseFields()

	return mapper
}

// parseFields 解析结构体字段
func (m *StructMapper[T]) parseFields() {
	for i := 0; i < m.typ.NumField(); i++ {
		field := m.typ.Field(i)

		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		// 解析 excel tag
		tag := field.Tag.Get("excel")
		if tag == "" {
			continue
		}

		// 解析 tag
		name, options := parseFieldTag(tag)

		// 如果是 "-"，表示忽略此字段
		if name == "-" {
			continue
		}

		fieldInfo := &FieldInfo{
			Name:    name,
			Index:   i,
			Type:    field.Type,
			Options: options,
		}

		m.fields[name] = fieldInfo
		m.fieldName[field.Name] = fieldInfo
	}
}

// MapRow 将行数据映射到结构体
func (m *StructMapper[T]) MapRow(headers []string, row []string) (T, error) {
	var zero T
	result := reflect.New(m.typ).Elem()

	// 创建列名到索引的映射
	colIndex := make(map[string]int)
	for i, h := range headers {
		colIndex[h] = i
	}

	// 遍历所有字段
	for excelName, fieldInfo := range m.fields {
		// 获取列索引
		idx, ok := colIndex[excelName]
		if !ok {
			// 列不存在，跳过
			continue
		}

		// 获取单元格值
		var valueStr string
		if idx < len(row) {
			valueStr = row[idx]
		}

		// 检查必填字段
		if _, required := fieldInfo.Options["required"]; required && valueStr == "" {
			return zero, NewConfigError("", 0, idx, excelName,
				fmt.Sprintf("缺少必填字段 '%s'", excelName), ErrRequiredField)
		}

		// 处理空值和默认值
		if valueStr == "" {
			if defaultValue, ok := fieldInfo.Options["default"]; ok {
				valueStr = defaultValue
			} else {
				// 没有默认值，使用零值
				continue
			}
		}

		// 类型转换
		value, err := convertValue(valueStr, fieldInfo.Type)
		if err != nil {
			return zero, NewConfigError("", 0, idx, excelName,
				fmt.Sprintf("无法将字符串 %q 转换为 %v 类型", valueStr, fieldInfo.Type), err)
		}

		// 设置字段值
		result.Field(fieldInfo.Index).Set(value)
	}

	return result.Interface().(T), nil
}

// MapRows 将多行数据映射到结构体切片
func (m *StructMapper[T]) MapRows(headers []string, rows [][]string) ([]T, error) {
	results := make([]T, 0, len(rows))

	for i, row := range rows {
		result, err := m.MapRow(headers, row)
		if err != nil {
			// 添加行号到错误信息
			if configErr, ok := err.(*ConfigError); ok {
				configErr.Row = i + 1 // 从 1 开始计数
				return nil, configErr
			}
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetFieldInfo 获取字段信息
func (m *StructMapper[T]) GetFieldInfo(excelName string) (*FieldInfo, bool) {
	info, ok := m.fields[excelName]
	return info, ok
}

// parseFieldTag 解析字段 tag
// 格式: excel:"name,opt1,opt2:value"
func parseFieldTag(tag string) (string, map[string]string) {
	// 去掉 `excel:"` 前缀和 `"` 后缀
	tag = strings.TrimPrefix(tag, `excel:"`)
	tag = strings.TrimSuffix(tag, `"`)

	if tag == "-" {
		return "-", map[string]string{"-": ""}
	}

	// 分割 tag
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return "", nil
	}

	name := parts[0]
	options := make(map[string]string)

	for _, part := range parts[1:] {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 1 {
			// 只有 key，没有 value
			options[kv[0]] = ""
		} else {
			// key:value
			options[kv[0]] = kv[1]
		}
	}

	return name, options
}

// convertValue 将字符串值转换为目标类型
func convertValue(value string, targetType reflect.Type) (reflect.Value, error) {
	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		if value == "" {
			// 空值返回 nil 指针
			return reflect.Zero(targetType), nil
		}
		// 递归处理指针指向的类型
		converted, err := convertValue(value, targetType.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		// 创建指针
		ptr := reflect.New(targetType.Elem())
		ptr.Elem().Set(converted)
		return ptr, nil
	}

	// 根据目标类型转换
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("无法解析整数: %w", err)
		}
		return reflect.ValueOf(intVal).Convert(targetType), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("无法解析无符号整数: %w", err)
		}
		return reflect.ValueOf(uintVal).Convert(targetType), nil

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("无法解析浮点数: %w", err)
		}
		return reflect.ValueOf(floatVal).Convert(targetType), nil

	case reflect.Bool:
		boolVal, err := parseBool(value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("无法解析布尔值: %w", err)
		}
		return reflect.ValueOf(boolVal), nil

	default:
		return reflect.Value{}, fmt.Errorf("不支持的目标类型: %v", targetType)
	}
}

// parseBool 解析布尔值
func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off", "":
		return false, nil
	default:
		return false, errors.New("无法识别的布尔值")
	}
}
