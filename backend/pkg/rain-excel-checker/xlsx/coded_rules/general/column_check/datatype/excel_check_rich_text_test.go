// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"encoding/json"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

func TestRichTextCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataRichTextCheckRule()

	bc := new(RichTextCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataRichTextCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.ALLOW_EMPTY)] = "false"
	params[string(json_rule.ALLOW_COMMIT)] = "false"
	return [][]string{
		{"你好F", "{0}<color=#FFFFFF>{1}</color>", "{0}<color=#FFFFFF>{2}{1}</color>", "{0}<color=#FFFFFF>{1}</colo>", "{0}<color=#FFFFFF>{1}<color>", "{0}<color=#FFFFF>{1}</color>", "{0}<color='#FFFFFF'>{1}</color>"},
	}, 0, 0, params, nil
}
