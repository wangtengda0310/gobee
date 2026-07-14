// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestUniqueCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataUniqueCheckRule()

	bc := new(UniqueCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataUniqueCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	return [][]string{
		{"true", "1", "456", "1"},
	}, 0, 0, nil, nil
}
