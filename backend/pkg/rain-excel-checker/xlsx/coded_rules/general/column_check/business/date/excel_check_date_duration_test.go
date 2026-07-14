// Package date 提供日期相关的列级校验规则
// 本包中的规则用于检查日期持续时间、日期范围等

package date

import (
	"encoding/json"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

func TestDateDurationCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataDateDurationCheckRule()

	bc := new(DateDurationCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataDateDurationCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params["endColOffset"] = "1"
	params["duration"] = "3h"
	params["displayUnit"] = "m"
	params[string(json_rule.TOLERANCE)] = "1m"

	return [][]string{
		{"2006-01-02 12:00:00", "2006-01-02 16:00:00"},
		{"2006-01-02 15:00:00", "2006-01-02 19:02:00"},
	}, 0, 0, params, nil
}
