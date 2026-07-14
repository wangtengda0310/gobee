// services.go 的单元测试，聚焦 UpdateCase（按 case 名原地更新，保持顺序）
package functiontest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTempCasesForUpdate 写入临时 JSON 用例文件，返回路径。供 UpdateCase 测试复用。
func writeTempCasesForUpdate(t *testing.T, cases []QAFuncCase) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cases.json")
	require.NoError(t, NewJsonCaseService(nil).SaveJSONFile(path, nil, cases))
	return path
}

func TestUpdateCase_Success(t *testing.T) {
	svc := NewJsonCaseService(nil)
	path := writeTempCasesForUpdate(t, []QAFuncCase{
		{Case: "TC-A", Desc: "old A"},
		{Case: "TC-B", Desc: "B"},
	})

	require.NoError(t, svc.UpdateCase(path, "TC-A", QAFuncCase{Case: "TC-A", Desc: "new A"}))

	got, _, err := svc.readJSONFile(path)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "new A", got[0].Desc, "目标用例内容应已更新")
	assert.Equal(t, "B", got[1].Desc, "其他用例不应受影响")
}

func TestUpdateCase_NotFound(t *testing.T) {
	svc := NewJsonCaseService(nil)
	path := writeTempCasesForUpdate(t, []QAFuncCase{{Case: "TC-A", Desc: "A"}})

	err := svc.UpdateCase(path, "TC-X", QAFuncCase{Case: "TC-X", Desc: "X"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不存在")
}

func TestUpdateCase_KeepsOrder(t *testing.T) {
	svc := NewJsonCaseService(nil)
	path := writeTempCasesForUpdate(t, []QAFuncCase{
		{Case: "TC-A"}, {Case: "TC-B"}, {Case: "TC-C"},
	})

	require.NoError(t, svc.UpdateCase(path, "TC-B", QAFuncCase{Case: "TC-B", Desc: "updated"}))

	got, _, err := svc.readJSONFile(path)
	require.NoError(t, err)
	require.Len(t, got, 3)
	assert.Equal(t, []string{"TC-A", "TC-B", "TC-C"},
		[]string{got[0].Case, got[1].Case, got[2].Case}, "用例顺序应保持不变")
	assert.Equal(t, "updated", got[1].Desc, "目标用例内容应已更新")
}
