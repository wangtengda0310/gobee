package config

import (
	"os"
	"path/filepath"
	"sync"
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
		// 获取当前程序所在目录
		execPath, err := os.Executable()
		if err != nil {
			panic("无法获取程序路径: " + err.Error())
		}
		execDir := filepath.Dir(execPath)

		// 默认命令目录为当前程序同目录下的cmd文件夹
		CommandDir = filepath.Join(execDir, "cmd")

		// 确保命令目录存在
		if err := os.MkdirAll(CommandDir, 0755); err != nil {
			panic("无法创建命令目录: " + err.Error())
		}
	})
}