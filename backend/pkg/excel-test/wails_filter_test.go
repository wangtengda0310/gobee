package exceltest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// createTestXlsxFile 在指定路径创建一个简单的 xlsx 文件并返回其路径
// data 用于写入单元格数据，确保多次调用产生不同文件内容
func createTestXlsxFile(t *testing.T, path string, sheetName string, data map[string]string) string {
	t.Helper()
	f := excelize.NewFile()
	// 确保 Sheet1 存在或使用指定名称
	if sheetName != "Sheet1" {
		f.NewSheet(sheetName)
	}
	// 写入单元格数据
	for cell, value := range data {
		err := f.SetCellValue(sheetName, cell, value)
		assert.NoError(t, err)
	}
	err := f.SaveAs(path)
	assert.NoError(t, err)
	f.Close()
	return path
}

// setupGitRepo 初始化一个 git 仓库并创建初始提交
func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	// git init
	runGit(t, dir, "init")
	// git config user
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
}

// runGit 在指定目录执行 git 命令
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "git %v failed: %s", args, string(output))
}

// TestFilterSheetMapByGitChanges_OnlyChangedFiles 测试只保留 git 变更文件对应的表
func TestFilterSheetMapByGitChanges_OnlyChangedFiles(t *testing.T) {
	// 创建临时目录和 git 仓库
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir)

	// 创建两个 xlsx 文件，写入初始数据
	file1Path := filepath.Join(tmpDir, "Hero.xlsx")
	file2Path := filepath.Join(tmpDir, "Arena.xlsx")
	createTestXlsxFile(t, file1Path, "Sheet1", map[string]string{"A1": "v1"})
	createTestXlsxFile(t, file2Path, "Sheet1", map[string]string{"A1": "v1"})

	// 初始提交所有文件
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial commit")

	// 修改 file1 的内容（写入不同的数据，确保 git 检测到变更）
	createTestXlsxFile(t, file1Path, "Sheet1", map[string]string{"A1": "v2", "B1": "new"})

	// 提交变更
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "modify Hero.xlsx")

	// 打开文件构建 sheetMap
	f1, err := excelize.OpenFile(file1Path)
	assert.NoError(t, err)
	f2, err := excelize.OpenFile(file2Path)
	assert.NoError(t, err)

	sheetMap := map[string]*excelize.File{
		"Hero":  f1,
		"Arena": f2,
	}

	// 执行过滤
	svc := &ExcelCheckService{}
	filtered := svc.filterSheetMapByGitChanges(sheetMap, tmpDir)

	// 验证只有 Hero 表被保留
	assert.Contains(t, filtered, "Hero", "变更的 Hero.xlsx 对应的表应该被保留")
	assert.NotContains(t, filtered, "Arena", "未变更的 Arena.xlsx 对应的表不应该被保留")

	// 清理
	f1.Close()
	f2.Close()
}

// TestFilterSheetMapByGitChanges_NoChanges 测试没有变更文件时返回空 map
func TestFilterSheetMapByGitChanges_NoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir)

	// 创建文件并提交
	filePath := filepath.Join(tmpDir, "Hero.xlsx")
	createTestXlsxFile(t, filePath, "Sheet1", map[string]string{"A1": "v1"})
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial commit")

	// 不做任何修改，直接再次提交（空提交）
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "empty commit")

	f1, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)

	sheetMap := map[string]*excelize.File{
		"Hero": f1,
	}

	svc := &ExcelCheckService{}
	filtered := svc.filterSheetMapByGitChanges(sheetMap, tmpDir)

	// 没有变更文件，应返回空 map
	assert.Empty(t, filtered, "没有变更文件时应返回空 map")

	f1.Close()
}

// TestFilterSheetMapByGitChanges_NotGitRepo 测试非 git 仓库时回退到检查所有表
func TestFilterSheetMapByGitChanges_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	// 不初始化 git 仓库

	filePath := filepath.Join(tmpDir, "Hero.xlsx")
	createTestXlsxFile(t, filePath, "Sheet1", map[string]string{"A1": "v1"})

	f1, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)

	sheetMap := map[string]*excelize.File{
		"Hero": f1,
	}

	svc := &ExcelCheckService{}
	filtered := svc.filterSheetMapByGitChanges(sheetMap, tmpDir)

	// 非 git 仓库，应回退返回原始 sheetMap
	assert.Equal(t, sheetMap, filtered, "非 git 仓库应返回原始 sheetMap")

	f1.Close()
}

// TestFilterSheetMapByGitChanges_RelativePath 测试相对路径能正确匹配
func TestFilterSheetMapByGitChanges_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitRepo(t, tmpDir)

	// 创建子目录
	excelDir := filepath.Join(tmpDir, "excel")
	os.MkdirAll(excelDir, 0755)

	// 创建文件并提交
	filePath := filepath.Join(excelDir, "Hero.xlsx")
	createTestXlsxFile(t, filePath, "Sheet1", map[string]string{"A1": "v1"})
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial commit")

	// 修改文件
	createTestXlsxFile(t, filePath, "Sheet1", map[string]string{"A1": "v2"})
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "modify Hero.xlsx")

	// 使用相对路径打开文件（模拟实际场景）
	relPath := filepath.Join(".", "excel", "Hero.xlsx")
	f1, err := excelize.OpenFile(filepath.Join(tmpDir, relPath))
	assert.NoError(t, err)

	sheetMap := map[string]*excelize.File{
		"Hero": f1,
	}

	svc := &ExcelCheckService{}
	// 传入 tmpDir 而非 excelDir，模拟实际使用场景
	filtered := svc.filterSheetMapByGitChanges(sheetMap, tmpDir)

	// 相对路径也应该能正确匹配
	assert.Contains(t, filtered, "Hero", "相对路径应该能正确匹配 git 变更文件")

	f1.Close()
}
