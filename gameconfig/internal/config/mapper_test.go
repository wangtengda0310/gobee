package config

import (
	"errors"
	"reflect"
	"testing"
)

// TestParseFieldTag 测试解析字段 tag
func TestParseFieldTag(t *testing.T) {
	tests := []struct {
		tag     string
		name    string
		options map[string]string
	}{
		{
			tag:     `excel:"id"`,
			name:    "id",
			options: map[string]string{},
		},
		{
			tag:     `excel:"name,required"`,
			name:    "name",
			options: map[string]string{"required": ""},
		},
		{
			tag:     `excel:"value,default:100"`,
			name:    "value",
			options: map[string]string{"default": "100"},
		},
		{
			tag:     `excel:"field,required,default:0,opt1,opt2"`,
			name:    "field",
			options: map[string]string{"required": "", "default": "0", "opt1": "", "opt2": ""},
		},
		{
			tag:     `excel:"-"`,
			name:    "-",
			options: map[string]string{"-": ""}, // 忽略字段
		},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			name, opts := parseFieldTag(tt.tag)
			if name != tt.name {
				t.Errorf("name = %s, want %s", name, tt.name)
			}
			if !reflect.DeepEqual(opts, tt.options) {
				t.Errorf("options = %v, want %v", opts, tt.options)
			}
		})
	}
}

// TestMapRowToStruct 测试将行数据映射到结构体
func TestMapRowToStruct(t *testing.T) {
	type TestStruct struct {
		ID     int    `excel:"id"`
		Name    string `excel:"name"`
		Value   int    `excel:"value,default:0"`
		Invalid string `excel:"-"` // 应该被忽略
	}

	// 准备测试数据
	headers := []string{"id", "name", "value", "invalid"}
	row := []string{"1", "test", "100", "should_be_ignored"}

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 映射数据
	result, err := mapper.MapRow(headers, row)
	if err != nil {
		t.Fatalf("MapRow() error = %v", err)
	}

	// 验证结果
	if result.ID != 1 {
		t.Errorf("ID = %d, want 1", result.ID)
	}
	if result.Name != "test" {
		t.Errorf("Name = %s, want test", result.Name)
	}
	if result.Value != 100 {
		t.Errorf("Value = %d, want 100", result.Value)
	}
}

// TestMapRowToStruct_RequiredField 测试必填字段验证
func TestMapRowToStruct_RequiredField(t *testing.T) {
	type TestStruct struct {
		ID   int    `excel:"id,required"`
		Name string `excel:"name,required"`
	}

	// 准备测试数据（缺少 name 字段）
	headers := []string{"id", "name"}
	row := []string{"1", ""} // name 为空

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 映射数据应该返回错误
	_, err := mapper.MapRow(headers, row)
	if err == nil {
		t.Error("MapRow() should return error for missing required field")
	}

	if !errors.Is(err, ErrRequiredField) {
		t.Errorf("error = %v, want ErrRequiredField", err)
	}
}

// TestMapRowToStruct_DefaultValue 测试默认值处理
func TestMapRowToStruct_DefaultValue(t *testing.T) {
	type TestStruct struct {
		ID    int     `excel:"id"`
		Name  string  `excel:"name"`
		Value int     `excel:"value,default:999"`
		Flag  bool    `excel:"flag,default:true"`
		Count uint32  `excel:"count,default:100"`
	}

	// 准备测试数据（缺少 value 和 flag 字段）
	headers := []string{"id", "name", "value", "flag", "count"}
	row := []string{"1", "test", "", "", ""}

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 映射数据
	result, err := mapper.MapRow(headers, row)
	if err != nil {
		t.Fatalf("MapRow() error = %v", err)
	}

	// 验证默认值
	if result.Value != 999 {
		t.Errorf("Value = %d, want 999 (default)", result.Value)
	}
	if result.Flag != true {
		t.Errorf("Flag = %v, want true (default)", result.Flag)
	}
	if result.Count != 100 {
		t.Errorf("Count = %d, want 100 (default)", result.Count)
	}
}

// TestMapRowToStruct_TypeConversion 测试类型转换
func TestMapRowToStruct_TypeConversion(t *testing.T) {
	type TestStruct struct {
		IntVal     int32   `excel:"int_val"`
		UintVal    uint32  `excel:"uint_val"`
		FloatVal    float64 `excel:"float_val"`
		BoolVal     bool     `excel:"bool_val"`
	}

	// 准备测试数据
	headers := []string{"int_val", "uint_val", "float_val", "bool_val"}
	row := []string{"-123", "456", "123.45", "true"}

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 映射数据
	result, err := mapper.MapRow(headers, row)
	if err != nil {
		t.Fatalf("MapRow() error = %v", err)
	}

	// 验证类型转换
	if result.IntVal != -123 {
		t.Errorf("IntVal = %d, want -123", result.IntVal)
	}
	if result.UintVal != 456 {
		t.Errorf("UintVal = %d, want 456", result.UintVal)
	}
	if result.FloatVal != 123.45 {
		t.Errorf("FloatVal = %f, want 123.45", result.FloatVal)
	}
	if result.BoolVal != true {
		t.Errorf("BoolVal = %v, want true", result.BoolVal)
	}
}

// TestMapRowToStruct_TypeMismatch 测试类型不匹配错误
func TestMapRowToStruct_TypeMismatch(t *testing.T) {
	type TestStruct struct {
		ID int `excel:"id"`
	}

	// 准备测试数据（无效的整数值）
	headers := []string{"id"}
	row := []string{"not_a_number"}

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 映射数据应该返回错误
	_, err := mapper.MapRow(headers, row)
	if err == nil {
		t.Error("MapRow() should return error for type mismatch")
	}

	// 验证错误包含位置信息
	configErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("error type = %T, want *ConfigError", err)
	}

	if configErr.ColName != "id" {
		t.Errorf("error.ColName = %s, want id", configErr.ColName)
	}
}

// TestMapRowToStruct_UnknownColumn 测试忽略未知列
func TestMapRowToStruct_UnknownColumn(t *testing.T) {
	type TestStruct struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	// 准备测试数据（包含未知列）
	headers := []string{"id", "unknown_column", "name"}
	row := []string{"1", "ignored", "test"}

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 映射数据应该成功（忽略未知列）
	result, err := mapper.MapRow(headers, row)
	if err != nil {
		t.Fatalf("MapRow() error = %v", err)
	}

	if result.ID != 1 {
		t.Errorf("ID = %d, want 1", result.ID)
	}
	if result.Name != "test" {
		t.Errorf("Name = %s, want test", result.Name)
	}
}

// TestMapRowToStruct_NestedStruct 测试嵌套结构体（暂不支持）
func TestMapRowToStruct_NestedStruct(t *testing.T) {
	type Inner struct {
		Value int `excel:"value"`
	}

	type TestStruct struct {
		ID    int   `excel:"id"`
		Inner Inner `excel:"inner"` // 嵌套结构体
	}

	// 准备测试数据
	headers := []string{"id", "inner"}
	row := []string{"1", "ignored"}

	// 创建映射器
	mapper := NewStructMapper[TestStruct]()

	// 嵌套结构体暂不支持，应该返回错误
	_, err := mapper.MapRow(headers, row)
	if err == nil {
		t.Error("MapRow() should return error for nested struct")
	}
}
