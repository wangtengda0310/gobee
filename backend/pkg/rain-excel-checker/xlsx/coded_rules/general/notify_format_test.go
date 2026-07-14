package coded_rules

import (
	"strings"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"github.com/stretchr/testify/assert"
)

// ==================== diffTextHighlight 测试 ====================

func TestDiffTextHighlight_Identical(t *testing.T) {
	old, new_ := diffTextHighlight("hello world", "hello world")
	assert.Equal(t, "hello world", old)
	assert.Equal(t, "hello world", new_)
}

func TestDiffTextHighlight_CompletelyDifferent(t *testing.T) {
	old, new_ := diffTextHighlight("abc", "xyz")
	assert.Contains(t, old, "<font color='red'>abc</font>")
	assert.Contains(t, new_, "<font color='red'>xyz</font>")
}

func TestDiffTextHighlight_SuffixChanged(t *testing.T) {
	// "price:100" vs "price:200"：公共前缀 "price:"，公共后缀 "00"
	// 所以只有 "1" vs "2" 被标红
	old, new_ := diffTextHighlight("price:100", "price:200")
	assert.Equal(t, "price:<font color='red'>1</font>00", old)
	assert.Equal(t, "price:<font color='red'>2</font>00", new_)
}

func TestDiffTextHighlight_PrefixChanged(t *testing.T) {
	old, new_ := diffTextHighlight("old_value", "new_value")
	assert.Contains(t, old, "<font color='red'>old</font>")
	assert.Contains(t, new_, "<font color='red'>new</font>")
	// 后缀 "_value" 是公共部分
	assert.Contains(t, old, "_value")
	assert.Contains(t, new_, "_value")
}

func TestDiffTextHighlight_MiddleChanged(t *testing.T) {
	old, new_ := diffTextHighlight("abc123def", "abc456def")
	assert.Equal(t, "abc<font color='red'>123</font>def", old)
	assert.Equal(t, "abc<font color='red'>456</font>def", new_)
}

func TestDiffTextHighlight_EmptyOld(t *testing.T) {
	old, new_ := diffTextHighlight("", "something")
	assert.Equal(t, "(空)", old)
	assert.Equal(t, "<font color='red'>something</font>", new_)
}

func TestDiffTextHighlight_EmptyNew(t *testing.T) {
	old, new_ := diffTextHighlight("something", "")
	assert.Equal(t, "<font color='red'>something</font>", old)
	assert.Equal(t, "(空)", new_)
}

func TestDiffTextHighlight_BothEmpty(t *testing.T) {
	old, new_ := diffTextHighlight("", "")
	assert.Equal(t, "", old)
	assert.Equal(t, "", new_)
}

func TestDiffTextHighlight_ChineseContent(t *testing.T) {
	oldVal := "活动开启时间2026/04/30 5点结束"
	newVal := "活动开启时间2026/05/14 结束"
	old, new_ := diffTextHighlight(oldVal, newVal)
	// 公共前缀是 "活动开启时间"
	assert.Contains(t, old, "活动开启时间")
	assert.Contains(t, new_, "活动开启时间")
	// 公共后缀包含 "结束"
	assert.Contains(t, old, "结束")
	assert.Contains(t, new_, "结束")
	// 变化部分有红色标注
	assert.Contains(t, old, "<font color='red'>")
	assert.Contains(t, new_, "<font color='red'>")
}

func TestDiffTextHighlight_LongDescription(t *testing.T) {
	// 模拟实际活动描述的变更场景
	prefix := "1.活动开启时间"
	suffix := "23点59分59秒。\\n2.玩家每日可组成一支3人小队开启活动"
	oldVal := prefix + "2026/04/30 5点-" + suffix
	newVal := prefix + "2026/05/14 -" + suffix

	old, new_ := diffTextHighlight(oldVal, newVal)
	assert.Contains(t, old, prefix)
	assert.Contains(t, new_, prefix)
	assert.Contains(t, old, "<font color='red'>")
	assert.Contains(t, new_, "<font color='red'>")
	// 确认后缀保留
	assert.Contains(t, old, suffix)
	assert.Contains(t, new_, suffix)
}

func TestDiffTextHighlight_OneCharDifferent(t *testing.T) {
	old, new_ := diffTextHighlight("1", "2")
	assert.Contains(t, old, "<font color='red'>1</font>")
	assert.Contains(t, new_, "<font color='red'>2</font>")
}

func TestDiffTextHighlight_SingleCharSame(t *testing.T) {
	old, new_ := diffTextHighlight("a", "a")
	assert.Equal(t, "a", old)
	assert.Equal(t, "a", new_)
}

// TestDiffTextHighlight_UTF8RuneBoundary 验证 rune 级别比较不会切断多字节字符
// 修复前：按字节切片会切在 UTF-8 中间产生乱码
func TestDiffTextHighlight_UTF8RuneBoundary(t *testing.T) {
	// "小程序回流是否完成登录" vs "完成登录"
	// 公共后缀 "完成登录"，公共前缀为空
	// 修复前按字节切片：公共后缀 12 bytes 切在 "小"(3字节) 的中间
	old, new_ := diffTextHighlight("小程序回流是否完成登录", "完成登录")
	assert.Contains(t, old, "小程序回流是否")
	assert.Contains(t, old, "完成登录")
	assert.Contains(t, old, "<font color='red'>小程序回流是否</font>")
	assert.Equal(t, "完成登录", new_)

	// "小程序回流是否加入公会" vs "成功加入公会"
	old2, new2 := diffTextHighlight("小程序回流是否加入公会", "成功加入公会")
	assert.Contains(t, old2, "<font color='red'>小程序回流是否</font>")
	assert.Contains(t, new2, "<font color='red'>成功</font>")
	assert.Contains(t, old2, "加入公会")
	assert.Contains(t, new2, "加入公会")
	// 确认没有乱码
	for _, r := range old2 {
		assert.NotEqual(t, 0xFFFD, r, "old2 包含 Unicode 替换字符(乱码)")
	}
	for _, r := range new2 {
		assert.NotEqual(t, 0xFFFD, r, "new2 包含 Unicode 替换字符(乱码)")
	}
}

// ==================== 格式化函数测试 ====================

func TestFormatFieldChangeReason_WithColorHighlight(t *testing.T) {
	records := []*fieldChangeRecord{
		{
			rowId:   "39",
			rowName: "TestRow",
			lineNo:  6,
			changes: []struct {
				colName  string
				oldValue string
				newValue string
			}{
				{colName: "Description", oldValue: "price:100", newValue: "price:200"},
			},
		},
	}
	ctx := &formatNotifyContext{
		sheetName:   "活动表|Activity",
		idColName:   "Id",
		nameColName: "Name",
		commitTime:  "2026-05-14 10:00:00",
		committer:   "testuser",
		baseCommit:  "abc1234",
		headCommit:  "def5678",
	}

	result := formatFieldChangeReason(records, ctx)
	assert.Contains(t, result, "活动表|Activity")
	// "price:100" vs "price:200"：公共前缀 "price:"，公共后缀 "00"，只有 "1"/"2" 被标红
	assert.Contains(t, result, "<font color='red'>1</font>00")
	assert.Contains(t, result, "<font color='red'>2</font>00")
}

func TestFormatAddedRowsReason_GreenColor(t *testing.T) {
	rows := []*diff.RowChange{
		{RowId: "39", RowName: "TestRow"},
	}
	ctx := &formatNotifyContext{
		sheetName:   "活动表|Activity",
		idColName:   "Id",
		nameColName: "Name",
		commitTime:  "2026-05-14 10:00:00",
		committer:   "testuser",
	}

	result := formatAddedRowsReason(rows, ctx)
	assert.Contains(t, result, "<font color='green'>")
	assert.Contains(t, result, "新增行")
	assert.Contains(t, result, "活动表|Activity")
}

func TestFormatRemovedRowsReason_Strikethrough(t *testing.T) {
	rows := []*diff.RowChange{
		{RowId: "39", RowName: "TestRow"},
	}
	ctx := &formatNotifyContext{
		sheetName:   "活动表|Activity",
		idColName:   "Id",
		nameColName: "Name",
		commitTime:  "2026-05-14 10:00:00",
		committer:   "testuser",
	}

	result := formatRemovedRowsReason(rows, ctx)
	assert.Contains(t, result, "~~")
	assert.Contains(t, result, "删除行")
	assert.Contains(t, result, "活动表|Activity")
	// 确认有开始和结束的中划线标记
	assert.Equal(t, 2, strings.Count(result, "~~"))
}

func TestFormatFieldChangeReason_CompleteOutput(t *testing.T) {
	records := []*fieldChangeRecord{
		{
			rowId:   "39",
			rowName: "TestRow",
			lineNo:  6,
			changes: []struct {
				colName  string
				oldValue string
				newValue string
			}{
				{colName: "Description", oldValue: "old desc", newValue: "new desc"},
			},
		},
	}
	ctx := &formatNotifyContext{
		sheetName:   "活动表|Activity",
		idColName:   "Id",
		nameColName: "Name",
		commitTime:  "2026-05-14 10:00:00",
		committer:   "testuser",
		baseCommit:  "abc1234",
		headCommit:  "def5678",
	}

	result := formatFieldChangeReason(records, ctx)
	assert.Contains(t, result, "📋 表格名称: 活动表|Activity")
	assert.Contains(t, result, "🔄 变更类型: 修改行")
	assert.Contains(t, result, "变更范围: 共 1 行记录发生变更")
	assert.Contains(t, result, "第 6 行，Id 39")
	assert.Contains(t, result, "✏️ Description")
	assert.Contains(t, result, "⏰ 变更时间: 2026-05-14 10:00:00")
	assert.Contains(t, result, "testuser@ztgame.com")
	assert.Contains(t, result, "abc1234 → def5678")
}
