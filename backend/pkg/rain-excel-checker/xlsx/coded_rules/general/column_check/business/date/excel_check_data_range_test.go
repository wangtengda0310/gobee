// Package date 提供日期相关的列级校验规则
// 本包中的规则用于检查日期持续时间、日期范围等

package date

import (
	"encoding/json"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

func TestDateRangeCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataDateRangeCheckRule()

	bc := new(DateRangeCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataDateRangeCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.START_DATE)] = "2025-10-27 00:00:00"
	params[string(json_rule.END_DATE)] = "2025-10-28 00:00:00"
	params["includeBoundary"] = "true"
	return [][]string{
		{"2025-10-27 00:00:00", "2025-10-27 05:00:00", "2025-10-28 01:00:00"},
	}, 0, 0, params, nil
}
