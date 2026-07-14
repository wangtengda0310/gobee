// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestNumericCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataNumericCheckRule()

	bc := new(NumericCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataNumericCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	return [][]string{
		{"true", "1"},
	}, 0, 0, nil, nil
}
