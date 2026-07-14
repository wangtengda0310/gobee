// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性
package datatype

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// CellTypeCheckRuleAdaptor 单元格类型检查（使用 adaptor 模式）
//
// 功能：检测策划将数值型改为文本型（如 123 → "123"）
//
// 这是使用新接口（通过 adaptor 获取上下文）的实现版本
// 旧版本见 excel_check_cell_type.go（使用 6 参数接口）
//
// 使用场景：
//   - 检查 ID 列是否保持数值类型
//   - 检查数量字段是否被错误改为文本格式
//
// 检测原理：
//
//	通过 excelize.GetCellType() 获取实际单元格类型
//	- CellTypeSharedString (7): 文本格式（会被错误标记）
//	- CellTypeUnset (0): 数值格式（正确格式）
//	- CellTypeNumber (6): 数值格式（正确格式）
//
// 前端配置示例：
//
//	{
//	  "sheet": "物品|Item",
//	  "rules": {
//	    "Id": [{
//	      "type": "CELL_TYPE_CHECK_ADAPTOR",
//	      "displayName": "单元格类型检查（新版）",
//	      "uuid": "...",
//	      "params": {
//	        "allowEmpty": "true",
//	        "allowCommit": "true"
//	      }
//	    }]
//	  }
//	}
type CellTypeCheckRuleAdaptor struct{}

// Check 执行单元格类型检查
// 使用 adaptor 模式，通过 adaptor.GlobalAdaptor.GetContext() 获取完整上下文
//
// 参数：
//   - data: map[int]string，key 为相对行号（从0开始），value 为单元格值
//
// 返回：
//   - ColCheckResult: 检查结果
//
// 执行流程：
//  1. 从全局 adaptor 获取 Excel 上下文（表名、列索引、数据等）
//  2. 解析规则参数（允许空值、允许提交等）
//  3. 获取 Excel 文件对象以读取单元格类型
//  4. 遍历数据行，检查每个单元格的实际类型
//  5. 记录所有文本格式的单元格作为错误
func (c *CellTypeCheckRuleAdaptor) Check(data map[int]string) ColCheckResultAdaptor {
	// 1. 从全局 adaptor 获取 Excel 上下文
	// 包含：CheckParam（规则参数）、SheetName（表名）、ColIndex（列索引）等
	ctx, err := diff.GlobalAdaptor.GetContext("", "")
	if err != nil {
		return NewColCheckResultAdaptor(false, fmt.Sprintf("获取上下文失败: %v", err))
	}

	param := ctx.CheckParam
	sheetName := ctx.SheetName
	colIdx := ctx.ColIndex

	// 2. 解析规则参数
	// BREAK_LINE: 连续空行数阈值，用于自动检测数据结束位置
	breakLine := 3
	if break_, ok := param.Params[string(json_rule.BREAK_LINE)]; ok {
		if maxSpace, err := strconv.Atoi(break_); err == nil && maxSpace >= 1 {
			breakLine = maxSpace
		}
	}

	// ALLOW_COMMIT: 是否允许提交（通常用于开发阶段）
	// true: 即使有错误也允许通过，false: 有错误则阻止提交
	allowCommit := false
	if allow, ok := param.Params[string(json_rule.ALLOW_COMMIT)]; ok {
		allowCommit = strings.ToLower(allow) == "true"
	}

	// 3. 获取 Excel 文件对象以读取单元格类型
	// 注意：SheetMap 可能包含多个 Excel 文件（跨表检查时）
	var xlsxFile *excelize.File
	var ok bool
	if param.SheetMap != nil {
		xlsxFile, ok = param.SheetMap[sheetName]
	}
	if !ok || xlsxFile == nil {
		return NewColCheckResultAdaptor(false, "无法获取 Excel 文件对象")
	}

	// 4. 执行检查
	errorCount := 0
	firstError := ""

	// 获取数据结束位置（基于连续空行自动检测）
	// breakLine 参数控制连续多少个空行后认为数据结束
	endIdx := helpers.GetColEndIndex(param.Cols, colIdx, param.StartRowIdx, breakLine, param.Params)

	// 遍历数据行进行检查
	for rowIdx := param.StartRowIdx; rowIdx < endIdx; rowIdx++ {
		cellValue := param.Cols[colIdx][rowIdx]
		dataRowIdx := rowIdx - param.StartRowIdx

		// 处理空值和注释（自动跳过空行或标记为可提交错误）
		errCells := make([]*json_rule.CellError, 0)
		if helpers.SolveEmptyAndCommit(&errCells, param.Cols, param.StartRowIdx, colIdx, dataRowIdx, true, allowCommit) {
			continue
		}

		// 构造单元格坐标（如 "A5"，便于在 Excel 中定位）
		cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
		if err != nil {
			if firstError == "" {
				firstError = fmt.Sprintf("无法构造单元格坐标: %v", err)
			}
			errorCount++
			continue
		}

		// 获取单元格的实际类型
		// 这是检查的核心：通过 excelize 直接读取单元格存储格式
		cellType, err := xlsxFile.GetCellType(sheetName, cell)
		if err != nil {
			// 获取失败时跳过检查（可能是空单元格或格式问题）
			continue
		}

		// 检查是否为文本类型（CellTypeSharedString = 7）
		// 如果单元格被格式化为文本（即使是数字），也会被存储为 SharedString
		// 数值型通常是 CellTypeUnset(0) 或 CellTypeNumber(6)
		if cellType == excelize.CellTypeSharedString {
			if firstError == "" {
				firstError = fmt.Sprintf("单元格类型为文本格式(SharedString)，应为数值型: 值=%s, 坐标=%s", cellValue, cell)
			}
			errorCount++
		}
	}

	// 5. 返回检查结果
	// errorCount > 0 时返回错误信息和数量
	// errorCount = 0 时返回成功消息
	if errorCount == 0 {
		return NewColCheckResultAdaptor(true, "检查通过")
	}

	return NewColCheckResultAdaptor(false, fmt.Sprintf("发现 %d 个单元格类型错误: %s", errorCount, firstError))
}

// NewColCheckResultAdaptor 创建列检查结果（adaptor 版本）
// 封装检查结果的布尔状态和错误原因
func NewColCheckResultAdaptor(ok bool, reason string) ColCheckResultAdaptor {
	return &colCheckResultImplAdaptor{
		ok:     ok,
		reason: reason,
	}
}

// ColCheckResultAdaptor 列检查结果接口（adaptor 版本）
// 定义了列级检查结果的统一接口
// IsOk(): 返回检查是否通过（true=通过，false=失败）
// GetReason(): 返回错误原因（通过时通常返回"检查通过"）
type ColCheckResultAdaptor interface {
	IsOk() bool
	GetReason() string
}

// colCheckResultImplAdaptor ColCheckResult 的实现（adaptor 版本）
// 结构体实现 ColCheckResultAdaptor 接口
// ok: 检查是否通过
// reason: 错误原因或成功消息
type colCheckResultImplAdaptor struct {
	ok     bool
	reason string
}

// IsOk 返回检查是否通过
func (r *colCheckResultImplAdaptor) IsOk() bool {
	return r.ok
}

// GetReason 返回错误原因或成功消息
func (r *colCheckResultImplAdaptor) GetReason() string {
	return r.reason
}
