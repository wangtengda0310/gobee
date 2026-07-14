// Package excel_internal 提供 Excel 列数据处理的辅助函数
// 本包包含时间解析、列数据读取、消息格式化、布尔值解析等工具函数
package excelio

import (
	"fmt"
	"strings"
	"time"
)

// ==================== 时间解析工具 ====================

// DateFormats 通用日期格式列表
// 支持多种常见的日期和时间格式
var DateFormats = []string{
	"2006-01-02 15:04:05", // 标准日期时间格式（带秒）
	"2006/01/02 15:04:05", // 斜杠分隔的日期时间格式（带秒）
	"2006-01-02",          // 标准日期格式
	"2006/01/02",          // 斜杠分隔的日期格式
}

// ParseDate 解析日期字符串（支持多种格式）
//
// 此函数会按顺序尝试所有预定义的日期格式，直到成功解析为止。
// 如果所有格式都失败，返回零值时间。
//
// 参数：
//   - dateStr: 日期字符串
//
// 返回值：
//   - time.Time: 解析后的时间对象，失败返回零值
//
// 支持的格式示例：
//   - "2024-03-25 14:30:00"
//   - "2024/03/25 14:30:00"
//   - "2024-03-25"
//   - "2024/03/25"
func ParseDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}
	}

	for _, format := range DateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

// TimeEquals 判断两个时间是否精确相等（精确到秒）
//
// 此函数比较完整的时间戳（年月日时分秒）。
// 如果任一时间为零值，返回 false。
//
// 参数：
//   - date1: 第一个时间
//   - date2: 第二个时间
//
// 返回值：
//   - bool: 如果两个时间完全相同返回 true，否则返回 false
func TimeEquals(date1, date2 time.Time) bool {
	if date1.IsZero() || date2.IsZero() {
		return false
	}
	return date1.Format("2006-01-02 15:04:05") == date2.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期为标准格式
//
// 将时间对象格式化为 "2006-01-02" 格式的字符串。
// 如果时间为零值，返回空字符串。
//
// 参数：
//   - t: 时间对象
//
// 返回值：
//   - string: 格式化后的日期字符串，零值时间返回空字符串
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDateTime 格式化日期时间为标准格式
//
// 将时间对象格式化为 "2006-01-02 15:04:05" 格式的字符串。
// 如果时间为零值，返回空字符串。
//
// 参数：
//   - t: 时间对象
//
// 返回值：
//   - string: 格式化后的日期时间字符串，零值时间返回空字符串
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// ==================== 列数据读取工具 ====================

// GetColIndexByName 根据列名获取列索引
//
// 在 Excel 表格的列定义中查找指定列名所在的列索引。
// 列名位于第 MJS_FIXED_ROWS_NAME 行（字段名行）。
//
// 参数：
//   - cols: Excel 列数据（按列组织的二维数组）
//   - colName: 要查找的列名
//
// 返回值：
//   - int: 列索引，如果未找到返回 -1
func GetColIndexByName(cols [][]string, colName string) int {
	for i, col := range cols {
		if len(col) > MJS_FIXED_ROWS_NAME && col[MJS_FIXED_ROWS_NAME] == colName {
			return i
		}
	}
	return -1
}

// GetColValue 获取指定列指定行的值
//
// 从二维数组中获取指定列和行的值，并去除首尾空格。
// 如果索引越界，返回空字符串。
//
// 参数：
//   - cols: Excel 列数据（按列组织的二维数组）
//   - colIndex: 列索引
//   - rowIndex: 行索引
//
// 返回值：
//   - string: 单元格的值（去除首尾空格），越界返回空字符串
func GetColValue(cols [][]string, colIndex, rowIndex int) string {
	if colIndex < 0 || colIndex >= len(cols) {
		return ""
	}
	col := cols[colIndex]
	if rowIndex < 0 || rowIndex >= len(col) {
		return ""
	}
	return strings.TrimSpace(col[rowIndex])
}

// ==================== 通知格式化工具 ====================

// FormatChangeMessage 格式化变更消息
//
// 生成统一的字段变更通知消息格式。
//
// 格式："{rowName}行，id为{rowId}，字段{colName}从{oldValue}改成了{newValue}"
//
// 参数：
//   - rowName: 行名称（如"武将"、"技能"等）
//   - rowId: 行 ID
//   - colName: 列名（字段名）
//   - oldValue: 旧值
//   - newValue: 新值
//
// 返回值：
//   - string: 格式化后的变更消息
//
// 示例：
//
//	FormatChangeMessage("武将", "10001", "Name", "关羽", "张飞")
//	返回："武将行，id为10001，字段Name从关羽改成了张飞"
func FormatChangeMessage(rowName, rowId, colName, oldValue, newValue string) string {
	return fmt.Sprintf("%s行，id为%s，字段%s从%s改成了%s", rowName, rowId, colName, oldValue, newValue)
}

// FormatAddRowMessage 格式化新增行消息
//
// 生成统一的新增行通知消息格式。
//
// 格式："ID={rowId}, 名称={rowName}"
//
// 参数：
//   - rowId: 行 ID
//   - rowName: 行名称
//
// 返回值：
//   - string: 格式化后的新增行消息
//
// 示例：
//
//	FormatAddRowMessage("10001", "关羽")
//	返回："ID=10001, 名称=关羽"
func FormatAddRowMessage(rowId, rowName string) string {
	return fmt.Sprintf("ID=%s, 名称=%s", rowId, rowName)
}

// ==================== 布尔值解析工具 ====================

// ParseBool 解析布尔值
//
// 将字符串解析为布尔值，支持多种常见格式：
//   - 真值：true、1、yes（不区分大小写）
//   - 其他值均为 false
//
// 参数：
//   - s: 要解析的字符串
//
// 返回值：
//   - bool: 解析后的布尔值
//
// 示例：
//
//	ParseBool("true")   -> true
//	ParseBool("TRUE")   -> true
//	ParseBool("1")      -> true
//	ParseBool("yes")    -> true
//	ParseBool("false")  -> false
//	ParseBool("0")      -> false
//	ParseBool("no")     -> false
func ParseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}
