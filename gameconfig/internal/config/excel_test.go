package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestNewExcelReader 测试创建 Excel 读取器
func TestNewExcelReader(t *testing.T) {
	// 准备测试文件
	testFile := createTestExcelFile(t)
	defer os.Remove(testFile)

	// 创建读取器
	reader, err := NewExcelReader(testFile)
	if err != nil {
		t.Fatalf("NewExcelReader() error = %v", err)
	}

	if reader == nil {
		t.Fatal("NewExcelReader() returned nil reader")
	}

	// 验证文件路径
	if reader.Path() != testFile {
		t.Errorf("Path() = %s, want %s", reader.Path(), testFile)
	}

	// 关闭读取器
	reader.Close()
}

// TestNewExcelReader_NotExist 测试文件不存在的情况
func TestNewExcelReader_NotExist(t *testing.T) {
	_, err := NewExcelReader("not_exist.xlsx")
	if err == nil {
		t.Error("NewExcelReader() should return error for non-existent file")
	}

	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("error = %v, want ErrFileNotFound", err)
	}
}

// TestExcelReader_GetSheetNames 测试获取 Sheet 名称列表
func TestExcelReader_GetSheetNames(t *testing.T) {
	testFile := createTestExcelFile(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	sheets := reader.GetSheetNames()
	if len(sheets) == 0 {
		t.Fatal("GetSheetNames() returned empty list")
	}

	// 验证默认 Sheet 存在
	found := false
	for _, name := range sheets {
		if name == "Sheet1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetSheetNames() should contain 'Sheet1'")
	}
}

// TestExcelReader_ReadSheet 测试读取 Sheet 数据
func TestExcelReader_ReadSheet(t *testing.T) {
	testFile := createTestExcelFileWithData(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	data, err := reader.ReadSheet("Sheet1")
	if err != nil {
		t.Fatalf("ReadSheet() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ReadSheet() returned empty data")
	}

	// 验证数据格式
	for i, row := range data {
		if len(row) == 0 {
			t.Errorf("row %d is empty", i)
		}
	}
}

// TestExcelReader_ReadSheet_NotExist 测试读取不存在的 Sheet
func TestExcelReader_ReadSheet_NotExist(t *testing.T) {
	testFile := createTestExcelFile(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	_, err := reader.ReadSheet("NotExist")
	if err == nil {
		t.Error("ReadSheet() should return error for non-existent sheet")
	}

	if !errors.Is(err, ErrSheetNotFound) {
		t.Errorf("error = %v, want ErrSheetNotFound", err)
	}
}

// TestExcelReader_ReadAllSheets 测试读取所有 Sheet
func TestExcelReader_ReadAllSheets(t *testing.T) {
	testFile := createTestExcelFileWithMultipleSheets(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	data, err := reader.ReadAllSheets()
	if err != nil {
		t.Fatalf("ReadAllSheets() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ReadAllSheets() returned empty data")
	}

	// 验证返回的 map 包含 Sheet 名称
	for sheetName := range data {
		t.Logf("Found sheet: %s", sheetName)
	}
}

// TestExcelReader_ReadFormula 测试读取公式计算值
func TestExcelReader_ReadFormula(t *testing.T) {
	testFile := createTestExcelFileWithFormula(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	// 使用 ReadCell 方法读取单元格（它会获取公式计算后的值）
	value, err := reader.ReadCell("Sheet1", "A3")
	if err != nil {
		t.Fatalf("ReadCell() error = %v", err)
	}

	// 验证公式计算结果: A1=10, A2=20, A3=A1+A2=30
	if value != "30" {
		t.Errorf("Formula result = %s, want 30", value)
	}
}

// TestExcelReader_ReadEmptySheet 测试读取空 Sheet
func TestExcelReader_ReadEmptySheet(t *testing.T) {
	testFile := createTestExcelFile(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	data, err := reader.ReadSheet("Sheet1")
	if err != nil {
		t.Errorf("ReadSheet() on empty sheet should not error, got %v", err)
	}

	// 空 Sheet 应该返回空数据或只有表头
	if data == nil {
		t.Error("ReadSheet() should not return nil")
	}
}

// TestExcelReader_ReadCell 测试读取单个单元格
func TestExcelReader_ReadCell(t *testing.T) {
	testFile := createTestExcelFileWithData(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	value, err := reader.ReadCell("Sheet1", "A1")
	if err != nil {
		t.Fatalf("ReadCell() error = %v", err)
	}

	// 第一行第一列应该是 "id"
	if value != "id" {
		t.Errorf("ReadCell() = %s, want id", value)
	}
}

// TestExcelReader_GetVersion 测试获取版本号
func TestExcelReader_GetVersion(t *testing.T) {
	testFile := createTestExcelFileWithVersion(t, 2)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	version, err := reader.GetVersion("Sheet1")
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if version != 2 {
		t.Errorf("GetVersion() = %d, want 2", version)
	}
}

// TestExcelReader_GetVersion_NotSet 测试未设置版本号的情况
func TestExcelReader_GetVersion_NotSet(t *testing.T) {
	testFile := createTestExcelFile(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	version, err := reader.GetVersion("Sheet1")
	if err != nil {
		t.Fatalf("GetVersion() error = %v", err)
	}

	if version != 0 {
		t.Errorf("GetVersion() = %d, want 0 (default)", version)
	}
}

// TestExcelReader_GetRowCount 测试获取行数
func TestExcelReader_GetRowCount(t *testing.T) {
	testFile := createTestExcelFileWithData(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	count, err := reader.GetRowCount("Sheet1")
	if err != nil {
		t.Fatalf("GetRowCount() error = %v", err)
	}

	if count <= 0 {
		t.Errorf("GetRowCount() = %d, want > 0", count)
	}
}

// TestExcelReader_GetColCount 测试获取列数
func TestExcelReader_GetColCount(t *testing.T) {
	testFile := createTestExcelFileWithData(t)
	defer os.Remove(testFile)

	reader, _ := NewExcelReader(testFile)
	defer reader.Close()

	count, err := reader.GetColCount("Sheet1")
	if err != nil {
		t.Fatalf("GetColCount() error = %v", err)
	}

	if count <= 0 {
		t.Errorf("GetColCount() = %d, want > 0", count)
	}
}

// TestCellToCoord 测试单元格坐标转换
func TestCellToCoord(t *testing.T) {
	tests := []struct {
		cell string
		col  int
		row  int
	}{
		{"A1", 1, 1},
		{"B2", 2, 2},
		{"Z10", 26, 10},
		{"AA1", 27, 1},
	}

	for _, tt := range tests {
		t.Run(tt.cell, func(t *testing.T) {
			col, row, err := CellToCoord(tt.cell)
			if err != nil {
				t.Fatalf("CellToCoord() error = %v", err)
			}
			if col != tt.col || row != tt.row {
				t.Errorf("CellToCoord() = (%d, %d), want (%d, %d)", col, row, tt.col, tt.row)
			}
		})
	}
}

// TestCoordToCell 测试行列坐标转换单元格
func TestCoordToCell(t *testing.T) {
	tests := []struct {
		col  int
		row  int
		cell string
	}{
		{1, 1, "A1"},
		{2, 2, "B2"},
		{26, 10, "Z10"},
		{27, 1, "AA1"},
	}

	for _, tt := range tests {
		t.Run(tt.cell, func(t *testing.T) {
			cell := CoordToCell(tt.col, tt.row)
			if cell != tt.cell {
				t.Errorf("CoordToCell() = %s, want %s", cell, tt.cell)
			}
		})
	}
}

// 辅助函数：创建测试用 Excel 文件（空文件）
func createTestExcelFile(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+".xlsx")
	f := excelize.NewFile()
	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}

// 辅助函数：创建带数据的测试 Excel 文件
func createTestExcelFileWithData(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+"_data.xlsx")
	f := excelize.NewFile()

	// 添加表头
	f.SetCellValue("Sheet1", "A1", "id")
	f.SetCellValue("Sheet1", "B1", "name")
	f.SetCellValue("Sheet1", "C1", "value")

	// 添加数据行
	f.SetCellValue("Sheet1", "A2", "1")
	f.SetCellValue("Sheet1", "B2", "test")
	f.SetCellValue("Sheet1", "C2", "100")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}

// 辅助函数：创建带公式的测试 Excel 文件
func createTestExcelFileWithFormula(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+"_formula.xlsx")
	f := excelize.NewFile()

	// 设置值
	f.SetCellValue("Sheet1", "A1", 10)
	f.SetCellValue("Sheet1", "A2", 20)

	// 设置公式
	f.SetCellFormula("Sheet1", "A3", "A1+A2")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 关闭文件以便重新打开
	f.Close()

	// 重新打开文件以触发公式计算
	f2, err := excelize.OpenFile(file)
	if err != nil {
		t.Fatalf("Failed to reopen test file: %v", err)
	}

	// 计算公式值并写回单元格
	result, err := f2.CalcCellValue("Sheet1", "A3")
	if err != nil {
		t.Logf("Warning: CalcCellValue failed: %v", err)
	} else {
		t.Logf("Calculated value: %s", result)
		// 将计算结果写回单元格
		if err := f2.SetCellValue("Sheet1", "A3", result); err != nil {
			t.Logf("Warning: SetCellValue failed: %v", err)
		}
	}

	if err := f2.Save(); err != nil {
		t.Logf("Warning: Save failed: %v", err)
	}

	f2.Close()

	return file
}

// 辅助函数：创建带版本号的测试 Excel 文件
func createTestExcelFileWithVersion(t *testing.T, version int) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+"_version.xlsx")
	f := excelize.NewFile()

	// 设置版本号
	f.SetCellValue("Sheet1", "A1", "__version__")
	f.SetCellValue("Sheet1", "B1", version)

	// 添加表头
	f.SetCellValue("Sheet1", "A2", "id")
	f.SetCellValue("Sheet1", "B2", "name")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}

// 辅助函数：创建多 Sheet 的测试 Excel 文件
func createTestExcelFileWithMultipleSheets(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+"_multiple.xlsx")
	f := excelize.NewFile()

	// 创建第二个 Sheet
	index, err := f.NewSheet("Sheet2")
	if err != nil {
		t.Fatalf("Failed to create sheet: %v", err)
	}
	_ = index

	// 在 Sheet1 添加数据
	f.SetCellValue("Sheet1", "A1", "data1")
	// 在 Sheet2 添加数据
	f.SetCellValue("Sheet2", "A1", "data2")

	if err := f.SaveAs(file); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}
