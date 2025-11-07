package main

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"

	"github.com/wangtengda0310/gobee/lvan/pkg/compare"
	"github.com/wangtengda0310/gobee/lvan/pkg/merkle"
)

// 嵌入CLI文档
//
//go:embed cli-doc.txt
var cliDoc string

// 版本信息
const (
	Version = "1.0.0"
)

func main() {
	// 定义命令行参数
	var (
		algorithm       string
		outputFormat    string
		excludePatterns []string
		includeEmptyDir bool
		parallelMode    bool
		maxWorkers      int
		verbose         bool
		showVersion     bool
		showHelp        bool
		compareMode     bool // 新增对比模式标志
		source          string
		target          string
	)

	// 设置命令行参数
	pflag.StringVarP(&algorithm, "hash", "H", "sha256", "哈希算法: md5, sha1, sha256, sha512, crc32")
	pflag.StringVarP(&outputFormat, "format", "f", "hex", "输出格式: hex, raw")
	pflag.StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "排除模式，可使用通配符，多个模式用空格分隔")
	pflag.BoolVarP(&includeEmptyDir, "include-empty", "i", false, "包含空目录")
	pflag.BoolVarP(&parallelMode, "parallel", "p", false, "使用并行模式")
	pflag.IntVarP(&maxWorkers, "workers", "w", 0, "最大工作线程数，0表示使用系统核心数")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "显示详细信息")
	pflag.BoolVarP(&showVersion, "version", "V", false, "显示版本信息")
	pflag.BoolVarP(&showHelp, "help", "h", false, "显示帮助信息")
	pflag.BoolVarP(&compareMode, "compare", "c", false, "比较两个目录")
	pflag.StringVarP(&source, "source", "s", "", "源目录路径")
	pflag.StringVarP(&target, "target", "t", "", "目标目录路径")

	// 解析命令行参数
	pflag.Parse()

	// 显示版本信息
	if showVersion {
		fmt.Printf("merkle %s\n", Version)
		return
	}

	// 显示帮助信息
	if showHelp {
		fmt.Print(cliDoc)
		return
	}

	// 转换哈希算法
	var hashAlgo merkle.HashAlgorithm
	switch algorithm {
	case "md5":
		hashAlgo = merkle.MD5
	case "sha1":
		hashAlgo = merkle.SHA1
	case "sha256":
		hashAlgo = merkle.SHA256
	case "sha512":
		hashAlgo = merkle.SHA512
	case "crc32":
		hashAlgo = merkle.CRC32
	default:
		fmt.Fprintf(os.Stderr, "错误: 不支持的哈希算法: %s\n", algorithm)
		os.Exit(1)
	}

	// 处理比较模式
	if compareMode {
		handleCompareMode(source, target, verbose)
		return
	}

	// 获取要处理的目录路径
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "错误: 未指定目录路径\n")
		fmt.Fprintf(os.Stderr, "使用 -h 或 --help 查看帮助信息\n")
		os.Exit(1)
	}

	dirPath := args[0]

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 目录不存在: %s\n", dirPath)
		os.Exit(1)
	}

	// 记录开始时间
	startTime := time.Now()

	// 根据运行模式选择计算方法
	var rootHashStr string
	var err error

	if parallelMode {
		// 使用并行模式
		rootHashStr, err = calculateMerkleHashParallel(dirPath, hashAlgo, excludePatterns, outputFormat)
	} else {
		// 使用串行模式
		rootHashStr, err = calculateMerkleHashSerial(dirPath, hashAlgo, excludePatterns, includeEmptyDir, outputFormat)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 输出哈希值
	fmt.Printf("%s\n", rootHashStr)

	// 计算执行时间
	elapsedTime := time.Since(startTime)

	// 如果启用了详细模式，输出详细信息
	if verbose {
		fmt.Println("\n详细信息:")
		fmt.Printf("算法: %s\n", hashAlgo)
		fmt.Printf("执行时间: %v\n", elapsedTime)
		fmt.Printf("并行模式: %v\n", parallelMode)
		if parallelMode && maxWorkers > 0 {
			fmt.Printf("线程数: %d\n", maxWorkers)
		}
	}
}

// handleCompareMode 处理目录比较模式
func handleCompareMode(source, target string, verbose bool) {
	// 检查源目录和目标目录是否指定
	if source == "" || target == "" {
		fmt.Fprintf(os.Stderr, "错误: 比较模式需要同时指定源目录和目标目录\n")
		fmt.Fprintf(os.Stderr, "使用 --source/-s 指定源目录，使用 --target/-t 指定目标目录\n")
		os.Exit(1)
	}

	// 检查源目录和目标目录是否存在
	if _, err := os.Stat(source); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 源目录不存在: %s\n", source)
		os.Exit(1)
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 目标目录不存在: %s\n", target)
		os.Exit(1)
	}

	// 比较目录
	result, err := compare.CompareDirectories(source, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "比较目录时发生错误: %v\n", err)
		os.Exit(1)
	}

	// 输出比较结果
	fmt.Println("===== 目录比较结果 =====")
	fmt.Printf("源目录: %s\n", source)
	fmt.Printf("目标目录: %s\n", target)
	fmt.Println()

	if len(result.OnlyInSource) > 0 {
		fmt.Println("只在源目录中存在的文件:")
		for _, file := range result.OnlyInSource {
			fmt.Printf("  %s\n", file)
		}
		fmt.Println()
	}

	if len(result.OnlyInTarget) > 0 {
		fmt.Println("只在目标目录中存在的文件:")
		for _, file := range result.OnlyInTarget {
			fmt.Printf("  %s\n", file)
		}
		fmt.Println()
	}

	if len(result.DifferentContent) > 0 {
		fmt.Println("内容不同的文件:")
		for _, file := range result.DifferentContent {
			fmt.Printf("  %s\n", file)
		}
		fmt.Println()
	}

	if verbose && len(result.SameContent) > 0 {
		fmt.Println("内容相同的文件:")
		for _, file := range result.SameContent {
			fmt.Printf("  %s\n", file)
		}
		fmt.Println()
	}

	if len(result.Errors) > 0 {
		fmt.Println("比较过程中发生的错误:")
		for file, errMsg := range result.Errors {
			fmt.Printf("  %s: %s\n", file, errMsg)
		}
	}

	// 打印统计信息
	fmt.Println("===== 统计信息 =====")
	fmt.Printf("总文件数: %d\n", len(result.OnlyInSource)+len(result.OnlyInTarget)+len(result.DifferentContent)+len(result.SameContent))
	fmt.Printf("只在源目录中: %d\n", len(result.OnlyInSource))
	fmt.Printf("只在目标目录中: %d\n", len(result.OnlyInTarget))
	fmt.Printf("内容不同: %d\n", len(result.DifferentContent))
	fmt.Printf("内容相同: %d\n", len(result.SameContent))
	fmt.Printf("处理错误: %d\n", len(result.Errors))
}

// calculateMerkleHashParallel 使用并行模式计算Merkle树哈希
func calculateMerkleHashParallel(dirPath string, hashAlgo merkle.HashAlgorithm, excludePatterns []string, outputFormat string) (string, error) {
	// 使用并行构建Merkle树方法
	root, rootHashStr, err := merkle.BuildMerkleTreeParallel(dirPath, hashAlgo, excludePatterns)
	if err != nil {
		return "", fmt.Errorf("构建Merkle树时出错: %v", err)
	}

	// 根据输出格式处理
	if outputFormat == "raw" {
		return string(root.Hash), nil
	}
	return rootHashStr, nil
}

// calculateMerkleHashSerial 使用串行模式计算Merkle树哈希
func calculateMerkleHashSerial(dirPath string, hashAlgo merkle.HashAlgorithm, excludePatterns []string, includeEmptyDir bool, outputFormat string) (string, error) {
	// 收集文件信息
	fileInfos, err := collectFileInfos(dirPath, hashAlgo, excludePatterns, includeEmptyDir)
	if err != nil {
		return "", err
	}

	// 构建Merkle树
	root, err := merkle.BuildMerkleTree(fileInfos, hashAlgo)
	if err != nil {
		return "", fmt.Errorf("构建Merkle树时出错: %v", err)
	}

	// 格式化输出根哈希值
	if outputFormat == "hex" {
		return hex.EncodeToString(root.Hash), nil
	} else if outputFormat == "raw" {
		return string(root.Hash), nil
	} else {
		return "", fmt.Errorf("不支持的输出格式: %s", outputFormat)
	}
}

// collectFileInfos 收集目录中所有文件的信息
func collectFileInfos(dirPath string, hashAlgo merkle.HashAlgorithm, excludePatterns []string, includeEmptyDir bool) ([]merkle.FileInfo, error) {
	var fileInfos []merkle.FileInfo

	// 遍历目录中的所有文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录自身
		if info.IsDir() && path != dirPath && !includeEmptyDir {
			return nil
		}

		// 检查是否应排除此文件
		if merkle.ShouldExclude(path, excludePatterns) {
			return nil
		}

		if !info.IsDir() {
			// 处理普通文件
			hash, err := merkle.CalculateFileHash(path, hashAlgo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 计算文件哈希出错 %s: %v\n", path, err)
				return nil
			}

			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 获取相对路径出错 %s: %v\n", path, err)
				return nil
			}

			// 确保使用统一的路径分隔符
			relPath = filepath.ToSlash(relPath)

			fileInfos = append(fileInfos, merkle.FileInfo{
				Path: relPath,
				Hash: hash,
			})
		} else if path != dirPath && includeEmptyDir {
			// 处理目录（只有在包含空目录选项开启时）
			handleDirectory(path, dirPath, excludePatterns, &fileInfos)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录时出错: %v", err)
	}

	return fileInfos, nil
}

// handleDirectory 处理目录（检查是否为空目录，并添加相应记录）
func handleDirectory(dirPath, basePath string, excludePatterns []string, fileInfos *[]merkle.FileInfo) {
	// 检查是否为空目录
	isEmpty := true
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 读取目录出错 %s: %v\n", dirPath, err)
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if !merkle.ShouldExclude(entryPath, excludePatterns) {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		relPath, err := filepath.Rel(basePath, dirPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: 获取相对路径出错 %s: %v\n", dirPath, err)
			return
		}

		// 确保使用统一的路径分隔符
		relPath = filepath.ToSlash(relPath)

		// 为空目录计算特殊哈希
		hash := calculateEmptyDirHash(relPath)
		*fileInfos = append(*fileInfos, merkle.FileInfo{
			Path: relPath + "/",
			Hash: hash,
		})
	}
}

// calculateEmptyDirHash 为空目录计算特殊哈希
func calculateEmptyDirHash(dirPath string) []byte {
	h := merkle.CalculateEmptyDirHash(dirPath)
	return h
}
