package base

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 正则格式2：匹配格式 {123;234} 捕获{}内;分隔后第一个值
// pattern: `{\\d+;(\\d+)}` 不对，应该是 `{\\(\\d+\\);\\d+\\}`
// 实际是：`{(\\d+);\\d+}` 捕获第一个数字
// ============================================================

// TestAllBase_RegexFormat2_Unique_Pass 正则格式2 + unique 通过
// 从 {id;count} 中提取 id，值唯一
func TestAllBase_RegexFormat2_Unique_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}", "{200;10}", "{300;15}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取id值100,200,300均唯一，应通过")
}

// TestAllBase_RegexFormat2_Unique_Fail 正则格式2 + unique 失败
// 从 {id;count} 中提取 id，值重复
func TestAllBase_RegexFormat2_Unique_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}", "{200;10}", "{100;15}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取id值100重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
	assert.Contains(t, res[0].Reason, "100")
}

// TestAllBase_RegexFormat2_ChsOnly_Pass 正则格式2 + chsOnly 通过
// 从 {name;count} 中提取 name（中文），值均为中文
func TestAllBase_RegexFormat2_ChsOnly_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{张三;5}", "{李四;10}", "{王五;15}"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `\{([\p{Han}]+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取中文值均通过chsOnly检查")
}

// TestAllBase_RegexFormat2_ChsOnly_Fail 正则格式2 + chsOnly 失败
// 从 {name;count} 中提取 name，包含非中文
// 注意：当正则是 `\{([\p{Han}]+);\d+\}` 时，{LiSi;10} 不匹配，因为LiSi不是中文
// 但当前实现中，如果正则不匹配且fullMatch=false，不会报错"未匹配"
// 而是跳过该值。所以此测试需要调整：使用能匹配到非中文的正则
func TestAllBase_RegexFormat2_ChsOnly_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{张三;5}", "{LiSi;10}", "{王五;15}"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `\{([^;]+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "LiSi非中文，应失败")
	assert.Contains(t, res[0].Reason, "非中文")
}

// TestAllBase_RegexFormat2_Increase_Pass 正则格式2 + increase 通过
// 从 {id;count} 中提取 id，值自增
func TestAllBase_RegexFormat2_Increase_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{1;100}", "{2;200}", "{3;300}", "{4;400}"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取id值1,2,3,4自增，应通过")
}

// TestAllBase_RegexFormat2_Increase_Fail 正则格式2 + increase 失败
// 从 {id;count} 中提取 id，值不自增
func TestAllBase_RegexFormat2_Increase_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{1;100}", "{3;200}", "{3;300}", "{4;400}"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取id值3重复且不自增，应失败")
	assert.Contains(t, res[0].Reason, "自增")
}

// TestAllBase_RegexFormat2_MultipleMatches_Pass 正则格式2 + 多值匹配
// 单元格中有多个 {id;count}，提取所有id检查唯一性
func TestAllBase_RegexFormat2_MultipleMatches_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}{200;10}", "{300;15}{400;20}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取所有id值100,200,300,400均唯一，应通过")
}

// TestAllBase_RegexFormat2_MultipleMatches_Fail 正则格式2 + 多值匹配失败
// 单元格中有多个 {id;count}，提取的id有重复
func TestAllBase_RegexFormat2_MultipleMatches_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}{200;10}", "{200;15}{300;20}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取id值200重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
}

// TestAllBase_RegexFormat2_NoMatch_Fail 正则格式2 + 未匹配到值
// 单元格内容不符合 {id;count} 格式
func TestAllBase_RegexFormat2_NoMatch_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "abc", "def", "ghi"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "正则不匹配abc等值，应失败")
	assert.Contains(t, res[0].Reason, "未匹配")
}

// TestAllBase_RegexFormat2_EmptyCell_Pass 正则格式2 + 空单元格
// 空单元格应跳过
func TestAllBase_RegexFormat2_EmptyCell_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "", "", ""},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{(\d+);\d+\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "true",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "空值应跳过")
}
