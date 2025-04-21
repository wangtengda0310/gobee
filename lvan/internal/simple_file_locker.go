package internal

// 使用文件来实现简单的资源竞争机制
// 1. 在workdir/tasks/任务下创建以{cmd}命名的文件夹(现有的业务{cmd}可能有不同的version请求)
// 2. 任务如果有配置meta并且指定了resources字段则遍历resources中的每个{resource}寻找workdir/tasks/{cmd}/{resource}文件
// 找到则代表资源被占用,尝试寻找resources中配置的其他资源
// 没找到则创建一个workdir/tasks/{cmd}/{resource}文件独占资源
// 任务执行完成后删除workdir/task/{cmd}/{resource}文件代表释放资源
// todo 如果是因为任务异常退出导致资源未释放,则会在workdir/tasks/{cmd}/{resource}文件中写入异常信息
// 3. meta的配置参考CommandMeta结构体,其中resources字段是一个字符串数组,代表需要竞争的资源
// 任务执行的逻辑参考pkg包

import (
	"errors"
	"github.com/gofrs/flock"
	"github.com/wangtengda/gobee/lvan/pkg/logger"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

func TryMeta(metaFile string) *CommandMeta {
	file, err := os.ReadFile(metaFile)
	if err != nil {
		return nil
	}
	meta := &CommandMeta{}
	err = yaml.Unmarshal(file, meta)
	if err != nil {
		logger.Warn("解析 meta.yaml 失败: %v %s", err, file)
		return nil
	}
	return meta
}

// 资源锁目录
var ResourceLockDir string

// ExclusiveOneResource 从CommandMeta中获取资源列表并尝试独占一个资源
// 如果没有可用资源则阻塞等待其他任务释放资源
func ExclusiveOneResource(resources []string, lockDir string, maxRetries int) (string, error, *flock.Flock) {
	logger.Info("随机选择资源 %v", resources)
	// 如果meta为空或没有配置资源，直接返回空字符串，表示不需要资源锁
	if len(resources) == 0 {
		return "", nil, nil
	}
	ResourceLockDir = lockDir

	// 创建命令资源锁目录
	cmdLockDir := lockDir
	os.MkdirAll(cmdLockDir, 0755)

	// 尝试获取资源锁
	for retry := 0; retry < maxRetries; retry++ {
		// 遍历资源列表，尝试获取一个可用资源
		for _, lockfile := range resources {

			// 构建资源锁文件路径
			lockFilePath := filepath.Join(cmdLockDir, lockfile)

			logger.Info("尝试获取资源锁: %s", lockfile)

			// 临界区代码（保证互斥）
			// 检查资源锁文件是否存在
			if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
				lock := flock.New(lockFilePath)
				ok, err := lock.TryLock()
				if err != nil {
					return "", err, nil
				}
				if !ok {
					// 其他进程持有锁
					return "", errors.New("其他进程持有锁"), nil
				}
				//defer lock.Unlock()

				// create file
				file, err := os.Create(lockFilePath)
				if err != nil {
					return "", err, nil
				}
				file.Close()

				logger.Info("成功获取资源锁: %s", lockfile)
				return lockfile, nil, lock
			}

			// 资源已被占用，尝试下一个资源
			logger.Debug("资源 %s 已被占用，尝试下一个资源", lockfile)
		}

		retryInterval := 5 * time.Second
		// 所有资源都被占用，等待一段时间后重试
		logger.Info("所有资源都被占用，等待 %v 后重试 (尝试 %d/%d)", retryInterval, retry+1, maxRetries)
		time.Sleep(retryInterval)
	}

	return "", errors.New("无法获取资源锁，所有资源都被占用且超过最大重试次数"), nil
}

// ReleaseResource 释放指定的资源锁
func ReleaseResource(resource string, lock *flock.Flock) error {
	if resource == "" {
		return nil // 没有资源需要释放
	}

	// 构建资源锁文件路径
	lockFilePath := filepath.Join(ResourceLockDir, resource)

	// 检查资源锁文件是否存在
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		return nil // 文件不存在，可能已被释放
	}

	// 删除资源锁文件
	if err := os.Remove(lockFilePath); err != nil {
		return err
	}

	return nil
}
