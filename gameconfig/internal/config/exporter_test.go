package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestExcelExporter_ExportSingleSheet 测试导出单个 Sheet
func TestExcelExporter_ExportSingleSheet(t *testing.T) {
	// 创建测试 Excel 文件
	excelFile := createTestExcelForExport(t)
	defer os.Remove(excelFile)

	// 创建临时输出目录
	outputDir := filepath.Join(os.TempDir(), t.Name()+"_output")
	defer os.RemoveAll(outputDir)

	// 导出为 CSV
	exporter := NewExcelExporter(excelFile, outputDir)
	err := exporter.Export()
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 验证输出文件
	baseName := strings.TrimSuffix(filepath.Base(excelFile), ".xlsx")
	csvFile := filepath.Join(outputDir, baseName, "Sheet1.csv")

	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		t.Errorf("CSV file not created: %s", csvFile)
	}

	// 验证 CSV 内容
	content, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "id,name,value") {
		t.Errorf("CSV content missing header, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "1,test,100") {
		t.Errorf("CSV content missing data, got: %s", contentStr)
	}
}

// TestExcelExporter_ExportMultipleSheets 测试导出多个 Sheet
func TestExcelExporter_ExportMultipleSheets(t *testing.T) {
	// 创建包含多个 Sheet 的 Excel 文件
	excelFile := createTestExcelWithMultipleSheets(t)
	defer os.Remove(excelFile)

	// 创建临时输出目录
	outputDir := filepath.Join(os.TempDir(), t.Name()+"_output")
	defer os.RemoveAll(outputDir)

	// 导出为 CSV
	exporter := NewExcelExporter(excelFile, outputDir)
	err := exporter.Export()
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 验证输出文件
	baseName := strings.TrimSuffix(filepath.Base(excelFile), ".xlsx")

	// 检查两个 Sheet 都导出了
	sheet1CSV := filepath.Join(outputDir, baseName, "Sheet1.csv")
	sheet2CSV := filepath.Join(outputDir, baseName, "Sheet2.csv")

	if _, err := os.Stat(sheet1CSV); os.IsNotExist(err) {
		t.Errorf("Sheet1 CSV not created: %s", sheet1CSV)
	}
	if _, err := os.Stat(sheet2CSV); os.IsNotExist(err) {
		t.Errorf("Sheet2 CSV not created: %s", sheet2CSV)
	}
}

// TestExcelExporter_ExportWithVersion 测试导出带版本号的 Sheet
func TestExcelExporter_ExportWithVersion(t *testing.T) {
	// 创建带版本号的 Excel 文件
	excelFile := createTestExcelWithVersion(t, 2)
	defer os.Remove(excelFile)

	// 创建临时输出目录
	outputDir := filepath.Join(os.TempDir(), t.Name()+"_output")
	defer os.RemoveAll(outputDir)

	// 导出为 CSV
	exporter := NewExcelExporter(excelFile, outputDir)
	err := exporter.Export()
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 验证输出文件
	baseName := strings.TrimSuffix(filepath.Base(excelFile), ".xlsx")
	csvFile := filepath.Join(outputDir, baseName, "Sheet1.csv")

	content, err := os.ReadFile(csvFile)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// 版本行应该被跳过（CSV 从实际数据行开始）
	contentStr := string(content)
	if strings.Contains(contentStr, "__version__") {
		t.Error("Version row should be skipped in CSV export")
	}
}

// TestExcelExporter_CreateOutputDirectory 测试自动创建输出目录
func TestExcelExporter_CreateOutputDirectory(t *testing.T) {
	// 创建测试 Excel 文件
	excelFile := createTestExcelForExport(t)
	defer os.Remove(excelFile)

	// 使用不存在的输出目录
	outputDir := filepath.Join(os.TempDir(), t.Name()+"_nonexistent")
	defer os.RemoveAll(outputDir)

	// 确保目录不存在
	os.RemoveAll(outputDir)

	// 导出应该自动创建目录
	exporter := NewExcelExporter(excelFile, outputDir)
	err := exporter.Export()
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 验证目录已创建
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Output directory not created")
	}
}

// TestExcelExporter_ExportSpecificSheets 测试导出指定的 Sheet
func TestExcelExporter_ExportSpecificSheets(t *testing.T) {
	// 创建包含多个 Sheet 的 Excel 文件
	excelFile := createTestExcelWithMultipleSheets(t)
	defer os.Remove(excelFile)

	// 创建临时输出目录
	outputDir := filepath.Join(os.TempDir(), t.Name()+"_output")
	defer os.RemoveAll(outputDir)

	// 只导出 Sheet1
	exporter := NewExcelExporter(excelFile, outputDir)
	exporter.SetSheets([]string{"Sheet1"})
	err := exporter.Export()
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	// 验证只导出了 Sheet1
	baseName := strings.TrimSuffix(filepath.Base(excelFile), ".xlsx")

	sheet1CSV := filepath.Join(outputDir, baseName, "Sheet1.csv")
	sheet2CSV := filepath.Join(outputDir, baseName, "Sheet2.csv")

	if _, err := os.Stat(sheet1CSV); os.IsNotExist(err) {
		t.Error("Sheet1 CSV should be created")
	}
	if _, err := os.Stat(sheet2CSV); !os.IsNotExist(err) {
		t.Error("Sheet2 CSV should not be created")
	}
}

// 辅助函数：创建用于导出测试的 Excel 文件
func createTestExcelForExport(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+".xlsx")
	f := excelize.NewFile()

	// 添加表头
	f.SetCellValue("Sheet1", "A1", "id")
	f.SetCellValue("Sheet1", "B1", "name")
	f.SetCellValue("Sheet1", "C1", "value")

	// 添加数据行
	f.SetCellValue("Sheet1", "A2", "1")
	f.SetCellValue("Sheet1", "B2", "test")
	f.SetCellValue("Sheet1", "C2", "100")

	f.SetCellValue("Sheet1", "A3", "2")
	f.SetCellValue("Sheet1", "B3", "another")
	f.SetCellValue("Sheet1", "C3", "200")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}

// 辅助函数：创建包含多个 Sheet 的 Excel 文件
func createTestExcelWithMultipleSheets(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+"_multiple.xlsx")
	f := excelize.NewFile()

	// Sheet1 数据
	f.SetCellValue("Sheet1", "A1", "id")
	f.SetCellValue("Sheet1", "B1", "name")
	f.SetCellValue("Sheet1", "A2", "1")
	f.SetCellValue("Sheet1", "B2", "data1")

	// 创建 Sheet2
	index, _ := f.NewSheet("Sheet2")
	_ = index

	f.SetCellValue("Sheet2", "A1", "key")
	f.SetCellValue("Sheet2", "B1", "value")
	f.SetCellValue("Sheet2", "A2", "k1")
	f.SetCellValue("Sheet2", "B2", "v1")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}

// 辅助函数：创建带版本号的 Excel 文件
func createTestExcelWithVersion(t *testing.T, version int) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+"_version.xlsx")
	f := excelize.NewFile()

	// 设置版本号
	f.SetCellValue("Sheet1", "A1", "__version__")
	f.SetCellValue("Sheet1", "B1", version)

	// 添加表头
	f.SetCellValue("Sheet1", "A2", "id")
	f.SetCellValue("Sheet1", "B2", "name")

	// 添加数据行
	f.SetCellValue("Sheet1", "A3", "1")
	f.SetCellValue("Sheet1", "B3", "test")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}
