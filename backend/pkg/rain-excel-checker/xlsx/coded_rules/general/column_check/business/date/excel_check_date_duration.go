// Package date 提供日期相关的列级校验规则
// 本包中的规则用于检查日期持续时间、日期范围等

package date

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DateDurationCheckRule 日期持续时间检查规则（统一使用秒匹配）
type DateDurationCheckRule struct{}

func (c *DateDurationCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 获取结束日期列的索引
	endColOffset, err := strconv.Atoi(params["endColOffset"])
	if err != nil || endColOffset == 0 {
		// 参数错误
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("结束日期列索引参数错误: %s", params["endColOffset"]),
			},
		}
	}
	if len(cols) < colIdx+endColOffset || colIdx+endColOffset < 0 {
		// 参数错误
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("结束日期列索引参数错误: 越界[%d], 实际:(%d)", colIdx+endColOffset, len(cols)),
			},
		}
	}
	endIdx2 := helpers.GetColEndIndex(cols, colIdx+endColOffset, startRowIdx, breakLine, params)
	endData := cols[colIdx+endColOffset][startRowIdx:endIdx2]

	// 支持的日期格式
	dateFormats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-1-2",
		"2006/1/2",
		"20060102",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006年01月02日",
		"2006年1月2日",
		"2006年01月02日 15时04分05秒",
		"2006年1月2日 15时4分5秒",
	}

	// 解析目标持续时间（转换为秒）
	var targetSeconds float64
	if durationStr, ok := params["duration"]; ok && durationStr != "" {
		// 尝试解析复杂的时间字符串
		duration, err := parseDurationString(durationStr)
		if err != nil {
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("持续时间格式错误: %s (错误: %v)", durationStr, err),
				},
			}
		}
		targetSeconds = duration.Seconds()
	} else if valueStr, ok := params["value"]; ok && valueStr != "" {
		// 向后兼容：使用value参数指定天数（转换为秒）
		if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
			// 默认按天数计算，转换为秒
			targetSeconds = val * 24 * 3600
		}
	}

	// 比较规则：eq（等于）、gt（大于）、lt（小于）、ge（大于等于）、le（小于等于）
	compareRule := "eq"
	if rule, ok := params[string(json_rule.COMPARE_RULE)]; ok {
		compareRule = rule
	}

	// 容差范围（秒）
	var toleranceSeconds float64
	if toleranceStr, ok := params[string(json_rule.TOLERANCE)]; ok && toleranceStr != "" {
		if tol, err := parseDurationString(toleranceStr); err == nil {
			toleranceSeconds = tol.Seconds()
		} else {
			// 尝试解析为纯数字（按秒处理）
			if tolVal, err := strconv.ParseFloat(toleranceStr, 64); err == nil {
				toleranceSeconds = tolVal
			}
		}
	}

	// 是否允许开始日期晚于结束日期
	allowReverse := false
	if allow, ok := params["allowReverse"]; ok {
		allowReverse = strings.ToLower(allow) == "true"
	}

	// 显示单位（仅用于错误信息展示）
	displayUnit := "秒"
	if unit, ok := params["displayUnit"]; ok {
		displayUnit = unit
	}

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	for i := range myColData {
		startStr := myColData[i]
		if len(endData) <= i {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("结束日期列长度不足: %d, 索引: %d", len(endData), i),
			})
			continue
		}
		endStr := endData[i]

		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) || helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx+endColOffset, i, allowEmpty, allowCommit) {
			continue
		}

		// 解析开始日期和结束日期
		startDate, err1 := helpers.ParseDateWithFormats(startStr, dateFormats)
		endDate, err2 := helpers.ParseDateWithFormats(endStr, dateFormats)

		if err1 != nil || err2 != nil {
			// 日期解析失败
			if err1 != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("开始日期格式错误: %s", startStr),
				})
			}
			if err2 != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("结束日期格式错误: %s", endStr),
				})
			}
			continue
		}

		// 检查开始日期是否晚于结束日期
		if !allowReverse && startDate.After(endDate) {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("开始日期晚于结束日期: %s > %s", startStr, endStr),
			})
			continue
		}

		// 计算实际持续时间（秒）
		actualSeconds := endDate.Sub(startDate).Seconds()

		// 根据比较规则进行检查（全部使用秒）
		isValid := false
		var reason string

		// 转换显示值（根据displayUnit）
		displayActual := formatDurationForDisplay(actualSeconds, displayUnit)
		displayTarget := formatDurationForDisplay(targetSeconds, displayUnit)
		displayTolerance := formatDurationForDisplay(toleranceSeconds, displayUnit)

		switch compareRule {
		case "eq": // 等于
			if toleranceSeconds > 0 {
				// 使用容差范围
				isValid = math.Abs(actualSeconds-targetSeconds) <= toleranceSeconds
				if !isValid {
					reason = fmt.Sprintf("持续时间不在目标值附近: %s (目标: %s±%s)",
						displayActual, displayTarget, displayTolerance)
				}
			} else {
				// 精确等于
				isValid = math.Abs(actualSeconds-targetSeconds) < 0.001 // 毫秒级精度
				if !isValid {
					reason = fmt.Sprintf("持续时间不等于目标值: %s (目标: %s)",
						displayActual, displayTarget)
				}
			}

		case "gt": // 大于
			isValid = actualSeconds > targetSeconds
			if !isValid {
				reason = fmt.Sprintf("持续时间不大于目标值: %s ≤ %s",
					displayActual, displayTarget)
			}

		case "lt": // 小于
			isValid = actualSeconds < targetSeconds
			if !isValid {
				reason = fmt.Sprintf("持续时间不小于目标值: %s ≥ %s",
					displayActual, displayTarget)
			}

		case "ge": // 大于等于
			isValid = actualSeconds >= targetSeconds
			if !isValid {
				reason = fmt.Sprintf("持续时间小于目标值: %s < %s",
					displayActual, displayTarget)
			}

		case "le": // 小于等于
			isValid = actualSeconds <= targetSeconds
			if !isValid {
				reason = fmt.Sprintf("持续时间大于目标值: %s > %s",
					displayActual, displayTarget)
			}

		case "between": // 在两个值之间
			minValue := targetSeconds
			maxValueStr, ok := params["maxValue"]
			maxValue := targetSeconds
			if ok && maxValueStr != "" {
				if maxDur, err := parseDurationString(maxValueStr); err == nil {
					maxValue = maxDur.Seconds()
				} else if maxVal, err := strconv.ParseFloat(maxValueStr, 64); err == nil {
					maxValue = maxVal
				}
			}

			displayMin := formatDurationForDisplay(minValue, displayUnit)
			displayMax := formatDurationForDisplay(maxValue, displayUnit)

			isValid = actualSeconds >= minValue && actualSeconds <= maxValue
			if !isValid {
				reason = fmt.Sprintf("持续时间不在范围内: %s (范围: %s-%s)",
					displayActual, displayMin, displayMax)
			}

		case "multiple": // 是某个值的倍数
			if targetSeconds > 0 {
				remainder := math.Mod(actualSeconds, targetSeconds)
				isValid = remainder < 0.001 || math.Abs(remainder-targetSeconds) < 0.001
				if !isValid {
					reason = fmt.Sprintf("持续时间不是%s的倍数: %s",
						displayTarget, displayActual)
				}
			}

		default:
			// 默认按eq处理
			isValid = math.Abs(actualSeconds-targetSeconds) < 0.001
			if !isValid {
				reason = fmt.Sprintf("持续时间不等于目标值: %s (目标: %s)",
					displayActual, displayTarget)
			}
		}

		if !isValid && reason != "" {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   reason,
			})
		}
	}

	return res
}

// parseDurationString 解析复杂的时间间隔字符串
// 支持格式: "2d3h30m", "1w2d", "1h30m", "30m", "2.5d" 等
func parseDurationString(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, nil
	}

	// 如果字符串可以解析为纯数字，按秒处理
	if val, err := strconv.ParseFloat(durationStr, 64); err == nil {
		return time.Duration(val * float64(time.Second)), nil
	}

	var totalDuration time.Duration

	// 正则表达式匹配时间单位
	re := regexp.MustCompile(`([\d\.]+)([ywdhms])`)
	matches := re.FindAllStringSubmatch(durationStr, -1)

	if len(matches) == 0 {
		// 尝试直接解析为标准Duration格式
		if d, err := time.ParseDuration(durationStr); err == nil {
			return d, nil
		}
		return 0, fmt.Errorf("无法解析时间格式: %s", durationStr)
	}

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		valueStr := match[1]
		unit := match[2]

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, fmt.Errorf("无效的数字: %s", valueStr)
		}

		switch unit {
		case "y": // 年（按365天计算）
			seconds := value * 365 * 24 * 3600
			totalDuration += time.Duration(seconds * float64(time.Second))
		case "w": // 周
			seconds := value * 7 * 24 * 3600
			totalDuration += time.Duration(seconds * float64(time.Second))
		case "d": // 天
			seconds := value * 24 * 3600
			totalDuration += time.Duration(seconds * float64(time.Second))
		case "h": // 小时
			seconds := value * 3600
			totalDuration += time.Duration(seconds * float64(time.Second))
		case "m": // 分钟
			seconds := value * 60
			totalDuration += time.Duration(seconds * float64(time.Second))
		case "s": // 秒
			totalDuration += time.Duration(value * float64(time.Second))
		default:
			return 0, fmt.Errorf("未知的时间单位: %s", unit)
		}
	}

	return totalDuration, nil
}

// formatDurationForDisplay 格式化持续时间用于显示
func formatDurationForDisplay(seconds float64, unit string) string {
	if seconds == 0 {
		return "0"
	}

	switch strings.ToLower(unit) {
	case "y", "year", "years":
		years := seconds / (365 * 24 * 3600)
		return fmt.Sprintf("%.2f年", years)
	case "w", "week", "weeks":
		weeks := seconds / (7 * 24 * 3600)
		return fmt.Sprintf("%.2f周", weeks)
	case "d", "day", "days":
		days := seconds / (24 * 3600)
		return fmt.Sprintf("%.2f天", days)
	case "h", "hour", "hours":
		hours := seconds / 3600
		return fmt.Sprintf("%.2f小时", hours)
	case "m", "min", "minute", "minutes":
		minutes := seconds / 60
		return fmt.Sprintf("%.2f分钟", minutes)
	default: // 秒
		return fmt.Sprintf("%.2f秒", seconds)
	}
}
