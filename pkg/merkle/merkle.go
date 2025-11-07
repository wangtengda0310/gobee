package merkle

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// 哈希算法枚举
type HashAlgorithm int

const (
	// MD5 算法
	MD5 HashAlgorithm = iota
	// SHA1 算法
	SHA1
	// SHA256 算法
	SHA256
	// SHA512 算法
	SHA512
	// CRC32 算法
	CRC32
)

// String 返回哈希算法的字符串表示
func (a HashAlgorithm) String() string {
	switch a {
	case MD5:
		return "md5"
	case SHA1:
		return "sha1"
	case SHA256:
		return "sha256"
	case SHA512:
		return "sha512"
	case CRC32:
		return "crc32"
	default:
		return "unknown"
	}
}

// MerkleNode 表示Merkle树中的一个节点
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  []byte
	Path  string // 文件路径（如果是叶子节点）
}

// FileInfo 包含文件路径和对应的哈希值
type FileInfo struct {
	Path string
	Hash []byte
}

// newHash 根据指定的哈希算法创建一个新的哈希计算实例
func newHash(algo HashAlgorithm) (hash.Hash, error) {
	switch algo {
	case MD5:
		return md5.New(), nil
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case SHA512:
		return sha512.New(), nil
	case CRC32:
		return crc32.NewIEEE(), nil
	default:
		return nil, fmt.Errorf("不支持的哈希算法: %v", algo)
	}
}

// CalculateFileHash 计算文件的哈希值
func CalculateFileHash(filePath string, algo HashAlgorithm) ([]byte, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 创建哈希计算实例
	h, err := newHash(algo)
	if err != nil {
		return nil, err
	}

	// 计算哈希
	if _, err := io.Copy(h, file); err != nil {
		return nil, fmt.Errorf("计算哈希时出错: %w", err)
	}

	return h.Sum(nil), nil
}

// CombineHashes 组合两个哈希值为一个新的哈希值
func CombineHashes(left, right []byte, algo HashAlgorithm) ([]byte, error) {
	h, err := newHash(algo)
	if err != nil {
		return nil, err
	}

	h.Write(left)
	h.Write(right)
	return h.Sum(nil), nil
}

// BuildMerkleTree 从文件信息列表构建Merkle树
func BuildMerkleTree(fileInfos []FileInfo, algo HashAlgorithm) (*MerkleNode, error) {
	if len(fileInfos) == 0 {
		// 返回空数据的哈希
		h, err := newHash(algo)
		if err != nil {
			return nil, err
		}
		emptyHash := h.Sum(nil)
		return &MerkleNode{Hash: emptyHash}, nil
	}

	// 创建叶子节点
	var nodes []*MerkleNode
	for _, fileInfo := range fileInfos {
		node := &MerkleNode{
			Hash: fileInfo.Hash,
			Path: fileInfo.Path,
		}
		nodes = append(nodes, node)
	}

	// 构建树
	for len(nodes) > 1 {
		var level []*MerkleNode

		// 处理每一层的节点对
		for i := 0; i < len(nodes); i += 2 {
			// 如果是最后一个节点且总数为奇数，则复制该节点
			if i+1 == len(nodes) {
				combinedHash, err := CombineHashes(nodes[i].Hash, nodes[i].Hash, algo)
				if err != nil {
					return nil, err
				}
				parent := &MerkleNode{
					Left:  nodes[i],
					Right: nodes[i], // 使用同一个节点
					Hash:  combinedHash,
				}
				level = append(level, parent)
			} else {
				combinedHash, err := CombineHashes(nodes[i].Hash, nodes[i+1].Hash, algo)
				if err != nil {
					return nil, err
				}
				parent := &MerkleNode{
					Left:  nodes[i],
					Right: nodes[i+1],
					Hash:  combinedHash,
				}
				level = append(level, parent)
			}
		}

		nodes = level
	}

	return nodes[0], nil
}

// BuildMerkleTreeParallel 以并行方式构建Merkle树
func BuildMerkleTreeParallel(dirPath string, algo HashAlgorithm, excludePatterns []string) (*MerkleNode, string, error) {
	// 设置工作池大小
	workers := runtime.NumCPU()

	// 创建任务通道和结果通道
	filesChan := make(chan string, workers*2)
	resultsChan := make(chan FileInfo, workers*2)
	errorsChan := make(chan error, workers)

	// 创建WaitGroup以等待所有工作完成
	var wg sync.WaitGroup

	// 启动工作线程
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range filesChan {
				// 检查是否应该排除此文件
				if ShouldExclude(filePath, excludePatterns) {
					continue
				}

				// 获取文件信息
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					errorsChan <- fmt.Errorf("无法获取文件信息 %s: %w", filePath, err)
					continue
				}

				if fileInfo.IsDir() {
					continue
				}

				// 计算文件哈希
				hash, err := CalculateFileHash(filePath, algo)
				if err != nil {
					errorsChan <- fmt.Errorf("计算文件哈希出错 %s: %w", filePath, err)
					continue
				}

				// 获取相对路径
				relPath, err := filepath.Rel(dirPath, filePath)
				if err != nil {
					errorsChan <- fmt.Errorf("获取相对路径出错 %s: %w", filePath, err)
					continue
				}

				// 确保使用统一的路径分隔符
				relPath = filepath.ToSlash(relPath)

				// 发送结果
				resultsChan <- FileInfo{
					Path: relPath,
					Hash: hash,
				}
			}
		}()
	}

	// 收集所有文件路径
	go func() {
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 跳过目录
			if info.IsDir() {
				return nil
			}

			// 发送文件路径到通道
			filesChan <- path
			return nil
		})

		if err != nil {
			errorsChan <- fmt.Errorf("遍历目录时出错: %w", err)
		}

		// 关闭文件通道，表示没有更多文件
		close(filesChan)
	}()

	// 等待所有工作完成，然后关闭结果通道
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 收集所有文件信息
	var fileInfos []FileInfo
	for result := range resultsChan {
		fileInfos = append(fileInfos, result)
	}

	// 检查是否有错误
	select {
	case err := <-errorsChan:
		return nil, "", err
	default:
		// 继续处理
	}

	// 构建Merkle树
	root, err := BuildMerkleTree(fileInfos, algo)
	if err != nil {
		return nil, "", err
	}

	// 将哈希值转换为十六进制字符串
	hashStr := fmt.Sprintf("%x", root.Hash)

	return root, hashStr, nil
}

// ShouldExclude 检查文件路径是否应该被排除
func ShouldExclude(path string, patterns []string) bool {
	// 如果没有排除模式，则不排除任何文件
	if len(patterns) == 0 {
		return false
	}

	// 对每个模式进行检查
	for _, pattern := range patterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// IsEqualHash 比较两个哈希值是否相等
func IsEqualHash(hash1, hash2 []byte) bool {
	return bytes.Equal(hash1, hash2)
}

// CalculateEmptyDirHash 为空目录计算特殊哈希
func CalculateEmptyDirHash(dirPath string) []byte {
	h := md5.New()
	h.Write([]byte("empty_dir:" + dirPath))
	return h.Sum(nil)
}
