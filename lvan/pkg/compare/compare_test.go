package compare

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试比较目录功能
func TestCompareDirectories(t *testing.T) {
	// 创建测试目录结构
	baseDir, err := os.MkdirTemp("", "compare-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// 创建源目录和目标目录
	sourceDir := filepath.Join(baseDir, "source")
	targetDir := filepath.Join(baseDir, "target")

	err = os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("无法创建源目录: %v", err)
	}
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("无法创建目标目录: %v", err)
	}

	// 创建源目录中的文件
	sourceFiles := map[string]string{
		"common.txt":                   "相同内容",
		"different.txt":                "源内容",
		"only_in_source.txt":           "源独有",
		"json_files/source_order.json": `{"b": 2, "a": 1}`,
		"csv_files/source_order.csv":   "b,2\na,1",
	}

	// 创建目标目录中的文件
	targetFiles := map[string]string{
		"common.txt":                   "相同内容",
		"different.txt":                "目标内容",
		"only_in_target.txt":           "目标独有",
		"json_files/target_order.json": `{"a": 1, "b": 2}`,
		"csv_files/target_order.csv":   "a,1\nb,2",
	}

	// 写入源目录文件
	for path, content := range sourceFiles {
		fullPath := filepath.Join(sourceDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("无法创建目录: %v", err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("无法创建文件: %v", err)
		}
	}

	// 写入目标目录文件
	for path, content := range targetFiles {
		fullPath := filepath.Join(targetDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("无法创建目录: %v", err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("无法创建文件: %v", err)
		}
	}

	// 执行目录比较
	result, err := CompareDirectories(sourceDir, targetDir)
	if err != nil {
		t.Fatalf("比较目录出错: %v", err)
	}

	// 验证结果
	// 检查仅在源目录中的文件
	if !containsPath(result.OnlyInSource, "only_in_source.txt") {
		t.Errorf("未能检测到仅在源目录中的文件: only_in_source.txt")
	}

	// 检查仅在目标目录中的文件
	if !containsPath(result.OnlyInTarget, "only_in_target.txt") {
		t.Errorf("未能检测到仅在目标目录中的文件: only_in_target.txt")
	}

	// 检查内容不同的文件
	if !containsPath(result.DifferentContent, "different.txt") {
		t.Errorf("未能检测到内容不同的文件: different.txt")
	}

	// 检查内容相同的文件
	if !containsPath(result.SameContent, "common.txt") {
		t.Errorf("未能检测到内容相同的文件: common.txt")
	}

	// 检查JSON文件比较
	sourceJsonPath := filepath.Join(sourceDir, "json_files/source_order.json")
	targetJsonPath := filepath.Join(targetDir, "json_files/target_order.json")
	jsonEqual, err := compareJsonFiles(sourceJsonPath, targetJsonPath)
	if err != nil {
		t.Errorf("比较JSON文件出错: %v", err)
	}
	if !jsonEqual {
		t.Errorf("JSON文件比较失败，应当识别为相同内容但顺序不同")
	}

	// 检查CSV文件比较
	sourceCsvPath := filepath.Join(sourceDir, "csv_files/source_order.csv")
	targetCsvPath := filepath.Join(targetDir, "csv_files/target_order.csv")
	csvEqual, err := compareCsvFiles(sourceCsvPath, targetCsvPath)
	if err != nil {
		t.Errorf("比较CSV文件出错: %v", err)
	}
	if !csvEqual {
		t.Errorf("CSV文件比较失败，应当识别为相同内容但顺序不同")
	}
}

// 测试文件类型识别
func TestDetermineFileType(t *testing.T) {
	testCases := []struct {
		path     string
		expected FileType
	}{
		{"file.txt", FileTypeRegular},
		{"file.json", FileTypeJSON},
		{"file.csv", FileTypeCSV},
		{"file.JSON", FileTypeJSON}, // 测试大写扩展名
		{"file.CSV", FileTypeCSV},   // 测试大写扩展名
		{"file.tar.gz", FileTypeRegular},
		{"file", FileTypeRegular},
		{"path/to/config.json", FileTypeJSON},
	}

	for _, tc := range testCases {
		actual := DetermineFileType(tc.path)
		if actual != tc.expected {
			t.Errorf("文件类型识别错误 - 文件: %s, 期望: %v, 实际: %v", tc.path, tc.expected, actual)
		}
	}
}

// 测试对比两个文件
func TestCompareFiles(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "compare-files-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试场景
	testCases := []struct {
		name      string
		file1Data string
		file2Data string
		fileType  string
		expected  bool
	}{
		{"常规文件相同", "内容相同", "内容相同", "txt", true},
		{"常规文件不同", "内容A", "内容B", "txt", false},
		{"JSON文件-相同顺序", `{"a":1,"b":2}`, `{"a":1,"b":2}`, "json", true},
		{"JSON文件-不同顺序", `{"a":1,"b":2}`, `{"b":2,"a":1}`, "json", true},
		{"JSON文件-不同内容", `{"a":1,"b":2}`, `{"a":1,"b":3}`, "json", false},
		{"CSV文件-相同顺序", "a,1\nb,2", "a,1\nb,2", "csv", true},
		{"CSV文件-不同顺序", "a,1\nb,2", "b,2\na,1", "csv", true},
		{"CSV文件-不同内容", "a,1\nb,2", "a,1\nb,3", "csv", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建测试文件
			file1Path := filepath.Join(tempDir, "file1."+tc.fileType)
			file2Path := filepath.Join(tempDir, "file2."+tc.fileType)

			err := os.WriteFile(file1Path, []byte(tc.file1Data), 0644)
			if err != nil {
				t.Fatalf("创建测试文件1失败: %v", err)
			}
			err = os.WriteFile(file2Path, []byte(tc.file2Data), 0644)
			if err != nil {
				t.Fatalf("创建测试文件2失败: %v", err)
			}

			// 比较文件
			equal, err := CompareFiles(file1Path, file2Path)
			if err != nil {
				t.Fatalf("比较文件失败: %v", err)
			}

			if equal != tc.expected {
				t.Errorf("文件比较结果不符合预期 - 期望: %v, 实际: %v", tc.expected, equal)
			}
		})
	}
}

// 测试深度相等比较
func TestDeepEqual(t *testing.T) {
	testCases := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "基本类型相同",
			a:        123,
			b:        123,
			expected: true,
		},
		{
			name:     "基本类型不同",
			a:        123,
			b:        456,
			expected: false,
		},
		{
			name:     "不同类型",
			a:        123,
			b:        "123",
			expected: false,
		},
		{
			name:     "相同map顺序相同",
			a:        map[string]interface{}{"a": 1, "b": 2},
			b:        map[string]interface{}{"a": 1, "b": 2},
			expected: true,
		},
		{
			name:     "相同map顺序不同",
			a:        map[string]interface{}{"a": 1, "b": 2},
			b:        map[string]interface{}{"b": 2, "a": 1},
			expected: true,
		},
		{
			name:     "不同map",
			a:        map[string]interface{}{"a": 1, "b": 2},
			b:        map[string]interface{}{"a": 1, "b": 3},
			expected: false,
		},
		{
			name:     "相同数组顺序相同",
			a:        []interface{}{1, 2, 3},
			b:        []interface{}{1, 2, 3},
			expected: true,
		},
		{
			name:     "相同数组顺序不同",
			a:        []interface{}{1, 2, 3},
			b:        []interface{}{3, 2, 1},
			expected: true,
		},
		{
			name:     "不同数组",
			a:        []interface{}{1, 2, 3},
			b:        []interface{}{1, 2, 4},
			expected: false,
		},
		{
			name:     "嵌套结构",
			a:        map[string]interface{}{"a": 1, "b": []interface{}{1, 2}},
			b:        map[string]interface{}{"b": []interface{}{2, 1}, "a": 1},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DeepEqual(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("深度相等比较结果不符合预期 - 期望: %v, 实际: %v", tc.expected, result)
			}
		})
	}
}

// 辅助函数：检查路径是否在列表中
func containsPath(paths []string, path string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}
