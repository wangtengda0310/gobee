package reference

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// RegexExtractRangeCheckRule 正则提取值范围校验规则
// 当前列用正则提取多个值后，验证这些值是否都在配置范围内
//
// 参数：
//   - pattern: 正则表达式
//   - groups: 捕获组索引（逗号分隔，默认 "1"）
//   - checkMode: 范围类型（enum/numeric），对应 ERuleParam.CHECK_MODE
//   - enums: 枚举值列表（checkMode=enum 时）
//   - min: 最小值（checkMode=numeric 时）
//   - max: 最大值（checkMode=numeric 时）
type RegexExtractRangeCheckRule struct{}

func (r *RegexExtractRangeCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseIntParam(params, string(json_rule.BREAK_LINE), 3)
	allowEmpty := helpers.ParseBool(params[string(json_rule.ALLOW_EMPTY)])
	allowCommit := helpers.ParseBool(params[string(json_rule.ALLOW_COMMIT)])

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	patternStr, ok := params[string(json_rule.PATTERN)]
	if !ok || patternStr == "" {
		return []*json_rule.CellError{{Index: 0, Reason: "必须提供正则表达式参数 'pattern'"}}
	}
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		return []*json_rule.CellError{{Index: 0, Reason: fmt.Sprintf("正则表达式编译错误: %v", err)}}
	}

	rangeType, ok := params[string(json_rule.CHECK_MODE)]
	if !ok || rangeType == "" {
		return []*json_rule.CellError{{Index: 0, Reason: "必须指定范围类型参数 'checkMode' (enum/numeric)"}}
	}

	var enumSet map[string]bool
	var minVal, maxVal float64
	hasMin, hasMax := false, false

	switch rangeType {
	case "enum":
		enumsStr := params[string(json_rule.ENUMS)]
		if enumsStr == "" {
			return []*json_rule.CellError{{Index: 0, Reason: "checkMode=enum 时必须提供枚举值参数 'enums'"}}
		}
		enumSet = make(map[string]bool)
		for _, v := range strings.Split(enumsStr, ",") {
			v = strings.TrimSpace(v)
			if v != "" {
				enumSet[v] = true
			}
		}
	case "numeric":
		if minStr, ok := params[string(json_rule.MIN)]; ok && minStr != "" {
			minVal, err = strconv.ParseFloat(minStr, 64)
			if err != nil {
				return []*json_rule.CellError{{Index: 0, Reason: fmt.Sprintf("min 参数不是有效数值: %s", minStr)}}
			}
			hasMin = true
		}
		if maxStr, ok := params[string(json_rule.MAX)]; ok && maxStr != "" {
			maxVal, err = strconv.ParseFloat(maxStr, 64)
			if err != nil {
				return []*json_rule.CellError{{Index: 0, Reason: fmt.Sprintf("max 参数不是有效数值: %s", maxStr)}}
			}
			hasMax = true
		}
		if !hasMin && !hasMax {
			return []*json_rule.CellError{{Index: 0, Reason: "checkMode=numeric 时必须提供 min 或 max 参数"}}
		}
	default:
		return []*json_rule.CellError{{Index: 0, Reason: fmt.Sprintf("不支持的 checkMode: %s (支持: enum, numeric)", rangeType)}}
	}

	groups := helpers.ParseCaptureGroups(params[string(json_rule.GROUPS)])

	for i, s := range myColData {
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

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

			switch rangeType {
			case "enum":
				if !enumSet[val] {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("值 %s 不在枚举范围内 (原值: %s)", val, s),
					})
				}
			case "numeric":
				numVal, err := strconv.ParseFloat(val, 64)
				if err != nil {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("提取值 %s 不是有效数值 (原值: %s)", val, s),
					})
					continue
				}
				if hasMin && numVal < minVal {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("值 %.0f 小于最小值 %.0f (原值: %s)", numVal, minVal, s),
					})
				}
				if hasMax && numVal > maxVal {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("值 %.0f 大于最大值 %.0f (原值: %s)", numVal, maxVal, s),
					})
				}
			}
		}
	}

	return res
}
