// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// StringCheckRule 字符串类型检查
//
// 功能：检测策划将字符串类型错误地配成数值型（如 "abc" 所在列出现 123）
//
// 前端对应：EColRule.STRING — 检测字符串类型（参数组件：CellTypeCheckParams.vue）
//
// 检测原理：
//
//	通过 excelize.GetCellType() 获取实际单元格类型
//	- CellTypeSharedString(7)/CellTypeInlineString(5): 文本格式（正确格式）
//	- CellTypeUnset(0)/CellTypeNumber(6): 数值格式（会被标记为错误）
//
// 校验成功示例：
//   - 单元格值为 "1010804" 且存储格式为文本（CellTypeSharedString）
//   - 单元格值为 "abc" 且存储格式为文本
//
// 校验失败示例：
//   - 单元格值为 1010804 且存储格式为数值（CellTypeNumber）
//   - 单元格显示为 "1010804" 但实际存储为数值格式（CellTypeUnset）
//
// 参数说明：
//   - allowEmpty: 是否允许空值（空单元格跳过检查）
//   - allowCommit: 是否允许提交（开发阶段可设为 true）
//   - breakLine: 连续空行数量阈值，用于自动检测数据结束位置
//
// 使用场景：
//   - 检查声明为 string 类型的字段是否保持文本格式
//   - 检查 ID/条件字段是否被错误改为数值格式
type StringCheckRule struct{}

// Check 执行字符串类型检查
// 使用传统的 Checker 接口，直接传入完整的列数据上下文
//
// 参数:
//   - sheetName: 表名（如"物品|Item"）
//   - cols: 所有列数据（列主序二维数组）
//   - colIdx: 当前检查的列索引
//   - startRowIdx: 数据起始行索引（通常为4，第5行开始）
//   - params: 规则参数（allowEmpty、allowCommit等）
//   - sheetMap: 其他表的数据（用于跨表检查，本规则用于获取原始 excelize.File）
//
// 返回:
//   - 错误单元格列表：每个错误包含行索引和原因描述
//
// 执行流程：
//  1. 解析规则参数（breakLine、allowEmpty、allowCommit）
//  2. 自动检测数据结束位置（基于连续空行）
//  3. 遍历当前列的所有数据行
//  4. 对每个单元格执行类型检查
//  5. 记录所有非文本格式的单元格作为错误
func (c *StringCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 获取 excelize.File 对象
	var xlsxFile *excelize.File
	var ok bool
	if sheetMap != nil {
		xlsxFile, ok = sheetMap[sheetName]
	}
	if !ok || xlsxFile == nil {
		// 如果无法获取 excelize.File，返回空结果（降级处理）
		return res
	}

	for i, cellValue := range myColData {
		actualRowIdx := startRowIdx + i

		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 构造单元格坐标（如 "A5"）
		cell, err := excelize.CoordinatesToCellName(colIdx+1, actualRowIdx+1)
		if err != nil {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("无法构造单元格坐标: %v", err),
			})
			continue
		}

		// 获取单元格类型
		cellType, err := xlsxFile.GetCellType(sheetName, cell)
		if err != nil {
			// 获取失败时跳过检查（可能是空单元格或格式问题）
			continue
		}

		// 检查是否为非文本类型
		// 文本格式通常是 CellTypeSharedString(7) 或 CellTypeInlineString(5)
		// 数值型通常是 CellTypeUnset(0) 或 CellTypeNumber(6)
		if cellType != excelize.CellTypeSharedString && cellType != excelize.CellTypeInlineString {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("单元格类型为数值格式，应为字符串: 值=%s, 坐标=%s", cellValue, cell),
			})
		}
	}
	return res
}
