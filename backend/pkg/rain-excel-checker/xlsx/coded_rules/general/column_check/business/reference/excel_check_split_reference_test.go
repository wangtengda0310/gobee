// Package reference 提供引用关系相关的校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package reference

import (
	"encoding/json"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestSplitReferenceCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataSplitReferenceCheckRule()

	bc := new(SplitReferenceCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataSplitReferenceCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	return [][]string{
		{"true", "1"},
	}, 0, 0, nil, nil
}
