package reference

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
)

func newBaseCols() [][]string {
	return [][]string{
		{"", "", "Items", "", "{100;5}{200;10}", "{999;1}", "{50;3}", "", "{200;5}"},
	}
}

func TestRegexExtractRange_Enum_Pass(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "enum",
		"enums":     "100,200,50,999",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "提取值100,200,50,999都在枚举中，应通过")
}

func TestRegexExtractRange_Enum_Fail(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "enum",
		"enums":     "100,200,999",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res, "提取值50不在枚举中，应失败")
	assert.Contains(t, res[0].Reason, "枚举")
	assert.Contains(t, res[0].Reason, "50")
}

func TestRegexExtractRange_Numeric_Pass(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "numeric",
		"min":       "1",
		"max":       "1000",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "所有提取值在1-1000范围内，应通过")
}

func TestRegexExtractRange_Numeric_BelowMin(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "numeric",
		"min":       "100",
		"max":       "1000",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "50")
}

func TestRegexExtractRange_Numeric_AboveMax(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "numeric",
		"min":       "1",
		"max":       "100",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	// 第一行包含 200，第二行包含 999，都超过最大值 100
	// 可能是 200 或 999 先被检测到
	assert.True(t,
		len(res) > 0 && (res[0].Reason == "值 200 大于最大值 100 (原值: {100;5}{200;10})" ||
			res[0].Reason == "值 999 大于最大值 100 (原值: {999;1})"),
		"期望错误信息包含 200 或 999，实际: %s", res[0].Reason)
}

func TestRegexExtractRange_MissingPattern(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"checkMode": "enum",
		"enums":     "100",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "pattern")
}

func TestRegexExtractRange_MissingRangeType(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern": `\d+`,
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "checkMode")
}

func TestRegexExtractRange_NoMatch(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "no match here", "", ""},
	}
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `X(\d+)X`,
		"groups":    "1",
		"checkMode": "enum",
		"enums":     "100",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "未匹配")
}

func TestRegexExtractRange_AllowEmpty(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "", "", ""},
	}
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":    `\d+`,
		"checkMode":  "enum",
		"enums":      "100",
		"allowEmpty": "true",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "空值应跳过")
}

func TestRegexExtractRange_Enum_WithoutEnumsParam(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "enum",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "enums")
}

func TestRegexExtractRange_Numeric_WithoutMinOrMax(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"checkMode": "numeric",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "min 或 max")
}

func TestRegexExtractRange_InvalidRangeType(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `\d+`,
		"groups":    "1",
		"checkMode": "invalid",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "不支持的 checkMode")
}

func TestRegexExtractRange_InvalidPattern(t *testing.T) {
	cols := newBaseCols()
	colIdx := 0
	startRowIdx := excelio.MJS_FIXED_ROWS_NUM
	params := map[string]string{
		"pattern":   `[invalid`,
		"groups":    "1",
		"checkMode": "enum",
		"enums":     "100",
	}
	rule := &RegexExtractRangeCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "正则表达式编译错误")
}
