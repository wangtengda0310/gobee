package coded_rules

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
)

// makeTestNotifyCtx 构造测试用的通知格式化上下文
func makeTestNotifyCtx() *formatNotifyContext {
	return &formatNotifyContext{
		sheetName:   "TestSheet",
		idColName:   "Id",
		nameColName: "Name",
		commitTime:  "2025-03-30 14:23:15",
		committer:   "提交人",
		baseCommit:  "HEAD~1",
		headCommit:  "HEAD",
	}
}

func TestNewRowNotifyDetail_AddedRow(t *testing.T) {
	detail := &diff.RowChange{RowId: "100", RowName: "测试新增"}
	ctx := makeTestNotifyCtx()
	result := formatAddedRowsReason([]*diff.RowChange{detail}, ctx)
	assert.Contains(t, result, "📋 工作表变更通知 - TestSheet")
	assert.Contains(t, result, "🔄 新增行")
	assert.Contains(t, result, "+1 行")
	assert.Contains(t, result, "100")
	t.Logf("✓ 新增行通知格式:\n%s", result)
}

func TestNewRowNotifyDetail_RemovedRow(t *testing.T) {
	detail := &diff.RowChange{RowId: "50", RowName: "已删除数据"}
	ctx := makeTestNotifyCtx()
	result := formatRemovedRowsReason([]*diff.RowChange{detail}, ctx)
	assert.Contains(t, result, "📋 工作表变更通知 - TestSheet")
	assert.Contains(t, result, "🔄 删除行")
	assert.Contains(t, result, "-1 行")
	assert.Contains(t, result, "50")
	t.Logf("✓ 删除行通知格式:\n%s", result)
}

func TestNewRowNotifyDetail_ColumnChange(t *testing.T) {
	ctx := makeTestNotifyCtx()

	addedDetail := &json_rule.ColumnChangeDetail{ChangeType: "added", ColName: "NewField"}
	addedResult := formatAddedColsReason([]string{"NewField"}, ctx)
	assert.Contains(t, addedResult, "📋 工作表变更通知 - TestSheet")
	assert.Contains(t, addedResult, "🔄 新增列")
	assert.Contains(t, addedResult, "+1 列")
	assert.Contains(t, addedResult, "📝 NewField")

	removedDetail := &json_rule.ColumnChangeDetail{ChangeType: "removed", ColName: "OldField"}
	removedResult := formatRemovedColsReason([]string{"OldField"}, ctx)
	assert.Contains(t, removedResult, "🔄 删除列")
	assert.Contains(t, removedResult, "-1 列")
	assert.Contains(t, removedResult, "📝 OldField")

	t.Logf("✓ 新增列格式:\n%s", addedResult)
	t.Logf("✓ 删除列格式:\n%s", removedResult)

	// 保持 Detail 结构验证
	assert.Equal(t, "added", addedDetail.ChangeType)
	assert.Equal(t, "NewField", addedDetail.ColName)
	assert.Equal(t, "removed", removedDetail.ChangeType)
	assert.Equal(t, "OldField", removedDetail.ColName)
}
