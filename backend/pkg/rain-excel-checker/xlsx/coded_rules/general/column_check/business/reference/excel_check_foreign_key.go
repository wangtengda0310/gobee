// Package reference 提供引用关系相关的校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package reference

import (
	"fmt"
	"path"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ForeignKeyCheckRule 外键检查（某个字段的值在另外一个表中）
type ForeignKeyCheckRule struct{}

// Check 执行外键检查（验证某个字段的值在另外一个表中是否存在）
//
// 执行流程：
//  1. 解析 BREAK_LINE 参数，设置空行连续阈值（默认3行）
//  2. 自动检测数据结束位置并提取列数据
//  3. 解析目标表参数（targetSheet 和 targetCol）
//  4. 从 sheetMap 中使用后缀匹配查找目标表文件对象
//  5. 解析 ALLOW_EMPTY 和 ALLOW_COMMIT 参数
//  6. 获取目标列的所有有效值（构建查找map）
//  7. 遍历列数据，对每个单元格：
//     a. 处理空值和注释（根据配置决定是否跳过）
//     b. 检查值是否在目标列的值集合中
//  8. 返回所有错误信息
func (c *ForeignKeyCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	// 步骤1: 解析 BREAK_LINE 参数
	breakLine := helpers.ParseBreakLine(params)

	// 步骤2: 自动检测数据结束位置并提取列数据
	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]

	// 初始化检查结果切片
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 步骤3: 解析目标表参数
	targetSheetSuffix := params["targetSheet"]

	// 步骤4: 从 sheetMap 中使用后缀匹配查找目标表文件对象
	// sheetMap 的 key 是完整的 sheet 名（如"掉落分组|DropGroup"），需要用 FindSheetBySuffix 匹配
	var targetFile *excelize.File
	var actualSheetName string
	if f, sn, ok := helpers.FindSheetBySuffix(sheetMap, targetSheetSuffix); ok {
		targetFile = f
		actualSheetName = sn
	}

	// 步骤5: 解析 ALLOW_EMPTY 和 ALLOW_COMMIT 参数
	allowEmpty := helpers.ParseAllowEmpty(params)

	allowCommit := helpers.ParseAllowCommit(params)

	// 步骤6: 获取目标列的所有值（构建查找map）
	// 使用 FindSheetBySuffix 查找目标表（支持"中文|英文"格式的 sheet 名），
	// 然后从目标表的 Id 列读取所有值填充 map
	targetValues := make(map[string]bool)
	if targetFile != nil && actualSheetName != "" {
		targetColName := params["targetCol"]
		if targetColName == "" {
			targetColName = "Id"
		}

		targetCols, err := targetFile.GetCols(actualSheetName)
		if err == nil {
			var isEnumTable = strings.Contains(path.Base(targetFile.Path), "_enum.xlsx")

			for _, col := range targetCols {
				if len(col) == 0 {
					continue
				}

				var isTargetCol bool
				var refColData []string
				if isEnumTable {
					if len(col) > excelio.MJS_FIXED_ENUM_ROWS_CHS && col[excelio.MJS_FIXED_ENUM_ROWS_CHS] == targetColName {
						isTargetCol = true
						refColData = col[excelio.MJS_FIXED_ENUM_ROWS_CHS:]
					}
				} else {
					if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == targetColName {
						isTargetCol = true
						refColData = col[excelio.MJS_FIXED_ROWS_NUM:]
					}
				}

				if isTargetCol {
					for _, value := range refColData {
						if value != "" {
							targetValues[value] = true
						}
					}
					break
				}
			}
		}
	}

	// 步骤7: 遍历列数据
	for i, s := range myColData {
		// 步骤7a: 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 步骤7b: 检查值是否在目标列的值集合中
		if !targetValues[s] {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("值%s在目标表%s中不存在", s, targetSheetSuffix),
			})
		}
	}

	// 步骤8: 返回所有错误信息
	return res
}
