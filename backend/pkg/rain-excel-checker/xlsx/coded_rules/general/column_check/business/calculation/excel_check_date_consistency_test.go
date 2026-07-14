// Package calculation 提供计算类列校验规则（权重和、日期一致性等）
// 本包中的规则用于检查涉及多列计算逻辑的数据一致性

package calculation

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestDateConsistencyCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataDateConsistencyCheckRule()

	bc := new(DateConsistencyCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataDateConsistencyCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params["endTimeCol"] = "1"
	params["descSheet"] = ""
	params["descColAttrName"] = "desc"
	params["reportNoTimeInDesc"] = "true"

	return [][]string{
		{"", "", "", "", "2025-01-01 12:30:55", "2025-05-05 12:30:55"},
		{"", "", "", "", "2025-01-01 12:35:55", "2025-07-01 12:30:55"},
		{"", "", "desc", "", "活动开始时间2025年1月1日12点30分55到2025年1月1日12点35分55秒结束",
			"活动开始时间2025年5月5日12点30分55到2025年7月1日12点30分55秒结束"},
	}, 0, 4, params, nil
}
