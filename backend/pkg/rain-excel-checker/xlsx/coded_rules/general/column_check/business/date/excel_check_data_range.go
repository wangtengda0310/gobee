// Package date 提供日期相关的列级校验规则
// 本包中的规则用于检查日期持续时间、日期范围等

package date

import (
	"fmt"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DateRangeCheckRule 日期范围检查规则
type DateRangeCheckRule struct{}

func (c *DateRangeCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

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
	}

	// 解析开始日期
	var startDate, endDate time.Time
	var err error

	if startStr, ok := params["startDate"]; ok && startStr != "" {
		startDate, err = helpers.ParseDateWithFormats(startStr, dateFormats)
		if err != nil {
			// 如果无法解析开始日期，返回错误
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("配置的开始日期格式错误: %s", startStr),
				},
			}
		}
	}

	// 解析结束日期
	if endStr, ok := params["endDate"]; ok && endStr != "" {
		endDate, err = helpers.ParseDateWithFormats(endStr, dateFormats)
		if err != nil {
			// 如果无法解析结束日期，返回错误
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("配置的结束日期格式错误: %s", endStr),
				},
			}
		}
	}

	// 检查类型：within（在范围内）、outside（在范围外）、before（在之前）、after（在之后）
	checkType := "within"
	if t, ok := params["checkType"]; ok {
		checkType = t
	}

	// 是否包含边界
	includeBoundary := true
	if include, ok := params["includeBoundary"]; ok {
		includeBoundary = strings.ToLower(include) == "true"
	}

	for i, dateStr := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 解析单元格中的日期
		cellDate, err := helpers.ParseDateWithFormats(dateStr, dateFormats)
		if err != nil {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("日期格式错误: %s", dateStr),
			})
			continue
		}

		// 根据检查类型进行验证
		isValid := false
		var reason string

		switch checkType {
		case "within": // 在指定范围内
			if !startDate.IsZero() && !endDate.IsZero() {
				if includeBoundary {
					isValid = !cellDate.Before(startDate) && !cellDate.After(endDate)
				} else {
					isValid = cellDate.After(startDate) && cellDate.Before(endDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期不在范围内[%s, %s]",
							formatDateForDisplay(startDate), formatDateForDisplay(endDate))
					} else {
						reason = fmt.Sprintf("日期不在范围内(%s, %s)",
							formatDateForDisplay(startDate), formatDateForDisplay(endDate))
					}
				}
			} else if !startDate.IsZero() { // 只有开始日期，检查是否在开始日期之后
				if includeBoundary {
					isValid = !cellDate.Before(startDate)
				} else {
					isValid = cellDate.After(startDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期早于%s", formatDateForDisplay(startDate))
					} else {
						reason = fmt.Sprintf("日期不晚于%s", formatDateForDisplay(startDate))
					}
				}
			} else if !endDate.IsZero() { // 只有结束日期，检查是否在结束日期之前
				if includeBoundary {
					isValid = !cellDate.After(endDate)
				} else {
					isValid = cellDate.Before(endDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期晚于%s", formatDateForDisplay(endDate))
					} else {
						reason = fmt.Sprintf("日期不早于%s", formatDateForDisplay(endDate))
					}
				}
			} else {
				// 没有指定范围，跳过检查
				continue
			}

		case "outside": // 在指定范围外
			if !startDate.IsZero() && !endDate.IsZero() {
				if includeBoundary {
					isValid = cellDate.Before(startDate) || cellDate.After(endDate)
				} else {
					isValid = cellDate.Before(startDate) || cellDate.After(endDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期在范围内[%s, %s]，应在范围外",
							formatDateForDisplay(startDate), formatDateForDisplay(endDate))
					} else {
						reason = fmt.Sprintf("日期在范围内(%s, %s)，应在范围外",
							formatDateForDisplay(startDate), formatDateForDisplay(endDate))
					}
				}
			}

		case "before": // 在某个日期之前
			if !endDate.IsZero() {
				if includeBoundary {
					isValid = !cellDate.After(endDate)
				} else {
					isValid = cellDate.Before(endDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期晚于%s", formatDateForDisplay(endDate))
					} else {
						reason = fmt.Sprintf("日期不早于%s", formatDateForDisplay(endDate))
					}
				}
			}

		case "after": // 在某个日期之后
			if !startDate.IsZero() {
				if includeBoundary {
					isValid = !cellDate.Before(startDate)
				} else {
					isValid = cellDate.After(startDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期早于%s", formatDateForDisplay(startDate))
					} else {
						reason = fmt.Sprintf("日期不晚于%s", formatDateForDisplay(startDate))
					}
				}
			}

		case "not_between": // 不在两个日期之间（与within相反）
			if !startDate.IsZero() && !endDate.IsZero() {
				if includeBoundary {
					isValid = cellDate.Before(startDate) || cellDate.After(endDate)
				} else {
					isValid = cellDate.Before(startDate) || cellDate.After(endDate)
				}
				if !isValid {
					if includeBoundary {
						reason = fmt.Sprintf("日期在[%s, %s]之间，应不在此区间",
							formatDateForDisplay(startDate), formatDateForDisplay(endDate))
					} else {
						reason = fmt.Sprintf("日期在(%s, %s)之间，应不在此区间",
							formatDateForDisplay(startDate), formatDateForDisplay(endDate))
					}
				}
			}

		default:
			// 默认按within处理
			if !startDate.IsZero() && !endDate.IsZero() {
				isValid = !cellDate.Before(startDate) && !cellDate.After(endDate)
				if !isValid {
					reason = fmt.Sprintf("日期不在范围内[%s, %s]",
						formatDateForDisplay(startDate), formatDateForDisplay(endDate))
				}
			}
		}

		if !isValid && reason != "" {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   fmt.Sprintf("%s (当前日期: %s)", reason, formatDateForDisplay(cellDate)),
			})
		}
	}

	return res
}

// 辅助函数：格式化日期用于显示
func formatDateForDisplay(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		// 如果没有时间部分，只显示日期
		return t.Format("2006-01-02")
	}
	return t.Format("2006-01-02 15:04:05")
}
