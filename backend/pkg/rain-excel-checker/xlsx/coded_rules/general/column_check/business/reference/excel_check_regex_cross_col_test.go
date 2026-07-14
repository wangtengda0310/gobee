package reference

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/stretchr/testify/assert"
)

func newCrossColTestData() (cols [][]string, colIdx int, startRowIdx int, params map[string]string) {
	startRowIdx = excelio.MJS_FIXED_ROWS_NUM
	colIdx = 0
	params = map[string]string{
		"pattern":   `\{(\d+);\d+\}`,
		"groups":    "1",
		"targetCol": "Name",
		"checkMode": "exists",
	}
	cols = [][]string{
		{"", "", "HeroRef", "", "{100;5}", "{200;1}", ""},
		{"", "", "Name", "", "100", "200", ""},
	}
	return
}

func TestRegexCrossCol_Exists_Pass(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "exists"
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	if len(res) > 0 {
		t.Logf("意外的错误: %+v", res)
		for _, r := range res {
			t.Logf("  - Index: %d, Reason: %s", r.Index, r.Reason)
		}
	}
	assert.Empty(t, res, "提取值100在目标列中存在，应通过")
}

func TestRegexCrossCol_Exists_Fail(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "exists"
	cols[0][5] = "{999;1}"
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res, "提取值999不在目标列中，应失败")
	assert.Contains(t, res[0].Reason, "exists")
}

func TestRegexCrossCol_Equals_Pass(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "equals"
	params["pattern"] = `(\w+)`
	params["groups"] = "1"
	// 只保留第一行数据
	cols[0] = []string{"", "", "HeroRef", "", "100", ""}
	cols[1] = []string{"", "", "Name", "", "100", ""}
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	if len(res) > 0 {
		t.Logf("意外的错误: %+v", res)
		for _, r := range res {
			t.Logf("  - Index: %d, Reason: %s", r.Index, r.Reason)
		}
	}
	assert.Empty(t, res, "提取值与同行目标列值相等，应通过")
}

func TestRegexCrossCol_Equals_Fail(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "equals"
	params["pattern"] = `(\w+)`
	params["groups"] = "1"
	cols[0][4] = "100"
	cols[1][4] = "200"
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "equals")
}

func TestRegexCrossCol_Pinyin_Pass(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "pinyin"
	params["pattern"] = `(\w+)`
	params["groups"] = "1"
	// 只保留第一行数据，清除第二行
	cols[0] = []string{"", "", "HeroRef", "", "ZhangFei", ""}
	cols[1] = []string{"", "", "Name", "", "张飞", ""}
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	if len(res) > 0 {
		t.Logf("意外的错误: %+v", res)
		for _, r := range res {
			t.Logf("  - Index: %d, Reason: %s", r.Index, r.Reason)
		}
	}
	assert.Empty(t, res, "提取值ZhangFei等于张飞转拼音结果，应通过")
}

func TestRegexCrossCol_Pinyin_Fail(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "pinyin"
	params["pattern"] = `(\w+)`
	params["groups"] = "1"
	cols[0][4] = "WrongValue"
	cols[1][4] = "张飞"
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
}

func TestRegexCrossCol_PinyinContains_Pass(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "pinyin_contains"
	params["pattern"] = `(\w+_\w+_\d+)`
	params["groups"] = "1"
	// 只保留第一行数据
	cols[0] = []string{"", "", "HeroRef", "", "hero_ZhangFei_01", ""}
	cols[1] = []string{"", "", "Name", "", "张飞", ""}
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	if len(res) > 0 {
		t.Logf("意外的错误: %+v", res)
		for _, r := range res {
			t.Logf("  - Index: %d, Reason: %s", r.Index, r.Reason)
		}
	}
	assert.Empty(t, res, "提取值包含张飞转拼音结果，应通过")
}

func TestRegexCrossCol_MissingPattern(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	delete(params, "pattern")
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "pattern")
}

func TestRegexCrossCol_MissingTargetCol(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	delete(params, "targetCol")
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "targetCol")
}

func TestRegexCrossCol_TargetColNotFound(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["targetCol"] = "NonExist"
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.NotEmpty(t, res)
	assert.Contains(t, res[0].Reason, "未找到目标列")
}

func TestRegexCrossCol_AllowEmpty(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["allowEmpty"] = "true"
	cols[0][4] = ""
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "空值应跳过")
}

func TestRegexCrossCol_AllowCommit(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["allowCommit"] = "true"
	cols[0][4] = "# 这是注释"
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	assert.Empty(t, res, "注释值应跳过")
}

// TestRegexCrossCol_Polyphone 多音字测试："长驱烈风" 的 "长" 可读 zhang 或 chang
func TestRegexCrossCol_Polyphone_Pass(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "pinyin_contains"
	params["pattern"] = `(\w+)`
	params["groups"] = "1"
	// 只保留第一行数据
	cols[0] = []string{"", "", "HeroRef", "", "zhangqu_liefeng", ""}
	cols[1] = []string{"", "", "Name", "", "长驱烈风", ""}
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	if len(res) > 0 {
		t.Logf("意外的错误: %+v", res)
		for _, r := range res {
			t.Logf("  - Index: %d, Reason: %s", r.Index, r.Reason)
		}
	}
	assert.Empty(t, res, "提取值 zhangqu_liefeng 应匹配 长驱烈风 的多音字变体 ZhangQuLieFeng")
}

// TestRegexCrossCol_Polyphone_AltVariant 多音字另一读音：chang 也应匹配
func TestRegexCrossCol_Polyphone_AltVariant_Pass(t *testing.T) {
	cols, colIdx, startRowIdx, params := newCrossColTestData()
	params["checkMode"] = "pinyin_contains"
	params["pattern"] = `(\w+)`
	params["groups"] = "1"
	cols[0] = []string{"", "", "HeroRef", "", "changqu_liefeng", ""}
	cols[1] = []string{"", "", "Name", "", "长驱烈风", ""}
	rule := &RegexCrossColCheckRule{}
	res := rule.Check("test", cols, colIdx, startRowIdx, params, nil)
	if len(res) > 0 {
		t.Logf("意外的错误: %+v", res)
		for _, r := range res {
			t.Logf("  - Index: %d, Reason: %s", r.Index, r.Reason)
		}
	}
	assert.Empty(t, res, "提取值 changqu_liefeng 应匹配 长驱烈风 的多音字变体 ChangQuLieFeng")
}
