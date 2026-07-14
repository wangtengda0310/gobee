// Package numeric 提供枚举和数值范围的列级通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package numeric

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// NumericRangeCheckRule 数值范围检查
type NumericRangeCheckRule struct{}

func (c *NumericRangeCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	isArr := strings.HasSuffix(cols[colIdx][excelio.MJS_FIXED_ROWS_TYPE], "[]")
	res := make([]*json_rule.CellError, 0, len(myColData))

	minVal, _ := strconv.ParseFloat(params["minValue"], 64)
	maxVal, _ := strconv.ParseFloat(params["maxValue"], 64)

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	if isArr {
		for i, s := range myColData {
			// 处理空值和注释
			if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
				continue
			}
			// 处理数组分割
			arr := strings.Split(s, ",")
			for i2, s2 := range arr {

				val, err := strconv.ParseFloat(s2, 64)
				if err != nil {
					continue // 不是数值跳过
				}

				if val < minVal || val > maxVal {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("值%.2f[%d]不在范围[%.2f, %.2f]内", val, i2, minVal, maxVal),
					})
				}

			}
		}
	} else {
		for i, s := range myColData {
			// 处理空值和注释
			if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
				continue
			}

			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				continue // 不是数值跳过
			}

			if val < minVal || val > maxVal {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("值%.2f不在范围[%.2f, %.2f]内", val, minVal, maxVal),
				})
			}
		}
	}
	return res
}
