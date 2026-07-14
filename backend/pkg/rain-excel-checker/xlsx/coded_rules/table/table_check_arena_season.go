// Package table 提供表级别的校验规则
// 本包中的规则针对单个 Excel 表的特定业务逻辑进行校验

package coded_rules

import (
	"fmt"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ArenaSeasonCheckRule 竞技场赛季检查规则
// 检查 ArenaSeason.xlsx 最后一条数据的 SeasonEndTime 字段与当前时间的关系
type ArenaSeasonCheckRule struct{}

// Meta 返回规则元数据
func (c *ArenaSeasonCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.ARENA_SEASON_CHECK,
		DisplayName:  "竞技场赛季不足",
		Description:  "检查 ArenaSeason.xlsx 最后一条数据的 SeasonEndTime 字段，如果距离当前时间不足指定天数则报警",
		TargetSheets: []string{"ArenaSeason"}, // 仅适用于 ArenaSeason 表
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.SEASON_END_TIME_COL,
				Label:       "结束时间列名",
				Description: "赛季结束时间的列名",
				Type:        "string",
				Default:     "SeasonEndTime",
				Required:    false,
			},
			{
				Key:         json_rule.TIME_RANGE_BEFORE,
				Label:       "提前警告时间",
				Description: "距离结束多少时间开始警告（如 168h 表示7天）",
				Type:        "duration",
				Default:     "168h",
				Required:    false,
			},
		},
	}
}

// Check 执行表级检查
func (c *ArenaSeasonCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		DisplayName: "竞技场赛季不足",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 获取参数（调用方已从 ParamDefs 填充默认值，无需回退）
	seasonEndTimeCol := param.Params[string(json_rule.SEASON_END_TIME_COL)]
	timeRangeBefore := param.Params[string(json_rule.TIME_RANGE_BEFORE)]

	// 解析警告时间（容错处理：解析失败时使用默认值 7天）
	warnDuration, err := time.ParseDuration(timeRangeBefore)
	if err != nil {
		// 解析失败时使用默认值 7天，与 DanQingGeTimeActiveCheckRule 保持一致的容错策略
		warnDuration = 7 * 24 * time.Hour
	}

	// 查找 SeasonEndTime 列
	colIdx := -1
	for i, col := range param.Cols {
		if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == seasonEndTimeCol {
			colIdx = i
			break
		}
	}

	if colIdx == -1 {
		result.Ok = false
		result.Reason = fmt.Sprintf("未找到列: %s", seasonEndTimeCol)
		return result
	}

	// 获取该列数据（跳过表头）
	if param.EndIndex <= param.StartRowIdx {
		result.Ok = false
		result.Reason = "没有有效的数据行"
		return result
	}
	colData := param.Cols[colIdx][param.StartRowIdx:]

	// 找到最后一条有效数据
	var lastTimeStr string
	var lastRowIdx int
	for i := len(colData) - 1; i >= 0; i-- {
		if strings.TrimSpace(colData[i]) != "" {
			lastTimeStr = colData[i]
			lastRowIdx = i
			break
		}
	}

	if lastTimeStr == "" {
		result.Ok = false
		result.Reason = "未找到有效的赛季结束时间"
		return result
	}

	// 解析时间（支持多种格式）
	dateFormats := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02",
		"2006/01/02",
	}

	var seasonEndTime time.Time
	for _, format := range dateFormats {
		seasonEndTime, err = time.Parse(format, strings.TrimSpace(lastTimeStr))
		if err == nil {
			break
		}
	}

	if err != nil {
		result.Ok = false
		result.Reason = fmt.Sprintf("无法解析时间: %s", lastTimeStr)
		return result
	}

	// 检查时间
	now := helpers.ResolveNow(param.Now) // 单元测试可通过 param.Now 注入固定时间
	timeUntilEnd := seasonEndTime.Sub(now)

	if timeUntilEnd <= 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("竞技场赛季已结束，请立即配置新赛季（结束时间: %s）", seasonEndTime.Format("2006-01-02 15:04:05"))
		result.ErrCells = append(result.ErrCells, &json_rule.CellError{
			Index:  lastRowIdx + param.StartRowIdx,
			Reason: fmt.Sprintf("赛季已于 %s 结束", seasonEndTime.Format("2006-01-02 15:04:05")),
		})
	} else if timeUntilEnd < warnDuration {
		result.Ok = false
		result.Reason = fmt.Sprintf("竞技场赛季即将结束，剩余 %v，建议及时配置新赛季", timeUntilEnd.Round(time.Hour))
		result.ErrCells = append(result.ErrCells, &json_rule.CellError{
			Index:  lastRowIdx + param.StartRowIdx,
			Reason: fmt.Sprintf("赛季结束时间: %s", seasonEndTime.Format("2006-01-02 15:04:05")),
		})
	}

	return result
}
