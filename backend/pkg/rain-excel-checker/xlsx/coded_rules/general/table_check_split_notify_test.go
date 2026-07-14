package coded_rules

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// testTableChecker 本地测试接口，避免导入 check_manager 造成循环依赖
type testTableChecker interface {
	Check(param json_rule.CheckParam) *json_rule.TableCheckResult
}

// testSheetName 测试使用的 sheet 名称（excelize.NewFile() 默认创建 "Sheet1"）
const testSheetName = "Sheet1"

// createTestSheetMap 创建测试用的 sheetMap
// 构建包含标准 4 行表头 + 2 行数据的测试 Excel 文件
func createTestSheetMap() (map[string]*excelize.File, *excelize.File) {
	f := excelize.NewFile()
	// 第1行：中文名
	f.SetCellValue(testSheetName, "A1", "ID")
	f.SetCellValue(testSheetName, "B1", "名称")
	// 第2行：类型
	f.SetCellValue(testSheetName, "A2", "int")
	f.SetCellValue(testSheetName, "B2", "string")
	// 第3行：字段名
	f.SetCellValue(testSheetName, "A3", "Id")
	f.SetCellValue(testSheetName, "B3", "Name")
	// 第4行：导出标识
	f.SetCellValue(testSheetName, "A4", "server")
	f.SetCellValue(testSheetName, "B4", "server/client")
	// 第5行开始：数据
	f.SetCellValue(testSheetName, "A5", "1")
	f.SetCellValue(testSheetName, "B5", "测试行1")
	f.SetCellValue(testSheetName, "A6", "2")
	f.SetCellValue(testSheetName, "B6", "测试行2")
	f.Path = "/test/test.xlsx"

	sheetMap := map[string]*excelize.File{testSheetName: f}
	return sheetMap, f
}

// commonTestParams 构建通用测试参数
// Params 中包含规则解析所需参数，使规则构建的缓存键与预填充的键一致
func commonTestParams(sheetMap map[string]*excelize.File) json_rule.CheckParam {
	return json_rule.CheckParam{
		SheetName: testSheetName,
		Params: map[string]string{
			"baseCommit":  "HEAD~1",
			"nameColName": "Name",
			"idColName":   "Id",
		},
		SheetMap:    sheetMap,
		StartRowIdx: 4, // MJS_FIXED_ROWS_NUM
	}
}

// prefillWithAddedRows 预填充包含新增行的 diff 缓存
func prefillWithAddedRows(excelPath, baseCommit, sheetName string) {
	key := diff.DiffCacheKey{
		ExcelPath:   excelPath,
		BaseCommit:  baseCommit,
		SheetName:   sheetName,
		IdColKey:    "Id",
		NameColName: "Name",
	}
	entry := &diff.DiffCacheEntry{
		DiffResult: &diff.ExcelDiffResult{
			AddedRows: []*diff.RowChange{
				{RowId: "100", RowName: "新增行100", RowIndex: 10},
				{RowId: "101", RowName: "新增行101", RowIndex: 11},
			},
		},
		GitCtx: &diff.GitNotifyContext{
			CommitTime: "2026-04-15 10:00:00",
			Committer:  "testuser",
			BaseCommit: "abc123",
			HeadCommit: "def456",
		},
		IsNewFile: false,
	}
	diff.PrefillDiffCache(key, entry)
}

// TestAddedRowNotifyRule_ShouldDetect 测试新增行通知规则
func TestAddedRowNotifyRule_ShouldDetect(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()
	prefillWithAddedRows(f.Path, "HEAD~1", testSheetName)

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	rule := &AddedRowNotifyRule{}
	result := rule.Check(param)

	assert.NotNil(t, result)
	assert.True(t, result.Ok) // 通知规则 Ok=true
	assert.NotEmpty(t, result.Reason)
	assert.Len(t, result.ErrCells, 2)
	assert.Equal(t, "100", result.ErrCells[0].Detail.(*json_rule.RowChangeDetail).RowId)
}

// TestRemovedRowNotifyRule_ShouldDetect 测试删除行通知规则
func TestRemovedRowNotifyRule_ShouldDetect(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()

	key := diff.DiffCacheKey{
		ExcelPath: f.Path, BaseCommit: "HEAD~1", SheetName: testSheetName,
		IdColKey: "Id", NameColName: "Name",
	}
	diff.PrefillDiffCache(key, &diff.DiffCacheEntry{
		DiffResult: &diff.ExcelDiffResult{
			RemovedRows: []*diff.RowChange{
				{RowId: "50", RowName: "已删除行", RowIndex: 5},
			},
		},
		GitCtx: &diff.GitNotifyContext{
			CommitTime: "2026-04-15 10:00:00", Committer: "testuser",
			BaseCommit: "abc123", HeadCommit: "def456",
		},
	})

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	rule := &RemovedRowNotifyRule{}
	result := rule.Check(param)

	assert.NotNil(t, result)
	assert.True(t, result.Ok)
	assert.Len(t, result.ErrCells, 1)
}

// TestAddedColNotifyRule_ShouldDetect 测试新增列通知规则
func TestAddedColNotifyRule_ShouldDetect(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()

	key := diff.DiffCacheKey{
		ExcelPath: f.Path, BaseCommit: "HEAD~1", SheetName: testSheetName,
		IdColKey: "Id", NameColName: "Name",
	}
	diff.PrefillDiffCache(key, &diff.DiffCacheEntry{
		DiffResult: &diff.ExcelDiffResult{
			AddedCols: []string{"NewField1", "NewField2"},
		},
		GitCtx: &diff.GitNotifyContext{
			CommitTime: "2026-04-15 10:00:00", Committer: "testuser",
			BaseCommit: "abc123", HeadCommit: "def456",
		},
	})

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	rule := &AddedColNotifyRule{}
	result := rule.Check(param)

	assert.NotNil(t, result)
	assert.True(t, result.Ok)
	assert.Len(t, result.ErrCells, 2)
}

// TestRemovedColNotifyRule_ShouldDetect 测试删除列通知规则
func TestRemovedColNotifyRule_ShouldDetect(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()

	key := diff.DiffCacheKey{
		ExcelPath: f.Path, BaseCommit: "HEAD~1", SheetName: testSheetName,
		IdColKey: "Id", NameColName: "Name",
	}
	diff.PrefillDiffCache(key, &diff.DiffCacheEntry{
		DiffResult: &diff.ExcelDiffResult{
			RemovedCols: []string{"OldField"},
		},
		GitCtx: &diff.GitNotifyContext{
			CommitTime: "2026-04-15 10:00:00", Committer: "testuser",
			BaseCommit: "abc123", HeadCommit: "def456",
		},
	})

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	rule := &RemovedColNotifyRule{}
	result := rule.Check(param)

	assert.NotNil(t, result)
	assert.True(t, result.Ok)
	assert.Len(t, result.ErrCells, 1)
}

// TestModifiedRowNotifyRule_ShouldDetect 测试修改行通知规则
func TestModifiedRowNotifyRule_ShouldDetect(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()

	key := diff.DiffCacheKey{
		ExcelPath: f.Path, BaseCommit: "HEAD~1", SheetName: testSheetName,
		IdColKey: "Id", NameColName: "Name",
	}
	diff.PrefillDiffCache(key, &diff.DiffCacheEntry{
		DiffResult: &diff.ExcelDiffResult{
			ModifiedRows: []*diff.RowChange{
				{
					RowId: "1", RowName: "测试行1", RowIndex: 4,
					Changes: []*diff.FieldChange{
						{ColName: "Name", OldValue: "旧名", NewValue: "新名"},
					},
				},
			},
		},
		GitCtx: &diff.GitNotifyContext{
			CommitTime: "2026-04-15 10:00:00", Committer: "testuser",
			BaseCommit: "abc123", HeadCommit: "def456",
		},
	})

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	rule := &ModifiedRowNotifyRule{}
	result := rule.Check(param)

	assert.NotNil(t, result)
	assert.True(t, result.Ok)
	assert.Len(t, result.ErrCells, 1)
	detail := result.ErrCells[0].Detail.(*json_rule.FieldChangeDetail)
	assert.Equal(t, "Name", detail.ColName)
	assert.Equal(t, "旧名", detail.OldValue)
	assert.Equal(t, "新名", detail.NewValue)
}

// TestNewRules_NoChange 测试无变更时返回空结果
func TestNewRules_NoChange(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()

	// 预填充空 diff
	key := diff.DiffCacheKey{
		ExcelPath: f.Path, BaseCommit: "HEAD~1", SheetName: testSheetName,
		IdColKey: "Id", NameColName: "Name",
	}
	diff.PrefillDiffCache(key, &diff.DiffCacheEntry{
		DiffResult: &diff.ExcelDiffResult{},
		GitCtx:     &diff.GitNotifyContext{},
	})

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	// 所有 5 个规则都应返回空结果
	for _, rule := range []testTableChecker{
		&AddedRowNotifyRule{},
		&RemovedRowNotifyRule{},
		&AddedColNotifyRule{},
		&RemovedColNotifyRule{},
		&ModifiedRowNotifyRule{},
	} {
		result := rule.Check(param)
		assert.NotNil(t, result)
		assert.Empty(t, result.ErrCells)
		assert.Empty(t, result.Reason)
	}
}

// TestNewRules_NewFile 测试新文件场景
// 新文件时 AddedRowNotifyRule 生成"新增文件"通知，其他 4 个规则返回空
func TestNewRules_NewFile(t *testing.T) {
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	sheetMap, f := createTestSheetMap()

	// 预填充 IsNewFile=true
	key := diff.DiffCacheKey{
		ExcelPath: f.Path, BaseCommit: "HEAD~1", SheetName: testSheetName,
		IdColKey: "Id", NameColName: "Name",
	}
	diff.PrefillDiffCache(key, &diff.DiffCacheEntry{
		IsNewFile:  true,
		NewFileMsg: "新增文件: test.xlsx",
		GitCtx:     &diff.GitNotifyContext{},
	})

	cols, _ := f.GetCols(testSheetName)
	param := commonTestParams(sheetMap)
	param.Cols = cols

	// AddedRowNotifyRule: 新文件场景，生成"新增文件"通知
	addedResult := (&AddedRowNotifyRule{}).Check(param)
	assert.NotNil(t, addedResult)
	assert.True(t, addedResult.Ok)
	assert.NotEmpty(t, addedResult.ErrCells) // 新文件有通知
	assert.Equal(t, "fileAdded", addedResult.ErrCells[0].Detail.(*json_rule.ColumnChangeDetail).ChangeType)

	// 其他规则：新文件场景都返回空
	for _, rule := range []testTableChecker{
		&RemovedRowNotifyRule{},
		&AddedColNotifyRule{},
		&RemovedColNotifyRule{},
		&ModifiedRowNotifyRule{},
	} {
		result := rule.Check(param)
		assert.NotNil(t, result)
		assert.Empty(t, result.ErrCells)
	}
}
