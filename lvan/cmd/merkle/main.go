package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/tjfoc/gmsm/sm3"
	"golang.org/x/crypto/blake2b"
)

// 嵌入CLI文档
//
//go:embed cli-doc.txt
var cliDoc string

// 版本信息
const (
	Version = "1.0.0"
)

// 哈希算法类型
type HashAlgorithm string

const (
	MD5     HashAlgorithm = "md5"
	SHA1    HashAlgorithm = "sha1"
	SHA256  HashAlgorithm = "sha256"
	SHA512  HashAlgorithm = "sha512"
	BLAKE2B HashAlgorithm = "blake2b"
	CRC32   HashAlgorithm = "crc32" // 新增CRC32算法
	SM3     HashAlgorithm = "sm3"   // 新增SM3国密算法
)

// FileInfo 文件信息结构
type FileInfo struct {
	Path string
	Hash []byte
}

// MerkleNode Merkle树节点
type MerkleNode struct {
	Hash  []byte
	Left  *MerkleNode
	Right *MerkleNode
	Data  *FileInfo
}

// calculateFileHash 计算文件哈希
func calculateFileHash(path string, algo HashAlgorithm) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hasher io.Writer

	switch algo {
	case MD5:
		h := md5.New()
		hasher = h
	case SHA1:
		h := sha1.New()
		hasher = h
	case SHA256:
		h := sha256.New()
		hasher = h
	case SHA512:
		h := sha512.New()
		hasher = h
	case BLAKE2B:
		h, err := blake2b.New512(nil)
		if err != nil {
			return nil, err
		}
		hasher = h
	case CRC32:
		h := crc32.NewIEEE() // 使用IEEE多项式的CRC32实现
		hasher = h
	case SM3:
		h := sm3.New() // 使用SM3国密算法
		hasher = h
	default:
		return nil, fmt.Errorf("不支持的哈希算法: %s", algo)
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}

	if h, ok := hasher.(interface{ Sum(b []byte) []byte }); ok {
		return h.Sum(nil), nil
	}
	return nil, fmt.Errorf("哈希器未实现Sum方法")
}

// buildMerkleTree 构建Merkle树
func buildMerkleTree(fileInfos []*FileInfo, algo HashAlgorithm) (*MerkleNode, error) {
	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("没有文件可以构建Merkle树")
	}

	var nodes []*MerkleNode

	// 创建叶子节点
	for _, fileInfo := range fileInfos {
		node := &MerkleNode{
			Hash: fileInfo.Hash,
			Data: fileInfo,
		}
		nodes = append(nodes, node)
	}

	// 如果节点数量为奇数，复制最后一个节点
	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	// 自底向上构建树
	for len(nodes) > 1 {
		levelUp := make([]*MerkleNode, 0)

		// 确保每一层都有偶数个节点
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		for i := 0; i < len(nodes); i += 2 {
			node := &MerkleNode{
				Left:  nodes[i],
				Right: nodes[i+1],
			}
			combinedHash, err := combineHashes(nodes[i].Hash, nodes[i+1].Hash, algo)
			if err != nil {
				return nil, err
			}
			node.Hash = combinedHash
			levelUp = append(levelUp, node)
		}
		nodes = levelUp
	}

	return nodes[0], nil
}

// combineHashes 组合两个哈希值
func combineHashes(left, right []byte, algo HashAlgorithm) ([]byte, error) {
	var hasher io.Writer

	switch algo {
	case MD5:
		h := md5.New()
		hasher = h
	case SHA1:
		h := sha1.New()
		hasher = h
	case SHA256:
		h := sha256.New()
		hasher = h
	case SHA512:
		h := sha512.New()
		hasher = h
	case BLAKE2B:
		h, err := blake2b.New512(nil)
		if err != nil {
			return nil, err
		}
		hasher = h
	case CRC32:
		h := crc32.NewIEEE() // 使用IEEE多项式的CRC32实现
		hasher = h
	case SM3:
		h := sm3.New() // 使用SM3国密算法
		hasher = h
	default:
		return nil, fmt.Errorf("不支持的哈希算法: %s", algo)
	}

	if _, err := hasher.Write(left); err != nil {
		return nil, err
	}
	if _, err := hasher.Write(right); err != nil {
		return nil, err
	}

	if h, ok := hasher.(interface{ Sum(b []byte) []byte }); ok {
		return h.Sum(nil), nil
	}
	return nil, fmt.Errorf("哈希器未实现Sum方法")
}

// shouldExclude 判断路径是否应当被排除
func shouldExclude(path string, excludePatterns []string) bool {
	// 预处理: 分割空格分隔的模式
	var processedPatterns []string
	for _, pattern := range excludePatterns {
		// 处理空格分隔的多个模式
		if strings.Contains(pattern, " ") {
			parts := strings.Fields(pattern)
			processedPatterns = append(processedPatterns, parts...)
		} else {
			processedPatterns = append(processedPatterns, pattern)
		}
	}

	// 路径标准化为系统分隔符
	normalizedPath := filepath.ToSlash(path)

	for _, pattern := range processedPatterns {
		// 目录名称精确匹配 (例如 .git 匹配 /path/to/.git/config)
		if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
			// 检查是否作为目录名或文件名出现
			pathParts := strings.Split(normalizedPath, "/")
			for _, part := range pathParts {
				if part == pattern {
					return true
				}
			}

			// 直接路径包含检查
			if strings.Contains(normalizedPath, pattern) {
				return true
			}
		}

		// 尝试使用基本文件名匹配
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}

		// 尝试使用通配符模式匹配路径的各个部分
		if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
			pathParts := strings.Split(normalizedPath, "/")
			for _, part := range pathParts {
				if part == "" {
					continue
				}
				matched, err := filepath.Match(pattern, part)
				if err == nil && matched {
					return true
				}
			}
		}
	}
	return false
}

// 并行构建Merkle树
func buildMerkleTreeParallel(fileInfos []*FileInfo, algo HashAlgorithm) (*MerkleNode, error) {
	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("没有文件可以构建Merkle树")
	}

	var nodes []*MerkleNode

	// 创建叶子节点
	for _, fileInfo := range fileInfos {
		node := &MerkleNode{
			Hash: fileInfo.Hash,
			Data: fileInfo,
		}
		nodes = append(nodes, node)
	}

	// 如果节点数量为奇数，复制最后一个节点
	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	// 并发处理每一层的哈希合并
	for len(nodes) > 1 {
		levelUp := make([]*MerkleNode, (len(nodes)+1)/2)
		var wg sync.WaitGroup

		// 计算最大并发数 - 使用处理器核心数或者节点数的一半，取较小值
		maxGoroutines := runtime.NumCPU()
		nodePairs := len(nodes) / 2
		if maxGoroutines > nodePairs {
			maxGoroutines = nodePairs
		}

		// 使用通道控制并发数
		semaphore := make(chan struct{}, maxGoroutines)

		var errMutex sync.Mutex
		var firstErr error

		for i := 0; i < len(nodes); i += 2 {
			if i+1 >= len(nodes) {
				// 如果是最后一个孤立节点，复制它
				nodes = append(nodes, nodes[i])
			}

			wg.Add(1)
			semaphore <- struct{}{} // 获取令牌

			go func(i int) {
				defer func() {
					<-semaphore // 释放令牌
					wg.Done()
				}()

				// 创建新的父节点
				node := &MerkleNode{
					Left:  nodes[i],
					Right: nodes[i+1],
				}

				// 计算合并哈希值
				combinedHash, err := combineHashes(nodes[i].Hash, nodes[i+1].Hash, algo)
				if err != nil {
					errMutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMutex.Unlock()
					return
				}

				node.Hash = combinedHash
				levelUp[i/2] = node
			}(i)
		}

		wg.Wait()

		// 检查是否有错误发生
		if firstErr != nil {
			return nil, firstErr
		}

		nodes = levelUp
	}

	return nodes[0], nil
}

// 并行计算文件哈希
func calculateFileHashesParallel(files []string, dirPath string, algo HashAlgorithm, excludePatterns []string, includeEmptyDir bool, maxWorkers int) ([]*FileInfo, error) {
	// 如果未指定最大工作线程数，则使用CPU核心数作为默认值
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	// 创建工作池
	filesChan := make(chan string, len(files))
	resultsChan := make(chan *FileInfo, len(files))
	errorsChan := make(chan error, len(files))

	var wg sync.WaitGroup

	// 启动工作线程
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filesChan {
				fileInfo, err := os.Stat(path)
				if err != nil {
					errorsChan <- fmt.Errorf("无法获取文件信息 %s: %v", path, err)
					continue
				}

				if !fileInfo.IsDir() {
					// 计算文件哈希
					hash, err := calculateFileHash(path, algo)
					if err != nil {
						errorsChan <- fmt.Errorf("计算文件哈希出错 %s: %v", path, err)
						continue
					}

					relPath, err := filepath.Rel(dirPath, path)
					if err != nil {
						errorsChan <- fmt.Errorf("获取相对路径出错 %s: %v", path, err)
						continue
					}

					resultsChan <- &FileInfo{
						Path: relPath,
						Hash: hash,
					}
				} else if path != dirPath && includeEmptyDir {
					// 检查是否为空目录
					isEmpty := true
					entries, err := os.ReadDir(path)
					if err != nil {
						errorsChan <- fmt.Errorf("读取目录出错 %s: %v", path, err)
						continue
					}
					for _, entry := range entries {
						entryPath := filepath.Join(path, entry.Name())
						if !shouldExclude(entryPath, excludePatterns) {
							isEmpty = false
							break
						}
					}

					if isEmpty {
						relPath, err := filepath.Rel(dirPath, path)
						if err != nil {
							errorsChan <- fmt.Errorf("获取相对路径出错 %s: %v", path, err)
							continue
						}

						// 为空目录创建一个特殊标记的哈希
						var hasher io.Writer
						switch algo {
						case MD5:
							h := md5.New()
							hasher = h
						case SHA1:
							h := sha1.New()
							hasher = h
						case SHA256:
							h := sha256.New()
							hasher = h
						case SHA512:
							h := sha512.New()
							hasher = h
						case BLAKE2B:
							h, err := blake2b.New512(nil)
							if err != nil {
								errorsChan <- fmt.Errorf("创建BLAKE2B哈希器出错: %v", err)
								continue
							}
							hasher = h
						case CRC32:
							h := crc32.NewIEEE()
							hasher = h
						case SM3:
							h := sm3.New()
							hasher = h
						}

						if _, err := hasher.Write([]byte("empty_dir:" + relPath)); err != nil {
							errorsChan <- fmt.Errorf("计算空目录哈希出错 %s: %v", path, err)
							continue
						}

						if h, ok := hasher.(interface{ Sum(b []byte) []byte }); ok {
							resultsChan <- &FileInfo{
								Path: relPath + "/",
								Hash: h.Sum(nil),
							}
						}
					}
				}
			}
		}()
	}

	// 发送文件到通道
	for _, file := range files {
		filesChan <- file
	}
	close(filesChan)

	// 等待所有工作线程完成
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// 收集结果和错误
	var fileInfos []*FileInfo
	var errs []error

	for fileInfo := range resultsChan {
		fileInfos = append(fileInfos, fileInfo)
	}

	for err := range errorsChan {
		errs = append(errs, err)
	}

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return fileInfos, nil
}

// 主函数
func main() {
	// 设置程序说明
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Merkle树计算工具 v%s\n\n", Version)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  merkle [选项] <目录路径>\n\n")
		fmt.Fprintf(os.Stderr, "选项:\n")
		pflag.PrintDefaults()
	}

	// 解析命令行参数
	hashAlgo := pflag.String("hash", string(SHA256), "指定要使用的哈希算法 (md5, sha1, sha256, sha512, blake2b, crc32, sm3)")
	outputFile := pflag.StringP("output", "o", "", "指定输出文件路径（默认：标准输出）")
	var excludePatterns []string
	pflag.StringSliceVar(&excludePatterns, "exclude", nil, "指定要排除的文件或目录模式（可多次使用）")
	includeEmptyDir := pflag.Bool("include-empty-dir", false, "是否包含空目录")
	verbose := pflag.BoolP("verbose", "v", false, "显示详细输出")
	showHelp := pflag.BoolP("help", "h", false, "显示帮助信息")
	showVersion := pflag.Bool("version", false, "显示版本信息")
	maxWorkers := pflag.Int("workers", runtime.NumCPU(), "指定并发工作线程数量（默认为CPU核心数）")
	disableParallel := pflag.Bool("disable-parallel", false, "禁用并行计算（串行模式）")

	pflag.Parse()

	// 处理版本信息
	if *showVersion {
		fmt.Printf("Merkle树计算工具 version %s\n", Version)
		return
	}

	// 处理帮助信息
	if *showHelp {
		fmt.Println(cliDoc)
		return
	}

	// 检查目录参数
	args := pflag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "错误: 缺少目录路径参数\n\n")
		pflag.Usage()
		os.Exit(1)
	}

	dirPath := args[0]
	fileInfo, dirErr := os.Stat(dirPath)
	if dirErr != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法访问目录 %s: %v\n", dirPath, dirErr)
		os.Exit(1)
	}

	if !fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "错误: %s 不是一个目录\n", dirPath)
		os.Exit(1)
	}

	// 验证哈希算法
	algo := HashAlgorithm(strings.ToLower(*hashAlgo))
	switch algo {
	case MD5, SHA1, SHA256, SHA512, BLAKE2B, CRC32, SM3: // 添加SM3支持
		// 支持的算法
	default:
		fmt.Fprintf(os.Stderr, "错误: 不支持的哈希算法 %s\n", *hashAlgo)
		os.Exit(1)
	}

	// 设置输出流
	var output io.Writer = os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无法创建输出文件 %s: %v\n", *outputFile, err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	if *verbose {
		fmt.Fprintf(output, "开始计算目录 %s 的Merkle树 (使用 %s 哈希算法)...\n", dirPath, algo)
		fmt.Fprintf(output, "工作线程数: %d\n", *maxWorkers)
	}

	// 收集所有文件路径
	var filePaths []string
	walkErr := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过排除的文件和目录
		if shouldExclude(path, excludePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 添加所有文件或目录到列表
		filePaths = append(filePaths, path)
		return nil
	})

	if walkErr != nil {
		fmt.Fprintf(os.Stderr, "错误: 遍历目录时出错: %v\n", walkErr)
		os.Exit(1)
	}

	// 使用并行或串行方式计算文件哈希
	var files []*FileInfo
	if *disableParallel {
		// 使用原有的串行处理方法
		files = []*FileInfo{}
		for _, path := range filePaths {
			fileInfo, err := os.Stat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: 无法获取文件信息 %s: %v\n", path, err)
				continue
			}

			// 只处理常规文件
			if !fileInfo.IsDir() {
				if *verbose {
					fmt.Fprintf(output, "正在处理文件: %s\n", path)
				}

				hash, err := calculateFileHash(path, algo)
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误: 计算文件哈希时出错 %s: %v\n", path, err)
					continue
				}

				relPath, err := filepath.Rel(dirPath, path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误: 获取相对路径时出错 %s: %v\n", path, err)
					continue
				}

				files = append(files, &FileInfo{
					Path: relPath,
					Hash: hash,
				})
			} else if path != dirPath && *includeEmptyDir {
				// 原有的空目录处理逻辑...
				isEmpty := true
				entries, err := os.ReadDir(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "错误: 读取目录时出错 %s: %v\n", path, err)
					continue
				}
				for _, entry := range entries {
					entryPath := filepath.Join(path, entry.Name())
					if !shouldExclude(entryPath, excludePatterns) {
						isEmpty = false
						break
					}
				}

				if isEmpty {
					relPath, err := filepath.Rel(dirPath, path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "错误: 获取相对路径时出错 %s: %v\n", path, err)
						continue
					}

					// 为空目录创建一个特殊标记的哈希
					var hasher io.Writer
					switch algo {
					case MD5:
						h := md5.New()
						hasher = h
					case SHA1:
						h := sha1.New()
						hasher = h
					case SHA256:
						h := sha256.New()
						hasher = h
					case SHA512:
						h := sha512.New()
						hasher = h
					case BLAKE2B:
						h, err := blake2b.New512(nil)
						if err != nil {
							fmt.Fprintf(os.Stderr, "错误: 创建BLAKE2B哈希器时出错: %v\n", err)
							continue
						}
						hasher = h
					case CRC32:
						h := crc32.NewIEEE()
						hasher = h
					case SM3:
						h := sm3.New()
						hasher = h
					}

					if _, err := hasher.Write([]byte("empty_dir:" + relPath)); err != nil {
						fmt.Fprintf(os.Stderr, "错误: 计算空目录哈希时出错 %s: %v\n", path, err)
						continue
					}

					if h, ok := hasher.(interface{ Sum(b []byte) []byte }); ok {
						files = append(files, &FileInfo{
							Path: relPath + "/",
							Hash: h.Sum(nil),
						})
					}
				}
			}
		}
	} else {
		// 使用并行处理方法
		if *verbose {
			fmt.Fprintf(output, "使用 %d 个工作线程并行计算文件哈希...\n", *maxWorkers)
		}

		var hashErr error
		files, hashErr = calculateFileHashesParallel(filePaths, dirPath, algo, excludePatterns, *includeEmptyDir, *maxWorkers)
		if hashErr != nil {
			fmt.Fprintf(os.Stderr, "错误: 并行计算文件哈希时出错: %v\n", hashErr)
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "错误: 目录为空或所有文件都被排除了\n")
		os.Exit(1)
	}

	// 按路径排序，确保结果一致性
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	// 构建Merkle树
	startTime := time.Now() // 记录开始时间
	var merkleRoot *MerkleNode
	var treeErr error

	if *disableParallel {
		// 使用原有的串行方法构建树
		merkleRoot, treeErr = buildMerkleTree(files, algo)
	} else {
		// 使用并行方法构建树
		if *verbose {
			fmt.Fprintf(output, "使用并行方法构建Merkle树...\n")
		}
		merkleRoot, treeErr = buildMerkleTreeParallel(files, algo)
	}

	if treeErr != nil {
		fmt.Fprintf(os.Stderr, "错误: 构建Merkle树时出错: %v\n", treeErr)
		os.Exit(1)
	}

	elapsedTime := time.Since(startTime)

	// 输出结果
	fmt.Fprintf(output, "Merkle树根哈希值: %s\n", hex.EncodeToString(merkleRoot.Hash))

	if *verbose {
		fmt.Fprintf(output, "计算耗时: %v\n", elapsedTime)
		fmt.Fprintf(output, "\n文件哈希值:\n")
		for _, file := range files {
			fmt.Fprintf(output, "%s: %s\n", file.Path, hex.EncodeToString(file.Hash))
		}
	}
}
