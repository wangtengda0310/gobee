package workdir

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

var (
	WorkDir     string // 工作目录
	workDirFlag *string
)

func Pflag(getEnvString func(key string, defaultVal string) string) *string {
	workDirFlag = pflag.StringP("workdir", "w", getEnvString("EXPORTER_WORKDIR", ""), "指定工作目录，默认为程序所在目录，支持环境变量 EXPORTER_WORKDIR")
	return workDirFlag
}

func Join(string ...string) string {
	return filepath.Join(string...)
}
func Path() {

	// 初始化工作目录
	if *workDirFlag != "" {
		// 使用命令行参数指定的工作目录
		WorkDir = *workDirFlag
	} else {
		// 默认使用可执行文件所在目录
		execPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("无法获取程序路径: %v\n", err)
			os.Exit(1)
		}
		WorkDir = execPath
	}

	// 确保工作目录存在
	if err := os.MkdirAll(WorkDir, 0755); err != nil {
		fmt.Printf("无法创建工作目录: %v\n", err)
		os.Exit(1)
	}
}
