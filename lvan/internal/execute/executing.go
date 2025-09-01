package execute

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
	"github.com/wangtengda0310/gobee/lvan/pkg/utf8"
)

type executing struct {
	env        []string
	encoding   utf8.Charset
	resource   string
	versionDir string
	task       *Task
}

func (execute *executing) outputTask(s string) {
	execute.task.AddOutput(s)
}

func (execute *executing) appendEnv(cmdName string, task *Task) {
	// 添加自定义环境变量
	execute.env = append(execute.env, fmt.Sprintf("exporter_cmd_%s_resource=%s", cmdName, execute.resource))

	execute.env = append(execute.env, fmt.Sprintf("exporter_cmd_%s_id=%s", cmdName, task.ID))

	// 设置环境变量
	if len(task.Request.Env) > 0 {

		// 添加自定义环境变量
		for key, value := range task.Request.Env {
			execute.env = append(execute.env, fmt.Sprintf("%s=%s", key, value))
			logger.Debug("设置环境变量: %s=%s", key, value)
		}
	}

}

// 返回 defer函数
func (execute *executing) excludeResource(task *Task, resources []string) func() {

	// 获取当前环境变量
	if len(resources) <= 0 {
		return func() {}
	}

	retries := 40
	logger.Info("默认重试次数为 %d 可以通过环境变量 exporter_retry_times 设置", retries)
	if os.Getenv("exporter_retry_times") != "" {
		retry, err := strconv.Atoi(os.Getenv("exporter_retry_times"))
		if err == nil {
			logger.Info("使用环境变量 exporter_retry_times 设置重试次数为 %d", retry)
			retries = retry
		}
	}
	a.Add(1)
	defer a.Add(-1)
	resource, err, lock := internal.ExclusiveOneResource(resources, TasksDir, retries)
	if err != nil {
		// 无法获取资源，记录错误并继续执行
		logger.Warn("无法获取资源锁: %v，任务将继续执行但可能影响性能", err)
		task.Result.Stderr = append(task.Result.Stderr, fmt.Sprintf("排队超时 当前排队人数 %d\n", a.Load()))
		task.Completef(Failed, exclusive, "排队超时 当前排队人数 %d\n", a.Load())
		return func() {}
	}
	execute.resource = resource
	return func() {
		if lock != nil {
			err := lock.Unlock()
			if err != nil {
				logger.Error("解锁失败: %v", err)
			}
		}

		// 释放资源锁
		if resource != "" {
			if err := internal.ReleaseResource(resource, lock); err != nil {
				logger.Error("释放资源锁失败: %s, %v", resource, err)
			} else {
				logger.Info("成功释放资源锁: %s", resource)
			}
		}
	}
}

func (execute *executing) CatchStdout(stdout io.ReadCloser, stderr io.ReadCloser) {
	var encoding = func(s []byte) string {
		if execute.encoding == "" {
			return string(s)
		}
		return utf8.From(s, execute.encoding)
	}
	CatchStdout(stdout, encoding, execute.task.AddOutput)
	CatchStderr(stderr, encoding, func(s string) {
		task := execute.task
		task.AddOutput(s)
		task.Result.Stderr = append(task.Result.Stderr, s)
	})
}

func Cmd(cmd *exec.Cmd, workdir string, env []string) (error, io.ReadCloser, io.ReadCloser) {
	cmd.Env = env

	// 设置工作目录（任务沙箱）
	cmd.Dir = workdir

	// 创建管道获取命令输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %s\n", err.Error()), nil, nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %s", err.Error()), nil, nil
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动命令失败: %s", err.Error()), nil, nil
	}

	return nil, stdout, stderr
}

func (execute *executing) taskTimeout(taskMeta *CommandMeta) (context.Context, context.CancelFunc) {
	var timeout = 10 * time.Minute
	if os.Getenv("exporter_timeout") != "" {
		m, err := strconv.Atoi(os.Getenv("exporter_timeout"))
		if err == nil {
			timeout = time.Duration(m) * time.Minute
			execute.outputTask(fmt.Sprintf("通过环境变量设置超时: %d 分钟\n", m))
		}
	}
	if taskMeta != nil {
		if minutes := taskMeta.Timeout; minutes > 0 {
			timeout = time.Duration(minutes) * time.Minute
			execute.outputTask(fmt.Sprintf("通过meta.yml设置超时: %d 分钟\n", minutes))
		}
	}
	return context.WithTimeout(context.Background(), timeout)

}
