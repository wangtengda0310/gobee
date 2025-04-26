package merkle

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
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

// HashAlgorithm 表示哈希算法类型
type HashAlgorithm int

const (
	MD5 HashAlgorithm = iota
	SHA1
	SHA256
	SHA512
	CRC32
)

// FileInfo 表示文件信息，包含文件路径和哈希值
type FileInfo struct {
	Path string
	Hash []byte
}

// MerkleNode 表示Merkle树中的节点
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  []byte
}

// String 将哈希算法转换为字符串
func (h HashAlgorithm) String() string {
	switch h {
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

// GetHashFunction 根据哈希算法类型返回对应的哈希函数
func GetHashFunction(algorithm HashAlgorithm) (hash.Hash, error) {
	switch algorithm {
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
		return nil, fmt.Errorf("不支持的哈希算法: %v", algorithm)
	}
}

// CalculateFileHash 计算文件的哈希值
func CalculateFileHash(filePath string, algorithm HashAlgorithm) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hashFunc, err := GetHashFunction(algorithm)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(hashFunc, file); err != nil {
		return nil, err
	}

	return hashFunc.Sum(nil), nil
}

// ShouldExclude 判断路径是否应该被排除
func ShouldExclude(path string, excludes []string) bool {
	for _, exclude := range excludes {
		if strings.Contains(path, exclude) {
			return true
		}
	}
	return false
}

// CalculateFileHashesParallel 并行计算多个文件的哈希值
func CalculateFileHashesParallel(rootDir string, algorithm HashAlgorithm, excludes []string) ([]FileInfo, error) {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !ShouldExclude(path, excludes) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	fileInfos := make([]FileInfo, 0, len(files))
	numWorkers := runtime.NumCPU()
	filesChan := make(chan string, len(files))
	resultsChan := make(chan FileInfo, len(files))
	errChan := make(chan error, len(files))
	wg := sync.WaitGroup{}

	// 启动工作协程
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range filesChan {
				hash, err := CalculateFileHash(filePath, algorithm)
				if err != nil {
					errChan <- fmt.Errorf("计算文件 %s 的哈希值错误: %v", filePath, err)
					continue
				}
				resultsChan <- FileInfo{
					Path: filePath,
					Hash: hash,
				}
			}
		}()
	}

	// 发送文件到处理队列
	for _, file := range files {
		filesChan <- file
	}
	close(filesChan)

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errChan)
	}()

	// 收集结果和错误
	for fileInfo := range resultsChan {
		fileInfos = append(fileInfos, fileInfo)
	}

	// 检查错误
	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return fileInfos, fmt.Errorf("计算文件哈希值时发生错误: %s", strings.Join(errs, "; "))
	}

	return fileInfos, nil
}

// CombineHashes 组合两个哈希值
func CombineHashes(left, right []byte, algorithm HashAlgorithm) ([]byte, error) {
	hashFunc, err := GetHashFunction(algorithm)
	if err != nil {
		return nil, err
	}

	hashFunc.Write(left)
	hashFunc.Write(right)
	return hashFunc.Sum(nil), nil
}

// BuildMerkleTree 构建Merkle树
func BuildMerkleTree(fileInfos []FileInfo, algorithm HashAlgorithm) (*MerkleNode, error) {
	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("没有文件信息，无法构建Merkle树")
	}

	var nodes []*MerkleNode
	for _, fileInfo := range fileInfos {
		nodes = append(nodes, &MerkleNode{
			Hash: fileInfo.Hash,
		})
	}

	// 如果节点数为奇数，复制最后一个节点
	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	// 构建树
	for len(nodes) > 1 {
		levelSize := len(nodes)
		var nextLevel []*MerkleNode

		for i := 0; i < levelSize; i += 2 {
			left := nodes[i]
			right := nodes[i+1]

			combinedHash, err := CombineHashes(left.Hash, right.Hash, algorithm)
			if err != nil {
				return nil, err
			}

			node := &MerkleNode{
				Left:  left,
				Right: right,
				Hash:  combinedHash,
			}
			nextLevel = append(nextLevel, node)
		}

		nodes = nextLevel
		// 如果节点数为奇数，复制最后一个节点
		if len(nodes)%2 != 0 && len(nodes) > 1 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}
	}

	return nodes[0], nil
}

// BuildMerkleTreeParallel 并行构建Merkle树
func BuildMerkleTreeParallel(rootDir string, algorithm HashAlgorithm, excludes []string) (*MerkleNode, string, error) {
	fileInfos, err := CalculateFileHashesParallel(rootDir, algorithm, excludes)
	if err != nil {
		return nil, "", err
	}

	root, err := BuildMerkleTree(fileInfos, algorithm)
	if err != nil {
		return nil, "", err
	}

	// 将根哈希转换为十六进制字符串
	rootHashHex := hex.EncodeToString(root.Hash)
	return root, rootHashHex, nil
}
