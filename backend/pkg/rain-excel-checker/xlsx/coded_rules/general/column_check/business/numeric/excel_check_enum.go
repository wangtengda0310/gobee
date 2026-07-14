// Package numeric 提供枚举和数值范围的列级通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package numeric

import (
	"fmt"
	"regexp"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// EnumCheckRule 枚举值检查
// 支持两种模式：
//  1. 直接枚举检查：单元格值直接匹配枚举列表
//  2. 正则提取后枚举检查：先用正则提取单元格中的值，再校验提取值是否在枚举范围内
//     参数：pattern（正则表达式）、groups（捕获组索引，默认"1"）
type EnumCheckRule struct{}

func (c *EnumCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	isArr := strings.HasSuffix(cols[colIdx][excelio.MJS_FIXED_ROWS_TYPE], "[]")
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 从参数获取枚举值，格式如 "1,2,3,4,5"
	enumValues := make(map[string]bool)
	if enumStr, ok := params[string(json_rule.ENUMS)]; ok {
		// 清理和分割字符串
		cleanedStr := strings.TrimSpace(enumStr)
		if cleanedStr != "" {
			// 替换多种换行符为逗号
			cleanedStr = strings.ReplaceAll(cleanedStr, "\n", ",")
			cleanedStr = strings.ReplaceAll(cleanedStr, "\r", ",")
			cleanedStr = strings.ReplaceAll(cleanedStr, " ", ",") // 空格也可作为分隔符

			// 分割并处理
			parts := strings.Split(cleanedStr, ",")
			for _, v := range parts {
				v = strings.TrimSpace(v)
				// 跳过空值（处理连续逗号的情况）
				if v == "" {
					continue
				}
				enumValues[v] = true
			}
		}
	}

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 检查是否配置了正则提取模式
	patternStr, hasPattern := params[string(json_rule.PATTERN)]
	var pattern *regexp.Regexp
	var patternErr error
	var groups []int
	if hasPattern && patternStr != "" {
		pattern, patternErr = regexp.Compile(patternStr)
		groups = helpers.ParseCaptureGroups(params[string(json_rule.GROUPS)])
	}

	if isArr {
		for i, s := range myColData {
			// 处理空值和注释
			if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
				continue
			}
			// 处理数组分割
			arr := strings.Split(s, ",")
			for i2, s2 := range arr {
				if hasPattern && patternErr == nil {
					// 正则提取模式：提取后校验枚举
					extracted := helpers.ExtractValuesByRegex(s2, pattern, groups)
					if len(extracted) == 0 {
						res = append(res, &json_rule.CellError{
							Index:    i,
							ExcelRow: startRowIdx + i + 1,
							Reason:   fmt.Sprintf("值%s[%d]正则未匹配到可提取的值", s, i2),
						})
						continue
					}
					for _, val := range extracted {
						val = strings.TrimSpace(val)
						if val == "" {
							continue
						}
						if !enumValues[val] {
							res = append(res, &json_rule.CellError{
								Index:    i,
								ExcelRow: startRowIdx + i + 1,
								Reason:   fmt.Sprintf("值%s[%d]提取值%s不在允许的枚举值范围内", s, i2, val),
							})
						}
					}
				} else {
					// 直接枚举检查
					if !enumValues[s2] {
						res = append(res, &json_rule.CellError{
							Index:    i,
							ExcelRow: startRowIdx + i + 1,
							Reason:   fmt.Sprintf("值%s[%d]不在允许的枚举值范围内", s, i2),
						})
					}
				}
			}
		}
	} else {
		for i, s := range myColData {
			// 处理空值和注释
			if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
				continue
			}
			if hasPattern && patternErr == nil {
				// 正则提取模式：提取后校验枚举
				extracted := helpers.ExtractValuesByRegex(s, pattern, groups)
				if len(extracted) == 0 {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("正则未匹配到任何可提取的值: %s", s),
					})
					continue
				}
				for _, val := range extracted {
					val = strings.TrimSpace(val)
					if val == "" {
						continue
					}
					if !enumValues[val] {
						res = append(res, &json_rule.CellError{
							Index:    i,
							ExcelRow: startRowIdx + i + 1,
							Reason:   fmt.Sprintf("提取值 %s 不在枚举范围内 (原值: %s)", val, s),
						})
					}
				}
			} else {
				// 直接枚举检查
				if !enumValues[s] {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("值%s不在允许的枚举值范围内", s),
					})
				}
			}
		}
	}
	return res
}
