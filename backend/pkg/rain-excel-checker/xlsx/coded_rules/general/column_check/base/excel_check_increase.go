// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"fmt"
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

type IncreaseCheckRule struct {
}

func (c *IncreaseCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 第一个应该为
	for i, s := range myColData {
		tempRes := &json_rule.CellError{
			Index:    i,
			ExcelRow: startRowIdx + i + 1,
		}

		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		if id, err := strconv.ParseInt(s, 10, 64); err == nil {
			if id != int64(i)+1 {
				tempRes.Reason = fmt.Sprintf("ID自增有问题, 当前ID：%d(%s), 应为ID: %d", id, s, i+1)
			}
		} else {
			tempRes.Reason = fmt.Sprintf("ID类型有问题, 当前id: %s", s)
		}
		if tempRes.Reason != "" {
			res = append(res, tempRes)
		}
	}
	return res
}
