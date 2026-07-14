package coded_rules

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// DropRuleDataValidityCheckRule DropRule基础数据检查
//
// ## 校验规则
// 1. Count 必须 > 0
// 2. DropGroup 不应为空数组
// 3. EnsureSmall > 0 且 EnsureBig > 0 时，EnsureSmall < EnsureBig
//
// ## 相关表结构
// - DropRule: Id, Name, Count, DropGroup, EnsureSmall, EnsureBig, EnsureItemCount
//
// ## 检查流程
// 1. 查找 DropRule 表中的关键列
// 2. 遍历数据行，逐行检查三条规则
// 3. 汇总错误并返回
//
// ## 已移除的检查
//   - ~~EnsureItemCount 不应 > Count~~：EnsureItemCount 是跨多次抽奖调用的累计保底计数器，
//     Count 是单次抽奖次数，两者是不同维度的概念，EnsureItemCount > Count 是正常保底机制。
//     服务端代码参考：roll.go:133-144（Count=单次循环次数，EnsureItemCount=累计保底计数器）
type DropRuleDataValidityCheckRule struct{}

// Meta 返回规则元数据
func (c *DropRuleDataValidityCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.DROP_RULE_DATA_VALIDITY_CHECK,
		DisplayName:  "DropRule基础数据检查",
		Description:  "检查DropRule表中Count、DropGroup、EnsureSmall/EnsureBig的基本数据有效性",
		TargetSheets: []string{"DropRule"},
		ParamDefs:    []json_rule.TableRuleParamDef{},
	}
}

// Check 执行DropRule基础数据检查
func (c *DropRuleDataValidityCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DropRule基础数据检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 查找关键列索引
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	countColIdx := helpers.GetColIndexByName(param.Cols, "Count")
	dropGroupColIdx := helpers.GetColIndexByName(param.Cols, "DropGroup")
	ensureSmallColIdx := helpers.GetColIndexByName(param.Cols, "EnsureSmall")
	ensureBigColIdx := helpers.GetColIndexByName(param.Cols, "EnsureBig")

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		rowName := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		var errors []string

		// 1. Count > 0
		if countColIdx >= 0 {
			countStr := helpers.GetColValue(param.Cols, countColIdx, rowIdx)
			if countStr != "" {
				count, err := strconv.Atoi(countStr)
				if err == nil && count <= 0 {
					errors = append(errors, fmt.Sprintf("Count=%d 必须 > 0", count))
				}
			}
		}

		// 2. DropGroup 不应为空
		if dropGroupColIdx >= 0 {
			dropGroupVal := helpers.GetColValue(param.Cols, dropGroupColIdx, rowIdx)
			if dropGroupVal == "" {
				errors = append(errors, "DropGroup 不应为空")
			}
		}

		// 3. EnsureSmall > 0 且 EnsureBig > 0 时，EnsureSmall < EnsureBig
		if ensureSmallColIdx >= 0 && ensureBigColIdx >= 0 {
			ensureSmallStr := helpers.GetColValue(param.Cols, ensureSmallColIdx, rowIdx)
			ensureBigStr := helpers.GetColValue(param.Cols, ensureBigColIdx, rowIdx)
			if ensureSmallStr != "" && ensureBigStr != "" {
				ensureSmall, err1 := strconv.Atoi(ensureSmallStr)
				ensureBig, err2 := strconv.Atoi(ensureBigStr)
				if err1 == nil && err2 == nil && ensureSmall > 0 && ensureBig > 0 && ensureSmall >= ensureBig {
					errors = append(errors, fmt.Sprintf("EnsureSmall=%d 应 < EnsureBig=%d（小保底次数应小于大保底次数）", ensureSmall, ensureBig))
				}
			}
		}

		if len(errors) > 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("掉落规则【%s】(ID=%s) %s", rowName, rowId, strings.Join(errors, "; ")),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个掉落规则基础数据问题", len(result.ErrCells))
	}

	return result
}
