package coded_rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// makeTestFieldChangeCtx 构造测试用的通知格式化上下文
func makeTestFieldChangeCtx() *formatNotifyContext {
	return &formatNotifyContext{
		sheetName:   "客户信息表",
		idColName:   "客户ID",
		nameColName: "客户名称",
		commitTime:  "2025-03-30 09:12:33",
		committer:   "王五",
		baseCommit:  "v2.3.1",
		headCommit:  "v2.4.0",
	}
}

// TestRowChangeNotifyDetail_FieldChange 测试单条字段变更通知格式
func TestRowChangeNotifyDetail_FieldChange(t *testing.T) {
	records := []*fieldChangeRecord{
		{
			rowId:   "CUST-10086",
			rowName: "客户信息表",
			lineNo:  42,
			changes: []struct {
				colName  string
				oldValue string
				newValue string
			}{
				{colName: "联系人", oldValue: "王建国", newValue: "王晓明"},
			},
		},
	}
	ctx := makeTestFieldChangeCtx()
	result := formatFieldChangeReason(records, ctx)

	assert.Contains(t, result, "📋 表格名称: 客户信息表")
	assert.Contains(t, result, "🔄 变更类型: 修改行")
	assert.Contains(t, result, "共 1 行记录发生变更")
	assert.Contains(t, result, "CUST-10086")
	assert.Contains(t, result, "第 42 行")
	assert.Contains(t, result, "联系人")
	// "王建国" vs "王晓明"：公共前缀 "王"，变化部分 "建国"/"晓明" 被红色标注
	assert.Contains(t, result, "[原值] 王<font color='red'>建国</font>")
	assert.Contains(t, result, "[新值] 王<font color='red'>晓明</font>")
	assert.Contains(t, result, "⏰ 变更时间: 2025-03-30 09:12:33")
	assert.Contains(t, result, "v2.3.1 → v2.4.0")
	t.Logf("✓ 字段变更通知格式:\n%s", result)
}

// TestRowChangeNotifyDetail_MultipleFields 测试多行多字段变更通知格式
func TestRowChangeNotifyDetail_MultipleFields(t *testing.T) {
	records := []*fieldChangeRecord{
		{
			rowId:   "CUST-10086",
			rowName: "客户信息表",
			lineNo:  42,
			changes: []struct {
				colName  string
				oldValue string
				newValue string
			}{
				{colName: "联系人", oldValue: "王建国", newValue: "王晓明"},
				{colName: "电话", oldValue: "13800001111", newValue: "13900002222"},
			},
		},
		{
			rowId:   "CUST-10087",
			rowName: "客户信息表",
			lineNo:  55,
			changes: []struct {
				colName  string
				oldValue string
				newValue string
			}{
				{colName: "地址", oldValue: "北京", newValue: "上海"},
			},
		},
	}
	ctx := makeTestFieldChangeCtx()
	result := formatFieldChangeReason(records, ctx)

	assert.Contains(t, result, "共 2 行记录发生变更")
	assert.Contains(t, result, "【变更记录 1】")
	assert.Contains(t, result, "【变更记录 2】")
	assert.Contains(t, result, "CUST-10086")
	assert.Contains(t, result, "CUST-10087")
	// "王建国" vs "王晓明"：公共前缀 "王"，变化部分 "建国"/"晓明" 被标红
	assert.Contains(t, result, "王<font color='red'>建国</font>")
	assert.Contains(t, result, "王<font color='red'>晓明</font>")
	// "13800001111" vs "13900002222"：公共前缀 "13"，变化部分被标红
	assert.Contains(t, result, "13<font color='red'>800001111</font>")
	assert.Contains(t, result, "13<font color='red'>900002222</font>")
	// "北京" vs "上海"：完全不同，整体标红
	assert.Contains(t, result, "<font color='red'>北京</font>")
	assert.Contains(t, result, "<font color='red'>上海</font>")
	t.Logf("✓ 多字段变更通知格式:\n%s", result)
}

// TestRowChangeNotifyDetail_SpecialCharacters 测试空值和特殊字符处理
func TestRowChangeNotifyDetail_SpecialCharacters(t *testing.T) {
	records := []*fieldChangeRecord{
		{
			rowId:   "CUST-10086",
			rowName: "客户信息表",
			lineNo:  1,
			changes: []struct {
				colName  string
				oldValue string
				newValue string
			}{
				// 原值为空，应显示为 "(空)"
				{colName: "备注", oldValue: "", newValue: "新增备注"},
			},
		},
	}
	ctx := makeTestFieldChangeCtx()
	result := formatFieldChangeReason(records, ctx)

	// 空值原值显示为 (空)，新值整体标红
	assert.Contains(t, result, "[原值] (空)")
	assert.Contains(t, result, "<font color='red'>新增备注</font>")
	t.Logf("✓ 空值处理格式:\n%s", result)
}

// TestRowChangeNotifyDetail_EmptyRecords 测试空记录列表
func TestRowChangeNotifyDetail_EmptyRecords(t *testing.T) {
	ctx := makeTestFieldChangeCtx()
	result := formatFieldChangeReason([]*fieldChangeRecord{}, ctx)

	assert.Contains(t, result, "📋 表格名称: 客户信息表")
	assert.Contains(t, result, "共 0 行记录发生变更")
	assert.Contains(t, result, "⏰ 变更时间: 2025-03-30 09:12:33")
	t.Logf("✓ 空记录格式:\n%s", result)
}
