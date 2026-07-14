// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestDateCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataDateCheckRule()

	bc := new(DateCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataDateCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params["format"] = "yyyy-MM-dd HH:mm:ss"
	return [][]string{
		{"2006-01-02 12:00:00", "1"},
	}, 0, 0, params, nil
}
