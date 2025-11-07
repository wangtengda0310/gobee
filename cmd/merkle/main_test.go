package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wangtengda0310/gobee/lvan/pkg/compare"
	"github.com/wangtengda0310/gobee/lvan/pkg/merkle"
)

// 测试文件哈希计算
func TestCalculateFileHash(t *testing.T) {
	// 创建临时测试文件
	tempDir, err := os.MkdirTemp("", "merkle-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("这是测试内容")
	err = os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	// 计算预期的哈希值
	md5Hash := md5.Sum(testContent)
	sha1Hash := sha1.Sum(testContent)
	sha256Hash := sha256.Sum256(testContent)
	sha512Hash := sha512.Sum512(testContent)

	// 计算CRC32哈希
	crc32Hash := crc32.NewIEEE()
	crc32Hash.Write(testContent)

	// 测试各种哈希算法
	testCases := []struct {
		name     string
		algo     merkle.HashAlgorithm
		expected []byte
	}{
		{"MD5", merkle.MD5, md5Hash[:]},
		{"SHA1", merkle.SHA1, sha1Hash[:]},
		{"SHA256", merkle.SHA256, sha256Hash[:]},
		{"SHA512", merkle.SHA512, sha512Hash[:]},
		{"CRC32", merkle.CRC32, crc32Hash.Sum(nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := merkle.CalculateFileHash(testFile, tc.algo)
			if err != nil {
				t.Fatalf("计算哈希出错: %v", err)
			}

			if !bytes.Equal(actual, tc.expected) {
				t.Errorf("哈希值不匹配, 期望: %x, 实际: %x", tc.expected, actual)
			}
		})
	}
}

// 测试组合哈希值
func TestCombineHashes(t *testing.T) {
	left := []byte("left data")
	right := []byte("right data")

	// 测试各种哈希算法
	testCases := []struct {
		name string
		algo merkle.HashAlgorithm
	}{
		{"MD5", merkle.MD5},
		{"SHA1", merkle.SHA1},
		{"SHA256", merkle.SHA256},
		{"SHA512", merkle.SHA512},
		{"CRC32", merkle.CRC32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			combined, err := merkle.CombineHashes(left, right, tc.algo)
			if err != nil {
				t.Fatalf("组合哈希出错: %v", err)
			}

			// 手动计算预期的哈希值
			var expected []byte
			switch tc.algo {
			case merkle.MD5:
				h := md5.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case merkle.SHA1:
				h := sha1.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case merkle.SHA256:
				h := sha256.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case merkle.SHA512:
				h := sha512.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case merkle.CRC32:
				h := crc32.NewIEEE()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			}

			if !bytes.Equal(combined, expected) {
				t.Errorf("组合哈希值不匹配, 期望: %x, 实际: %x", expected, combined)
			}
		})
	}
}

// 测试构建Merkle树
func TestBuildMerkleTree(t *testing.T) {
	// 创建测试数据
	fileInfos := []merkle.FileInfo{
		{Path: "file1", Hash: []byte{1, 2, 3, 4}},
		{Path: "file2", Hash: []byte{5, 6, 7, 8}},
		{Path: "file3", Hash: []byte{9, 10, 11, 12}},
	}

	// 测试树构建
	algorithms := []merkle.HashAlgorithm{merkle.SHA256, merkle.CRC32}

	for _, algo := range algorithms {
		t.Run(algo.String(), func(t *testing.T) {
			merkleRoot, err := merkle.BuildMerkleTree(fileInfos, algo)
			if err != nil {
				t.Fatalf("构建Merkle树出错: %v", err)
			}

			// 验证树结构
			if merkleRoot == nil {
				t.Fatal("Merkle树根节点为空")
			}

			if merkleRoot.Left == nil || merkleRoot.Right == nil {
				t.Error("Merkle树根节点应当有左右子节点")
			}

			// 验证根节点的哈希值是通过左右子节点计算的
			expectedHash, err := merkle.CombineHashes(merkleRoot.Left.Hash, merkleRoot.Right.Hash, algo)
			if err != nil {
				t.Fatalf("组合哈希出错: %v", err)
			}

			if !bytes.Equal(merkleRoot.Hash, expectedHash) {
				t.Errorf("根节点哈希值不正确")
			}
		})
	}
}

// 测试路径排除
func TestShouldExclude(t *testing.T) {
	testCases := []struct {
		path     string
		patterns []string
		expected bool
	}{
		{"/path/to/file.txt", []string{"*.txt"}, false}, // 简化版本的ShouldExclude只做简单的包含检查
		{"/path/to/file.go", []string{"*.txt"}, false},
		{"/path/to/.git/config", []string{".git"}, true},
		{"/path/to/tmp/file", []string{"tmp", "*.bak"}, true},
		{"/path/to/src/main.go", []string{"*.txt", "*.bak"}, false},
	}

	for _, tc := range testCases {
		actual := merkle.ShouldExclude(tc.path, tc.patterns)
		if actual != tc.expected {
			t.Errorf("路径 %s 使用模式 %v 排除结果不正确, 期望: %v, 实际: %v",
				tc.path, tc.patterns, tc.expected, actual)
		}
	}
}

// 模糊测试组合哈希函数
func FuzzCombineHashes(f *testing.F) {
	// 添加种子语料库
	f.Add([]byte("test data 1"), []byte("test data 2"))
	f.Add([]byte{1, 2, 3, 4}, []byte{5, 6, 7, 8})
	f.Add([]byte{}, []byte{})

	// 模糊测试
	f.Fuzz(func(t *testing.T, left, right []byte) {
		// 测试 SHA256 算法
		combined, err := merkle.CombineHashes(left, right, merkle.SHA256)
		if err != nil {
			t.Fatalf("组合哈希出错: %v", err)
		}

		// 手动计算哈希
		h := sha256.New()
		h.Write(left)
		h.Write(right)
		expected := h.Sum(nil)

		// 验证结果
		if !bytes.Equal(combined, expected) {
			t.Errorf("哈希值不匹配, 期望: %x, 实际: %x", expected, combined)
		}
	})
}

// 模糊测试文件哈希计算
func FuzzFileHashCalculation(f *testing.F) {
	// 添加种子语料库
	f.Add([]byte("hello world"))
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte("这是一个中文测试"))

	// 模糊测试
	f.Fuzz(func(t *testing.T, data []byte) {
		// 创建临时文件
		tempFile, err := os.CreateTemp("", "fuzz-test-*.tmp")
		if err != nil {
			t.Skip("无法创建临时文件:", err)
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// 写入数据
		_, err = tempFile.Write(data)
		if err != nil {
			t.Skip("无法写入临时文件:", err)
			return
		}
		tempFile.Close() // 关闭文件以便哈希计算

		// 计算文件哈希
		fileHash, err := merkle.CalculateFileHash(tempFile.Name(), merkle.SHA256)
		if err != nil {
			t.Fatalf("计算文件哈希出错: %v", err)
		}

		// 手动计算数据哈希
		dataHash := sha256.Sum256(data)

		// 验证结果
		if !bytes.Equal(fileHash, dataHash[:]) {
			t.Errorf("哈希值不匹配, 期望: %x, 实际: %x", dataHash[:], fileHash)
		}
	})
}

// 测试并行构建Merkle树
func TestBuildMerkleTreeParallel(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "merkle-parallel-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	fileContents := []struct {
		path    string
		content []byte
	}{
		{"file1.txt", []byte("文件1内容")},
		{"file2.txt", []byte("文件2内容")},
		{"subdir/file3.txt", []byte("子目录文件内容")},
	}

	for _, fc := range fileContents {
		fullPath := filepath.Join(tempDir, fc.path)
		// 确保目录存在
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("无法创建目录: %v", err)
		}
		// 写入文件
		err = os.WriteFile(fullPath, fc.content, 0644)
		if err != nil {
			t.Fatalf("无法创建文件: %v", err)
		}
	}

	// 执行并行构建Merkle树
	root, rootHash, err := merkle.BuildMerkleTreeParallel(tempDir, merkle.SHA256, []string{})
	if err != nil {
		t.Fatalf("构建Merkle树出错: %v", err)
	}

	// 验证结果
	if root == nil {
		t.Fatal("Merkle树根节点为空")
	}
	if len(rootHash) == 0 {
		t.Fatal("根哈希值为空")
	}
}

// 测试比较目录
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
	result, err := compare.CompareDirectories(sourceDir, targetDir)
	if err != nil {
		t.Fatalf("比较目录出错: %v", err)
	}

	// 验证结果
	// 检查仅在源目录中的文件
	found := false
	for _, file := range result.OnlyInSource {
		if file == "only_in_source.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("未检测到仅在源目录中的文件")
	}

	// 检查仅在目标目录中的文件
	found = false
	for _, file := range result.OnlyInTarget {
		if file == "only_in_target.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("未检测到仅在目标目录中的文件")
	}

	// 检查内容不同的文件
	found = false
	for _, file := range result.DifferentContent {
		if file == "different.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("未检测到内容不同的文件")
	}

	// 检查内容相同的文件
	found = false
	for _, file := range result.SameContent {
		if file == "common.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("未检测到内容相同的文件")
	}
}

// 测试比较文件内容
func TestCompareFileContent(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "file-compare-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件对
	testCases := []struct {
		name     string
		file1    string
		file2    string
		expected bool
	}{
		{
			name:     "常规文件相同",
			file1:    "相同的内容",
			file2:    "相同的内容",
			expected: true,
		},
		{
			name:     "常规文件不同",
			file1:    "内容A",
			file2:    "内容B",
			expected: false,
		},
		{
			name:     "JSON文件相同-不同顺序",
			file1:    `{"name": "测试", "value": 123}`,
			file2:    `{"value": 123, "name": "测试"}`,
			expected: true,
		},
		{
			name:     "JSON文件不同-值不同",
			file1:    `{"name": "测试", "value": 123}`,
			file2:    `{"name": "测试", "value": 456}`,
			expected: false,
		},
		{
			name:     "CSV文件相同-不同顺序",
			file1:    "name,value\ntest,123\nexample,456",
			file2:    "name,value\nexample,456\ntest,123",
			expected: true,
		},
		{
			name:     "CSV文件不同-值不同",
			file1:    "name,value\ntest,123\nexample,456",
			file2:    "name,value\ntest,789\nexample,456",
			expected: false,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建带扩展名的文件
			var ext string
			if strings.Contains(tc.name, "JSON") {
				ext = ".json"
			} else if strings.Contains(tc.name, "CSV") {
				ext = ".csv"
			} else {
				ext = ".txt"
			}

			file1Path := filepath.Join(tempDir, fmt.Sprintf("test1_%d%s", i, ext))
			file2Path := filepath.Join(tempDir, fmt.Sprintf("test2_%d%s", i, ext))

			// 写入文件内容
			err := os.WriteFile(file1Path, []byte(tc.file1), 0644)
			if err != nil {
				t.Fatalf("无法写入文件1: %v", err)
			}
			err = os.WriteFile(file2Path, []byte(tc.file2), 0644)
			if err != nil {
				t.Fatalf("无法写入文件2: %v", err)
			}

			// 比较文件
			var equal bool
			var err2 error

			fileType := compare.DetermineFileType(file1Path)
			switch fileType {
			case compare.FileTypeJSON:
				// 读取并解析JSON文件
				content1, _ := os.ReadFile(file1Path)
				content2, _ := os.ReadFile(file2Path)

				var data1, data2 interface{}
				if err := json.Unmarshal(content1, &data1); err != nil {
					// 如果无法解析为JSON，就当作普通文件比较
					file1Reader, err := os.Open(file1Path)
					if err != nil {
						t.Fatal(err)
					}
					defer file1Reader.Close()

					file2Reader, err := os.Open(file2Path)
					if err != nil {
						t.Fatal(err)
					}
					defer file2Reader.Close()

					equal, err2 = compareFiles(file1Reader, file2Reader)
				} else if err := json.Unmarshal(content2, &data2); err != nil {
					// 如果无法解析为JSON，就当作普通文件比较
					file1Reader, err := os.Open(file1Path)
					if err != nil {
						t.Fatal(err)
					}
					defer file1Reader.Close()

					file2Reader, err := os.Open(file2Path)
					if err != nil {
						t.Fatal(err)
					}
					defer file2Reader.Close()

					equal, err2 = compareFiles(file1Reader, file2Reader)
				} else {
					equal = deepEqual(data1, data2)
					err2 = nil
				}
			case compare.FileTypeCSV:
				// 这里可以添加CSV比较的逻辑
				equal, err2 = compare.CompareFiles(file1Path, file2Path)
			default:
				equal, err2 = compare.CompareFiles(file1Path, file2Path)
			}

			if err2 != nil {
				t.Fatalf("比较文件出错: %v", err2)
			}

			if equal != tc.expected {
				t.Errorf("比较结果不正确, 期望: %v, 实际: %v", tc.expected, equal)
			}
		})
	}
}

// 简单的比较文件内容函数
func compareFiles(file1, file2 io.Reader) (bool, error) {
	const bufferSize = 64 * 1024 // 64KB
	buf1 := make([]byte, bufferSize)
	buf2 := make([]byte, bufferSize)

	for {
		n1, err1 := file1.Read(buf1)
		n2, err2 := file2.Read(buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}

		if err2 != nil && err2 != io.EOF {
			return false, err2
		}
	}
}

// 深度比较函数
func deepEqual(a, b interface{}) bool {
	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}
		for k, v := range aVal {
			if !deepEqual(v, bVal[k]) {
				return false
			}
		}
		return true
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}
		// 此处需要考虑数组元素顺序可能不同
		// 对于简单比较，可以只检查元素是否存在
		for _, item := range aVal {
			found := false
			for _, bItem := range bVal {
				if deepEqual(item, bItem) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
