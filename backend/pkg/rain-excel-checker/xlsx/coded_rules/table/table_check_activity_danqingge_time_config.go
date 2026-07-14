package coded_rules

import (
	"fmt"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ActivityDanqinggeTimeConfigCheckRule 丹青阁活动时间配置检查规则
//
// ## 校验规则（ACT-04/05/06）
// 1. ACT-04: 丹青阁活动 TimeType=1 时，StartTime/EndTime/RewardTime/RewardEndTime 时间格式必须合法
// 2. ACT-05: EndTime 必须大于 StartTime（仅 TimeType=1 且都成功解析时）
// 3. ACT-06: RewardTime 应大于 EndTime，RewardEndTime 应大于 RewardTime（仅 TimeType=1 且都成功解析时）
type ActivityDanqinggeTimeConfigCheckRule struct{}

// Meta 返回规则元数据
func (c *ActivityDanqinggeTimeConfigCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.ACTIVITY_DANQINGGE_TIME_CONFIG_CHECK,
		DisplayName:  "丹青阁活动时间配置检查",
		Description:  "检查丹青阁活动(ActivityType=ActTypeSkinRaffle)的时间格式合法性、时间顺序正确性（ACT-04/05/06）",
		TargetSheets: []string{"Activity"},
		ParamDefs:    []json_rule.TableRuleParamDef{},
	}
}

// Check 执行丹青阁活动时间配置检查
func (c *ActivityDanqinggeTimeConfigCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "丹青阁活动时间配置检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 查找关键列索引
	activityTypeColIdx := excelio.GetColIndexByName(param.Cols, "ActivityType")
	timeTypeColIdx := excelio.GetColIndexByName(param.Cols, "TimeType")
	idColIdx := excelio.GetColIndexByName(param.Cols, "Id")
	nameColIdx := excelio.GetColIndexByName(param.Cols, "Name")
	startTimeColIdx := excelio.GetColIndexByName(param.Cols, "StartTime")
	endTimeColIdx := excelio.GetColIndexByName(param.Cols, "EndTime")
	rewardTimeColIdx := excelio.GetColIndexByName(param.Cols, "RewardTime")
	rewardEndTimeColIdx := excelio.GetColIndexByName(param.Cols, "RewardEndTime")

	if activityTypeColIdx == -1 {
		result.Ok = true
		result.Reason = "未找到 ActivityType 列，跳过检查"
		return result
	}

	errorCount := 0
	foundCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		// 筛选丹青阁活动
		activityType := strings.TrimSpace(excelio.GetColValue(param.Cols, activityTypeColIdx, rowIdx))
		if activityType != DanQingGeActivityType {
			continue
		}

		// 仅检查 TimeType=1（绝对时间）
		if timeTypeColIdx != -1 {
			timeType := strings.TrimSpace(excelio.GetColValue(param.Cols, timeTypeColIdx, rowIdx))
			if timeType != "1" {
				continue
			}
		}

		foundCount++
		rowLabel := c.getRowLabel(param.Cols, idColIdx, nameColIdx, rowIdx)

		// ACT-04: 解析各时间字段并检查格式
		startTime, startInvalid := c.parseField(param.Cols, startTimeColIdx, rowIdx)
		endTime, endInvalid := c.parseField(param.Cols, endTimeColIdx, rowIdx)
		rewardTime, rewardInvalid := c.parseField(param.Cols, rewardTimeColIdx, rowIdx)
		rewardEndTime, rewardEndInvalid := c.parseField(param.Cols, rewardEndTimeColIdx, rowIdx)

		// 报告格式错误
		if startInvalid != "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 StartTime【%s】不是合法日期格式", rowLabel, startInvalid),
			})
			errorCount++
		}
		if endInvalid != "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 EndTime【%s】不是合法日期格式", rowLabel, endInvalid),
			})
			errorCount++
		}
		if rewardInvalid != "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 RewardTime【%s】不是合法日期格式", rowLabel, rewardInvalid),
			})
			errorCount++
		}
		if rewardEndInvalid != "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 RewardEndTime【%s】不是合法日期格式", rowLabel, rewardEndInvalid),
			})
			errorCount++
		}

		// ACT-05: EndTime > StartTime（两者都成功解析时）
		if !startTime.IsZero() && !endTime.IsZero() && !endTime.After(startTime) {
			startStr := strings.TrimSpace(excelio.GetColValue(param.Cols, startTimeColIdx, rowIdx))
			endStr := strings.TrimSpace(excelio.GetColValue(param.Cols, endTimeColIdx, rowIdx))
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 EndTime【%s】不晚于 StartTime【%s】", rowLabel, endStr, startStr),
			})
			errorCount++
		}

		// ACT-06: RewardTime > EndTime 且 RewardEndTime > RewardTime
		if !rewardTime.IsZero() && !endTime.IsZero() && !rewardTime.After(endTime) {
			rewardStr := strings.TrimSpace(excelio.GetColValue(param.Cols, rewardTimeColIdx, rowIdx))
			endStr := strings.TrimSpace(excelio.GetColValue(param.Cols, endTimeColIdx, rowIdx))
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 RewardTime【%s】不晚于 EndTime【%s】", rowLabel, rewardStr, endStr),
			})
			errorCount++
		}
		if !rewardEndTime.IsZero() && !rewardTime.IsZero() && !rewardEndTime.After(rewardTime) {
			rewardEndStr := strings.TrimSpace(excelio.GetColValue(param.Cols, rewardEndTimeColIdx, rowIdx))
			rewardStr := strings.TrimSpace(excelio.GetColValue(param.Cols, rewardTimeColIdx, rowIdx))
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】的 RewardEndTime【%s】不晚于 RewardTime【%s】", rowLabel, rewardEndStr, rewardStr),
			})
			errorCount++
		}
	}

	if foundCount == 0 {
		result.Ok = true
		result.Reason = "未找到丹青阁绝对时间类型活动"
		return result
	}

	if errorCount > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("丹青阁活动时间配置检查发现 %d 个问题", errorCount)
	}

	return result
}

// parseField 解析时间字段，返回解析后的时间和格式错误原始值
// 如果字段为空返回 (zeroTime, "")
// 如果格式非法返回 (zeroTime, rawValue)
// 如果合法返回 (time, "")
func (c *ActivityDanqinggeTimeConfigCheckRule) parseField(cols [][]string, colIdx, rowIdx int) (time.Time, string) {
	if colIdx == -1 {
		return time.Time{}, ""
	}
	raw := strings.TrimSpace(excelio.GetColValue(cols, colIdx, rowIdx))
	if raw == "" {
		return time.Time{}, ""
	}
	t := excelio.ParseDate(raw)
	if t.IsZero() {
		return time.Time{}, raw
	}
	return t, ""
}

// getRowLabel 获取行的显示标签
func (c *ActivityDanqinggeTimeConfigCheckRule) getRowLabel(cols [][]string, idColIdx, nameColIdx, rowIdx int) string {
	name := ""
	if nameColIdx != -1 {
		name = strings.TrimSpace(excelio.GetColValue(cols, nameColIdx, rowIdx))
	}
	id := ""
	if idColIdx != -1 {
		id = strings.TrimSpace(excelio.GetColValue(cols, idColIdx, rowIdx))
	}
	if name != "" && id != "" {
		return fmt.Sprintf("%s(id:%s)", name, id)
	}
	if name != "" {
		return name
	}
	if id != "" {
		return fmt.Sprintf("id:%s", id)
	}
	return fmt.Sprintf("行%d", rowIdx)
}
