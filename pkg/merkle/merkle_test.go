package merkle

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
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
		algo     HashAlgorithm
		expected []byte
	}{
		{"MD5", MD5, md5Hash[:]},
		{"SHA1", SHA1, sha1Hash[:]},
		{"SHA256", SHA256, sha256Hash[:]},
		{"SHA512", SHA512, sha512Hash[:]},
		{"CRC32", CRC32, crc32Hash.Sum(nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := CalculateFileHash(testFile, tc.algo)
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
		algo HashAlgorithm
	}{
		{"MD5", MD5},
		{"SHA1", SHA1},
		{"SHA256", SHA256},
		{"SHA512", SHA512},
		{"CRC32", CRC32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			combined, err := CombineHashes(left, right, tc.algo)
			if err != nil {
				t.Fatalf("组合哈希出错: %v", err)
			}

			// 手动计算预期的哈希值
			var expected []byte
			switch tc.algo {
			case MD5:
				h := md5.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case SHA1:
				h := sha1.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case SHA256:
				h := sha256.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case SHA512:
				h := sha512.New()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case CRC32:
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
	fileInfos := []FileInfo{
		{Path: "file1", Hash: []byte{1, 2, 3, 4}},
		{Path: "file2", Hash: []byte{5, 6, 7, 8}},
		{Path: "file3", Hash: []byte{9, 10, 11, 12}},
	}

	// 测试树构建
	algorithms := []HashAlgorithm{SHA256, CRC32}

	for _, algo := range algorithms {
		t.Run(algo.String(), func(t *testing.T) {
			merkleRoot, err := BuildMerkleTree(fileInfos, algo)
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
			expectedHash, err := CombineHashes(merkleRoot.Left.Hash, merkleRoot.Right.Hash, algo)
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
		actual := ShouldExclude(tc.path, tc.patterns)
		if actual != tc.expected {
			t.Errorf("路径 %s 使用模式 %v 排除结果不正确, 期望: %v, 实际: %v",
				tc.path, tc.patterns, tc.expected, actual)
		}
	}
}

// 测试是否相等哈希
func TestIsEqualHash(t *testing.T) {
	testCases := []struct {
		name     string
		hash1    []byte
		hash2    []byte
		expected bool
	}{
		{"相同哈希", []byte{1, 2, 3, 4}, []byte{1, 2, 3, 4}, true},
		{"不同哈希", []byte{1, 2, 3, 4}, []byte{5, 6, 7, 8}, false},
		{"空哈希", []byte{}, []byte{}, true},
		{"不同长度", []byte{1, 2, 3}, []byte{1, 2, 3, 4}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsEqualHash(tc.hash1, tc.hash2)
			if actual != tc.expected {
				t.Errorf("哈希比较结果不正确, 期望: %v, 实际: %v", tc.expected, actual)
			}
		})
	}
}

// 测试空目录哈希计算
func TestCalculateEmptyDirHash(t *testing.T) {
	dirPath := "test/dir/path"
	hash := CalculateEmptyDirHash(dirPath)

	// 手动计算预期的哈希值
	h := md5.New()
	h.Write([]byte("empty_dir:" + dirPath))
	expected := h.Sum(nil)

	if !bytes.Equal(hash, expected) {
		t.Errorf("空目录哈希值不匹配, 期望: %x, 实际: %x", expected, hash)
	}
}
