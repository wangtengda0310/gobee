// Package numeric 提供枚举和数值范围的列级通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package numeric

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// 基础枚举检查测试数据
func fakeDataEnumCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.ENUMS)] = "AB,345,123 34,56,,,23 ABNN, 123"
	return [][]string{
		{"AB", "ABNN"},
	}, 0, 0, params, nil
}

func TestEnumCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataEnumCheckRule()

	bc := new(EnumCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	assert.Empty(t, res, "AB和ABNN都在枚举范围内，应通过")
}

// 正则提取+枚举校验测试
func newRegexEnumCols() [][]string {
	return [][]string{
		{"", "", "Items", "", "{100;5}{200;10}", "{999;1}", "{50;3}", "", "{200;5}"},
	}
}

// TestEnumCheckRule_RegexExtract_Pass 测试正则提取后枚举校验通过
func TestEnumCheckRule_RegexExtract_Pass(t *testing.T) {
	cols := newRegexEnumCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums":   "100,200,50,999",
		"pattern": `\{(\d+);\d+\}`,
		"groups":  "1",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "提取值100,200,999,50都在枚举中，应通过")
}

// TestEnumCheckRule_RegexExtract_Fail 测试正则提取后枚举校验失败
func TestEnumCheckRule_RegexExtract_Fail(t *testing.T) {
	cols := newRegexEnumCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums":   "100,200,999",
		"pattern": `\{(\d+);\d+\}`,
		"groups":  "1",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res, "提取值50不在枚举中，应失败")
	assert.Contains(t, res[0].Reason, "50")
	assert.Contains(t, res[0].Reason, "枚举")
}

// TestEnumCheckRule_RegexExtract_NoMatch 测试正则未匹配
func TestEnumCheckRule_RegexExtract_NoMatch(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "no match here", "", ""},
	}
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums":   "100",
		"pattern": `X(\d+)X`,
		"groups":  "1",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "未匹配")
}

// TestEnumCheckRule_RegexExtract_AllowEmpty 测试正则模式下允许空值
func TestEnumCheckRule_RegexExtract_AllowEmpty(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "", "", ""},
	}
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums":      "100",
		"pattern":    `\d+`,
		"allowEmpty": "true",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "空值应跳过")
}

// TestEnumCheckRule_RegexExtract_InvalidPattern 测试无效正则
func TestEnumCheckRule_RegexExtract_InvalidPattern(t *testing.T) {
	cols := newRegexEnumCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums":   "100",
		"pattern": `[invalid`,
		"groups":  "1",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	// 正则编译失败时回退到直接枚举检查，原值不在枚举中
	assert.NotEmpty(t, res)
}

// TestEnumCheckRule_DirectEnum 测试无pattern参数时回退到直接枚举检查
func TestEnumCheckRule_DirectEnum(t *testing.T) {
	cols := [][]string{
		{"", "", "Type", "", "A", "B", "C", "D"},
	}
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums": "A,B,C",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res, "D不在枚举中，应失败")
	assert.Contains(t, res[0].Reason, "D")
}

// TestEnumCheckRule_RegexExtract_ArrayMode 测试数组类型+正则提取
func TestEnumCheckRule_RegexExtract_ArrayMode(t *testing.T) {
	cols := [][]string{
		{"", "", "Items[]", "", "{100;1},{200;2}", "{50;3},{999;4}"},
	}
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"enums":   "100,200,50,999",
		"pattern": `\{(\d+);\d+\}`,
		"groups":  "1",
	}
	rule := &EnumCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "数组模式下提取值都在枚举中，应通过")
}
