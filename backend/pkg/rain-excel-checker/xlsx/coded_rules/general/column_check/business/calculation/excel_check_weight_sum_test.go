// Package calculation 提供计算类列校验规则（权重和、日期一致性等）
// 本包中的规则用于检查涉及多列计算逻辑的数据一致性

package calculation

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestWeightSumCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataWeightSumCheckRule()

	bc := new(WeightSumCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataWeightSumCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	return [][]string{
		{"0.5", "0.3", "0.3", "0"},
	}, 0, 0, nil, nil
}
