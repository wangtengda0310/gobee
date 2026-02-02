package redis

import (
	"io"
	"os"
	"runtime"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressManager 进度条管理器
type ProgressManager struct {
	bar    *progressbar.ProgressBar
	isTTY  bool
	quiet  bool
	writer io.Writer
}

// NewProgressManager 创建进度条管理器
func NewProgressManager(total int64, description string) *ProgressManager {
	// 检测是否在 TTY 终端中
	isTTY := isTerminal()

	// 检测是否启用安静模式（CI/CD 环境）
	quiet := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""

	// 确定输出目标
	writer := os.Stderr
	if !isTTY {
		writer = nil // 非终端环境不显示进度条
	}

	pm := &ProgressManager{
		isTTY:  isTTY,
		quiet:  quiet,
		writer: writer,
	}

	// 创建进度条
	if !quiet && writer != nil {
		pm.bar = progressbar.NewOptions64(
			total,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSetWriter(writer),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("keys"),
			progressbar.OptionOnCompletion(func() {
				// 进度条完成时不自动换行，由外部控制
			}),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionFullWidth(),
		)
	}

	return pm
}

// NewQuietProgressManager 创建安静模式进度管理器（不显示进度条）
func NewQuietProgressManager() *ProgressManager {
	return &ProgressManager{
		quiet: true,
	}
}

// Add 增加进度
func (pm *ProgressManager) Add(delta int) {
	if pm.bar != nil && !pm.quiet {
		pm.bar.Add64(int64(delta))
	}
}

// AddOne 增加一个进度
func (pm *ProgressManager) AddOne() {
	if pm.bar != nil && !pm.quiet {
		pm.bar.Add64(1)
	}
}

// Set 设置当前进度
func (pm *ProgressManager) Set(current int) {
	if pm.bar != nil && !pm.quiet {
		pm.bar.Set64(int64(current))
	}
}

// Close 关闭进度条
func (pm *ProgressManager) Close() {
	if pm.bar != nil && !pm.quiet {
		pm.bar.Close()
		pm.bar = nil
	}
}

// Finish 完成进度条并显示完成消息
func (pm *ProgressManager) Finish() {
	if pm.bar != nil && !pm.quiet {
		pm.bar.Finish()
		pm.bar = nil
	}
}

// IsQuiet 是否为安静模式
func (pm *ProgressManager) IsQuiet() bool {
	return pm.quiet
}

// IsTTY 是否在终端中运行
func (pm *ProgressManager) IsTTY() bool {
	return pm.isTTY
}

// LogQuiet 在安静模式下输出日志
func (pm *ProgressManager) LogQuiet(format string, args ...interface{}) {
	if pm.quiet || !pm.isTTY {
		// 安静模式或非终端环境：输出普通日志
		// 在进度条模式下，避免日志输出干扰进度条
		// 这里选择不输出，由外部控制日志
	}
}

// isTerminal 检测是否在终端中运行
func isTerminal() bool {
	// 检查是否在 Windows 终端
	if runtime.GOOS == "windows" {
		// Windows: 检查是否有控制台
		// 简单检测：如果 stderr 是终端
	 fileInfo, _ := os.Stderr.Stat()
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}

	// Unix/Linux: 检查 /dev/tty
	if _, err := os.Stat("/dev/tty"); err == nil {
		return true
	}

	// 检查 stderr 是否是终端
	fileInfo, _ := os.Stderr.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
