// Package calculation 提供计算类列校验规则（权重和、日期一致性等）
// 本包中的规则用于检查涉及多列计算逻辑的数据一致性

package calculation

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// WeightSumCheckRule 权重总和检查规则
type WeightSumCheckRule struct{}

func (c *WeightSumCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 获取目标总和，默认1.0
	targetSum := 1.0
	if sumStr, ok := params["targetSum"]; ok {
		if val, err := strconv.ParseFloat(sumStr, 64); err == nil {
			targetSum = val
		}
	}

	// 容差范围，默认0.0001
	tolerance := 0.0001
	if tolStr, ok := params[string(json_rule.TOLERANCE)]; ok {
		if val, err := strconv.ParseFloat(tolStr, 64); err == nil {
			tolerance = val
		}
	}

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 是否检查单行（默认false，检查整列）
	checkPerRow := false
	if perRow, ok := params["checkPerRow"]; ok {
		checkPerRow = strings.ToLower(perRow) == "true"
	}

	// 是否按分组检查（需要groupCol参数）
	groupColStr, hasGroup := params["groupCol"]
	var groupColIdx int
	var groupData []string
	if hasGroup {
		if idx, err := strconv.Atoi(groupColStr); err == nil {
			groupColIdx = idx
			groupData = cols[groupColIdx][startRowIdx:]
		}
	}

	if checkPerRow {
		// 逐行检查权重总和
		for i, s := range myColData {
			// 处理空值和注释
			if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
				continue
			}

			// 检查当前单元格是否为数值
			if val, err := strconv.ParseFloat(s, 64); err == nil {
				// 对于逐行检查，每个单元格就是单独的权重
				if math.Abs(val-targetSum) > tolerance {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("权重值%.4f不等于目标值%.4f", val, targetSum),
					})
				}
			} else if allowEmpty {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("权重值不是有效数值: %s", s),
				})
			}
		}
	} else if hasGroup {
		// 按分组检查权重总和
		groupWeights := make(map[string][]float64)
		groupRows := make(map[string][]int)

		// 收集每个分组的权重值
		for i := 0; i < len(myColData); i++ {
			s := myColData[i]
			group := ""
			if i < len(groupData) {
				group = strings.TrimSpace(groupData[i])
			}

			if group == "" && !allowEmpty {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   "单元格不能为空",
				})
				continue
			}

			if !allowEmpty && strings.TrimSpace(s) == "" {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   "单元格不能为空",
				})
				continue
			}

			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				if allowEmpty {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("权重值不是有效数值: %s", s),
					})
				}
				continue
			}

			groupWeights[group] = append(groupWeights[group], val)
			groupRows[group] = append(groupRows[group], i)
		}

		// 检查每个分组的权重总和
		for group, weights := range groupWeights {
			sum := 0.0
			for _, w := range weights {
				sum += w
			}

			if math.Abs(sum-targetSum) > tolerance {
				// 为分组中的每个行添加错误
				for _, rowIdx := range groupRows[group] {
					res = append(res, &json_rule.CellError{
						Index:  rowIdx,
						Reason: fmt.Sprintf("分组'%s'的权重总和%.4f不等于目标值%.4f", group, sum, targetSum),
					})
				}
			}
		}
	} else {
		// 检查整列的权重总和
		sum := 0.0
		validCount := 0

		for i, s := range myColData {
			if !allowEmpty && strings.TrimSpace(s) == "" {
				continue
			}

			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				if allowEmpty {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("权重值不是有效数值: %s", s),
					})
				}
				continue
			}

			sum += val
			validCount++
		}

		if validCount > 0 && math.Abs(sum-targetSum) > tolerance {
			// 为整个列添加错误
			res = append(res, &json_rule.CellError{
				Index:  0,
				Reason: fmt.Sprintf("权重总和%.4f不等于目标值%.4f", sum, targetSum),
			})
		}
	}

	return res
}
