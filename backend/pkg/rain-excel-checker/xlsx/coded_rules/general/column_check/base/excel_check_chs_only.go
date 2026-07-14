// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"fmt"
	"regexp"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

type CHSCheckRule struct {
}

func (c *CHSCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))
	// 更全面的中文字符范围
	pattern := `^[\p{Han}\p{P}\s]+$`

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 第一个应该为
	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		tempRes := &json_rule.CellError{
			Index:    i,
			ExcelRow: startRowIdx + i + 1,
		}
		if matched, _ := regexp.MatchString(pattern, s); !matched {
			tempRes.Reason = fmt.Sprintf("包含非中文字符: %s", s)
		}
		if tempRes.Reason != "" {
			res = append(res, tempRes)
		}
	}
	return res
}
