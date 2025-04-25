package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/tjfoc/gmsm/sm3"
	"golang.org/x/crypto/blake2b"
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
	blake2bHash, _ := blake2b.New512(nil)
	blake2bHash.Write(testContent)

	// 计算CRC32哈希
	crc32Hash := crc32.NewIEEE()
	crc32Hash.Write(testContent)

	// 计算SM3哈希
	sm3Hash := sm3.New()
	sm3Hash.Write(testContent)

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
		{"BLAKE2B", BLAKE2B, blake2bHash.Sum(nil)},
		{"CRC32", CRC32, crc32Hash.Sum(nil)},
		{"SM3", SM3, sm3Hash.Sum(nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := calculateFileHash(testFile, tc.algo)
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
		{"BLAKE2B", BLAKE2B},
		{"CRC32", CRC32},
		{"SM3", SM3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			combined, err := combineHashes(left, right, tc.algo)
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
			case BLAKE2B:
				h, _ := blake2b.New512(nil)
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case CRC32:
				h := crc32.NewIEEE()
				h.Write(left)
				h.Write(right)
				expected = h.Sum(nil)
			case SM3:
				h := sm3.New()
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
	fileInfos := []*FileInfo{
		{Path: "file1", Hash: []byte{1, 2, 3, 4}},
		{Path: "file2", Hash: []byte{5, 6, 7, 8}},
		{Path: "file3", Hash: []byte{9, 10, 11, 12}},
	}

	// 测试树构建 - 包括CRC32和SM3
	algorithms := []HashAlgorithm{SHA256, CRC32, SM3}

	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			merkleRoot, err := buildMerkleTree(fileInfos, algo)
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

			// 第二层节点
			if merkleRoot.Left.Left == nil || merkleRoot.Left.Right == nil {
				t.Error("Merkle树第二层左节点应当有子节点")
			}

			// 叶子节点
			if merkleRoot.Left.Left.Data.Path != "file1" {
				t.Errorf("叶子节点数据不正确, 期望: file1, 实际: %s", merkleRoot.Left.Left.Data.Path)
			}

			if merkleRoot.Left.Right.Data.Path != "file2" {
				t.Errorf("叶子节点数据不正确, 期望: file2, 实际: %s", merkleRoot.Left.Right.Data.Path)
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
		{"/path/to/file.txt", []string{"*.txt"}, true},
		{"/path/to/file.go", []string{"*.txt"}, false},
		{"/path/to/.git/config", []string{".git"}, true},
		{"/path/to/tmp/file", []string{"tmp", "*.bak"}, true},
		{"/path/to/src/main.go", []string{"*.txt", "*.bak"}, false},
	}

	for _, tc := range testCases {
		actual := shouldExclude(tc.path, tc.patterns)
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
		combined, err := combineHashes(left, right, SHA256)
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

// 模糊测试文件哈希计算（只测试哈希算法的正确性，不创建实际文件）
func FuzzFileHashCalculation(f *testing.F) {
	// 添加种子语料库
	f.Add([]byte("test file content"))
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{})

	// 模糊测试
	f.Fuzz(func(t *testing.T, content []byte) {
		// 创建临时文件
		tempDir, err := os.MkdirTemp("", "merkle-fuzz-test")
		if err != nil {
			t.Skipf("无法创建临时目录: %v", err)
			return
		}
		defer os.RemoveAll(tempDir)

		testFile := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFile, content, 0644)
		if err != nil {
			t.Skipf("无法创建测试文件: %v", err)
			return
		}

		// 测试SHA256算法
		fileHash, err := calculateFileHash(testFile, SHA256)
		if err != nil {
			t.Fatalf("计算SHA256哈希出错: %v", err)
		}

		// 手动计算哈希
		sha256Hash := sha256.Sum256(content)
		expected := sha256Hash[:]

		// 验证结果
		if !bytes.Equal(fileHash, expected) {
			t.Errorf("SHA256哈希值不匹配, 期望: %x, 实际: %x", expected, fileHash)
		}

		// 测试CRC32算法
		crc32FileHash, err := calculateFileHash(testFile, CRC32)
		if err != nil {
			t.Fatalf("计算CRC32哈希出错: %v", err)
		}

		// 手动计算CRC32哈希
		crc32Hash := crc32.NewIEEE()
		crc32Hash.Write(content)
		crc32Expected := crc32Hash.Sum(nil)

		// 验证结果
		if !bytes.Equal(crc32FileHash, crc32Expected) {
			t.Errorf("CRC32哈希值不匹配, 期望: %x, 实际: %x", crc32Expected, crc32FileHash)
		}
	})
}

// 测试并行构建Merkle树
func TestBuildMerkleTreeParallel(t *testing.T) {
	// 创建测试数据 - 使用较多的文件以便并行处理有意义
	fileInfos := make([]*FileInfo, 100)
	for i := 0; i < 100; i++ {
		fileInfos[i] = &FileInfo{
			Path: fmt.Sprintf("file%d", i),
			Hash: []byte{byte(i), byte(i + 1), byte(i + 2), byte(i + 3)},
		}
	}

	// 测试不同算法的并行树构建
	algorithms := []HashAlgorithm{SHA256, MD5, CRC32, SM3}
	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			// 构建串行树
			serialRoot, err := buildMerkleTree(fileInfos, algo)
			if err != nil {
				t.Fatalf("构建串行Merkle树出错: %v", err)
			}

			// 构建并行树
			parallelRoot, err := buildMerkleTreeParallel(fileInfos, algo)
			if err != nil {
				t.Fatalf("构建并行Merkle树出错: %v", err)
			}

			// 验证两种方法生成的根哈希相同
			if !bytes.Equal(serialRoot.Hash, parallelRoot.Hash) {
				t.Errorf("并行和串行构建的Merkle树根哈希不匹配，算法=%s\n串行: %x\n并行: %x",
					algo, serialRoot.Hash, parallelRoot.Hash)
			}
		})
	}
}

// 测试并行计算文件哈希
func TestCalculateFileHashesParallel(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "merkle-parallel-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建多个测试文件
	var filePaths []string
	fileContents := [][]byte{
		[]byte("文件内容1"),
		[]byte("文件内容2"),
		[]byte("文件内容3较长一些，以获得不同的哈希值"),
		[]byte("文件内容4更长一些，再次确保不同的哈希值"),
		[]byte("文件内容5非常长，这样可以有更多的差异性"),
	}

	for i, content := range fileContents {
		filePath := filepath.Join(tempDir, fmt.Sprintf("testfile%d.txt", i))
		err = os.WriteFile(filePath, content, 0644)
		if err != nil {
			t.Fatalf("无法创建测试文件 %s: %v", filePath, err)
		}
		filePaths = append(filePaths, filePath)
	}

	// 创建一个空子目录
	emptyDir := filepath.Join(tempDir, "empty-dir")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("无法创建空目录: %v", err)
	}
	filePaths = append(filePaths, emptyDir)

	// 测试不同的工作线程数
	workerCounts := []int{1, 2, 4, runtime.NumCPU()}
	for _, workerCount := range workerCounts {
		t.Run(fmt.Sprintf("Workers=%d", workerCount), func(t *testing.T) {
			// 使用并行方法计算哈希
			results, err := calculateFileHashesParallel(
				filePaths,
				tempDir,
				SHA256,
				[]string{},
				true,
				workerCount,
			)
			if err != nil {
				t.Fatalf("并行计算哈希出错 (workers=%d): %v", workerCount, err)
			}

			// 验证结果数量
			expectedCount := len(fileContents) + 1 // 文件 + 空目录
			if len(results) != expectedCount {
				t.Errorf("结果数量不正确，期望: %d, 实际: %d", expectedCount, len(results))
			}

			// 验证各个文件的哈希值正确性
			for _, fileInfo := range results {
				// 跳过空目录
				if filepath.Ext(fileInfo.Path) != ".txt" {
					continue
				}

				// 提取文件索引
				var fileIdx int
				if _, err := fmt.Sscanf(filepath.Base(fileInfo.Path), "testfile%d.txt", &fileIdx); err != nil {
					t.Errorf("无法解析文件名 %s: %v", fileInfo.Path, err)
					continue
				}

				if fileIdx >= 0 && fileIdx < len(fileContents) {
					// 手动计算哈希
					expectedHash := sha256.Sum256(fileContents[fileIdx])
					if !bytes.Equal(fileInfo.Hash, expectedHash[:]) {
						t.Errorf("文件 %s 的哈希值不匹配，期望: %x, 实际: %x",
							fileInfo.Path, expectedHash[:], fileInfo.Hash)
					}
				}
			}
		})
	}
}

// 测试并行计算在大量文件下的并发安全性
func TestParallelSafety(t *testing.T) {
	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "merkle-concurrency-test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建大量小文件
	fileCount := 500
	var filePaths []string
	var wg sync.WaitGroup

	// 并行创建文件以加速测试准备
	semaphore := make(chan struct{}, runtime.NumCPU())
	for i := 0; i < fileCount; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			filePath := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", idx))
			content := []byte(fmt.Sprintf("Content for file %d with some randomness: %d", idx, time.Now().UnixNano()))
			if err := os.WriteFile(filePath, content, 0644); err == nil {
				filePaths = append(filePaths, filePath)
			}
		}(i)
	}
	wg.Wait()

	// 使用多个工作线程测试并发安全性
	t.Run("ConcurrencySafety", func(t *testing.T) {
		var results []*FileInfo
		var resultErr error

		// 使用所有可用CPU进行测试
		results, resultErr = calculateFileHashesParallel(
			filePaths,
			tempDir,
			SHA256,
			[]string{},
			false,
			runtime.NumCPU()*2, // 使用比CPU核心更多的工作线程来增加并发压力
		)

		if resultErr != nil {
			t.Fatalf("并行计算出错: %v", resultErr)
		}

		// 验证结果数量
		if len(results) == 0 {
			t.Error("未返回任何结果")
		}

		// 检查是否有重复的文件路径
		pathMap := make(map[string]bool)
		for _, result := range results {
			if pathMap[result.Path] {
				t.Errorf("文件路径 %s 重复出现在结果中", result.Path)
			}
			pathMap[result.Path] = true
		}
	})
}

// 模糊测试 - 并行构建Merkle树
func FuzzBuildMerkleTreeParallel(f *testing.F) {
	// 添加种子语料库
	f.Add([]byte("test data 1"), []byte("test data 2"), []byte("test data 3"), 10)
	f.Add([]byte{1, 2, 3}, []byte{4, 5, 6}, []byte{7, 8, 9}, 4)

	// 模糊测试
	f.Fuzz(func(t *testing.T, data1, data2, data3 []byte, workerCount int) {
		// 修正工作线程数范围 (1-16)
		if workerCount < 1 {
			workerCount = 1
		}
		if workerCount > 16 {
			workerCount = 16
		}

		// 创建测试数据
		fileInfos := []*FileInfo{
			{Path: "file1", Hash: data1},
			{Path: "file2", Hash: data2},
			{Path: "file3", Hash: data3},
		}

		// 使用不同的哈希算法
		algos := []HashAlgorithm{SHA256, MD5}

		for _, algo := range algos {
			// 串行构建
			serialRoot, err := buildMerkleTree(fileInfos, algo)
			if err != nil {
				continue // 跳过无效输入
			}

			// 并行构建
			parallelRoot, err := buildMerkleTreeParallel(fileInfos, algo)
			if err != nil {
				t.Errorf("并行构建出错，但串行成功: %v", err)
				continue
			}

			// 验证两种方法生成的根哈希相同
			if !bytes.Equal(serialRoot.Hash, parallelRoot.Hash) {
				t.Errorf("并行和串行构建的Merkle树根哈希不匹配，算法=%s", algo)
			}
		}
	})
}

// 性能基准测试 - 比较串行和并行构建树
func BenchmarkBuildMerkleTree(b *testing.B) {
	// 创建大量测试数据 (1000个文件)
	fileCount := 1000
	fileInfos := make([]*FileInfo, fileCount)
	for i := 0; i < fileCount; i++ {
		hash := make([]byte, 32)
		for j := 0; j < len(hash); j++ {
			hash[j] = byte((i * j) % 256)
		}
		fileInfos[i] = &FileInfo{
			Path: fmt.Sprintf("file%d", i),
			Hash: hash,
		}
	}

	// 不同的哈希算法
	algos := []HashAlgorithm{SHA256, MD5, CRC32, SM3}

	for _, algo := range algos {
		// 串行构建基准测试
		b.Run(fmt.Sprintf("Serial-%s", algo), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := buildMerkleTree(fileInfos, algo)
				if err != nil {
					b.Fatalf("构建出错: %v", err)
				}
			}
		})

		// 不同工作线程的并行构建基准测试
		workerCounts := []int{1, 2, 4, 8, runtime.NumCPU()}
		for _, count := range workerCounts {
			b.Run(fmt.Sprintf("Parallel-%s-Workers%d", algo, count), func(b *testing.B) {
				runtime.GOMAXPROCS(count) // 限制可用线程数
				for i := 0; i < b.N; i++ {
					_, err := buildMerkleTreeParallel(fileInfos, algo)
					if err != nil {
						b.Fatalf("构建出错: %v", err)
					}
				}
				runtime.GOMAXPROCS(0) // 重置为默认值
			})
		}
	}
}

// 性能基准测试 - 比较串行和并行文件哈希计算
func BenchmarkFileHashCalculation(b *testing.B) {
	// 跳过实际运行，仅用于说明
	b.Skip("此基准测试需要文件系统操作，默认跳过。移除此行以启用测试。")

	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "merkle-bench")
	if err != nil {
		b.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	fileCount := 100
	fileSizes := []int{1024, 10240, 102400} // 1KB, 10KB, 100KB

	for _, size := range fileSizes {
		// 创建测试子目录
		subDir := filepath.Join(tempDir, fmt.Sprintf("size-%d", size))
		if err := os.Mkdir(subDir, 0755); err != nil {
			b.Fatalf("无法创建子目录: %v", err)
		}

		// 创建指定大小的测试文件
		var filePaths []string
		content := make([]byte, size)
		for i := 0; i < size; i++ {
			content[i] = byte(i % 256)
		}

		for i := 0; i < fileCount; i++ {
			filePath := filepath.Join(subDir, fmt.Sprintf("file%d.dat", i))
			if err := os.WriteFile(filePath, content, 0644); err != nil {
				b.Fatalf("无法创建测试文件: %v", err)
			}
			filePaths = append(filePaths, filePath)
		}

		// 串行计算基准测试
		b.Run(fmt.Sprintf("Serial-Size%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var results []*FileInfo
				for _, path := range filePaths {
					hash, err := calculateFileHash(path, SHA256)
					if err != nil {
						b.Fatalf("计算哈希出错: %v", err)
					}
					relPath, _ := filepath.Rel(tempDir, path)
					results = append(results, &FileInfo{
						Path: relPath,
						Hash: hash,
					})
				}
				if len(results) != fileCount {
					b.Fatalf("结果数量不正确: %d", len(results))
				}
			}
		})

		// 不同工作线程的并行计算基准测试
		workerCounts := []int{1, 2, 4, 8, runtime.NumCPU()}
		for _, count := range workerCounts {
			b.Run(fmt.Sprintf("Parallel-Size%d-Workers%d", size, count), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					results, err := calculateFileHashesParallel(
						filePaths,
						tempDir,
						SHA256,
						[]string{},
						false,
						count,
					)
					if err != nil {
						b.Fatalf("并行计算出错: %v", err)
					}
					if len(results) != fileCount {
						b.Fatalf("结果数量不正确: %d", len(results))
					}
				}
			})
		}
	}
}
