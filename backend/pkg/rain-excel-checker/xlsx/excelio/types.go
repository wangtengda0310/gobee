// Package excel_internal 提供 Excel 文件内部结构定义和类型解析
// 本包定义了 Excel 表格的内部数据结构，包括 Sheet、列属性、枚举类型等核心数据模型
package excelio

import "github.com/gogo/protobuf/proto"

// Sheet 表示 Excel 中的一个工作表
// 包含工作表名称、类型、表头信息和错误状态
type Sheet struct {
	// Name 工作表名称
	Name string
	// SheetType 工作表类型（普通表、枚举表等）
	SheetType SheetType
	// Header 工作表表头信息，包含列属性定义
	Header *SheetHeader
	// Error 错误信息，如果不为空说明解析过程中发现问题
	Error string
}

// SheetHeader 工作表表头信息
// 包含所有列的属性定义集合
type SheetHeader struct {
	// Col 列属性集合
	Col []*ColAttr
}

// ColAttr 列属性定义
// 描述 Excel 表格中某一列的类型、名称、状态等信息
type ColAttr struct {
	// AttrName 属性名称（字段名，如 Id、Name 等）
	AttrName string
	// AttrType 属性类型（如 int、string、bool、E#枚举名 等）
	AttrType string
	// AttrCHS 属性中文名称（用于显示和文档说明）
	AttrCHS string
	// ColStatus 列状态（正常列、注释列、枚举列、空列、错误列）
	ColStatus EColType
	// Error 错误信息，如果不为空说明该列存在解析问题
	Error string
}

// EColType 列状态枚举类型
// 定义 Excel 列的多种状态类型
type EColType int32

const (
	// NORMAL 正常列：包含有效的属性名和类型定义
	NORMAL EColType = iota
	// COMMENT 注释列：以 "#" 标记，用于策划备注，不参与数据导出
	COMMENT
	// ENUM 枚举列：以 "E#" 开头的类型定义
	ENUM
	// EMPTY 空列：类型和名称都为空的列
	EMPTY
	// ERROR 错误列：类型和名称不匹配或格式错误的列
	ERROR
)

// ColType_name 列状态枚举值到名称的映射
// 用于将枚举整数值转换为可读的字符串名称
var ColType_name = map[int32]string{
	0: "NORMAL",
	1: "COMMENT",
	2: "ENUM",
	3: "EMPTY",
	4: "ERROR",
}

// ColType_value 列状态名称到枚举值的映射
// 用于将字符串名称转换为枚举整数值
var ColType_value = map[string]int32{
	"NORMAL":  0,
	"COMMENT": 1,
	"ENUM":    2,
	"EMPTY":   3,
	"ERROR":   4,
}

// String 返回列状态的字符串表示
// 实现 proto.EnumName 接口，用于序列化和日志输出
func (x EColType) String() string {
	return proto.EnumName(ColType_name, int32(x))
}

// SheetType 工作表类型枚举
// 定义 Excel 工作表的分类类型
type SheetType int32

const (
	// NONE 未定义类型
	NONE SheetType = iota
	// MING_JIANG_SHA 普通名将杀配表（标准4行表头：中文、类型、字段名、导出标识）
	MING_JIANG_SHA
	// MING_JIANG_SHA_ENUM 名将杀枚举配表（特殊格式：name、value、description）
	MING_JIANG_SHA_ENUM
)

// SheetType_name 工作表类型枚举值到名称的映射
// 用于将枚举整数值转换为可读的字符串名称
var SheetType_name = map[int32]string{
	0: "NONE",
	1: "MING_JIANG_SHA",
	2: "MING_JIANG_SHA_ENUM",
}

// SheetType_value 工作表类型名称到枚举值的映射
// 用于将字符串名称转换为枚举整数值
var SheetType_value = map[string]int32{
	"NONE":                0,
	"MING_JIANG_SHA":      1,
	"MING_JIANG_SHA_ENUM": 2,
}

// String 返回工作表类型的字符串表示
// 实现 proto.EnumName 接口，用于序列化和日志输出
func (x SheetType) String() string {
	return proto.EnumName(SheetType_name, int32(x))
}
