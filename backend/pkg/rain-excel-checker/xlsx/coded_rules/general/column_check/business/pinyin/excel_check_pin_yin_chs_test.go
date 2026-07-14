// Package pinyin 提供拼音相关的列级校验规则
// 本包中的规则用于检查拼音与中文的对应关系

package pinyin

import (
	"encoding/json"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

func TestPinYinCHSCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataPinYinCHSCheckRule()

	bc := new(PinYinCHSCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataPinYinCHSCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.ALLOW_EMPTY)] = "false"
	params[string(json_rule.ALLOW_COMMIT)] = "false"
	params["chsColOffset"] = "1"
	return [][]string{
		{"NiHao", "BuHa_", "WanLe"},
		{"你好", "布豪", "丸了"},
	}, 0, 0, params, nil
}
