package base

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 正则格式3：匹配格式 {123;234} 捕获{}内;分隔后第二个值
// pattern: `{\d+;(\d+)}` 捕获第二个数字
// ============================================================

// TestAllBase_RegexFormat3_Unique_Pass 正则格式3 + unique 通过
// 从 {id;count} 中提取 count，值唯一
func TestAllBase_RegexFormat3_Unique_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}", "{200;10}", "{300;15}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取count值5,10,15均唯一，应通过")
}

// TestAllBase_RegexFormat3_Unique_Fail 正则格式3 + unique 失败
// 从 {id;count} 中提取 count，值重复
func TestAllBase_RegexFormat3_Unique_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}", "{200;10}", "{300;10}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取count值10重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
	assert.Contains(t, res[0].Reason, "10")
}

// TestAllBase_RegexFormat3_ChsOnly_Pass 正则格式3 + chsOnly 通过
// 从 {id;name} 中提取 name（中文），值均为中文
func TestAllBase_RegexFormat3_ChsOnly_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;张三}", "{200;李四}", "{300;王五}"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `\{\d+;([\p{Han}]+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取中文值均通过chsOnly检查")
}

// TestAllBase_RegexFormat3_ChsOnly_Fail 正则格式3 + chsOnly 失败
// 从 {id;name} 中提取 name，包含非中文
func TestAllBase_RegexFormat3_ChsOnly_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;张三}", "{200;LiSi}", "{300;王五}"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `\{\d+;([^}]+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "LiSi非中文，应失败")
	assert.Contains(t, res[0].Reason, "非中文")
}

// TestAllBase_RegexFormat3_Increase_Pass 正则格式3 + increase 通过
// 从 {id;count} 中提取 count，值自增
func TestAllBase_RegexFormat3_Increase_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;1}", "{200;2}", "{300;3}", "{400;4}"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取count值1,2,3,4自增，应通过")
}

// TestAllBase_RegexFormat3_Increase_Fail 正则格式3 + increase 失败
// 从 {id;count} 中提取 count，值不自增
func TestAllBase_RegexFormat3_Increase_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;1}", "{200;3}", "{300;3}", "{400;4}"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取count值3重复且不自增，应失败")
	assert.Contains(t, res[0].Reason, "自增")
}

// TestAllBase_RegexFormat3_MultipleMatches_Pass 正则格式3 + 多值匹配
// 单元格中有多个 {id;count}，提取所有count检查唯一性
func TestAllBase_RegexFormat3_MultipleMatches_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}{200;10}", "{300;15}{400;20}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取所有count值5,10,15,20均唯一，应通过")
}

// TestAllBase_RegexFormat3_MultipleMatches_Fail 正则格式3 + 多值匹配失败
// 单元格中有多个 {id;count}，提取的count有重复
func TestAllBase_RegexFormat3_MultipleMatches_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "{100;5}{200;10}", "{300;10}{400;20}"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取count值10重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
}

// TestAllBase_RegexFormat3_NoMatch_Fail 正则格式3 + 未匹配到值
// 单元格内容不符合 {id;count} 格式
func TestAllBase_RegexFormat3_NoMatch_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "abc", "def", "ghi"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "正则不匹配abc等值，应失败")
	assert.Contains(t, res[0].Reason, "未匹配")
}

// TestAllBase_RegexFormat3_EmptyCell_Pass 正则格式3 + 空单元格
// 空单元格应跳过
func TestAllBase_RegexFormat3_EmptyCell_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "", "", ""},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\{\d+;(\d+)\}`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "true",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "空值应跳过")
}
