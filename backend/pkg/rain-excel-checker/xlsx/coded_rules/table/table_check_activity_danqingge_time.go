// Package coded_rules 提供表级别的校验规则
// 本文件实现丹青阁活动时间校验规则

package coded_rules

import (
	"fmt"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// 丹青阁活动类型标识，对应服务端 EActivityType_ActTypeSkinRaffle (枚举值13)
// 服务端代码参考: useractManager.go:98 case excel.EActivityType_ActTypeSkinRaffle
const DanQingGeActivityType = "ActTypeSkinRaffle"

// DanQingGeTimeActiveCheckRule 丹青阁活动时间校验规则
// 检查 Activity_活动表.xlsx 中 ActivityType 为 ActTypeSkinRaffle 的活动行：
//   - 校验逻辑参考服务端 useractBase.go CheckConfig 方法
//   - 如果活动已结束（EndTime < 当前时间），报错
//   - 如果活动离结束不到指定天数（默认7天），警告
//   - 仅检查 TimeType=1（绝对时间）的活动行
type DanQingGeTimeActiveCheckRule struct{}

// Meta 返回规则元数据
func (c *DanQingGeTimeActiveCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.DANQINGGE_TIME_ACTIVE_CHECK,
		DisplayName:  "丹青阁活动时间校验",
		Description:  "检查 Activity 活动表中丹青阁活动(ActTypeSkinRaffle)是否在生效时间范围内，离结束不到指定天数则警告",
		TargetSheets: []string{"Activity"},
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.TIME_RANGE_BEFORE,
				Label:       "提前警告时间",
				Description: "距离活动结束多少时间开始警告（如 168h 表示7天）",
				Type:        "duration",
				Default:     "168h",
				Required:    false,
			},
		},
	}
}

// Check 执行丹青阁活动时间校验
// 遍历 Activity 表中 ActivityType 为 ActTypeSkinRaffle 的数据行，检查其 EndTime
// 校验逻辑参考服务端 useractBase.go:30 CheckConfig 方法
func (c *DanQingGeTimeActiveCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "丹青阁活动时间校验",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 解析警告时间阈值
	warnDuration := c.parseWarnDuration(param.Params)

	// 查找关键列
	activityTypeColIdx := excelio.GetColIndexByName(param.Cols, "ActivityType")
	timeTypeColIdx := excelio.GetColIndexByName(param.Cols, "TimeType")
	idColIdx := excelio.GetColIndexByName(param.Cols, "Id")
	nameColIdx := excelio.GetColIndexByName(param.Cols, "Name")
	startTimeColIdx := excelio.GetColIndexByName(param.Cols, "StartTime")
	endTimeColIdx := excelio.GetColIndexByName(param.Cols, "EndTime")

	if activityTypeColIdx == -1 {
		result.Ok = false
		result.Reason = "未找到 ActivityType 列"
		return result
	}

	if endTimeColIdx == -1 {
		result.Ok = false
		result.Reason = "未找到 EndTime 列"
		return result
	}

	// 确保有数据行可检查
	if param.EndIndex <= param.StartRowIdx {
		result.Ok = false
		result.Reason = "没有有效的数据行"
		return result
	}

	now := helpers.ResolveNow(param.Now) // 单元测试可通过 param.Now 注入固定时间
	warnCount := 0
	errorCount := 0
	skippedCount := 0

	// 遍历所有数据行，只处理丹青阁活动
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		activityType := strings.TrimSpace(excelio.GetColValue(param.Cols, activityTypeColIdx, rowIdx))
		if activityType != DanQingGeActivityType {
			continue
		}

		// 仅检查绝对时间类型的活动（TimeType=1）
		if timeTypeColIdx != -1 {
			timeType := strings.TrimSpace(excelio.GetColValue(param.Cols, timeTypeColIdx, rowIdx))
			if timeType != "1" && timeType != "" {
				skippedCount++
				continue
			}
		}

		endTimeStr := strings.TrimSpace(excelio.GetColValue(param.Cols, endTimeColIdx, rowIdx))
		if endTimeStr == "" {
			continue
		}

		endTime := excelio.ParseDate(endTimeStr)
		if endTime.IsZero() {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("无法解析结束时间: %s", endTimeStr),
			})
			errorCount++
			continue
		}

		rowLabel := c.getRowLabel(param.Cols, idColIdx, nameColIdx, rowIdx)
		timeUntilEnd := endTime.Sub(now)

		if timeUntilEnd <= 0 {
			// 活动已结束
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】已结束（结束时间: %s）", rowLabel, endTime.Format("2006-01-02 15:04:05")),
			})
			errorCount++
		} else if timeUntilEnd < warnDuration {
			// 活动即将结束
			daysLeft := int(timeUntilEnd.Hours() / 24)
			hoursLeft := int(timeUntilEnd.Hours()) % 24
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("丹青阁活动【%s】即将结束，剩余 %d天%d小时（结束时间: %s）", rowLabel, daysLeft, hoursLeft, endTime.Format("2006-01-02 15:04:05")),
			})
			warnCount++
		}

		// 检查开始时间是否在未来（尚未开始的活动）
		if startTimeColIdx != -1 {
			startTimeStr := strings.TrimSpace(excelio.GetColValue(param.Cols, startTimeColIdx, rowIdx))
			if startTimeStr != "" {
				startTime := excelio.ParseDate(startTimeStr)
				if !startTime.IsZero() && startTime.After(now) {
					result.ErrCells = append(result.ErrCells, &json_rule.CellError{
						Index:    rowIdx,
						ExcelRow: rowIdx + 1,
						Reason:   fmt.Sprintf("丹青阁活动【%s】尚未开始（开始时间: %s）", rowLabel, startTime.Format("2006-01-02 15:04:05")),
					})
					warnCount++
				}
			}
		}
	}

	// 汇总结果
	if errorCount > 0 || warnCount > 0 {
		result.Ok = false
		parts := make([]string, 0, 3)
		if errorCount > 0 {
			parts = append(parts, fmt.Sprintf("%d 个活动已结束", errorCount))
		}
		if warnCount > 0 {
			parts = append(parts, fmt.Sprintf("%d 个活动即将结束或尚未开始", warnCount))
		}
		result.Reason = fmt.Sprintf("丹青阁活动时间检查: %s", strings.Join(parts, "，"))
	}

	return result
}

// parseWarnDuration 解析警告时间阈值
func (c *DanQingGeTimeActiveCheckRule) parseWarnDuration(params map[string]string) time.Duration {
	if timeRangeBefore, ok := params[string(json_rule.TIME_RANGE_BEFORE)]; ok && timeRangeBefore != "" {
		if d, err := time.ParseDuration(timeRangeBefore); err == nil {
			return d
		}
	}
	return 7 * 24 * time.Hour
}

// getRowLabel 获取行的显示标签，用于错误提示
func (c *DanQingGeTimeActiveCheckRule) getRowLabel(cols [][]string, idColIdx, nameColIdx, rowIdx int) string {
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
