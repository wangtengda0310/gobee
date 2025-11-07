package encoding

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshaler 定义了将CSV数据转换为结构体的接口
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

// EncodingCSVUnmarshaler 使用标准库encoding/csv实现Unmarshaler接口
type EncodingCSVUnmarshaler struct{}

// NewEncodingCSVUnmarshaler 创建新的EncodingCSVUnmarshaler实例
func NewEncodingCSVUnmarshaler() *EncodingCSVUnmarshaler {
	return &EncodingCSVUnmarshaler{}
}

// Unmarshal 将CSV数据解析到指定的结构体切片中
func (e *EncodingCSVUnmarshaler) Unmarshal(data []byte, v interface{}) error {
	// 检查v是否为指向切片的指针
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("v must be a pointer to a slice")
	}

	// 解析CSV数据
	records, err := e.ParseCSV(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		// 只有header，没有数据行
		return nil
	}

	// 获取header行
	header := records[0]
	dataRows := records[1:]

	// 获取切片的元素类型
	sliceType := val.Elem().Type()
	elemType := sliceType.Elem()

	// 如果元素类型不是结构体，返回错误
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs")
	}

	// 创建结果切片
	resultSlice := reflect.MakeSlice(sliceType, 0, len(dataRows))

	// 处理每一行数据
	for _, row := range dataRows {
		elem := reflect.New(elemType).Elem()
		err := e.populateStruct(elem, header, row)
		if err != nil {
			return fmt.Errorf("failed to populate struct: %w", err)
		}
		resultSlice = reflect.Append(resultSlice, elem)
	}

	// 设置结果到原始切片
	val.Elem().Set(resultSlice)
	return nil
}

// ParseCSV 解析CSV字符串为记录数组
func (e *EncodingCSVUnmarshaler) ParseCSV(data string) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(data))
	return reader.ReadAll()
}

// populateStruct 根据header和row数据填充结构体
func (e *EncodingCSVUnmarshaler) populateStruct(structVal reflect.Value, header, row []string) error {
	structType := structVal.Type()

	// 创建字段名到索引的映射
	fieldMap := make(map[string]int)
	for i, fieldName := range header {
		fieldMap[strings.ToLower(fieldName)] = i
	}

	// 遍历结构体的每个字段
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		// 获取CSV标签
		csvTag := fieldType.Tag.Get("csv")
		if csvTag == "" {
			continue // 跳过没有csv标签的字段
		}

		// 查找对应的列索引
		colIndex, exists := fieldMap[strings.ToLower(csvTag)]
		if !exists || colIndex >= len(row) {
			continue // 跳过不存在的列
		}

		// 设置字段值
		err := e.setFieldValue(field, row[colIndex])
		if err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue 根据字段类型设置值
func (e *EncodingCSVUnmarshaler) setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field is not settable")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		field.SetInt(intVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}