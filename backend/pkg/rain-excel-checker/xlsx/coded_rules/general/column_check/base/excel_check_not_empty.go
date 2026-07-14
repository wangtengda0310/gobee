// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// NotEmptyCheckRule 不为空检查
type NotEmptyCheckRule struct{}

func (c *NotEmptyCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := 3
	if break_, ok := params[string(json_rule.BREAK_LINE)]; ok {
		if maxSpace, err := strconv.Atoi(break_); err == nil && maxSpace >= 1 {
			breakLine = maxSpace
		}
	}

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许注释
	allowCommit := false
	if allow, ok := params[string(json_rule.ALLOW_COMMIT)]; ok {
		allowCommit = strings.ToLower(allow) == "true"
	}

	for i, s := range myColData {
		// 使用公共函数统一处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, false, allowCommit) {
			continue
		}

		// 如果走到这里，说明单元格非空且不是注释
		// NOT_EMPTY规则不需要额外检查，因为SolveEmptyAndCommit已经处理了空值情况
		// 但为了保持与原有行为一致，如果单元格为空（在allowCommit=true时注释行会被跳过）
		// 这里实际上不会执行到，因为空值已经在SolveEmptyAndCommit中处理
		_ = s
	}
	return res
}
