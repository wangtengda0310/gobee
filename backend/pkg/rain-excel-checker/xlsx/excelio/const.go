// Package excel_internal 定义 Excel 配表解析的常量
// 包含名将杀配表和枚举配表的行索引常量
package excelio

const (
	// MJS_FIXED_ROWS_CHS 名将杀普通配表第1行（中文名称行）
	// 包含各列的中文说明，如"ID"、"名称"、"描述"等
	MJS_FIXED_ROWS_CHS = 0

	// MJS_FIXED_ROWS_TYPE 名将杀普通配表第2行（类型定义行）
	// 定义每列的数据类型，如 int、string、bool、E#枚举名 等
	MJS_FIXED_ROWS_TYPE = 1

	// MJS_FIXED_ROWS_NAME 名将杀普通配表第3行（属性名称行）
	// 定义每列的字段名，对应代码中的属性名，如 Id、Name、Description 等
	MJS_FIXED_ROWS_NAME = 2

	// MJS_FIXED_ROWS_CAS 名将杀普通配表第4行（导出标识行）
	// 标识该列数据导出到客户端、服务器还是两者都导出
	// 可选值：空、server、client、server/client、client/server
	MJS_FIXED_ROWS_CAS = 3

	// MJS_FIXED_ROWS_NUM 名将杀普通配表固定表头行数
	// 表示表头占据的总行数，实际数据从第5行开始（索引4）
	MJS_FIXED_ROWS_NUM = 4
)

const (
	// MJS_FIXED_ENUM_ROWS_CHS 名将杀枚举配表第1行（列定义行）
	// 枚举表只有1行表头，包含列名：name、value、(sign)、description
	MJS_FIXED_ENUM_ROWS_CHS = 0

	// MJS_FIXED_ENUM_ROWS_NUM 名将杀枚举配表固定表头行数
	// 表示枚举表表头占据的总行数，实际数据从第2行开始（索引1）
	MJS_FIXED_ENUM_ROWS_NUM = 1
)
