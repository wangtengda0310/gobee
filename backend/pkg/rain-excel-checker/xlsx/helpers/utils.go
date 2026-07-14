// Package helpers 提供校验规则的通用辅助工具
// 本包包含列检查、参数解析、表查找等通用辅助函数
package helpers

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ParseBreakLine 从参数中解析 breakLine（连续空单元格阈值）
// 默认值：3，最小值：1
func ParseBreakLine(params map[string]string) int {
	breakLine := 3
	if break_, ok := params[string(json_rule.BREAK_LINE)]; ok {
		if maxSpace, err := strconv.Atoi(break_); err == nil && maxSpace >= 1 {
			breakLine = maxSpace
		}
	}
	return breakLine
}

// ParseAllowEmpty 从参数中解析 allowEmpty（是否允许空值）
// 默认值：false
func ParseAllowEmpty(params map[string]string) bool {
	allowEmpty := false
	if allow, ok := params[string(json_rule.ALLOW_EMPTY)]; ok {
		allowEmpty = strings.ToLower(allow) == "true"
	}
	return allowEmpty
}

// ParseAllowCommit 从参数中解析 allowCommit（是否允许注释）
// 默认值：false
func ParseAllowCommit(params map[string]string) bool {
	allowCommit := false
	if allow, ok := params[string(json_rule.ALLOW_COMMIT)]; ok {
		allowCommit = strings.ToLower(allow) == "true"
	}
	return allowCommit
}

// FindSheetBySuffix 根据英文名后缀查找 Sheet
// 支持两种格式：
// 1. 精确匹配：sheetMap["Hero"]
// 2. 后缀匹配：sheetMap["武将|Hero"] 通过 "Hero" 也能找到
// 返回：文件指针、Sheet名称、是否找到
func FindSheetBySuffix(sheetMap map[string]*excelize.File, sheetName string) (*excelize.File, string, bool) {
	// 1. 先尝试精确匹配
	if file, ok := sheetMap[sheetName]; ok {
		return file, sheetName, true
	}

	// 2. 尝试后缀匹配（格式为 "中文|英文"）
	for fullSheetName, file := range sheetMap {
		// 检查是否以 "|sheetName" 结尾
		if strings.HasSuffix(fullSheetName, "|"+sheetName) {
			return file, fullSheetName, true
		}
	}

	return nil, "", false
}

// MatchSheetBySuffix 检查 sheetName 是否匹配 requiredFilter 中的任何一个要求
// 支持精确匹配和后缀匹配（格式为 "中文|英文"）
// 用于过滤不需要补充到 sheetMap 的表
//
// 参数：
//   - sheetName: 待检查的表名（如 "武将|Hero" 或 "Hero"）
//   - requiredFilter: 需要的表名集合（如 {"Hero": true, "Item": true}）
//
// 返回：
//   - true: sheetName 匹配 requiredFilter 中的某个表名
//   - false: sheetName 不匹配任何需要的表名
//
// 示例：
//
//	MatchSheetBySuffix("武将|Hero", {"Hero": true}) → true
//	MatchSheetBySuffix("Hero", {"Hero": true}) → true
//	MatchSheetBySuffix("Item", {"Hero": true}) → false
func MatchSheetBySuffix(sheetName string, requiredFilter map[string]bool) bool {
	// 1. 先尝试精确匹配
	if requiredFilter[sheetName] {
		return true
	}

	// 2. 尝试后缀匹配（格式为 "中文|英文"）
	// 提取 | 后的英文名部分
	if strings.Contains(sheetName, "|") {
		parts := strings.Split(sheetName, "|")
		if len(parts) >= 2 {
			suffix := parts[len(parts)-1] // 取最后一部分（英文名）
			if requiredFilter[suffix] {
				return true
			}
		}
	}

	return false
}

// 辅助函数：使用多种格式尝试解析日期
func ParseDateWithFormats(dateStr string, formats []string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("空字符串")
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析日期: %s", dateStr)
}

// AutoDetectEndIndex 自动检测数据结束位置（连续N行空单元格视为结束）
// 返回的位置下标为第一个连续空值的下标
//
// 执行流程：
//  1. 获取指定列的数据
//  2. 检查起始位置是否超出列长度，超出则返回列长度
//  3. 从起始位置开始遍历：
//     a. 遇到空单元格则累加空计数
//     b. 空计数达到阈值（spaceNum）时，返回第一个空单元格的位置
//     c. 遇到非空单元格则重置空计数
//  4. 遍历结束未达到阈值，返回列长度
func AutoDetectEndIndex(cols [][]string, cIdx, c1idx, spaceNum int) int {
	// 🆕 边界检查：防止列索引越界导致 panic
	if cIdx < 0 || cIdx >= len(cols) {
		return c1idx
	}

	// 步骤1: 获取指定列的数据
	colData := cols[cIdx]

	// 步骤2: 检查起始位置是否超出列长度
	if len(colData) <= c1idx {
		return len(colData)
	}

	// 步骤3: 从起始位置开始检查
	emptyCount := 0
	for i := c1idx; i < len(colData); i++ {
		// 步骤3a: 遇到空单元格则累加空计数
		if strings.TrimSpace(colData[i]) == "" {
			emptyCount++
			// 步骤3b: 空计数达到阈值，返回第一个空单元格的位置
			if emptyCount == spaceNum {
				return i - spaceNum + 1
			}
		} else {
			// 步骤3c: 非空单元格，重置计数
			emptyCount = 0
		}
	}

	// 步骤4: 遍历结束未达到阈值，返回列长度
	return len(colData)
}

// GetColEndIndex 获取列的数据结束位置
// 优先使用 ID 列确定统一的结束位置（避免不同列截断位置不一致）
// 如果 params 中指定了 useIdColForEnd=true，则使用 ID 列确定结束位置
// 否则使用当前列的 AutoDetectEndIndex
//
// 参数：
//   - cols: 所有列数据
//   - colIdx: 当前列索引
//   - startRowIdx: 数据起始行索引
//   - breakLine: 连续空单元格阈值
//   - params: 规则参数（可能包含 useIdColForEnd 和 idColName）
//
// 返回：
//   - 数据结束位置（相对于 cols 的索引）
func GetColEndIndex(cols [][]string, colIdx, startRowIdx, breakLine int, params map[string]string) int {
	// 检查是否使用 ID 列确定结束位置
	useIdCol := false
	if v, ok := params[string(json_rule.USE_ID_COL_FOR_END)]; ok {
		useIdCol = v == "true"
	}

	if useIdCol {
		// 获取 ID 列名（默认为 "Id"）
		idColName := "Id"
		if v, ok := params[string(json_rule.ID_COL_NAME)]; ok && v != "" {
			idColName = v
		}

		// 查找 ID 列索引
		idColIdx := -1
		for i, col := range cols {
			if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == idColName {
				idColIdx = i
				break
			}
		}

		// 如果找到 ID 列，使用 ID 列确定结束位置
		if idColIdx >= 0 {
			return AutoDetectEndIndex(cols, idColIdx, startRowIdx, breakLine)
		}
	}

	// 否则使用当前列确定结束位置
	return AutoDetectEndIndex(cols, colIdx, startRowIdx, breakLine)
}

// GetDataEndIndex 从列数据中自动检测数据结束位置
// 通过 ID 列（字段名为 "Id"）检测连续3个空行，找不到 ID 列时回退到第一列
// 用于表级规则的辅助函数（如 buildValidItemIdSet、checkOpenedHeroes 等），
// 这些函数从其他表（如 Item、Hero）读取数据时也需要知道数据结束位置
// GetDataEndIndex 从列数据中自动检测数据结束位置
func GetDataEndIndex(cols [][]string, startRowIdx int) int {
	idColIdx := excelio.GetColIndexByName(cols, "Id")
	if idColIdx < 0 {
		if len(cols) > 0 {
			return AutoDetectEndIndex(cols, 0, startRowIdx, 3)
		}
		return startRowIdx
	}
	return AutoDetectEndIndex(cols, idColIdx, startRowIdx, 3)
}

// SolveEmptyAndCommit 处理空值和注释
// 返回true表示跳过该单元格的检查，返回false表示需要检查
//
// 执行流程：
// 1. 获取当前单元格和该行第一个单元格的值
// 2. 检查整行是否为空：
//   - 如果全空则返回true（跳过检查）
//
// 3. 检查是否是注释行（第一个单元格以"#"开头）：
//   - 如果允许注释则跳过，否则添加错误
//
// 4. 检查当前单元格是否是注释（以"#"开头）：
//   - 如果允许注释则跳过，否则添加错误
//
// 5. 检查当前单元格是否为空：
//   - 如果允许空值则跳过，否则添加错误
//
// 6. 以上条件都不满足则返回false（需要检查该单元格）
func SolveEmptyAndCommit(res *[]*json_rule.CellError, cols [][]string, startRowIdx, col, row int, allowEmpty, allowCommit bool) bool {
	// 步骤1: 获取当前单元格和该行第一个单元格的值（使用安全函数防止不同列长度不一致导致越界）
	currentCellValue := excelio.GetColValue(cols, col, startRowIdx+row)
	firstCellInRow := excelio.GetColValue(cols, 0, startRowIdx+row)

	// 步骤2: 检查整行是否为空
	if isRowEmpty(cols, startRowIdx, row) {
		return true // 全空行直接跳过
	}

	// 步骤3: 检查是否是注释行（第一列有#）
	if strings.HasPrefix(firstCellInRow, "#") {
		return handleCommentRow(res, row, startRowIdx, allowCommit)
	}

	// 步骤4: 检查当前单元格是否是注释
	if strings.HasPrefix(currentCellValue, "#") {
		return handleCommentCell(res, row, startRowIdx, allowCommit)
	}

	// 步骤5: 检查当前单元格是否为空
	if currentCellValue == "" {
		return handleEmptyCell(res, row, startRowIdx, allowEmpty)
	}

	// 步骤6: 需要检查这个单元格
	return false
}

// 检查整行是否为空
// 使用 excelio.GetColValue 安全访问，防止不同列长度不一致导致越界
func isRowEmpty(cols [][]string, startRowIdx, row int) bool {
	for colIdx := range cols {
		if excelio.GetColValue(cols, colIdx, startRowIdx+row) != "" {
			return false
		}
	}
	return true
}

// 处理注释行
func handleCommentRow(res *[]*json_rule.CellError, row, startRowIdx int, allowCommit bool) bool {
	if allowCommit {
		return true // 允许注释，跳过检查
	}
	// 不允许注释行
	*res = append(*res, &json_rule.CellError{
		Index:    row,
		ExcelRow: startRowIdx + row + 1,
		Reason:   "不允许注释行",
	})
	return true // 跳过后续检查
}

// 处理注释单元格
func handleCommentCell(res *[]*json_rule.CellError, row, startRowIdx int, allowCommit bool) bool {
	if allowCommit {
		return true // 允许注释，跳过检查
	}
	// 不允许注释
	*res = append(*res, &json_rule.CellError{
		Index:    row,
		ExcelRow: startRowIdx + row + 1,
		Reason:   "不允许注释",
	})
	return true // 跳过后续检查
}

// 处理空单元格
func handleEmptyCell(res *[]*json_rule.CellError, row, startRowIdx int, allowEmpty bool) bool {
	if allowEmpty {
		return true // 允许空值，跳过检查
	}
	// 不允许空值
	*res = append(*res, &json_rule.CellError{
		Index:    row,
		ExcelRow: startRowIdx + row + 1,
		Reason:   "单元格不能为空",
	})
	return true // 跳过后续检查
}

// getTargetSheetCol 获取指定表的参照列作为一个map
//
// 执行流程：
//  1. 检查是否指定了参照表名称，未指定则返回错误
//  2. 从sheetMap中获取参照表的文件对象，不存在则返回错误
//  3. 检查是否指定了参照列名称，未指定则返回错误
//  4. 获取参照表的所有列数据
//  5. 判断表类型（枚举表或普通表）以确定列名所在的行
//  6. 遍历所有列查找目标参照列：
//     a. 枚举表：在第MJS_FIXED_ENUM_ROWS_CHS行查找列名
//     b. 普通表：在第MJS_FIXED_ROWS_NAME行查找列名
//  7. 找到参照列后，提取所有非空值存入map
//  8. 如果未找到参照列或列中没有数据，返回错误
//  9. 返回参照列的值map（用于快速查找）
func getTargetSheetCol(sheetMap map[string]*excelize.File, refSheetName, refColName string) (map[string]bool, []*json_rule.CellError) {
	// 初始化结果map
	refValues := make(map[string]bool)

	// 步骤1: 检查是否指定了参照表名称
	if refSheetName != "" {
		// 步骤2: 从sheetMap中获取参照表的文件对象
		refFile, refFileExist := sheetMap[refSheetName]
		if !refFileExist {
			return nil, []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("参照表不存在: %s", refSheetName),
				},
			}
		}

		// 步骤3: 检查是否指定了参照列名称
		if refColName == "" {
			return nil, []*json_rule.CellError{
				{
					Index:  0,
					Reason: "必须指定参照列参数 'refCol'",
				},
			}
		}

		// 步骤4: 获取参照表的所有列数据
		refCols, err := refFile.GetCols(refSheetName)
		if err != nil {
			return nil, []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("无法读取参照表列数据: %v", err),
				},
			}
		}

		// 步骤5: 判断表类型（枚举表或普通表）
		var refColData []string
		var isEnumTable = strings.Contains(path.Base(refFile.Path), "_enum.xlsx")

		// 步骤6: 遍历所有列查找目标参照列
		for _, col := range refCols {
			if len(col) == 0 {
				continue
			}

			// 步骤6a: 枚举表处理
			var isTargetCol bool
			if isEnumTable {
				// 枚举表：第MJS_FIXED_ENUM_ROWS_CHS行是中文列名
				if len(col) > excelio.MJS_FIXED_ENUM_ROWS_CHS && col[excelio.MJS_FIXED_ENUM_ROWS_CHS] == refColName {
					isTargetCol = true
					refColData = col[excelio.MJS_FIXED_ENUM_ROWS_CHS:]
				}
			} else {
				// 步骤6b: 普通表处理
				// 普通表：第MJS_FIXED_ROWS_NAME行是列名
				if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == refColName {
					isTargetCol = true
					refColData = col[excelio.MJS_FIXED_ROWS_NUM:]
				}
			}

			// 步骤7: 找到参照列后，提取所有非空值存入map
			if isTargetCol {
				for _, value := range refColData {
					if value != "" {
						refValues[value] = true
					}
				}
				break
			}
		}

		// 步骤8: 检查是否找到参照列
		if len(refValues) == 0 {
			return nil, []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("在表%s中未找到列: %s", refSheetName, refColName),
				},
			}
		}
	} else {
		// 步骤1: 未指定参照表名称，返回错误
		return nil, []*json_rule.CellError{
			{
				Index:  0,
				Reason: "必须指定参照表参数 'refSheet'",
			},
		}
	}

	// 步骤9: 返回参照列的值map
	return refValues, nil
}
