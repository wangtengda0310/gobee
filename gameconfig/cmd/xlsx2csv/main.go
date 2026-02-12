// Excel to CSV 导出工具
//
// 用法:
//   xlsx2csv -source <目录> -target <目录>
//
// 示例:
//   xlsx2csv -source ./config -target ./config/csv
//
// 功能:
//   - 递归扫描 source 目录下的所有 .xlsx 文件
//   - 将每个 Excel 文件的每个 Sheet 导出为 CSV 文件
//   - CSV 文件组织方式: {target}/{Excel名}/{Sheet名}.csv
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wangtengda0310/gobee/gameconfig/internal/config"
)

var (
	// source 源目录（包含 Excel 文件）
	source string
	// target 目标目录（存放 CSV 文件）
	target string
	// version 显示版本号
	version bool
)

func init() {
	flag.StringVar(&source, "source", "", "源目录（包含 Excel 文件）")
	flag.StringVar(&target, "target", "", "目标目录（存放 CSV 文件）")
	flag.BoolVar(&version, "version", false, "显示版本号")
}

func main() {
	flag.Parse()

	// 显示版本号
	if version {
		fmt.Println("xlsx2csv version 1.0.0")
		fmt.Println("Go Module: github.com/wangtengda0310/gobee/gameconfig")
		os.Exit(0)
	}

	// 验证参数
	if source == "" || target == "" {
		fmt.Println("用法: xlsx2csv -source <目录> -target <目录>")
		fmt.Println("示例: xlsx2csv -source ./config -target ./config/csv")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 检查源目录是否存在
	if _, err := os.Stat(source); os.IsNotExist(err) {
		log.Fatalf("源目录不存在: %s", source)
	}

	// 执行导出
	count, err := exportDirectory(source, target)
	if err != nil {
		log.Fatalf("导出失败: %v", err)
	}

	fmt.Printf("导出完成！共处理 %d 个 Excel 文件\n", count)
}

// exportDirectory 导出目录下的所有 Excel 文件
func exportDirectory(sourceDir, targetDir string) (int, error) {
	count := 0

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理 .xlsx 文件
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".xlsx") {
			return nil
		}

		// 导出单个文件
		if err := exportFile(path, targetDir); err != nil {
			log.Printf("导出文件 %s 失败: %v", path, err)
			return nil // 继续处理其他文件
		}

		count++
		fmt.Printf("已导出: %s\n", path)
		return nil
	})

	return count, err
}

// exportFile 导出单个 Excel 文件
func exportFile(excelPath, targetDir string) error {
	exporter := config.NewExcelExporter(excelPath, targetDir)
	return exporter.Export()
}
