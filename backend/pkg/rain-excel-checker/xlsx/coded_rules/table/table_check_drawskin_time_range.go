package coded_rules

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// DrawskinTimeRangeCheckRule DrawSkin时间范围检查（DSK-10/11/13）
// DSK-10: StartTime/EndTime 时间格式合法性（Warning）
// DSK-11: StartTime <= EndTime（Error）
// DSK-13: StartTime/EndTime 均为空（Error）
type DrawskinTimeRangeCheckRule struct{}

func (c *DrawskinTimeRangeCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.DRAWSKIN_TIME_RANGE_CHECK,
		DisplayName:  "DrawSkin时间范围检查",
		Description:  "检查DrawSkin表的StartTime/EndTime格式合法性(DSK-10)、时间顺序(DSK-11)和空值检查(DSK-13)",
		TargetSheets: []string{"DrawSkin"},
		ParamDefs:    []json_rule.TableRuleParamDef{},
	}
}

func (c *DrawskinTimeRangeCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "DrawSkin时间范围检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")
	startColIdx := helpers.GetColIndexByName(param.Cols, "StartTime")
	endColIdx := helpers.GetColIndexByName(param.Cols, "EndTime")

	warningCount := 0
	errorCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := ""
		if idColIdx != -1 {
			rowId = helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		}
		name := ""
		if nameColIdx != -1 {
			name = helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		}

		// 获取 StartTime 和 EndTime 的原始值
		startStr := ""
		endStr := ""
		if startColIdx != -1 {
			startStr = strings.TrimSpace(helpers.GetColValue(param.Cols, startColIdx, rowIdx))
		}
		if endColIdx != -1 {
			endStr = strings.TrimSpace(helpers.GetColValue(param.Cols, endColIdx, rowIdx))
		}

		// DSK-13: StartTime 和 EndTime 均为空时报错
		if startStr == "" && endStr == "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 StartTime 和 EndTime 均为空，至少需要配置一个时间", name, rowId),
			})
			errorCount++
			continue // 跳过后续检查
		}

		// 分别记录解析结果，避免 interface{} 存储 time.Time
		startOk := false // StartTime 是否成功解析
		endOk := false   // EndTime 是否成功解析

		// DSK-10: 检查 StartTime 格式
		if startColIdx != -1 && startStr != "" {
			t := excelio.ParseDate(startStr)
			if t.IsZero() {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 StartTime【%s】不是合法日期格式", name, rowId, startStr),
				})
				warningCount++
			} else {
				startOk = true
			}
		}

		// DSK-10: 检查 EndTime 格式
		if endColIdx != -1 && endStr != "" {
			t := excelio.ParseDate(endStr)
			if t.IsZero() {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 EndTime【%s】不是合法日期格式", name, rowId, endStr),
				})
				warningCount++
			} else {
				endOk = true
			}
		}

		// DSK-11: StartTime <= EndTime（两者都成功解析时比较）
		if startOk && endOk {
			s := excelio.ParseDate(startStr)
			e := excelio.ParseDate(endStr)
			if s.After(e) {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index:    rowIdx,
					ExcelRow: rowIdx + 1,
					Reason:   fmt.Sprintf("皮肤抽奖池【%s】(ID=%s) 的 StartTime【%s】晚于 EndTime【%s】", name, rowId, startStr, endStr),
				})
				errorCount++
			}
		}
	}

	if errorCount > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("DrawSkin时间范围检查发现 %d 个错误", errorCount)
	} else if warningCount > 0 {
		result.Ok = true
		result.Reason = fmt.Sprintf("DrawSkin时间范围检查发现 %d 个格式警告", warningCount)
	}

	return result
}
