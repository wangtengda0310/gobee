package coded_rules

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// DrawskinTenCostCheckRule DrawSkin十连消耗配置检查（DSK-09）
// 如果 TenDropRule > 0，则 TenItemCost 应非空配置
type DrawskinTenCostCheckRule struct{}

func (c *DrawskinTenCostCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.DRAWSKIN_TEN_COST_CHECK,
		DisplayName:  "DrawSkin十连消耗配置检查",
		Description:  "检查DrawSkin表中配置了TenDropRule但未配置TenItemCost的情况(DSK-09)",
		TargetSheets: []string{"DrawSkin"},
		ParamDefs:    []json_rule.TableRuleParamDef{},
	}
}

func (c *DrawskinTenCostCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DrawSkin十连消耗配置检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	tenDropColIdx := helpers.GetColIndexByName(param.Cols, "TenDropRule")
	tenCostColIdx := helpers.GetColIndexByName(param.Cols, "TenItemCost")

	if tenDropColIdx == -1 {
		result.Ok = true
		result.Reason = "未找到 TenDropRule 列，跳过检查"
		return result
	}

	warningCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := ""
		if idColIdx != -1 {
			rowId = helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		}
		name := ""
		if nameColIdx != -1 {
			name = helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		}

		tenDropStr := helpers.GetColValue(param.Cols, tenDropColIdx, rowIdx)
		if tenDropStr == "" {
			continue
		}
		tenDropRule, err := helpers.ParseIntWithError(tenDropStr)
		if err != nil || tenDropRule <= 0 {
			continue
		}

		// TenDropRule > 0，检查 TenItemCost 是否配置
		if tenCostColIdx == -1 {
			continue
		}
		tenCostStr := helpers.GetColValue(param.Cols, tenCostColIdx, rowIdx)
		if tenCostStr == "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 配置了 TenDropRule=%d 但 TenItemCost 未配置", name, rowId, tenDropRule),
			})
			warningCount++
		}
	}

	if warningCount > 0 {
		result.Ok = true // Warning 性质
		result.Reason = fmt.Sprintf("DrawSkin十连消耗配置检查发现 %d 个警告", warningCount)
	}

	return result
}
