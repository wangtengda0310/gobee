// Package excel_internal 提供 Excel 文件读写功能
// 本包负责从文件系统读取 Excel 文件，支持单文件和目录批量读取
package excelio

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

// ErrSkipFilteredFile 表示该文件不在文件名过滤列表中（非打开失败、非 ~$ 临时文件）。
// 单列检查按需加载时，scanDirectory 会静默跳过此错误，避免误报「文件被 Excel 打开」。
var ErrSkipFilteredFile = errors.New("skipped by filename filter")

// ReadXlsx 读取单个 Excel 文件
//
// 此函数会跳过以下文件：
//   - Excel 临时文件（以 "~$" 开头）
//   - 不在文件名过滤列表中的文件（当提供了 sheetNamesFilter 时）
//
// 参数：
//   - path: Excel 文件路径
//   - sheetNamesFilter: 可选的文件名过滤列表
//
// 返回值：
//   - *excelize.File: Excel 文件对象
//   - error: 读取过程中的错误
//
// 注意：调用方负责关闭返回的文件对象
func ReadXlsx(path string, sheetNamesFilter ...string) (*excelize.File, error) {
	// 跳过 Excel 临时文件（Excel 打开时自动生成的临时文件）
	if strings.HasPrefix(filepath.Base(path), "~$") {
		return nil, fmt.Errorf("skipping Excel temp file: %s", path)
	}

	// 按需加载：仅打开过滤列表中的文件（返回 ErrSkipFilteredFile，由 scanDirectory 静默跳过）
	if sheetNamesFilter != nil && !slices.Contains(sheetNamesFilter, filepath.Base(path)) {
		return nil, fmt.Errorf("%w: %s", ErrSkipFilteredFile, path)
	}

	// 打开并读取 Excel 文件
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}

	// 注意：不在这里关闭文件，由调用方负责关闭
	return f, nil
}

// ReadFileOrDir 读取文件或目录，返回所有 Excel 文件对象
//
// 支持两种模式：
//   - 单文件模式：直接读取指定的 Excel 文件
//   - 目录模式：递归扫描目录下的所有 Excel 文件
//
// 参数：
//   - filePath: 文件或目录路径
//   - sheetNamesFilter: 可选的文件名过滤列表
//
// 返回值：
//   - []*excelize.File: Excel 文件对象列表
//   - error: 读取过程中的错误
//
// 注意：调用方负责关闭返回的所有文件对象
func ReadFileOrDir(filePath string, sheetNamesFilter ...string) ([]*excelize.File, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	var allXlsx = make([]*excelize.File, 0)

	if fileInfo.IsDir() {
		// 目录模式：递归扫描目录
		allXlsx, err = scanDirectory(filePath, allXlsx, sheetNamesFilter...)
		if err != nil {
			log.Printf("scanDir(%s) fail: %v", filePath, err)
			return nil, err
		}
	} else {
		// 单文件模式：直接读取文件
		xlsx, err := ReadXlsx(filePath, sheetNamesFilter...)
		if err != nil {
			return nil, err
		}
		allXlsx = append(allXlsx, xlsx)
	}

	return allXlsx, nil
}

// scanDirectory 递归扫描目录结构，收集所有 Excel 文件
//
// 参数：
//   - dirPath: 目录路径
//   - allXlsx: 已收集的 Excel 文件列表
//   - sheetNamesFilter: 可选的文件名过滤列表
//
// 返回值：
//   - []*excelize.File: 更新后的 Excel 文件列表
//   - error: 扫描过程中的错误
func scanDirectory(dirPath string, allXlsx []*excelize.File, sheetNamesFilter ...string) ([]*excelize.File, error) {
	// 读取目录下的所有条目
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return allXlsx, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 递归处理子目录
			entryPath := filepath.Join(dirPath, entry.Name())
			allXlsx, err = scanDirectory(entryPath, allXlsx, sheetNamesFilter...)
			if err != nil {
				log.Printf("readDir(%s) fail: %v", entryPath, err)
				continue // 跳过处理失败的目录
			}
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".xlsx") {
			// 处理 Excel 文件
			filePath := filepath.Join(dirPath, entry.Name())
			xlsx, err := ReadXlsx(filePath, sheetNamesFilter...)
			if err != nil {
				if errors.Is(err, ErrSkipFilteredFile) {
					continue // 按需加载时不在过滤列表，静默跳过
				}
				// 输出警告，提示文件被跳过（可能被 Excel 锁定）
				log.Printf("[警告] 跳过文件 %s: %v (文件可能被Excel或其他程序打开)", filePath, err)
				continue
			}
			allXlsx = append(allXlsx, xlsx)
		}
	}

	return allXlsx, nil
}

// ListXlsxFileNames 递归列出目录下所有 Excel 文件名（仅 basename，含 .xlsx 后缀）
// 跳过 Excel 临时文件（~$ 开头），不打开文件内容
func ListXlsxFileNames(dir string) ([]string, error) {
	var names []string
	err := listXlsxFileNamesRecursive(dir, &names)
	return names, err
}

func listXlsxFileNamesRecursive(dirPath string, names *[]string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := listXlsxFileNamesRecursive(filepath.Join(dirPath, entry.Name()), names); err != nil {
				return err
			}
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".xlsx") {
			continue
		}
		if strings.HasPrefix(name, "~$") {
			continue
		}
		*names = append(*names, name)
	}
	return nil
}

// ReadFileOrDirConcurrent 并发读取文件或目录，返回所有 Excel 文件对象
//
// 与 ReadFileOrDir 的区别：
//   - 使用并发模式读取，提高大目录的扫描速度
//   - 不支持文件名过滤（sheetNamesFilter）
//
// 参数：
//   - filePath: 文件或目录路径
//
// 返回值：
//   - []*excelize.File: Excel 文件对象列表
//   - error: 读取过程中的错误
//
// 注意：调用方负责关闭返回的所有文件对象
func ReadFileOrDirConcurrent(filePath string) ([]*excelize.File, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		// 目录模式：并发扫描
		return scanDirectoryConcurrent(filePath)
	} else {
		// 单文件模式：直接读取
		return readSingleFile(filePath)
	}
}

// scanDirectoryConcurrent 并发扫描目录结构
//
// 使用工作池模式并发读取 Excel 文件，提高性能：
//   - Worker 数量 = max(8, CPU 核心数)
//   - 文件通道缓冲区大小 = 100
//   - 错误通道缓冲区大小 = 10
//
// 参数：
//   - dirPath: 目录路径
//
// 返回值：
//   - []*excelize.File: Excel 文件对象列表
//   - error: 扫描过程中的错误
func scanDirectoryConcurrent(dirPath string) ([]*excelize.File, error) {
	var (
		allXlsx  []*excelize.File
		mu       sync.Mutex               // 保护 allXlsx 的并发访问
		wg       sync.WaitGroup           // 等待所有 Worker 完成
		errChan  = make(chan error, 10)   // 错误收集通道
		fileChan = make(chan string, 100) // 文件路径通道
		doneChan = make(chan bool)        // 错误收集完成信号
	)

	// 启动 Worker 池（Worker 数量根据 CPU 核心数动态调整，最少 8 个）
	workerCount := max(8, runtime.NumCPU())
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 从文件通道读取路径并处理
			for filePath := range fileChan {
				xlsx, err := ReadXlsx(filePath)
				if err != nil {
					errChan <- err
					continue
				}

				// 线程安全地添加到结果列表
				mu.Lock()
				allXlsx = append(allXlsx, xlsx)
				mu.Unlock()
			}
		}()
	}

	// 启动错误收集协程
	var errors []error
	go func() {
		for err := range errChan {
			errors = append(errors, err)
		}
		doneChan <- true
	}()

	// 遍历目录，将所有 Excel 文件路径发送到文件通道
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 将 .xlsx 文件路径发送到通道
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".xlsx") {
			fileChan <- path
		}
		return nil
	})

	// 关闭通道并等待所有 Worker 完成
	close(fileChan)
	wg.Wait()
	close(errChan)
	<-doneChan

	// 处理遍历错误
	if err != nil {
		return allXlsx, err
	}

	// 记录读取错误（如果有）
	if len(errors) > 0 {
		for _, e := range errors {
			log.Printf("读取文件失败: %v", e)
		}
		// 可以根据需求决定是否返回错误
		// 如果需要严格处理，可以返回错误
		// return allXlsx, fmt.Errorf("有%d个文件读取失败", len(errors))
	}

	return allXlsx, nil
}

// readSingleFile 读取单个 Excel 文件
//
// 参数：
//   - filePath: Excel 文件路径
//
// 返回值：
//   - []*excelize.File: 包含单个 Excel 文件对象的列表
//   - error: 读取过程中的错误
//
// 注意：调用方负责关闭返回的文件对象
func readSingleFile(filePath string) ([]*excelize.File, error) {
	xlsx, err := ReadXlsx(filePath)
	if err != nil {
		return nil, err
	}
	return []*excelize.File{xlsx}, nil
}
