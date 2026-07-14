//go:build integration
// +build integration

// Package general 提供通用表级别校验规则的集成测试
// 本测试文件用于验证 git diff 变更检测功能的正确性
//
// 运行方式：
//   cd D:/work/xcard-qa-tools/rain-excel-checker
//   go test ./xlsx/coded_rules/general/... -v -tags=integration -run TestGitDiff

package coded_rules

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ============================================================================
// 测试配置 - 固化在代码中，保持测试稳定性
// ============================================================================

const (
	// Excel 文件配置
	testExcelDir  = "D:/work/config/excel" // Excel 配表目录（指向 config 仓库）
	testExcelFile = "Hero.xlsx"            // 指定的 Excel 文件名
	testSheetName = ""                     // Sheet 名称，空则自动解析

	// Git 提交配置 - 指定基准提交的 hash
	// 使用提交 41e36c1d 的父提交作为基准，该提交修改了 Hero.xlsx
	// 41e36c1d: 【0326版本-调整】限制黄巾兵出现的模式 (Hero.xlsx: 44025 -> 44067 字节)
	testBaseCommitHash = "41e36c1d^" // 父提交，确保有变更可检测

	// 预期变更断言（基于实际检测结果设置）
	// 通过运行 TestGitDiffBothNotify 获取实际变更数量后设置
	expectAddedRows    = 0  // 预期新增行数量
	expectRemovedRows  = 0  // 预期删除行数量
	expectModifiedRows = 24 // 预期修改字段数量（实际检测到 24 条）
)

// getBaseCommit 获取基准 commit（优先使用配置的 hash，否则使用 HEAD~1）
func getBaseCommit() string {
	if testBaseCommitHash != "" {
		return testBaseCommitHash
	}
	return "HEAD~1"
}

// resolveSheetName 解析 Sheet 名称
// 如果没有指定 Sheet 名称，尝试从 Excel 文件中获取
func resolveSheetName(t *testing.T, excelPath, excelFile string) string {
	xlsxFile, err := excelize.OpenFile(excelPath)
	assert.NoError(t, err, "无法打开 Excel 文件来获取 Sheet 列表")
	defer xlsxFile.Close()

	sheets := xlsxFile.GetSheetList()
	t.Logf("可用的 Sheet: %v", sheets)

	// 如果只有一个 Sheet，使用它
	if len(sheets) == 1 {
		return sheets[0]
	}

	// 尝试使用文件名（去掉扩展名）作为 Sheet 名称
	sheetName := strings.TrimSuffix(excelFile, filepath.Ext(excelFile))

	// 检查是否存在精确匹配或后缀匹配
	for _, s := range sheets {
		if s == sheetName || strings.HasSuffix(s, sheetName) {
			t.Logf("找到匹配的 Sheet: %s", s)
			return s
		}
	}

	// 默认返回第一个 Sheet
	assert.True(t, len(sheets) > 0, "没有可用的 Sheet")
	t.Logf("使用第一个 Sheet: %s", sheets[0])
	return sheets[0]
}

// ============================================================================
// 测试用例
// ============================================================================

// TestGitDiffNewRowNotify 使用 NewRowNotifyRule 进行回归测试
// 验证新增行、删除行、新增列、删除列的检测
func TestGitDiffNewRowNotify(t *testing.T) {
	baseCommit := getBaseCommit()

	// 构建文件路径
	excelPath := filepath.Join(testExcelDir, testExcelFile)

	sheetName := testSheetName
	if sheetName == "" {
		sheetName = resolveSheetName(t, excelPath, testExcelFile)
	}

	t.Logf("========== TestGitDiffNewRowNotify ==========")
	t.Logf("测试参数: Excel=%s, Sheet=%s, Commit=%s", testExcelFile, sheetName, baseCommit)

	// 读取当前版本 Excel（使用现有工具函数）
	xlsxFile, err := excelize.OpenFile(excelPath)
	assert.NoError(t, err, "打开 Excel 文件失败")
	defer xlsxFile.Close()

	cols, err := xlsxFile.GetCols(sheetName)
	assert.NoError(t, err, "读取 Sheet 失败")

	// 构建 sheetMap（规则检查需要）
	sheetMap := map[string]*excelize.File{sheetName: xlsxFile}

	// 构建规则参数（复用规则的默认参数结构）
	params := map[string]string{
		string(json_rule.ID_COL_NAME):         "Id",
		string(json_rule.NAME_COL_NAME):       "Name",
		string(json_rule.GIT_REPO_PATH):       testExcelDir,
		string(json_rule.BASE_COMMIT):         baseCommit,
		string(json_rule.NOTIFY_ADDED_ROWS):   "true",
		string(json_rule.NOTIFY_REMOVED_ROWS): "true",
		string(json_rule.NOTIFY_ADDED_COLS):   "true",
		string(json_rule.NOTIFY_REMOVED_COLS): "true",
	}

	// 执行规则检查（复用现有规则实现）
	rule := &NewRowNotifyRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   sheetName,
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	// 输出基本信息
	t.Logf("NewRowNotifyRule: Ok=%v, ErrCells=%d", result.Ok, len(result.ErrCells))

	// 详细输出变更内容
	if result.Reason != "" && result.Reason != "首次运行" {
		lines := strings.Split(result.Reason, "\n")
		for _, line := range lines {
			if line != "" {
				t.Logf("  %s", line)
			}
		}
	}

	// 断言：根据预期结果进行验证
	changeCount := len(result.ErrCells)
	totalExpected := expectAddedRows + expectRemovedRows

	// 断言变更状态
	if totalExpected == 0 {
		assert.True(t, result.Reason == "" || result.Reason == "首次运行",
			"预期无变更，但检测到变更: %s", result.Reason)
		return
	}

	// 预期有变更时验证
	assert.False(t, result.Reason == "" || result.Reason == "首次运行",
		"预期检测到 %d 项变更（新增=%d+删除=%d），但未检测到变更",
		totalExpected, expectAddedRows, expectRemovedRows)
	assert.Equal(t, totalExpected, changeCount,
		"预期变更数=%d (新增=%d+删除=%d), 实际=%d",
		totalExpected, expectAddedRows, expectRemovedRows, changeCount)
}

// TestGitDiffRowChangeNotify 使用 RowChangeNotifyRule 进行回归测试
// 验证字段值变更的检测
func TestGitDiffRowChangeNotify(t *testing.T) {
	baseCommit := getBaseCommit()

	// 构建文件路径
	excelPath := filepath.Join(testExcelDir, testExcelFile)

	sheetName := testSheetName
	if sheetName == "" {
		sheetName = resolveSheetName(t, excelPath, testExcelFile)
	}

	t.Logf("========== TestGitDiffRowChangeNotify ==========")
	t.Logf("测试参数: Excel=%s, Sheet=%s, Commit=%s", testExcelFile, sheetName, baseCommit)

	// 读取当前版本 Excel
	xlsxFile, err := excelize.OpenFile(excelPath)
	assert.NoError(t, err, "打开 Excel 文件失败")
	defer xlsxFile.Close()

	cols, err := xlsxFile.GetCols(sheetName)
	assert.NoError(t, err, "读取 Sheet 失败")

	// 构建 sheetMap
	sheetMap := map[string]*excelize.File{sheetName: xlsxFile}

	// 构建规则参数
	params := map[string]string{
		string(json_rule.ID_COL_NAME):   "Id",
		string(json_rule.NAME_COL_NAME): "Name",
		string(json_rule.GIT_REPO_PATH): testExcelDir,
		string(json_rule.BASE_COMMIT):   baseCommit,
	}

	// 执行规则检查（复用现有规则实现）
	rule := &RowChangeNotifyRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   sheetName,
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	// 输出基本信息
	t.Logf("RowChangeNotifyRule: Ok=%v, ErrCells=%d", result.Ok, len(result.ErrCells))

	// 详细输出变更内容
	if result.Reason != "" && result.Reason != "首次运行" {
		lines := strings.Split(result.Reason, "\n")
		for _, line := range lines {
			if line != "" {
				t.Logf("  %s", line)
			}
		}
	}

	// 断言：验证字段修改
	modifiedCount := len(result.ErrCells)

	// 断言变更状态
	if expectModifiedRows == 0 {
		assert.True(t, result.Reason == "" || result.Reason == "首次运行",
			"预期无字段变更，但检测到变更: %s", result.Reason)
		return
	}

	// 预期有变更时验证
	assert.False(t, result.Reason == "" || result.Reason == "首次运行",
		"预期检测到 %d 条字段修改，但未检测到变更", expectModifiedRows)
	assert.Equal(t, expectModifiedRows, modifiedCount,
		"预期修改字段数=%d, 实际=%d", expectModifiedRows, modifiedCount)

	// 显式断言：验证具体的字段修改
	// ID=19002 (山贼丙丁) 的 NotUseModeType 字段修改
	assertModifiedField(t, result, "19002", "NotUseModeType",
		"1,2,4,5,6,7,8,9,12,16", "1,2,4,5,6,7,8,9,12,15,16")
}

// TestGitDiffBothNotify 同时执行两个规则，验证完整变更检测
// 组合测试：先检测新增/删除，再检测修改
// 验证两个规则协同工作的正确性
func TestGitDiffBothNotify(t *testing.T) {
	baseCommit := getBaseCommit()

	// 构建文件路径
	excelPath := filepath.Join(testExcelDir, testExcelFile)

	sheetName := testSheetName
	if sheetName == "" {
		sheetName = resolveSheetName(t, excelPath, testExcelFile)
	}

	t.Logf("========== TestGitDiffBothNotify ==========")
	t.Logf("测试参数: Excel=%s, Sheet=%s, Commit=%s", testExcelFile, sheetName, baseCommit)

	// 读取当前版本 Excel
	xlsxFile, err := excelize.OpenFile(excelPath)
	assert.NoError(t, err, "打开 Excel 文件失败")
	defer xlsxFile.Close()

	cols, err := xlsxFile.GetCols(sheetName)
	assert.NoError(t, err, "读取 Sheet 失败")

	sheetMap := map[string]*excelize.File{sheetName: xlsxFile}

	// 执行 NewRowNotifyRule
	paramsRow := map[string]string{
		string(json_rule.ID_COL_NAME):         "Id",
		string(json_rule.NAME_COL_NAME):       "Name",
		string(json_rule.GIT_REPO_PATH):       testExcelDir,
		string(json_rule.BASE_COMMIT):         baseCommit,
		string(json_rule.NOTIFY_ADDED_ROWS):   "true",
		string(json_rule.NOTIFY_REMOVED_ROWS): "true",
		string(json_rule.NOTIFY_ADDED_COLS):   "true",
		string(json_rule.NOTIFY_REMOVED_COLS): "true",
	}

	rowRule := &NewRowNotifyRule{}
	rowResult := rowRule.Check(sheetName, cols, excelio.MJS_FIXED_ROWS_NUM, paramsRow, sheetMap)

	// 执行 RowChangeNotifyRule
	paramsField := map[string]string{
		string(json_rule.ID_COL_NAME):   "Id",
		string(json_rule.NAME_COL_NAME): "Name",
		string(json_rule.GIT_REPO_PATH): testExcelDir,
		string(json_rule.BASE_COMMIT):   baseCommit,
	}

	fieldRule := &RowChangeNotifyRule{}
	fieldResult := fieldRule.Check(sheetName, cols, excelio.MJS_FIXED_ROWS_NUM, paramsField, sheetMap)

	// 输出结果汇总
	t.Logf("========== 检测结果汇总 ==========")
	t.Logf("【NewRowNotifyRule】Ok=%v, ErrCells=%d", rowResult.Ok, len(rowResult.ErrCells))
	t.Logf("【RowChangeNotifyRule】Ok=%v, ErrCells=%d", fieldResult.Ok, len(fieldResult.ErrCells))

	// 统计断言
	totalChanges := len(rowResult.ErrCells) + len(fieldResult.ErrCells)
	totalExpected := expectAddedRows + expectRemovedRows + expectModifiedRows

	t.Logf("总计变更数: %d (预期: %d)", totalChanges, totalExpected)

	// 断言：验证总变更数
	assert.Equal(t, totalExpected, totalChanges,
		"预期总变更数=%d, 实际=%d", totalExpected, totalChanges)
}

// ============================================================================
// 断言辅助函数（使用 testify/assert，简化 if 逻辑）
// ============================================================================

// assertModifiedField 断言字段修改
// 用于验证检测到指定字段从旧值变成新值
func assertModifiedField(t *testing.T, result *json_rule.TableCheckResult, rowId, colName, expectedOld, expectedNew string) {
	// 处理通配符
	expectedOldPart := expectedOld
	if expectedOldPart == "*" {
		expectedOldPart = ""
	}
	expectedNewPart := expectedNew
	if expectedNewPart == "*" {
		expectedNewPart = ""
	}

	// 查找匹配的 ErrCell
	var foundCell *json_rule.CellError
	for _, cell := range result.ErrCells {
		if strings.Contains(cell.Reason, rowId) && strings.Contains(cell.Reason, colName) {
			foundCell = cell
			break
		}
	}

	// 断言：必须找到字段修改
	assert.NotNil(t, foundCell, "未找到字段修改: ID=%s, 列=%s", rowId, colName)

	// 断言：验证旧值
	if expectedOldPart != "" {
		assert.Contains(t, foundCell.Reason, expectedOldPart,
			"ID=%s 字段 %s 旧值不匹配: 期望包含 '%s', 实际 '%s'",
			rowId, colName, expectedOldPart, foundCell.Reason)
	}

	// 断言：验证新值
	if expectedNewPart != "" {
		assert.Contains(t, foundCell.Reason, expectedNewPart,
			"ID=%s 字段 %s 新值不匹配: 期望包含 '%s', 实际 '%s'",
			rowId, colName, expectedNewPart, foundCell.Reason)
	}
}
