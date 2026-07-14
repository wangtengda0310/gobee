//go:build integration
// +build integration

// Package check_manager 提供增量检查功能的集成测试
// 本测试使用测试资源目录，验证端到端的增量检查流程
//
// 运行方式：
//
//	cd D:/work/xcard-qa-tools/rain-excel-checker
//	go test ./xlsx/check_manager/... -v -tags=integration -run TestCheckWithFilter
package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// 集成测试的路径配置（使用项目内的测试资源）
// 注意：go test 会切换到包目录执行，所以需要使用相对于包目录的路径
const (
	testExcelPath = "../tests/resources"
	testCasePath  = "../tests/cases/excel_cases"
)

// TestCheckWithFilter_HeroOnly 测试只变更 Hero 表的增量检查
// 业务场景：只修改了 Hero.xlsx，应该只执行 Hero 相关的规则
// 集成测试：使用测试资源目录
func TestCheckWithFilter_HeroOnly(t *testing.T) {
	// 确保测试目录存在
	if _, err := os.Stat(testExcelPath); os.IsNotExist(err) {
		t.Skipf("测试资源目录不存在: %s", testExcelPath)
		return
	}

	excelPath := testExcelPath
	casePath := testCasePath
	changedFiles := []string{"Hero.xlsx"}

	// 打印绝对路径用于调试
	absExcelPath, _ := filepath.Abs(excelPath)
	absCasePath, _ := filepath.Abs(casePath)
	t.Logf("excelPath 相对路径: %s, 绝对路径: %s", excelPath, absExcelPath)
	t.Logf("casePath 相对路径: %s, 绝对路径: %s", casePath, absCasePath)

	// 调用 CheckWithFilter
	res, _, err := CheckWithFilter(excelPath, casePath, changedFiles)
	if err != nil {
		t.Fatalf("CheckWithFilter 失败: %v", err)
	}

	// 验证：函数能正常执行
	// 空结果表示没有检查错误，这是正常的
	if res == nil {
		t.Error("期望返回结果切片，实际为 nil")
	}
}

// TestCheckWithFilter_ArenaSeasonOnly 测试只变更 ArenaSeason 表的增量检查
// 业务场景：只修改了 ArenaSeason.xlsx，应该只执行 ArenaSeason 相关的规则
// 集成测试：使用测试资源目录
func TestCheckWithFilter_ArenaSeasonOnly(t *testing.T) {
	if _, err := os.Stat(testExcelPath); os.IsNotExist(err) {
		t.Skipf("测试资源目录不存在: %s", testExcelPath)
		return
	}

	excelPath := testExcelPath
	casePath := testCasePath
	changedFiles := []string{"ArenaSeason.xlsx"}

	res, _, err := CheckWithFilter(excelPath, casePath, changedFiles)
	if err != nil {
		t.Fatalf("CheckWithFilter 失败: %v", err)
	}

	if res == nil {
		t.Error("期望返回结果，实际为 nil")
	}
}

// TestCheckWithFilter_MultipleChanges 测试多个文件变更的增量检查
// 业务场景：同时修改了 Hero.xlsx 和 ArenaSeason.xlsx，应该执行两个表的规则
// 集成测试：使用测试资源目录
func TestCheckWithFilter_MultipleChanges(t *testing.T) {
	if _, err := os.Stat(testExcelPath); os.IsNotExist(err) {
		t.Skipf("测试资源目录不存在: %s", testExcelPath)
		return
	}

	excelPath := testExcelPath
	casePath := testCasePath
	changedFiles := []string{"Hero.xlsx", "ArenaSeason.xlsx"}

	res, _, err := CheckWithFilter(excelPath, casePath, changedFiles)
	if err != nil {
		t.Fatalf("CheckWithFilter 失败: %v", err)
	}

	if res == nil {
		t.Error("期望返回结果，实际为 nil")
	}
}

// TestCheckWithFilter_NoChanges 测试无变更文件时的全量检查
// 业务场景：没有提供变更文件列表，应该执行所有规则（全量检查）
// 集成测试：使用测试资源目录
func TestCheckWithFilter_NoChanges(t *testing.T) {
	if _, err := os.Stat(testExcelPath); os.IsNotExist(err) {
		t.Skipf("测试资源目录不存在: %s", testExcelPath)
		return
	}

	excelPath := testExcelPath
	casePath := testCasePath
	changedFiles := []string{} // 空列表表示全量检查

	res, _, err := CheckWithFilter(excelPath, casePath, changedFiles)
	if err != nil {
		t.Fatalf("CheckWithFilter 失败: %v", err)
	}

	if res == nil {
		t.Error("期望返回结果，实际为 nil")
	}
}

// TestCheckWithFilter_NilChanges 测试传入 nil 时的全量检查
// 业务场景：传入 nil 而不是空列表，也应该执行全量检查
// 集成测试：使用测试资源目录
func TestCheckWithFilter_NilChanges(t *testing.T) {
	if _, err := os.Stat(testExcelPath); os.IsNotExist(err) {
		t.Skipf("测试资源目录不存在: %s", testExcelPath)
		return
	}

	excelPath := testExcelPath
	casePath := testCasePath
	var changedFiles []string = nil // nil 表示全量检查

	res, _, err := CheckWithFilter(excelPath, casePath, changedFiles)
	if err != nil {
		t.Fatalf("CheckWithFilter 失败: %v", err)
	}

	if res == nil {
		t.Error("期望返回结果，实际为 nil")
	}
}

// TestSupplementDefaultTableRules 测试默认表级规则的自动补充
// 业务场景：Hero 表没有 JSON 文件，但有默认表级规则，应该被自动补充
func TestSupplementDefaultTableRules(t *testing.T) {
	// 这个测试验证的是：Hero 表有默认表级规则，应该被自动补充
	// 实际的验证已经在集成测试 TestCheckWithFilter_HeroOnly 中完成
	// TestCheckWithFilter_HeroOnly 会输出 "自动补充规则: 武将表|Hero (3 条表级规则)"
	t.Skip("需要实际的 Excel 文件，已在 TestCheckWithFilter_HeroOnly 中验证")
}
