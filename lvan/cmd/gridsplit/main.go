package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/pkg/gridsplit"
)

//go:embed cli-doc.txt
var cliDoc string

const Version = "1.0.0"

func main() {
	// 定义命令行参数
	var (
		mode           string
		rows           int
		cols           int
		lineColor      string
		colorTolerance int
		edgeThreshold  int
		outputFormat   string
		padding        int
		autoDetect     bool
		verbose        bool
		dryRun         bool
		showVersion    bool
		showHelp       bool
	)

	pflag.StringVarP(&mode, "mode", "m", "equal", "切分模式: equal(等间距), color(颜色检测), edge(边缘检测)")
	pflag.IntVarP(&rows, "rows", "r", 3, "行数 (0=自动检测)")
	pflag.IntVarP(&cols, "cols", "c", 3, "列数 (0=自动检测)")
	pflag.StringVar(&lineColor, "line-color", "#FFFFFF", "分割线颜色 (颜色检测模式)")
	pflag.IntVar(&colorTolerance, "color-tolerance", 30, "颜色容差 (0-255, 颜色检测模式)")
	pflag.IntVar(&edgeThreshold, "edge-threshold", 50, "边缘阈值 (0-255, 边缘检测模式)")
	pflag.StringVarP(&outputFormat, "format", "f", "png", "输出格式: png, jpg, jpeg")
	pflag.IntVar(&padding, "padding", 0, "切分边距 (像素)")
	pflag.BoolVarP(&autoDetect, "auto", "a", false, "自动检测网格")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	pflag.BoolVar(&dryRun, "dry-run", false, "预览模式, 不实际保存")
	pflag.BoolVarP(&showVersion, "version", "V", false, "显示版本信息")
	pflag.BoolVarP(&showHelp, "help", "h", false, "显示帮助信息")

	pflag.Parse()

	// 显示版本信息
	if showVersion {
		fmt.Printf("gridsplit %s\n", Version)
		return
	}

	// 显示帮助信息
	if showHelp {
		fmt.Print(cliDoc)
		return
	}

	// 获取输入路径
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "错误: 未指定输入路径\n")
		fmt.Fprintf(os.Stderr, "使用 -h 或 --help 查看帮助信息\n")
		os.Exit(1)
	}

	inputPath := args[0]
	outputPath := "output" // 默认输出目录
	if len(args) > 1 {
		outputPath = args[1]
	}

	// 解析切分模式
	splitMode, err := gridsplit.ParseSplitMode(mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 解析颜色
	var lineColorParsed color.Color
	if lineColor != "" {
		lineColorParsed, err = parseColor(lineColor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的颜色值: %v\n", err)
			os.Exit(1)
		}
	}

	// 创建配置
	config := gridsplit.SplitConfig{
		Mode:           splitMode,
		Rows:           rows,
		Cols:           cols,
		LineColor:      lineColorParsed,
		ColorTolerance: colorTolerance,
		EdgeThreshold:  edgeThreshold,
		OutputFormat:   outputFormat,
		Padding:        padding,
	}

	// 处理输入
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法访问路径 %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	startTime := time.Now()

	if fileInfo.IsDir() {
		// 处理目录
		if verbose {
			fmt.Printf("处理目录: %s\n", inputPath)
			fmt.Printf("输出目录: %s\n", outputPath)
		}
		err = processDirectory(inputPath, outputPath, config, autoDetect, verbose, dryRun)
	} else {
		// 处理单个文件
		if verbose {
			fmt.Printf("处理文件: %s\n", inputPath)
		}
		err = processFile(inputPath, outputPath, config, autoDetect, verbose, dryRun)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	elapsedTime := time.Since(startTime)
	if verbose {
		fmt.Printf("完成! 耗时: %v\n", elapsedTime)
	}
}

func processFile(inputPath, outputPath string, config gridsplit.SplitConfig, autoDetect, verbose, dryRun bool) error {
	// 加载图片
	img, err := gridsplit.LoadImage(inputPath)
	if err != nil {
		return fmt.Errorf("加载图片失败: %w", err)
	}

	// 自动检测网格
	if autoDetect {
		detectedRows, detectedCols, err := gridsplit.DetectGridSize(img)
		if err == nil && detectedRows > 0 && detectedCols > 0 {
			if verbose {
				fmt.Printf("自动检测到网格: %d行 x %d列\n", detectedRows, detectedCols)
			}
			config.Rows = detectedRows
			config.Cols = detectedCols
		} else if verbose {
			fmt.Printf("警告: 自动检测失败，使用配置的行列数\n")
		}
	}

	// 切分图片
	result, err := gridsplit.SplitImage(img, config)
	if err != nil {
		return fmt.Errorf("切分图片失败: %w", err)
	}

	if verbose {
		fmt.Printf("切分结果: %d行 x %d列 = %d张图片\n", result.Rows, result.Cols, len(result.Images))
	}

	// 预览模式
	if dryRun {
		fmt.Printf("[预览模式] 将生成 %d 张图片到: %s\n", len(result.Images), outputPath)
		for i, img := range result.Images {
			bounds := img.Bounds()
			fmt.Printf("  [%d] %dx%d\n", i, bounds.Dx(), bounds.Dy())
		}
		return nil
	}

	// 获取基础文件名
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)

	// 保存结果
	if err := gridsplit.SaveResult(result, outputPath, config.OutputFormat, baseName); err != nil {
		return fmt.Errorf("保存结果失败: %w", err)
	}

	fmt.Printf("已保存 %d 张图片到: %s\n", len(result.Images), outputPath)
	return nil
}

func processDirectory(inputDir, outputDir string, config gridsplit.SplitConfig, autoDetect, verbose, dryRun bool) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	processedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 检查是否是图片文件
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		inputPath := filepath.Join(inputDir, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ext)
		subOutputDir := filepath.Join(outputDir, baseName)

		if err := processFile(inputPath, subOutputDir, config, autoDetect, verbose, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 处理 %s 失败: %v\n", entry.Name(), err)
			continue
		}

		processedCount++
	}

	if processedCount == 0 {
		fmt.Println("未找到可处理的图片文件")
	} else {
		fmt.Printf("共处理 %d 个文件\n", processedCount)
	}

	return nil
}

// parseColor 解析颜色字符串
func parseColor(colorStr string) (color.Color, error) {
	colorStr = strings.TrimSpace(colorStr)

	// 支持 #RRGGBB 格式
	if strings.HasPrefix(colorStr, "#") {
		colorStr = colorStr[1:]
	}

	var r, g, b uint8

	if len(colorStr) == 6 {
		// #RRGGBB 格式
		r64, err := strconv.ParseUint(colorStr[0:2], 16, 8)
		if err != nil {
			return nil, err
		}
		r = uint8(r64)
		g64, err := strconv.ParseUint(colorStr[2:4], 16, 8)
		if err != nil {
			return nil, err
		}
		g = uint8(g64)
		b64, err := strconv.ParseUint(colorStr[4:6], 16, 8)
		if err != nil {
			return nil, err
		}
		b = uint8(b64)
	} else if len(colorStr) == 3 {
		// #RGB 格式
		rVal, _ := strconv.ParseUint(string(colorStr[0:1]), 16, 8)
		gVal, _ := strconv.ParseUint(string(colorStr[1:2]), 16, 8)
		bVal, _ := strconv.ParseUint(string(colorStr[2:3]), 16, 8)
		r = uint8(rVal * 17)
		g = uint8(gVal * 17)
		b = uint8(bVal * 17)
	} else {
		return nil, fmt.Errorf("无效的颜色格式: %s", colorStr)
	}

	return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
}
