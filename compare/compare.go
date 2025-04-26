package compare

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// ComparisonResult 比较结果类型
type ComparisonResult struct {
	OnlyInSource     []string          // 仅存在于源目录的文件
	OnlyInTarget     []string          // 仅存在于目标目录的文件
	DifferentContent []string          // 内容不同的文件
	SameContent      []string          // 内容相同的文件
	Errors           map[string]string // 比较过程中的错误信息
}

// FileType 文件类型
type FileType int

const (
	FileTypeRegular FileType = iota // 普通文件
	FileTypeJSON                    // JSON文件
	FileTypeCSV                     // CSV文件
)

// DetermineFileType 确定文件类型
func DetermineFileType(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return FileTypeJSON
	case ".csv":
		return FileTypeCSV
	default:
		return FileTypeRegular
	}
}

// CompareDirectories 比较两个目录的内容
func CompareDirectories(sourceDir, targetDir string) (*ComparisonResult, error) {
	result := &ComparisonResult{
		Errors: make(map[string]string),
	}

	// 获取源目录和目标目录中的所有文件
	sourceFiles, err := getFilesRecursively(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("获取源目录文件错误: %v", err)
	}

	targetFiles, err := getFilesRecursively(targetDir)
	if err != nil {
		return nil, fmt.Errorf("获取目标目录文件错误: %v", err)
	}

	// 创建目标文件的映射，方便查找
	targetFilesMap := make(map[string]bool)
	for _, file := range targetFiles {
		targetFilesMap[file] = true
	}

	// 遍历源目录中的所有文件
	for _, sourceFile := range sourceFiles {
		// 获取相对路径
		relPath, err := filepath.Rel(sourceDir, sourceFile)
		if err != nil {
			result.Errors[sourceFile] = fmt.Sprintf("获取相对路径错误: %v", err)
			continue
		}

		// 构建目标文件的完整路径
		targetFile := filepath.Join(targetDir, relPath)

		// 检查目标文件是否存在
		if _, ok := targetFilesMap[targetFile]; !ok {
			// 文件仅存在于源目录
			result.OnlyInSource = append(result.OnlyInSource, filepath.ToSlash(relPath))
			continue
		}

		// 删除已处理的目标文件
		delete(targetFilesMap, targetFile)

		// 比较文件内容
		equal, err := compareFiles(sourceFile, targetFile)
		if err != nil {
			result.Errors[relPath] = fmt.Sprintf("比较文件内容错误: %v", err)
			continue
		}

		// 根据比较结果记录
		if equal {
			result.SameContent = append(result.SameContent, filepath.ToSlash(relPath))
		} else {
			result.DifferentContent = append(result.DifferentContent, filepath.ToSlash(relPath))
		}
	}

	// 处理仅存在于目标目录的文件
	for targetFile := range targetFilesMap {
		relPath, err := filepath.Rel(targetDir, targetFile)
		if err != nil {
			result.Errors[targetFile] = fmt.Sprintf("获取相对路径错误: %v", err)
			continue
		}
		result.OnlyInTarget = append(result.OnlyInTarget, filepath.ToSlash(relPath))
	}

	// 排序结果列表，确保输出顺序一致
	sort.Strings(result.OnlyInSource)
	sort.Strings(result.OnlyInTarget)
	sort.Strings(result.DifferentContent)
	sort.Strings(result.SameContent)

	return result, nil
}

// getFilesRecursively 递归获取目录中的所有文件
func getFilesRecursively(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录自身
		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// compareFiles 比较两个文件的内容
func compareFiles(file1, file2 string) (bool, error) {
	// 确定文件类型
	fileType := DetermineFileType(file1)

	// 根据文件类型选择适当的比较方法
	switch fileType {
	case FileTypeJSON:
		return compareJsonFiles(file1, file2)
	case FileTypeCSV:
		return compareCsvFiles(file1, file2)
	default:
		return compareRegularFiles(file1, file2)
	}
}

// compareRegularFiles 比较普通文件内容
func compareRegularFiles(file1, file2 string) (bool, error) {
	content1, err := os.ReadFile(file1)
	if err != nil {
		return false, fmt.Errorf("读取文件 %s 错误: %v", file1, err)
	}

	content2, err := os.ReadFile(file2)
	if err != nil {
		return false, fmt.Errorf("读取文件 %s 错误: %v", file2, err)
	}

	return bytes.Equal(content1, content2), nil
}

// compareJsonFiles 比较JSON文件内容，忽略键的顺序
func compareJsonFiles(file1, file2 string) (bool, error) {
	// 读取并解析第一个文件
	content1, err := os.ReadFile(file1)
	if err != nil {
		return false, fmt.Errorf("读取文件 %s 错误: %v", file1, err)
	}

	var data1 interface{}
	if err := json.Unmarshal(content1, &data1); err != nil {
		// 如果解析失败，则作为普通文件比较
		return compareRegularFiles(file1, file2)
	}

	// 读取并解析第二个文件
	content2, err := os.ReadFile(file2)
	if err != nil {
		return false, fmt.Errorf("读取文件 %s 错误: %v", file2, err)
	}

	var data2 interface{}
	if err := json.Unmarshal(content2, &data2); err != nil {
		// 如果解析失败，则作为普通文件比较
		return compareRegularFiles(file1, file2)
	}

	// 深度比较JSON对象，忽略顺序
	return deepEqual(data1, data2), nil
}

// compareCsvFiles 比较CSV文件内容，忽略行的顺序
func compareCsvFiles(file1, file2 string) (bool, error) {
	// 读取第一个CSV文件
	rows1, err := readCsvFile(file1)
	if err != nil {
		// 如果读取失败，则作为普通文件比较
		return compareRegularFiles(file1, file2)
	}

	// 读取第二个CSV文件
	rows2, err := readCsvFile(file2)
	if err != nil {
		// 如果读取失败，则作为普通文件比较
		return compareRegularFiles(file1, file2)
	}

	// 比较行数
	if len(rows1) != len(rows2) {
		return false, nil
	}

	// 转换为可比较的字符串集合
	rowStrings1 := make(map[string]bool)
	for _, row := range rows1 {
		rowStrings1[strings.Join(row, ",")] = true
	}

	// 检查第二个文件的每一行是否存在于第一个文件中
	for _, row := range rows2 {
		rowString := strings.Join(row, ",")
		if !rowStrings1[rowString] {
			return false, nil
		}
	}

	return true, nil
}

// readCsvFile 读取CSV文件内容
func readCsvFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var rows [][]string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, record)
	}

	return rows, nil
}

// deepEqual 深度比较两个对象，忽略映射键的顺序
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 获取底层类型
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// 如果类型不同，直接不等
	if va.Kind() != vb.Kind() {
		return false
	}

	// 根据类型进行不同的比较
	switch va.Kind() {
	case reflect.Map:
		// 检查大小
		if va.Len() != vb.Len() {
			return false
		}

		// 检查每个键值对
		for _, key := range va.MapKeys() {
			vbVal := vb.MapIndex(key)
			if !vbVal.IsValid() {
				// 键在a中存在但在b中不存在
				return false
			}
			if !deepEqual(va.MapIndex(key).Interface(), vbVal.Interface()) {
				// 键存在但值不相等
				return false
			}
		}
		return true

	case reflect.Slice, reflect.Array:
		// 检查大小
		if va.Len() != vb.Len() {
			return false
		}

		// 对于JSON数组，我们应该考虑顺序
		for i := 0; i < va.Len(); i++ {
			if !deepEqual(va.Index(i).Interface(), vb.Index(i).Interface()) {
				return false
			}
		}
		return true

	default:
		// 对于基本类型、指针等，使用标准库的Equal函数
		return reflect.DeepEqual(a, b)
	}
}
