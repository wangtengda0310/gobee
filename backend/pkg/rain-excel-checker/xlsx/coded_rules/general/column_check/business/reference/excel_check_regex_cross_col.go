package reference

import (
	"fmt"
	"regexp"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// RegexCrossColCheckRule 正则提取跨列校验规则
// 同一行中，当前列用正则提取文本后，与同表另一列按可扩展匹配规则比较
//
// 参数：
//   - pattern: 正则表达式（提取当前列内容）
//   - groups: 捕获组索引（逗号分隔，默认 "1"）
//   - targetCol: 同表目标列名
//   - matchType: 匹配规则（exists/equals/pinyin/pinyin_contains）
//   - pinyinFormat: 拼音输出格式（camel/lower/snake，默认 camel）
type RegexCrossColCheckRule struct{}

// validateCrossCol 核心校验逻辑，独立于 cols 结构，便于未来迁移到 rowRule
func validateCrossCol(extracted []string, targetValue string, allTargetValues map[string]bool, matchType string, pinyinFormat helpers.PinyinFormat) (bool, string) {
	for _, val := range extracted {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		switch matchType {
		case "exists":
			if !allTargetValues[val] {
				return false, fmt.Sprintf("值 %s 未通过 exists 检查（目标列中不存在）", val)
			}
		case "equals":
			if val != targetValue {
				return false, fmt.Sprintf("值 %s 与同行目标列值 %s 不相等 (equals)", val, targetValue)
			}
		case "pinyin":
			variants := helpers.GetPinyinVariants(targetValue, pinyinFormat)
			matched := false
			// 统一去掉下划线后比较，支持 snake_case 和 camelCase 混合匹配
			valNormalized := strings.ReplaceAll(strings.ToLower(val), "_", "")
			for _, expected := range variants {
				expectedNormalized := strings.ReplaceAll(strings.ToLower(expected), "_", "")
				if valNormalized == expectedNormalized {
					matched = true
					break
				}
			}
			if !matched {
				return false, fmt.Sprintf("值 %s 与目标列拼音 %v (原值:%s) 均不相等 (pinyin)", val, variants, targetValue)
			}
		case "pinyin_contains":
			variants := helpers.GetPinyinVariants(targetValue, pinyinFormat)
			matched := false
			// 统一去掉下划线后比较，支持 snake_case 和 camelCase 混合匹配
			valNormalized := strings.ReplaceAll(strings.ToLower(val), "_", "")
			for _, expected := range variants {
				expectedNormalized := strings.ReplaceAll(strings.ToLower(expected), "_", "")
				if strings.Contains(valNormalized, expectedNormalized) {
					matched = true
					break
				}
			}
			if !matched {
				return false, fmt.Sprintf("值 %s 不包含目标列拼音 %v (原值:%s) (pinyin_contains)", val, variants, targetValue)
			}
		default:
			return false, fmt.Sprintf("不支持的匹配类型: %s", matchType)
		}
	}
	return true, ""
}

func (r *RegexCrossColCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
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

	targetColName, ok := params[string(json_rule.TARGET_COL)]
	if !ok || targetColName == "" {
		return []*json_rule.CellError{{Index: 0, Reason: "必须指定目标列参数 'targetCol'"}}
	}

	targetColIdx := excelio.GetColIndexByName(cols, targetColName)
	if targetColIdx == -1 {
		var available []string
		for _, col := range cols {
			if len(col) > excelio.MJS_FIXED_ROWS_NAME {
				available = append(available, col[excelio.MJS_FIXED_ROWS_NAME])
			}
		}
		return []*json_rule.CellError{{
			Index:  0,
			Reason: fmt.Sprintf("未找到目标列: %s，可用列: %v", targetColName, available),
		}}
	}

	groups := helpers.ParseCaptureGroups(params[string(json_rule.GROUPS)])

	matchType := params[string(json_rule.CHECK_MODE)]
	if matchType == "" {
		return []*json_rule.CellError{{Index: 0, Reason: "必须指定匹配类型参数 'matchType' (exists/equals/pinyin/pinyin_contains)"}}
	}

	// 解析拼音格式参数（仅 pinyin/pinyin_contains 时生效）
	pinyinFormat := helpers.PinyinCamel
	if pf, ok := params["pinyinFormat"]; ok {
		switch pf {
		case "lower":
			pinyinFormat = helpers.PinyinLower
		case "snake":
			pinyinFormat = helpers.PinyinSnake
		default:
			pinyinFormat = helpers.PinyinCamel
		}
	}

	allTargetValues := make(map[string]bool)
	if matchType == "exists" {
		targetEndIdx := helpers.GetColEndIndex(cols, targetColIdx, startRowIdx, breakLine, params)
		targetData := cols[targetColIdx][startRowIdx:targetEndIdx]
		for _, v := range targetData {
			if v != "" {
				allTargetValues[v] = true
			}
		}
	}

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

		targetRowIdx := i
		var targetValue string
		targetColData := cols[targetColIdx]
		if startRowIdx+targetRowIdx < len(targetColData) {
			targetValue = targetColData[startRowIdx+targetRowIdx]
		}

		ok, reason := validateCrossCol(extracted, targetValue, allTargetValues, matchType, pinyinFormat)
		if !ok {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   reason,
			})
		}
	}

	return res
}
