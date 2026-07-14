package coded_rules

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// DrawskinDataValidityCheckRule DrawSkin基础数据检查（DSK-01/12）
// DSK-01: Id > 0 且唯一
// DSK-12: DailyLimit >= 0
type DrawskinDataValidityCheckRule struct{}

func (c *DrawskinDataValidityCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.DRAWSKIN_DATA_VALIDITY_CHECK,
		DisplayName:  "DrawSkin基础数据检查",
		Description:  "检查DrawSkin表的ID唯一性(DSK-01)和DailyLimit值合法性(DSK-12)",
		TargetSheets: []string{"DrawSkin"},
		ParamDefs:    []json_rule.TableRuleParamDef{},
	}
}

func (c *DrawskinDataValidityCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DrawSkin基础数据检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	dailyLimitColIdx := helpers.GetColIndexByName(param.Cols, "DailyLimit")

	if idColIdx == -1 {
		result.Ok = false
		result.Reason = "DrawSkin 表中未找到 Id 列"
		return result
	}

	errorCount := 0
	seenIds := make(map[int]int) // Id值 -> 首次出现的行索引

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		idStr := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}

		name := ""
		if nameColIdx != -1 {
			name = helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		}

		id, err := helpers.ParseIntWithError(idStr)
		if err != nil {
			continue
		}

		// DSK-01: Id > 0 且唯一
		if id <= 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】的 ID=%d 必须 > 0", name, id),
			})
			errorCount++
			continue
		}

		if firstRow, exists := seenIds[id]; exists {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(行%d) 的 ID=%d 与行%d 重复", name, rowIdx, id, firstRow),
			})
			errorCount++
		} else {
			seenIds[id] = rowIdx
		}

		// DSK-12: DailyLimit >= 0
		if dailyLimitColIdx != -1 {
			limitStr := helpers.GetColValue(param.Cols, dailyLimitColIdx, rowIdx)
			if limitStr != "" {
				limit, err := helpers.ParseIntWithError(limitStr)
				if err == nil && limit < 0 {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    rowIdx,
						ExcelRow: rowIdx + 1,
						Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%d) 的 DailyLimit=%d 必须 >= 0", name, id, limit),
					})
					errorCount++
				}
			}
		}
	}

	if errorCount > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("DrawSkin基础数据检查发现 %d 个问题", errorCount)
	}

	return result
}
