// Package engine 提供 Excel 检查的核心功能
// 本包包含读取 Excel 文件、执行列级和表级检查的核心函数
package engine

import (
	"fmt"
	"log"
	"path"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ReadAndCheckXlsx 读取 Excel 文件并执行列级检查
// 参数:
//   - excelPath: Excel 文件路径
//   - rule: 表规则配置
//   - sheetMap: 其他表的数据映射（用于跨表检查）
//
// 返回:
//   - 列检查结果列表
//   - 错误信息
func ReadAndCheckXlsx(excelPath string, rule *json_rule.SheetRule, sheetMap map[string]*excelize.File) (res []*json_rule.ColCheckResult, err error) {
	// 如果 SheetRule 里没有 ColRule 则返回
	if len(rule.Rules) == 0 {
		return
	}

	// 检查是否存在非空规则
	hasNonZero := false
	for _, colRule := range rule.Rules {
		if len(colRule.PropRules) != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		return
	}

	// 如果有规则就打开这张表开始检查
	xlsx, err := excelio.ReadXlsx(excelPath)
	if err != nil {
		return
	}
	defer xlsx.Close()

	return CheckSheetCols(xlsx, rule, sheetMap)
}

// CheckTableRules 检查表级规则
// 参数:
//   - xlsx: Excel 文件对象
//   - rule: 表规则配置
//   - sheetMap: 其他表的数据映射（用于跨表检查）
//
// 返回:
//   - 表检查结果列表
//   - 错误信息
func CheckTableRules(xlsx *excelize.File, rule *json_rule.SheetRule, sheetMap map[string]*excelize.File) (res []*json_rule.TableCheckResult, err error) {
	// 如果没有表级规则则直接返回
	if len(rule.TableRules) == 0 {
		return res, nil
	}

	fmt.Printf("-- CheckTableRules rule.Sheet %s \n", rule.Sheet)
	// 获取所有列数据
	cols, err := xlsx.GetCols(rule.Sheet)
	if err != nil {
		return nil, err
	}

	res = make([]*json_rule.TableCheckResult, 0)

	// 使用 ID 列确定数据结束位置（连续3个空行视为注释区，注释区数据不参与检查）
	endIndex := helpers.GetColEndIndex(cols, 0, excelio.MJS_FIXED_ROWS_NUM,
		3, nil)

	for _, tableRule := range rule.TableRules {
		fmt.Printf(">>> CheckTableRules tableRule.DisplayName %s \n", tableRule.DisplayName)
		// 跳过未启用的规则
		if !tableRule.Enabled {
			continue
		}

		// 获取对应的检查器
		checker := TableManager.GetChecker(tableRule.Type)
		if checker == nil {
			log.Printf("未找到表级检查器: %s", tableRule.Type)
			continue
		}

		// 检查规则是否适用于当前表
		meta := checker.Meta()
		if !isRuleApplicableToSheet(rule.Sheet, meta.TargetSheets) {
			continue
		}

		// 执行检查
		result := checker.Check(json_rule.CheckParam{
			SheetName:   rule.Sheet,
			Cols:        cols,
			StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
			EndIndex:    endIndex,
			Params:      tableRule.Params,
			SheetMap:    sheetMap,
		})
		if result != nil {
			result.SheetName = &rule.Sheet
			result.RuleType = tableRule.Type
			result.DisplayName = tableRule.DisplayName
			res = append(res, result)
		}
		fmt.Printf("<<< CheckTableRules tableRule.DisplayName %s \n", tableRule.DisplayName)
	}

	return res, nil
}

// isRuleApplicableToSheet 检查规则是否适用于指定的表
// 支持后缀匹配：targetSheet "ArenaSeason" 可以匹配 "竞技场赛季|ArenaSeason"
// 参数:
//   - sheetName: 表名（如 "竞技场赛季|ArenaSeason"）
//   - targetSheets: 目标表列表（如 ["ArenaSeason"]）
//
// 返回:
//   - 是否适用于该表
func isRuleApplicableToSheet(sheetName string, targetSheets []string) bool {
	if len(targetSheets) == 0 {
		// 如果没有指定目标表，则适用于所有表
		return true
	}

	for _, target := range targetSheets {
		// 精确匹配
		if sheetName == target {
			return true
		}
		// 后缀匹配：支持 "中文|英文" 格式
		if strings.HasSuffix(sheetName, "|"+target) {
			return true
		}
	}

	return false
}

// CheckSingleColumn 检查单个列
// 只执行指定列的列级规则检查，用于前端"执行此字段检查"功能
// 参数:
//   - xlsx: Excel 文件对象
//   - sheetName: 表名
//   - colRule: 列规则配置
//   - sheetMap: 其他表的数据映射（用于跨表检查）
//
// 返回:
//   - 列检查结果（nil 表示无规则或检查通过无错误）
//   - 错误信息
func CheckSingleColumn(xlsx *excelize.File, sheetName string, colRule *json_rule.SheetColRule, sheetMap map[string]*excelize.File) (*json_rule.ColCheckResult, error) {
	// 获取所有列数据
	cols, err := xlsx.GetCols(sheetName)
	if err != nil {
		return nil, err
	}

	// 截取合法列（与 CheckSheetCols 相同逻辑）
	emptyTypeCount := 0
	maxColIndex := len(cols)
	for i, colData := range cols {
		if len(colData) < excelio.MJS_FIXED_ROWS_NUM || strings.TrimSpace(colData[excelio.MJS_FIXED_ROWS_TYPE]) == "" {
			emptyTypeCount++
			if emptyTypeCount == 2 {
				maxColIndex = i - 1
				break
			}
		} else {
			emptyTypeCount = 0
		}
	}
	cols = cols[:maxColIndex]

	excelName := path.Base(xlsx.Path)

	// 初始化列检查结果
	colRes := &json_rule.ColCheckResult{
		TableName: &excelName,
		SheetName: &sheetName,
		ColName:   &colRule.PropName,
		ErrCells:  make([]*json_rule.CellError, 0),
	}

	// 找到对应列（根据属性名）
	index := -1
	for i, col := range cols {
		if len(col) < excelio.MJS_FIXED_ROWS_NUM {
			continue
		}
		if col[excelio.MJS_FIXED_ROWS_NAME] == colRule.PropName {
			index = i
			break
		}
	}

	if index == -1 {
		colRes.Ok = false
		colRes.Reason = fmt.Sprintf("未找到%s对应的列", colRule.PropName)
		return colRes, nil
	}
	colRes.ColIndex = &index

	// 验证列类型
	if colRule.PropType != cols[index][excelio.MJS_FIXED_ROWS_TYPE] {
		colRes.Ok = false
		colRes.Reason = fmt.Sprintf("列类型不正确, 实际类型:%s, 期望类型:%s", cols[index][excelio.MJS_FIXED_ROWS_TYPE], colRule.PropType)
		return colRes, nil
	}

	// 执行所有列规则检查
	for _, cRule := range colRule.PropRules {
		checker := Manager.GetChecker(cRule.Type)
		if checker == nil {
			log.Printf("没有对应的检查器: %s", cRule.Type)
			continue
		}
		// 创建参数副本
		checkParams := make(map[string]string)
		for k, v := range cRule.Params {
			checkParams[k] = v
		}
		// 默认启用 ID 列统一检测
		if _, exists := checkParams[string(json_rule.USE_ID_COL_FOR_END)]; !exists {
			checkParams[string(json_rule.USE_ID_COL_FOR_END)] = "true"
		}

		colRes.ErrCells = append(colRes.ErrCells, checker.Check(sheetName, cols, index, excelio.MJS_FIXED_ROWS_NUM, checkParams, sheetMap)...)
	}

	// 根据错误单元格数量设置检查结果状态
	if len(colRes.ErrCells) > 0 {
		colRes.Ok = false
		colRes.Reason = "单元格内存在错误"
	} else {
		colRes.Ok = true
		colRes.Reason = ""
	}

	return colRes, nil
}

// CheckSheetCols 检查表列
// 对表中的每一列执行配置的检查规则
// 参数:
//   - xlsx: Excel 文件对象
//   - rule: 表规则配置
//   - sheetMap: 其他表的数据映射（用于跨表检查）
//
// 返回:
//   - 列检查结果列表
//   - 错误信息
func CheckSheetCols(xlsx *excelize.File, rule *json_rule.SheetRule, sheetMap map[string]*excelize.File) (res []*json_rule.ColCheckResult, err error) {
	fmt.Printf("-- CheckSheetCols %s \n", rule.Sheet)
	// 获取所有列数据
	cols, err := xlsx.GetCols(rule.Sheet)
	if err != nil {
		return
	}

	// 截取合法列，即类型存在的列
	// 通过检测连续空类型列来确定有效列范围
	emptyTypeCount := 0
	maxColIndex := len(cols)
	for i, colData := range cols {
		// 检查列数据是否完整（至少有固定行数）且类型非空
		if len(colData) < excelio.MJS_FIXED_ROWS_NUM || strings.TrimSpace(colData[excelio.MJS_FIXED_ROWS_TYPE]) == "" {
			emptyTypeCount++

			if emptyTypeCount == 2 {
				// 找到连续2个空类型列，返回结束位置
				maxColIndex = i - 1 // 返回第一个空列的位置
				break
			}
		} else {
			emptyTypeCount = 0
		}
	}
	cols = cols[:maxColIndex]

	excelName := path.Base(xlsx.Path)
	res = make([]*json_rule.ColCheckResult, 0, len(cols))

	// 遍历每个属性规则进行检查
	for attrName, colRule := range rule.Rules {
		if len(colRule.PropRules) > 0 {
			fmt.Printf("DEBUG: Column %s has %d rules\n", attrName, len(colRule.PropRules))
		}
		fmt.Printf("-- cRule.Type %s \n", colRule.PropName)
		// 初始化列检查结果
		colRes := &json_rule.ColCheckResult{
			TableName: &excelName,
			SheetName: &rule.Sheet,
			ColName:   &attrName,
			ErrCells:  make([]*json_rule.CellError, 0),
		}

		// 找到对应列（根据属性名）
		index := -1
		for i, col := range cols {
			// 检查列是否有完整的表头
			if len(col) < excelio.MJS_FIXED_ROWS_NUM {
				colRes.Ok = false
				colRes.Reason = fmt.Sprintf("表头行数不足%d行", excelio.MJS_FIXED_ROWS_NUM)
				continue
			}
			// 通过属性名匹配列
			if col[excelio.MJS_FIXED_ROWS_NAME] == colRule.PropName {
				index = i
				break
			}
		}

		// 没找到对应列则报错
		if index == -1 {
			colRes.Ok = false
			colRes.Reason = fmt.Sprintf("未找到%s对应的列", colRule.PropName)
			continue
		}
		colRes.ColIndex = &index

		// 验证列类型是否正确
		if colRule.PropType != cols[index][excelio.MJS_FIXED_ROWS_TYPE] {
			fmt.Printf("DEBUG: Column type mismatch for %s: actual=%s, expected=%s\n", colRule.PropName, cols[index][excelio.MJS_FIXED_ROWS_TYPE], colRule.PropType)
			colRes.Ok = false
			colRes.Reason = fmt.Sprintf("列类型不正确, 实际类型:%s, 期望类型:%s", cols[index][excelio.MJS_FIXED_ROWS_TYPE], colRule.PropType)
			continue
		}

		// 执行所有列规则检查
		fmt.Printf("--- Entering PropRules loop for column %s, rule count: %d\n", colRule.PropName, len(colRule.PropRules))
		for _, cRule := range colRule.PropRules {
			fmt.Printf("--- cRule.Type %s \n", cRule.Type)
			checker := Manager.GetChecker(cRule.Type)
			if checker == nil {
				log.Printf("没有对应的检查器: %s", cRule.Type)
			} else {
				// 创建参数副本，设置 useIdColForEnd=true（使用 ID 列统一确定数据结束位置）
				checkParams := make(map[string]string)
				for k, v := range cRule.Params {
					checkParams[k] = v
				}
				// 默认启用 ID 列统一检测（如果规则未明确设置）
				if _, exists := checkParams[string(json_rule.USE_ID_COL_FOR_END)]; !exists {
					checkParams[string(json_rule.USE_ID_COL_FOR_END)] = "true"
				}

				colRes.ErrCells = append(colRes.ErrCells, Manager.GetChecker(cRule.Type).Check(rule.Sheet, cols, index, excelio.MJS_FIXED_ROWS_NUM, checkParams, sheetMap)...)
			}
		}

		// 根据错误单元格数量设置检查结果状态
		if len(colRes.ErrCells) > 0 {
			colRes.Ok = false
			colRes.Reason = "单元格内存在错误"
		} else {
			colRes.Ok = true
			colRes.Reason = ""
		}

		res = append(res, colRes)
	}

	return res, nil
}
