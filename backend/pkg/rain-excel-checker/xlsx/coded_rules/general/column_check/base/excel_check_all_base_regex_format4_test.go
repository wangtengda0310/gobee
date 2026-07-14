package base

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 正则格式4：匹配格式 1,A,鸡 捕获组: 所有逗号分隔的元素
// pattern: `\s*([^,\s]+)\s*` 捕获逗号分隔的元素
// ============================================================

// TestAllBase_RegexFormat4_Unique_Pass 正则格式4 + unique 通过
// 从逗号分隔字符串中提取元素，所有元素唯一
func TestAllBase_RegexFormat4_Unique_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "100,200,300", "400,500,600"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取值100,200,300,400,500,600均唯一，应通过")
}

// TestAllBase_RegexFormat4_Unique_Fail 正则格式4 + unique 失败
// 从逗号分隔字符串中提取元素，有重复值
func TestAllBase_RegexFormat4_Unique_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "100,200,300", "400,200,600"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取值200重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
	assert.Contains(t, res[0].Reason, "200")
}

// TestAllBase_RegexFormat4_ChsOnly_Pass 正则格式4 + chsOnly 通过
// 从逗号分隔字符串中提取中文元素
func TestAllBase_RegexFormat4_ChsOnly_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Names", "", "张三,李四", "王五,赵六"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取中文值均通过chsOnly检查")
}

// TestAllBase_RegexFormat4_ChsOnly_Fail 正则格式4 + chsOnly 失败
// 从逗号分隔字符串中提取元素，包含非中文
func TestAllBase_RegexFormat4_ChsOnly_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Names", "", "张三,李四", "王五,LiSi"},
	}
	params := map[string]string{
		"chsOnly":    "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "LiSi非中文，应失败")
	assert.Contains(t, res[0].Reason, "非中文")
}

// TestAllBase_RegexFormat4_Increase_Pass 正则格式4 + increase 通过
// 从逗号分隔字符串中提取数值，值自增
func TestAllBase_RegexFormat4_Increase_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "1", "2", "3", "4"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取值1,2,3,4自增，应通过")
}

// TestAllBase_RegexFormat4_Increase_Fail 正则格式4 + increase 失败
// 从逗号分隔字符串中提取数值，值不自增
func TestAllBase_RegexFormat4_Increase_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Ids", "", "1", "3", "3", "4"},
	}
	params := map[string]string{
		"increase":   "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取值3重复且不自增，应失败")
	assert.Contains(t, res[0].Reason, "自增")
}

// TestAllBase_RegexFormat4_MultipleValuesInCell_Pass 正则格式4 + 多值匹配
// 单个单元格中有多个逗号分隔值，提取所有值检查唯一性
func TestAllBase_RegexFormat4_MultipleValuesInCell_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "100,200,300", "400,500,600"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "提取所有值均唯一，应通过")
}

// TestAllBase_RegexFormat4_MultipleValuesInCell_Fail 正则格式4 + 多值匹配失败
// 单个单元格中有多个逗号分隔值，提取的值有重复
func TestAllBase_RegexFormat4_MultipleValuesInCell_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "100,200,300", "400,200,600"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "提取值200重复，应失败")
	assert.Contains(t, res[0].Reason, "重复")
}

// TestAllBase_RegexFormat4_NoMatch_Fail 正则格式4 + 未匹配到值
// 单元格内容无法匹配（使用非空值但正则不匹配）
func TestAllBase_RegexFormat4_NoMatch_Fail(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "!!!", "@@@", "###"},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `^\d+$`,
		"groups":     "1",
		"fullMatch":  "true",
		"allowEmpty": "false",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.NotEmpty(t, res, "非空值不匹配正则，应失败")
	assert.Contains(t, res[0].Reason, "未匹配")
}

// TestAllBase_RegexFormat4_EmptyCell_Pass 正则格式4 + 空单元格
// 空单元格应跳过
func TestAllBase_RegexFormat4_EmptyCell_Pass(t *testing.T) {
	cols := [][]string{
		{"", "", "Items", "", "", "", ""},
	}
	params := map[string]string{
		"unique":     "true",
		"pattern":    `\s*([^,\s]+)\s*`,
		"groups":     "1",
		"fullMatch":  "false",
		"allowEmpty": "true",
	}
	rule := &AllBaseCheckRule{}
	res := rule.Check("test", cols, 0, excelio.MJS_FIXED_ROWS_NUM, params, nil)
	assert.Empty(t, res, "空值应跳过")
}
