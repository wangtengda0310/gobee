// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"encoding/json"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

func TestAllBaseCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataAllBaseCheckRule()

	bc := new(AllBaseCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataAllBaseCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.ALLOW_EMPTY)] = "false"
	params[string(json_rule.ALLOW_COMMIT)] = "true"
	params["unique"] = "true"
	params["chsOnly"] = "false"
	params["increase"] = "false"
	return [][]string{
		{"true", "", "1"},
		{"true", "", "1"},
		{"true", "", "1"},
	}, 1, 0, params, nil
}
