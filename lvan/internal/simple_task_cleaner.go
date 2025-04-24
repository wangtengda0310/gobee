package internal

import (
	"os"
	"path/filepath"
	"time"

	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

func Create(f string) {

	// create file
	file, err := os.Create(f)
	if err != nil {
		logger.Warn("create file %s failed: %v", f, err)
		return
	}

	// 确保文件被关闭
	defer file.Close()

	// 向文件写入当前时间戳 人类友好的格式
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	_, err = file.WriteString(timeStr)
	if err != nil {
		logger.Warn("write file %s failed: %v", file.Name(), err)
	}
}

func durationAfterCreate(f string) time.Duration {

	// 获取文件创建时间
	fileInfo, err := os.Stat(f)
	if err != nil {
		logger.Warn("get file %s info failed: %v", f, err)
		return 0
	}

	// 计算文件创建时间到现在的时间差
	now := time.Now()
	createTime := fileInfo.ModTime()
	return now.Sub(createTime)
}

func ScheduleCleaner(dir string, duration time.Duration) {
	// 每隔duration时间执行一次markCleaner
	// 创建定时器，每隔duration时间执行一次markCleaner
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for range ticker.C {
		// 执行清理任务
		markCleaner(dir, duration)
	}

}
func markCleaner(dir string, duration time.Duration) {
	// 扫描dir目录下的子文件夹，子文件夹中包含名为"createtimestamp"的文件
	//,读取文件内容，计算文件创建时间到现在的时间差，如果文件读取失败尝试使用durationAfterCreate方法
	// 不包含"createtimestamp"的文件不使用Create方法创建
	// 如果时间差大于duration，删除子文件夹
	// 读取目录下的所有子目录
	files, err := os.ReadDir(dir)
	if err != nil {
		logger.Warn("read dir %s failed: %v", dir, err)
		return
	}

	// 遍历所有子目录
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		subDir := filepath.Join(dir, file.Name())
		timestampFile := filepath.Join(subDir, "createtimestamp")

		// 检查createtimestamp文件是否存在
		if _, err := os.Stat(timestampFile); os.IsNotExist(err) {
			// 如果文件不存在，创建该文件
			Create(timestampFile)
			continue
		}

		// 计算文件创建时间到现在的时间差
		d := durationAfterCreate(timestampFile)
		if d > duration {
			// 如果时间差大于duration，删除整个子目录
			logger.Info("remove dir %s", subDir)
			err := os.RemoveAll(subDir)
			if err != nil {
				logger.Warn("remove dir %s failed: %v", subDir, err)
			}
		}
	}

}
