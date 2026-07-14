// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// SpecialFormatCheckRule 道具ID+数量格式检查
type SpecialFormatCheckRule struct{}

func (c *SpecialFormatCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 是否为数组模式（逗号分隔的多组 {id;count}）
	isArray := false
	if isArrayStr, ok := params["isArray"]; ok {
		isArray = strings.ToLower(isArrayStr) == "true"
	}

	// 单值模式正则：{1000005;1}{1000011;10}{9000001;5}
	singlePattern := regexp.MustCompile(`^\{(?:\d+;\d+)(?:}\{(?:\d+;\d+))*}$`)

	// 数组模式正则：{2000001;1},{2000001;1},{2000001;1}
	// 即逗号分隔的多个 {id;count} 组
	arrayPattern := regexp.MustCompile(`^\{\d+;\d+\}(?:,\{\d+;\d+\})*$`)

	// 提取单个 {id;count} 的正则
	itemPattern := regexp.MustCompile(`\{(\d+);(\d+)\}`)

	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		var pattern *regexp.Regexp
		if isArray {
			pattern = arrayPattern
		} else {
			pattern = singlePattern
		}

		if !pattern.MatchString(s) {
			if isArray {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("格式不正确，应为{itemId;count},{itemId;count}...数组格式，当前: %s", s),
				})
			} else {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("格式不正确，应为{itemId;count}格式，当前: %s", s),
				})
			}
			continue
		}

		// 进一步验证每个道具项
		matches := itemPattern.FindAllStringSubmatch(s, -1)

		for _, match := range matches {
			itemID := match[1]
			count := match[2]

			// 检查ID是否为数值
			if _, err := strconv.ParseInt(itemID, 10, 64); err != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("道具ID不是有效数值: %s", itemID),
				})
			}

			// 检查数量是否为数值
			if _, err := strconv.ParseInt(count, 10, 64); err != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("道具数量不是有效数值: %s", count),
				})
			}
		}
	}
	return res
}
