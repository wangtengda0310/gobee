package config

import (
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
func Init(workDir string) {
	// 确保只初始化一次
	once.Do(func() {
		// 设置命令目录为工作目录下的cmd文件夹
		CommandDir = filepath.Join(workDir, "cmd")
		logger.Info("命令管理目录: %s", CommandDir)

		// 确保命令目录存在
		if err := os.MkdirAll(CommandDir, 0755); err != nil {
			panic("无法创建命令目录: " + err.Error())
		}
	})
}
