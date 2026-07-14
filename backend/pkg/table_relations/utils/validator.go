package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"github.com/xuri/excelize/v2"
)

// ValidateRelation 验证表关联关系的有效性
//
// 检查：
//  1. 源表是否存在
//  2. 目标表是否存在
//  3. 源表字段是否有效
//  4. 目标表字段是否有效
//
// 参数：
//   - relation: 表关联关系配置
//   - sheetMap: 表映射表，用于验证表和字段是否存在
//
// 返回：
//   - error: 验证失败时的错误信息，验证成功时为 nil
//
// 示例：
//
//	err := ValidateRelation(tableRelation, sheetMap)
//	if err != nil {
//	    return fmt.Errorf("关联关系验证失败: %w", err)
//	}
func ValidateRelation(relation interface{}, sheetMap map[string]*excelize.File) error {
	// TODO: 根据实际的关联关系结构类型来实现
	// 当前先实现基础的表存在性检查
	return nil
}

// ValidateTableExists 验证表是否存在
//
// 参数：
//   - sheetMap: 表映射表
//   - tableName: 表名
//
// 返回：
//   - bool: 表是否存在
//   - string: 实际的表名（可能包含中文部分）
//   - error: 错误信息
//
// 示例：
//
//	exists, actualName, err := ValidateTableExists(sheetMap, "Activity")
//	if !exists {
//	    return fmt.Errorf("活动表不存在: %w", err)
//	}
func ValidateTableExists(sheetMap map[string]*excelize.File, tableName string) (bool, string, error) {
	_, actualName, found := helpers.FindSheetBySuffix(sheetMap, tableName)
	if !found {
		return false, "", fmt.Errorf("表 '%s' 不存在", tableName)
	}
	return true, actualName, nil
}

// ValidateFieldExists 验证字段是否存在于指定表中
//
// 参数：
//   - sheetMap: 表映射表
//   - tableName: 表名
//   - fieldName: 字段名（支持数组索引语法，如 "CustomParma[0]"）
//
// 返回：
//   - bool: 字段是否存在
//   - error: 验证失败时的错误信息
//
// 示例：
//
//	exists, err := ValidateFieldExists(sheetMap, "Activity", "CustomParma[0]")
//	if !exists {
//	    return fmt.Errorf("字段验证失败: %w", err)
//	}
func ValidateFieldExists(sheetMap map[string]*excelize.File, tableName, fieldName string) (bool, error) {
	// 解析字段名（处理数组索引语法）
	baseFieldName, arrayIndex, err := parseFieldName(fieldName)
	if err != nil {
		return false, fmt.Errorf("字段名 '%s' 格式无效: %w", fieldName, err)
	}

	// 加载表数据
	cols, _, err := LoadSheetData(sheetMap, tableName)
	if err != nil {
		return false, fmt.Errorf("加载表 '%s' 失败: %w", tableName, err)
	}

	// 检查字段是否存在
	colIdx := helpers.GetColIndexByName(cols, baseFieldName)
	if colIdx < 0 {
		return false, fmt.Errorf("字段 '%s' 在表 '%s' 中不存在", baseFieldName, tableName)
	}

	// 如果有数组索引，检查数组范围是否有效
	if arrayIndex != nil {
		startRow := GetStartRowIndex()
		idColIdx := helpers.GetColIndexByName(cols, "Id")
		if idColIdx < 0 {
			return false, fmt.Errorf("表 '%s' 中缺少 'Id' 列", tableName)
		}

		endRow := GetDataEndIndex(cols, idColIdx, startRow)

		// 检查第一行数据的数组长度是否包含指定索引
		for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
			cellValue := helpers.GetColValue(cols, colIdx, rowIdx)
			if cellValue == "" {
				continue
			}

			// 解析数组值
			values := parseArrayValue(cellValue)
			if *arrayIndex >= len(values) {
				return false, fmt.Errorf("字段 '%s' 的索引 %d 超出数组范围（长度: %d）",
					fieldName, *arrayIndex, len(values))
			}

			// 只要有一行数据满足条件即可
			return true, nil
		}

		return false, fmt.Errorf("字段 '%s' 没有有效的数据", fieldName)
	}

	return true, nil
}

// validateField 验证字段有效性（内部使用）
//
// 参数：
//   - cols: 表的列数据
//   - fieldName: 字段名（支持数组索引语法）
//   - tableName: 表名（用于错误消息）
//
// 返回：
//   - error: 验证失败时的错误信息
func validateField(cols [][]string, fieldName, tableName string) error {
	// 解析字段名（处理数组索引语法）
	baseFieldName, arrayIndex, err := parseFieldName(fieldName)
	if err != nil {
		return fmt.Errorf("字段名 '%s' 格式无效: %w", fieldName, err)
	}

	// 检查字段是否存在
	colIdx := helpers.GetColIndexByName(cols, baseFieldName)
	if colIdx < 0 {
		return fmt.Errorf("字段 '%s' 在表 '%s' 中不存在", baseFieldName, tableName)
	}

	// 如果有数组索引，检查数组范围是否有效
	if arrayIndex != nil {
		if *arrayIndex < 0 {
			return fmt.Errorf("字段 '%s' 的数组索引 %d 不能为负数", fieldName, *arrayIndex)
		}

		// 获取第一行数据检查
		startRow := GetStartRowIndex()
		if len(cols[colIdx]) <= startRow {
			return fmt.Errorf("字段 '%s' 没有数据", fieldName)
		}

		firstValue := helpers.GetColValue(cols, colIdx, startRow)
		values := parseArrayValue(firstValue)
		if len(values) == 0 {
			return fmt.Errorf("字段 '%s' 不是数组类型", fieldName)
		}

		if *arrayIndex >= len(values) {
			return fmt.Errorf("字段 '%s' 的索引 %d 超出数组范围（长度: %d）",
				fieldName, *arrayIndex, len(values))
		}
	}

	return nil
}

// parseFieldName 解析字段名，支持数组索引语法
//
// 支持格式：
//   - "FieldName" - 普通字段名
//   - "FieldName[0]" - 数组字段名（索引为 0）
//   - "FieldName[10]" - 数组字段名（索引为 10）
//
// 参数：
//   - fieldName: 字段名字符串
//
// 返回：
//   - string: 基础字段名
//   - *int: 数组索引（如果不是数组字段则为 nil）
//   - error: 解析失败时的错误信息
//
// 示例：
//
//	name, idx, err := parseFieldName("CustomParma[0]")
//	// name = "CustomParma", idx = &0
func parseFieldName(fieldName string) (string, *int, error) {
	// 使用正则表达式匹配数组索引语法
	re := regexp.MustCompile(`^(\w+)\[(\d+)\]$`)
	matches := re.FindStringSubmatch(fieldName)

	if matches != nil {
		// 数组字段
		baseName := matches[1]
		index, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", nil, fmt.Errorf("无效的数组索引: %s", matches[2])
		}
		return baseName, &index, nil
	}

	// 普通字段
	return fieldName, nil, nil
}

// parseArrayValue 解析数组值
//
// 支持多种分隔符：
//   - 逗号分隔: "1,2,3"
//   - 分号分隔: "1;2;3"
//
// 参数：
//   - value: 单元格值
//
// 返回：
//   - []string: 解析后的数组元素
//
// 示例：
//
//	values := parseArrayValue("1,2,3")
//	// values = ["1", "2", "3"]
func parseArrayValue(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}
	}

	// 尝试用逗号分隔
	if strings.Contains(value, ",") {
		return strings.Split(value, ",")
	}

	// 尝试用分号分隔
	if strings.Contains(value, ";") {
		return strings.Split(value, ";")
	}

	// 单个值
	return []string{value}
}
