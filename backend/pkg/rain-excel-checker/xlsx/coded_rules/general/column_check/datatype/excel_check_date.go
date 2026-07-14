// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DateCheckRule 日期格式检查（根据指定格式检查）
type DateCheckRule struct{}

func (c *DateCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 支持的日期格式映射表
	formatMap := map[string]string{
		// 标准格式
		"yyyy-MM-dd":            "2006-01-02",
		"yyyy/MM/dd":            "2006/01/02",
		"yyyy-M-d":              "2006-1-2",
		"yyyy/M/d":              "2006/1/2",
		"yyyyMMdd":              "20060102",
		"yyyy-MM-dd HH:mm:ss":   "2006-01-02 15:04:05",
		"yyyy/MM/dd HH:mm:ss":   "2006/01/02 15:04:05",
		"yyyy年MM月dd日":           "2006年01月02日",
		"yyyy年M月d日":             "2006年1月2日",
		"yyyy年MM月dd日 HH时mm分ss秒": "2006年01月02日 15时04分05秒",
		"yyyy年M月d日 H时m分s秒":      "2006年1月2日 15时4分5秒",

		// 带有时区的格式
		"yyyy-MM-ddTHH:mm:ssZ":    "2006-01-02T15:04:05Z",
		"yyyy-MM-dd HH:mm:ss.SSS": "2006-01-02 15:04:05.000",

		// 时间戳格式
		"timestamp":    "", // 特殊处理
		"timestamp_ms": "", // 特殊处理

		// 英文格式
		"Mon, 02 Jan 2006 15:04:05 MST": "Mon, 02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05":          "02 Jan 2006 15:04:05",
	}

	// 获取要检查的格式
	var checkFormats []string
	if formatParam, ok := params["format"]; ok && formatParam != "" {
		// 支持多种格式，用逗号分隔
		formatStrs := strings.Split(formatParam, ",")
		for _, fmtStr := range formatStrs {
			fmtStr = strings.TrimSpace(fmtStr)
			if fmtStr == "" {
				continue
			}

			// 检查是否为特殊格式
			switch fmtStr {
			case "timestamp":
				// 时间戳格式（秒）
				checkFormats = append(checkFormats, "timestamp")
			case "timestamp_ms":
				// 时间戳格式（毫秒）
				checkFormats = append(checkFormats, "timestamp_ms")
			default:
				// 从映射表中获取Go的时间格式
				if goFormat, exists := formatMap[fmtStr]; exists {
					checkFormats = append(checkFormats, goFormat)
				} else {
					// 如果不在映射表中，尝试直接使用作为Go格式
					checkFormats = append(checkFormats, fmtStr)
				}
			}
		}
	} else {
		// 如果没有指定格式，使用默认格式
		checkFormats = []string{
			"2006-01-02 15:04:05",
			"2006/01/02 15:04:05",
		}
	}

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 是否严格检查（必须完全匹配指定格式，不尝试其他格式）
	strictCheck := false
	if strict, ok := params[string(json_rule.STRICT)]; ok {
		strictCheck = strings.ToLower(strict) == "true"
	}

	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		valid := false
		var matchedFormat string

		// 根据检查模式进行处理
		if strictCheck && len(checkFormats) > 0 {
			// 严格模式：只检查指定的格式
			for _, format := range checkFormats {
				if isValidDateFormat(s, format) {
					valid = true
					matchedFormat = format
					break
				}
			}
		} else {
			// 非严格模式：先检查指定格式，再检查常见格式
			for _, format := range checkFormats {
				if isValidDateFormat(s, format) {
					valid = true
					matchedFormat = format
					break
				}
			}

			// 如果指定格式都没有匹配，且不是严格模式，尝试常见格式
			if !valid && len(checkFormats) > 0 {
				for _, format := range getCommonFormats() {
					if isValidDateFormat(s, format) {
						valid = true
						matchedFormat = format
						break
					}
				}
			}
		}

		if !valid {
			// 生成错误信息
			var expectedFormats []string
			for _, format := range checkFormats {
				if format == "timestamp" {
					expectedFormats = append(expectedFormats, "时间戳（秒）")
				} else if format == "timestamp_ms" {
					expectedFormats = append(expectedFormats, "时间戳（毫秒）")
				} else {
					// 从Go格式反向查找显示格式
					for display, goFmt := range formatMap {
						if goFmt == format {
							expectedFormats = append(expectedFormats, display)
							break
						}
					}
					if len(expectedFormats) == 0 || expectedFormats[len(expectedFormats)-1] != format {
						expectedFormats = append(expectedFormats, format)
					}
				}
			}

			reason := fmt.Sprintf("不是有效的日期格式: %s", s)
			if len(expectedFormats) > 0 {
				if len(expectedFormats) == 1 {
					reason = fmt.Sprintf("不是有效的日期格式: %s (应为: %s)", s, expectedFormats[0])
				} else {
					reason = fmt.Sprintf("不是有效的日期格式: %s (应为以下格式之一: %s)",
						s, strings.Join(expectedFormats, ", "))
				}
			}

			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   reason,
			})
		} else if formatParam, ok := params["expectedFormat"]; ok && formatParam != "" {
			// 如果指定了期望的显示格式，检查是否匹配该格式
			if expectedGoFormat, exists := formatMap[formatParam]; exists && expectedGoFormat != matchedFormat {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason: fmt.Sprintf("日期格式不匹配: %s (应为: %s, 实际: %s)",
						s, formatParam, getFormatDisplayName(matchedFormat)),
				})
			}
		}
	}
	return res
}

// 检查日期格式是否有效
func isValidDateFormat(dateStr, format string) bool {
	if dateStr == "" {
		return false
	}

	// 特殊格式处理
	switch format {
	case "timestamp":
		// 检查是否为时间戳（秒）
		if _, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
			// 简单验证：时间戳应该在合理范围内（2000年-2100年）
			if ts, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
				// 2000-01-01 的时间戳大约是946684800
				// 2100-01-01 的时间戳大约是4102444800
				return ts >= 946684800 && ts <= 4102444800
			}
		}
		return false

	case "timestamp_ms":
		// 检查是否为时间戳（毫秒）
		if _, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
			// 简单验证：时间戳应该在合理范围内（2000年-2100年）
			if ts, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
				// 转换为秒进行验证
				tsSec := ts / 1000
				return tsSec >= 946684800 && tsSec <= 4102444800
			}
		}
		return false

	default:
		// 普通格式检查
		_, err := time.Parse(format, dateStr)
		return err == nil
	}
}

// 获取常见日期格式
func getCommonFormats() []string {
	return []string{
		"2006-01-02",
		"2006/01/02",
		"2006-1-2",
		"2006/1/2",
		"20060102",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006年01月02日",
		"2006年1月2日",
		"2006-01-02T15:04:05Z",
		"Mon, 02 Jan 2006 15:04:05 MST",
	}
}

// 获取格式的显示名称
func getFormatDisplayName(goFormat string) string {
	// 反向映射
	reverseMap := map[string]string{
		"2006-01-02":                    "yyyy-MM-dd",
		"2006/01/02":                    "yyyy/MM/dd",
		"2006-1-2":                      "yyyy-M-d",
		"2006/1/2":                      "yyyy/M/d",
		"20060102":                      "yyyyMMdd",
		"2006-01-02 15:04:05":           "yyyy-MM-dd HH:mm:ss",
		"2006/01/02 15:04:05":           "yyyy/MM/dd HH:mm:ss",
		"2006年01月02日":                   "yyyy年MM月dd日",
		"2006年1月2日":                     "yyyy年M月d日",
		"2006年01月02日 15时04分05秒":         "yyyy年MM月dd日 HH时mm分ss秒",
		"2006年1月2日 15时4分5秒":             "yyyy年M月d日 H时m分s秒",
		"2006-01-02T15:04:05Z":          "yyyy-MM-ddTHH:mm:ssZ",
		"2006-01-02 15:04:05.000":       "yyyy-MM-dd HH:mm:ss.SSS",
		"Mon, 02 Jan 2006 15:04:05 MST": "Mon, 02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05":          "02 Jan 2006 15:04:05",
	}

	if displayName, exists := reverseMap[goFormat]; exists {
		return displayName
	}
	return goFormat
}
