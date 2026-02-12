package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestItem 测试用结构体
type TestItem struct {
	ID    int    `excel:"id"`
	Name  string `excel:"name,required"`
	Value int    `excel:"value,default:0"`
}

// TestLoader_AutoModeExcel 测试自动模式选择 Excel
func TestLoader_AutoModeExcel(t *testing.T) {
	// 创建测试 Excel 文件
	excelFile := createTestExcelForLoader(t)
	defer os.Remove(excelFile)

	// 创建加载器
	loader := NewLoader[TestItem](excelFile, "Sheet1", LoadOptions{
		Mode:      ModeAuto,
		HeaderRow: 0,
	})

	// 加载数据
	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Load() returned empty data")
	}

	// 验证数据
	if items[0].ID != 1 {
		t.Errorf("ID = %d, want 1", items[0].ID)
	}
	if items[0].Name != "test" {
		t.Errorf("Name = %s, want test", items[0].Name)
	}
}

// TestLoader_AutoModeCSV 测试自动模式选择 CSV
func TestLoader_AutoModeCSV(t *testing.T) {
	// 创建测试 CSV 文件
	csvFile := createTestCSVForLoader(t)
	defer os.RemoveAll(filepath.Dir(csvFile))

	// 创建加载器（使用 CSV 文件路径）
	loader := NewLoader[TestItem](csvFile, "data", LoadOptions{
		Mode:      ModeAuto,
		HeaderRow: 0,
	})

	// 加载数据
	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Load() returned empty data")
	}
}

// TestLoader_ExcelMode 测试强制 Excel 模式
func TestLoader_ExcelMode(t *testing.T) {
	excelFile := createTestExcelForLoader(t)
	defer os.Remove(excelFile)

	loader := NewLoader[TestItem](excelFile, "Sheet1", LoadOptions{
		Mode:      ModeExcel,
		HeaderRow: 0,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Load() returned empty data")
	}
}

// TestLoader_CSVMode 测试强制 CSV 模式
func TestLoader_CSVMode(t *testing.T) {
	csvFile := createTestCSVForLoader(t)
	defer os.RemoveAll(filepath.Dir(csvFile))

	loader := NewLoader[TestItem](csvFile, "data", LoadOptions{
		Mode:      ModeCSV,
		HeaderRow: 0,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Load() returned empty data")
	}
}

// TestLoader_SheetName 测试指定 Sheet 名称
func TestLoader_SheetName(t *testing.T) {
	excelFile := createTestExcelWithMultipleSheetsForLoader(t)
	defer os.Remove(excelFile)

	// 加载 Sheet2
	loader := NewLoader[TestItem](excelFile, "Sheet2", LoadOptions{
		Mode:      ModeExcel,
		HeaderRow: 0,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Sheet2 的数据应该来自不同的 Sheet
	if len(items) == 0 {
		t.Fatal("Load() returned empty data")
	}
}

// TestLoader_HeaderRow 测试自定义表头行
func TestLoader_HeaderRow(t *testing.T) {
	excelFile := createTestExcelWithHeaderRow(t)
	defer os.Remove(excelFile)

	// 表头在第 2 行（索引 1）
	loader := NewLoader[TestItem](excelFile, "Sheet1", LoadOptions{
		Mode:      ModeExcel,
		HeaderRow: 1,
	})

	items, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Load() returned empty data")
	}
}

// TestLoader_Reload 测试重新加载
func TestLoader_Reload(t *testing.T) {
	excelFile := createTestExcelForLoader(t)
	defer os.Remove(excelFile)

	loader := NewLoader[TestItem](excelFile, "Sheet1", LoadOptions{
		Mode:      ModeExcel,
		HeaderRow: 0,
	})

	// 第一次加载
	items1, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 重新加载
	items2, err := loader.Reload()
	if err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	// 两次加载结果应该相同
	if len(items1) != len(items2) {
		t.Errorf("Reload() returned %d items, want %d", len(items2), len(items1))
	}
}

// TestLoader_NotFound 测试文件不存在
func TestLoader_NotFound(t *testing.T) {
	loader := NewLoader[TestItem]("not_exist.xlsx", "Sheet1", LoadOptions{
		Mode: ModeAuto,
	})

	_, err := loader.Load()
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

// 辅助函数：创建测试 Excel 文件
func createTestExcelForLoader(t *testing.T) string {
	t.Helper()
	file := filepath.Join(os.TempDir(), t.Name()+".xlsx")
	f := NewTestExcelFile()
	defer f.Close()

	// 添加表头
	f.SetCellValue("Sheet1", "A1", "id")
	f.SetCellValue("Sheet1", "B1", "name")
	f.SetCellValue("Sheet1", "C1", "value")

	// 添加数据
	f.SetCellValue("Sheet1", "A2", "1")
	f.SetCellValue("Sheet1", "B2", "test")
	f.SetCellValue("Sheet1", "C2", "100")

	return f.Save(file)
}

// 辅助函数：创建带多个 Sheet 的测试 Excel 文件（loader 专用）
func createTestExcelWithMultipleSheetsForLoader(t *testing.T) string {
	t.Helper()
	file := filepath.Join(os.TempDir(), t.Name()+"_multiple.xlsx")
	f := NewTestExcelFile()

	// Sheet1 数据
	f.SetCellValue("Sheet1", "A1", "id")
	f.SetCellValue("Sheet1", "B1", "name")
	f.SetCellValue("Sheet1", "A2", "1")
	f.SetCellValue("Sheet1", "B2", "data1")

	// 创建并设置 Sheet2
	// 注意：NewTestExcelFile 默认创建 Sheet1，需要手动创建 Sheet2
	f.NewSheet("Sheet2")
	f.SetCellValue("Sheet2", "A1", "id")
	f.SetCellValue("Sheet2", "B1", "name")
	f.SetCellValue("Sheet2", "A2", "2")
	f.SetCellValue("Sheet2", "B2", "data2")

	result := f.Save(file)
	f.Close()
	return result
}

// 辅助函数：创建带自定义表头行的 Excel 文件
func createTestExcelWithHeaderRow(t *testing.T) string {
	t.Helper()
	file := filepath.Join(os.TempDir(), t.Name()+"_header.xlsx")
	f := NewTestExcelFile()
	defer f.Close()

	// 第一行：元数据
	f.SetCellValue("Sheet1", "A1", "metadata")
	f.SetCellValue("Sheet1", "B1", "value")

	// 第二行：表头
	f.SetCellValue("Sheet1", "A2", "id")
	f.SetCellValue("Sheet1", "B2", "name")

	// 第三行：数据
	f.SetCellValue("Sheet1", "A3", "1")
	f.SetCellValue("Sheet1", "B3", "test")

	return f.Save(file)
}

// 辅助函数：创建测试 CSV 文件
func createTestCSVForLoader(t *testing.T) string {
	t.Helper()

	// 创建目录结构: {filename}/{sheet}.csv
	baseDir := filepath.Join(os.TempDir(), t.Name())
	excelDir := filepath.Join(baseDir, "test_file")
	os.MkdirAll(excelDir, 0755)

	csvFile := filepath.Join(excelDir, "data.csv")
	content := `id,name,value
1,test,100
2,another,200`

	err := os.WriteFile(csvFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	return csvFile
}
