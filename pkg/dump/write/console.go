package write

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// formatValue 根据值的长度格式化输出
// 如果长度小于20字符，直接转换为字符串
// 否则输出字节长度，并支持单位自动转换
func formatValue(value []byte) string {
	length := len(value)
	if length < 20 {
		return string(value)
	}
	return formatByteSize(length)
}

// formatByteSize 将字节大小格式化为人类可读的形式
// 支持自动转换单位（B, KB, MB, GB, TB）
func formatByteSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := int64(bytes) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(div), units[exp])
}

// getWidth 获取字符串的显示宽度
func getWidth(s string) int {
	return utf8.RuneCountInString(s)
}

// Console 将数据输出到控制台，使用表格美化展示
// records: 要输出的数据记录
// pks: 主键字段名，用于确定记录的唯一标识
func Console(records []dump.Record, pks ...string) string {
	if len(records) == 0 {
		fmt.Println("没有数据记录")
		return ""
	}

	// 处理每条记录
	for i, record := range records {
		// 确定记录标识（如果有主键则使用主键，否则使用索引）
		recordID := fmt.Sprintf("%d", i+1)
		if len(pks) > 0 {
			for _, pk := range pks {
				if pkValue, exists := record[pk]; exists {
					recordID = formatValue(pkValue)
					break
				}
			}
		}

		// 找出所有字段名和值，计算最大字段名和值长度
		maxKeyWidth := 0
		maxValueWidth := 0
		var fields []struct {
			key   string
			value string
		}

		// 首先收集所有字段并计算最大宽度
		for key, value := range record {
			formattedValue := formatValue(value)
			fields = append(fields, struct {
				key   string
				value string
			}{key, formattedValue})

			// 更新最大字段名宽度（确保完全显示字段名）
			keyWidth := getWidth(key)
			if keyWidth > maxKeyWidth {
				maxKeyWidth = keyWidth
			}

			// 更新最大值宽度
			valueWidth := getWidth(formattedValue)
			if valueWidth > maxValueWidth {
				maxValueWidth = valueWidth
			}
		}

		// 确保有最小宽度以提高可读性
		if maxKeyWidth < 6 {
			maxKeyWidth = 6
		}
		if maxValueWidth < 6 {
			maxValueWidth = 6
		}

		// 计算总宽度（ID行）
		totalWidth := maxKeyWidth + maxValueWidth + 3 // 3是分隔符和空格的总宽度（1个分隔符+2个空格）

		// 添加记录标识行
		fmt.Println("┌" + strings.Repeat("─", totalWidth) + "┐")
		// 确保序号居中对齐
		padding := (totalWidth - getWidth(recordID)) / 2
		fmt.Printf("│%s%s%s│\n", strings.Repeat(" ", padding), recordID, strings.Repeat(" ", totalWidth-padding-getWidth(recordID)))
		// 确保边框连接正确对齐
		fmt.Println("├" + strings.Repeat("─", maxKeyWidth) + "┬" + strings.Repeat("─", maxValueWidth) + "┤")

		// 添加字段行
		for _, field := range fields {
			// 格式化字段名和值（左对齐，确保完全显示）
			fmt.Printf("│ %-*s │ %-*s │\n", maxKeyWidth-1, field.key, maxValueWidth-1, field.value)
		}

		// 添加分隔行或结束行
		if i < len(records)-1 {
			// 确保记录间分隔线正确对齐
			fmt.Println("├" + strings.Repeat("─", maxKeyWidth) + "┴" + strings.Repeat("─", maxValueWidth) + "┤")
		} else {
			// 确保最后一行分隔线正确闭合和对齐
			fmt.Println("└" + strings.Repeat("─", maxKeyWidth) + "┴" + strings.Repeat("─", maxValueWidth) + "┘")
		}
	}
	return ""
}
