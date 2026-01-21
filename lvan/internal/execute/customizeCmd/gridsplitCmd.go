package customizeCmd

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/cmd/exporter/customCmd"
	"github.com/wangtengda0310/gobee/lvan/pkg/gridsplit"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

func init() {
	customCmd.RegisterCommand("gridsplit", gridsplitCommand)
	customCmd.RegisterCommand("gs", gridsplitCommand)
}

// gridsplitCommand implements the gridsplit command for the exporter.
func gridsplitCommand(args []string) {
	flags := pflag.NewFlagSet("gridsplit", pflag.ExitOnError)

	mode := flags.StringP("mode", "m", "equal", "切分模式: equal, color, edge")
	rows := flags.IntP("rows", "r", 3, "行数 (0=自动检测)")
	cols := flags.IntP("cols", "c", 3, "列数 (0=自动检测)")
	lineColor := flags.String("line-color", "#FFFFFF", "分割线颜色")
	colorTolerance := flags.Int("color-tolerance", 30, "颜色容差 (0-255)")
	edgeThreshold := flags.Int("edge-threshold", 50, "边缘阈值 (0-255)")
	outputFormat := flags.StringP("format", "f", "png", "输出格式: png, jpg, jpeg")
	padding := flags.Int("padding", 0, "切分边距 (像素)")
	autoDetect := flags.BoolP("auto", "a", false, "自动检测网格")
	verbose := flags.BoolP("verbose", "v", false, "详细输出")
	help := flags.BoolP("help", "h", false, "显示帮助信息")

	if err := flags.Parse(args); err != nil {
		logger.Warn("参数解析失败: %v", err)
		return
	}

	if *help {
		printGridsplitHelp()
		return
	}

	// Get remaining arguments for input/output paths
	remaining := flags.Args()
	if len(remaining) == 0 {
		logger.Error("错误: 未指定输入路径")
		printGridsplitHelp()
		return
	}

	inputPath := remaining[0]
	outputPath := "output"
	if len(remaining) > 1 {
		outputPath = remaining[1]
	}

	// Parse split mode
	splitMode, err := gridsplit.ParseSplitMode(*mode)
	if err != nil {
		logger.Error("错误: %v", err)
		return
	}

	// Parse color
	var lineColorParsed color.Color
	if *lineColor != "" {
		parsed, err := parseColor(*lineColor)
		if err != nil {
			logger.Error("错误: 无效的颜色值: %v", err)
			return
		}
		lineColorParsed = parsed
	}

	// Create config
	config := gridsplit.SplitConfig{
		Mode:           splitMode,
		Rows:           *rows,
		Cols:           *cols,
		LineColor:      lineColorParsed,
		ColorTolerance: *colorTolerance,
		EdgeThreshold:  *edgeThreshold,
		OutputFormat:   *outputFormat,
		Padding:        *padding,
	}

	// Process input
	if err := processGridsplitInput(inputPath, outputPath, config, *autoDetect, *verbose); err != nil {
		logger.Error("处理失败: %v", err)
	}
}

func processGridsplitInput(inputPath, outputPath string, config gridsplit.SplitConfig, autoDetect, verbose bool) error {
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("无法访问路径 %s: %w", inputPath, err)
	}

	if fileInfo.IsDir() {
		if verbose {
			logger.Info("处理目录: %s", inputPath)
		}
		return processGridsplitDirectory(inputPath, outputPath, config, autoDetect, verbose)
	}

	if verbose {
		logger.Info("处理文件: %s", inputPath)
	}
	return processGridsplitFile(inputPath, outputPath, config, autoDetect, verbose)
}

func processGridsplitFile(inputPath, outputPath string, config gridsplit.SplitConfig, autoDetect, verbose bool) error {
	img, err := gridsplit.LoadImage(inputPath)
	if err != nil {
		return fmt.Errorf("加载图片失败: %w", err)
	}

	if autoDetect {
		detectedRows, detectedCols, err := gridsplit.DetectGridSize(img)
		if err == nil && detectedRows > 0 && detectedCols > 0 {
			if verbose {
				logger.Info("自动检测到网格: %d行 x %d列", detectedRows, detectedCols)
			}
			config.Rows = detectedRows
			config.Cols = detectedCols
		} else if verbose {
			logger.Warn("自动检测失败，使用配置的行列数")
		}
	}

	result, err := gridsplit.SplitImage(img, config)
	if err != nil {
		return fmt.Errorf("切分图片失败: %w", err)
	}

	if verbose {
		logger.Info("切分结果: %d行 x %d列 = %d张图片", result.Rows, result.Cols, len(result.Images))
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if err := gridsplit.SaveResult(result, outputPath, config.OutputFormat, baseName); err != nil {
		return fmt.Errorf("保存结果失败: %w", err)
	}

	logger.Info("已保存 %d 张图片到: %s", len(result.Images), outputPath)
	return nil
}

func processGridsplitDirectory(inputDir, outputDir string, config gridsplit.SplitConfig, autoDetect, verbose bool) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	processedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		inputPath := filepath.Join(inputDir, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ext)
		subOutputDir := filepath.Join(outputDir, baseName)

		if err := processGridsplitFile(inputPath, subOutputDir, config, autoDetect, verbose); err != nil {
			logger.Warn("处理 %s 失败: %v", entry.Name(), err)
			continue
		}

		processedCount++
	}

	if processedCount == 0 {
		logger.Info("未找到可处理的图片文件")
	} else {
		logger.Info("共处理 %d 个文件", processedCount)
	}

	return nil
}

func parseColor(colorStr string) (color.Color, error) {
	colorStr = strings.TrimSpace(colorStr)

	if strings.HasPrefix(colorStr, "#") {
		colorStr = colorStr[1:]
	}

	var r, g, b uint8

	if len(colorStr) == 6 {
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

func printGridsplitHelp() {
	logger.Info(`gridsplit - 九宫格图片切分工具

用法:
    gridsplit [选项] <输入路径> [输出路径]

选项:
    -m, --mode <模式>           切分模式: equal, color, edge (默认: equal)
    -r, --rows <行数>           行数，0=自动检测 (默认: 3)
    -c, --cols <列数>           列数，0=自动检测 (默认: 3)
    --line-color <颜色>         分割线颜色，格式: #RRGGBB (默认: #FFFFFF)
    --color-tolerance <值>      颜色容差 0-255 (默认: 30)
    --edge-threshold <值>       边缘阈值 0-255 (默认: 50)
    -f, --format <格式>         输出格式: png, jpg, jpeg (默认: png)
    --padding <像素>            切分边距 (默认: 0)
    -a, --auto                  自动检测网格
    -v, --verbose               详细输出
    -h, --help                  显示此帮助信息

示例:
    gridsplit input.png output/ -r 3 -c 3
    gridsplit input.png output/ --auto
    gridsplit input.png output/ --mode color --line-color "#FFFFFF"`)
}
