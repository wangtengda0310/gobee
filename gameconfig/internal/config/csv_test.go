package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestCSVReader_Read 测试读取 CSV 文件
func TestCSVReader_Read(t *testing.T) {
	// 创建测试 CSV 文件
	testFile := createTestCSVFile(t)
	defer os.Remove(testFile)

	// 创建读取器
	reader := NewCSVReader(testFile)
	defer reader.Close()

	// 读取数据
	data, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Read() returned empty data")
	}

	// 验证表头
	if len(data) < 1 {
		t.Fatal("No header row")
	}
	header := data[0]
	if len(header) < 3 {
		t.Errorf("Header has %d columns, want at least 3", len(header))
	}

	// 验证数据行
	if len(data) < 2 {
		t.Fatal("No data rows")
	}
}

// TestCSVReader_SkipHeader 测试跳过表头行
func TestCSVReader_SkipHeader(t *testing.T) {
	// 创建测试 CSV 文件
	testFile := createTestCSVFile(t)
	defer os.Remove(testFile)

	// 创建读取器，设置跳过表头
	reader := NewCSVReaderWithOptions(testFile, CSVReaderOptions{
		SkipHeader: true,
	})
	defer reader.Close()

	// 读取数据
	data, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// 表头应该被跳过
	if len(data) < 1 {
		t.Fatal("No data rows after skipping header")
	}

	// 第一行应该是数据行，不是表头
	firstRow := data[0]
	if len(firstRow) > 0 && firstRow[0] == "id" {
		t.Error("Header was not skipped")
	}
}

// TestCSVReader_EmptyFile 测试读取空文件
func TestCSVReader_EmptyFile(t *testing.T) {
	// 创建空 CSV 文件
	testFile := filepath.Join(os.TempDir(), t.Name()+"_empty.csv")
	file, _ := os.Create(testFile)
	file.Close()
	defer os.Remove(testFile)

	// 创建读取器
	reader := NewCSVReader(testFile)
	defer reader.Close()

	// 读取数据
	data, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Empty file should return empty data, got %d rows", len(data))
	}
}

// TestCSVReader_NotExist 测试文件不存在的情况
func TestCSVReader_NotExist(t *testing.T) {
	// 创建读取器
	reader := NewCSVReader("not_exist.csv")
	defer reader.Close()

	// 读取应该返回错误
	_, err := reader.Read()
	if err == nil {
		t.Error("Read() should return error for non-existent file")
	}

	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("error type = %T, want ErrFileNotFound", err)
	}
}

// TestCSVReader_SpecialCharacters 测试包含特殊字符的 CSV
func TestCSVReader_SpecialCharacters(t *testing.T) {
	// 创建包含特殊字符的 CSV 文件
	testFile := filepath.Join(os.TempDir(), t.Name()+"_special.csv")
	content := `id,name,description
1,"test, with comma","line1
line2"
2,"quote""test","special: ; \t"`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// 创建读取器
	reader := NewCSVReader(testFile)
	defer reader.Close()

	// 读取数据
	data, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// 验证数据正确解析
	if len(data) < 3 {
		t.Fatalf("Expected 3 rows (header + 2 data), got %d", len(data))
	}

	// 检查第一行数据
	row1 := data[1]
	if len(row1) < 3 {
		t.Errorf("Row 1 has %d columns, want 3", len(row1))
	}
	if row1[1] != "test, with comma" {
		t.Errorf("Row 1 col 2 = %s, want 'test, with comma'", row1[1])
	}
}

// TestCSVReader_TypeInference 测试类型推断
func TestCSVReader_TypeInference(t *testing.T) {
	// 测试各种类型的值推断
	tests := []struct {
		value    string
		expected string
	}{
		{"123", "int"},
		{"123.45", "float"},
		{"true", "bool"},
		{"text", "string"},
		{"", "null"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			inferred := inferType(tt.value)
			if inferred != tt.expected {
				t.Errorf("inferType(%q) = %s, want %s", tt.value, inferred, tt.expected)
			}
		})
	}
}

// TestIsInt 测试整数检查
func TestIsInt(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"123", true},
		{"-123", true},
		{"0", true},
		{"12.3", false},
		{"abc", false},
		{"", false},
		{"-", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := isInt(tt.value); got != tt.want {
				t.Errorf("isInt(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestIsFloat 测试浮点数检查
func TestIsFloat(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"123.45", true},
		{"-123.45", true},
		{"0.5", true},
		{"123", false},
		{"abc", false},
		{"", false},
		{"-.", false},
		{"1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := isFloat(tt.value); got != tt.want {
				t.Errorf("isFloat(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// TestCSVReader_ReadWithComments 测试读取包含注释的 CSV
func TestCSVReader_ReadWithComments(t *testing.T) {
	// 创建包含注释行的 CSV 文件
	testFile := filepath.Join(os.TempDir(), t.Name()+"_comment.csv")
	content := `# This is a comment
id,name,value
1,test,100
2,another,200`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// 创建读取器，启用注释行支持
	reader := NewCSVReaderWithOptions(testFile, CSVReaderOptions{
		Comment: '#',
	})
	defer reader.Close()

	// 读取数据
	data, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// 注释行应该被跳过
	if len(data) < 2 {
		t.Errorf("Expected 2 rows (header + 1 data), got %d", len(data))
	}

	// 第一行应该是表头
	if data[0][0] != "id" {
		t.Errorf("First row = %v, want header", data[0])
	}
}

// 辅助函数：创建测试 CSV 文件
func createTestCSVFile(t *testing.T) string {
	t.Helper()

	file := filepath.Join(os.TempDir(), t.Name()+".csv")
	content := `id,name,value
1,test,100
2,another,200
3,third,300`

	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}
