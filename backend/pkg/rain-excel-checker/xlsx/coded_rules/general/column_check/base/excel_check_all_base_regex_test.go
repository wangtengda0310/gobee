package base

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 正则格式1：无特定格式，整个单元格内容匹配
// ============================================================

// TestAllBase_RegexFormat1_Unique_Pass 正则格式1 + unique 通过
// 整个单元格匹配，值唯一
func TestAllBase_RegexFormat1_Unique_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "100", "200", "300", "400"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `^(\d+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取值100,200,300,400均唯一，应通过")
}

// TestAllBase_RegexFormat1_Unique_Fail 正则格式1 + unique 失败
// 整个单元格匹配，值重复
func TestAllBase_RegexFormat1_Unique_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "100", "200", "100", "400"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `^(\d+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取值100重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
	assert.Contains(t, res[0].Reason, "100")
}

// TestAllBase_RegexFormat1_ChsOnly_Pass 正则格式1 + chsOnly 通过
// 整个单元格匹配，仅中文
func TestAllBase_RegexFormat1_ChsOnly_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Names", "", "张三", "李四", "王五"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `^([\p{Han}]+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取值均为中文，应通过")
}

// TestAllBase_RegexFormat1_ChsOnly_Fail 正则格式1 + chsOnly 失败
// 整个单元格匹配，包含非中文（正则先匹配到中文字符串，然后chsOnly检查通过；
// 对于LiSi，正则不匹配，在fullMatch模式下报错"未匹配"）
// 此测试验证：当正则不匹配时，会报"未匹配"错误
func TestAllBase_RegexFormat1_ChsOnly_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Names", "", "张三", "LiSi", "王五"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `^([\p{Han}]+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "LiSi不匹配中文正则，应失败")
	assert.Contains(t, res[0].Reason, "未匹配")
}

// TestAllBase_RegexFormat1_Increase_Pass 正则格式1 + increase 通过
// 整个单元格匹配，值自增
func TestAllBase_RegexFormat1_Increase_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "1", "2", "3", "4"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `^(\d+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取值1,2,3,4自增，应通过")
}

// TestAllBase_RegexFormat1_Increase_Fail 正则格式1 + increase 失败
// 整个单元格匹配，值不自增
func TestAllBase_RegexFormat1_Increase_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "1", "3", "3", "4"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `^(\d+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取值3重复且不自增，应失败")
	assert.Contains(t, res[0].Reason, "自增")
}

// TestAllBase_RegexFormat1_NoMatch_Fail 正则格式1 + 未匹配到值
// 完全匹配模式下，正则未匹配到值应报错
func TestAllBase_RegexFormat1_NoMatch_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "abc", "def", "ghi"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `^(\d+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "正则未匹配到数值，应失败")
	assert.Contains(t, res[0].Reason, "未匹配")
}

// TestAllBase_RegexFormat1_InvalidPattern_Fail 正则格式1 + 无效正则
// 正则编译失败应返回错误
func TestAllBase_RegexFormat1_InvalidPattern_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "100", "200"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `[invalid`,
		"groups":     "1",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "无效正则应返回编译错误")
	assert.Contains(t, res[0].Reason, "编译错误")
}

// TestAllBase_RegexFormat1_AllowEmpty_Pass 正则格式1 + allowEmpty
// 空值应跳过
func TestAllBase_RegexFormat1_AllowEmpty_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "", "", ""},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `^(\d+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "true",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "空值应跳过")
}

// TestAllBase_RegexFormat1_NotFullMatch_Pass 正则格式1 + 非完全匹配
// 部分匹配模式下，从文本中提取数值
func TestAllBase_RegexFormat1_NotFullMatch_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Texts", "", "id_100", "id_200", "id_300"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\d+`,
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "部分匹配提取100,200,300均唯一，应通过")
}

// TestAllBase_RegexFormat1_Combined_Checks 正则格式1 + 多检查组合
// 同时启用 unique + chsOnly
func TestAllBase_RegexFormat1_Combined_Checks(t *testing.T) {
	cols := [][]string{
		{"", "", "Names", "", "张三", "李四", "张三"},
	}
	params := map[string]string{
		"unique":     "true",
		"chsOnly":    "true",
		"pattern":    `^([\p{Han}]+)$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "张三重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
}
