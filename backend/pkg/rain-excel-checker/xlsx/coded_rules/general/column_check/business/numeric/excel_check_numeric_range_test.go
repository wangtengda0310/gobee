// Package numeric 提供枚举和数值范围的列级通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package numeric

import (
	"encoding/json"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

func TestNumericRangeCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataNumericRangeCheckRule()

	bc := new(NumericRangeCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataNumericRangeCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.MIN)] = "0"
	params[string(json_rule.MAX)] = "1"
	return [][]string{
		{"0.5", "1", "2"},
	}, 0, 0, params, nil
}

// TestNumericRangeCheckRule_OldKeyCompatibility 测试旧版键名 minValue/maxValue 兼容性
func TestNumericRangeCheckRule_OldKeyCompatibility(t *testing.T) {
	// 使用旧版键名（前端曾用 minValue/maxValue）
	params := map[string]string{
		"minValue": "2",
		"maxValue": "4",
	}
	cols := [][]string{
		{"1", "2", "3", "4", "5", "8"},
	}

	bc := new(NumericRangeCheckRule)
	res := bc.Check("", cols, 0, 0, params, nil)

	// 期望 1、5 和 8 超出范围 [2,4]
	if len(res) != 3 {
		t.Errorf("期望发现 3 个错误，实际发现 %d 个", len(res))
	}

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}
