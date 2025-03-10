package config

import (
	"flag"
	"os"
	"path/filepath"
	"sync"

	"github.com/wangtengda/gobee/lvan/exporter/logger"
)

// 全局配置
var (
	// CommandDir 命令执行目录，默认为当前程序同目录下的cmd文件夹
	CommandDir string
	// 确保配置只初始化一次
	once sync.Once
)

// Init 初始化配置
func Init() {
	// 确保只初始化一次
	once.Do(func() {
		var execDir string
		// 优先使用环境变量
		if envDir := os.Getenv("COMMAND_DIR"); envDir != "" {
			execDir = envDir
			CommandDir = execDir
		} else if flag.Lookup("test.v") != nil || flag.Lookup("build") != nil {
			// 如果是测试环境，使用工作目录
			workDir, err := os.Getwd()
			if err != nil {
				panic("无法获取工作目录: " + err.Error())
			}
			execDir = workDir
			// 默认命令目录为当前程序同目录下的cmd文件夹
			CommandDir = filepath.Join(execDir, "cmd")
		} else {
			// 获取可执行文件路径
			execPath, err := os.Executable()
			if err != nil {
				panic("无法获取程序路径: " + err.Error())
			}
			execDir = filepath.Dir(execPath)
			// 默认命令目录为当前程序同目录下的cmd文件夹
			CommandDir = filepath.Join(execDir, "cmd")
		}

		logger.Info("命令管理目录: %s", CommandDir)
		// 确保命令目录存在
		if err := os.MkdirAll(CommandDir, 0755); err != nil {
			panic("无法创建命令目录: " + err.Error())
		}
	})
}
