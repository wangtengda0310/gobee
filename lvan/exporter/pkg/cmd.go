package pkg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var a atomic.Int32

// 执行命令
func ExecuteCommand(task *Task) {
	// 记录开始执行
	cmdName := task.Request.Cmd
	cmdVersion := task.Request.Version
	cmdArgs := task.Request.Args
	task.AddOutput(fmt.Sprintf("Starting command: %s\n", cmdName))
	task.AddOutput(fmt.Sprintf("Version: %s\n", cmdVersion))
	task.AddOutput(fmt.Sprintf("Arguments: %s\n", strings.Join(cmdArgs, ", ")))

	// 记录日志
	logger.Info("执行命令: %s, 版本: %s, 参数: %s", cmdName, cmdVersion, strings.Join(cmdArgs, ", "))

	// 使用版本管理获取可执行文件路径
	versionDir, err := GetCommandVersionPath(cmdName, cmdVersion)
	if err != nil {
		errMsg := fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		logger.Error(errMsg)
		task.AddOutput(errMsg)
		task.Complete(Failed, cmdNotExit)
		return
	}
	task.CmdMeta = internal.TryMeta(filepath.Join(versionDir, "meta.yaml"))

	// 查找可执行文件
	executable, err := FindExecutable(versionDir, cmdName)
	if err != nil {
		errMsg := fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		logger.Error(errMsg)
		task.AddOutput(errMsg)
		task.Complete(Failed, cmdNotExit)
		return
	}

	task.CmdPath = executable

	// 记录使用的可执行文件路径
	task.AddOutput(fmt.Sprintf("使用可执行文件: %s\n", executable))

	var timeout = 10 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	task.Cancel = cancel

	var cmd *exec.Cmd
	if task.CmdMeta != nil && task.CmdMeta.Shell != nil {
		newArgs := append(task.CmdMeta.Shell[1:], task.CmdPath)
		newArgs = append(newArgs, cmdArgs...)
		cmd = exec.CommandContext(ctx, task.CmdMeta.Shell[0], newArgs...)

	} else if runtime.GOOS == "windows" {
		// 检查是否是 Windows 平台 尝试使用cmd执行.bat和.cmd
		// 检查文件扩展名是否为批处理文件
		ext := strings.ToLower(filepath.Ext(task.CmdPath))
		if ext == ".bat" || ext == ".cmd" {
			// 使用 cmd /c 执行批处理文件
			newArgs := append([]string{"/c", task.CmdPath}, cmdArgs...)
			cmd = exec.CommandContext(ctx, "cmd", newArgs...)
		} else {
			// 非批处理文件直接执行
			cmd = exec.CommandContext(ctx, task.CmdPath, cmdArgs...)
		}
	} else {
		// 非 Windows 平台直接执行命令
		cmd = exec.CommandContext(ctx, task.CmdPath, cmdArgs...)
	}

	// 获取当前环境变量
	var env = os.Environ()
	if task.CmdMeta != nil && len(task.CmdMeta.Resources) > 0 {
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
		resource, err, lock := internal.ExclusiveOneResource(task.CmdMeta.Resources, TasksDir, retries)
		if err != nil {
			// 无法获取资源，记录错误并继续执行
			logger.Warn("无法获取资源锁: %v，任务将继续执行但可能影响性能", err)
			message := fmt.Sprintf("排队超时 当前排队人数 %d\n", a.Load())
			task.Result.Stderr = append(task.Result.Stderr, message)
			task.AddOutput(message)
			task.Complete(Failed, exclusive)
			return
		}
		defer func(resource string) {
			if lock != nil {
				lock.Unlock()
			}

			// 释放资源锁
			if resource != "" {
				if err := internal.ReleaseResource(resource, lock); err != nil {
					logger.Error("释放资源锁失败: %s, %v", resource, err)
				} else {
					logger.Info("成功释放资源锁: %s", resource)
				}
			}
		}(resource)

		// 添加自定义环境变量
		env = append(env, fmt.Sprintf("exporter_cmd_%s_resource=%s", cmdName, resource))
	}

	env = append(env, fmt.Sprintf("exporter_cmd_%s_id=%s", cmdName, task.ID))

	// 设置环境变量
	if len(task.Request.Env) > 0 {

		// 添加自定义环境变量
		for key, value := range task.Request.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
			logger.Debug("设置环境变量: %s=%s", key, value)
		}
	}

	var encodingf func([]byte) string
	var encoding internal.Charset
	if task.CmdMeta != nil && task.CmdMeta.Encoding != "" {
		encoding = task.CmdMeta.Encoding
		encodingf = func(s []byte) string {
			return UtfFrom(s, encoding)
		}
	}
	status, err, stdout, stderr := Cmd(cmd, task.WorkDir, env)
	task.Status = status
	if err != nil {
		return
	}

	CacthStdout(stdout, encodingf, task.AddOutput)

	log := func(s string) {
		task.AddOutput(s)
		task.Result.Stderr = append(task.Result.Stderr, s)
	}
	CacthStderr(stderr, encodingf, log)

	logger.Info("等待命令完成")
	err = cmd.Wait()
	if err != nil {
		var exitCode int
		// 处理超时错误
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			task.AddOutput(fmt.Sprintf("命令执行超时"))
			exitCode = 124 // 通常用 124 表示超时退出码
			task.AddOutput(fmt.Sprintf("命令执行超时，退出码 %d: %s", exitCode, err.Error()))
		} else if exitErr, ok := err.(*exec.ExitError); ok { // 尝试获取退出码
			exitCode = exitErr.ExitCode()
			task.AddOutput(fmt.Sprintf("命令执行失败，退出码 %d: %s", exitCode, err.Error()))
		} else {
			exitCode = 1 // 默认错误码
			task.AddOutput(fmt.Sprintf("命令执行未知错误，退出码 %d: %s", exitCode, err.Error()))
		}
		task.Complete(Failed, exitCode)
	} else {
		task.Complete(Completed, success)
	}

}

func Cmd(cmd *exec.Cmd, workdir string, env []string) (TaskStatus, error, io.ReadCloser, io.ReadCloser) {
	cmd.Env = env

	// 设置工作目录（任务沙箱）
	cmd.Dir = workdir

	// 创建管道获取命令输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Failed, fmt.Errorf("创建stdout管道失败: %s\n", err.Error()), nil, nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Failed, fmt.Errorf("创建stderr管道失败: %s", err.Error()), nil, nil
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return Failed, fmt.Errorf("启动命令失败: %s", err.Error()), nil, nil
	}

	return Running, nil, stdout, stderr
}

func CacthStderr(stderr io.ReadCloser, encodingFunc func([]byte) string, log func(string)) {
	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, cap(buf))

		for scanner.Scan() {
			var s string
			if encodingFunc != nil {
				s = encodingFunc(scanner.Bytes())
			} else {
				s = scanner.Text()
			}
			log(s)
		}

		if err := scanner.Err(); err != nil {
			logger.Error("标准错误扫描错误: %v", err)
			log(fmt.Sprintf("\n[SYSTEM ERROR] stderr扫描失败: %v\n", err))
		}
	}()
}

func CacthStdout(stdout io.ReadCloser, encodingFunc func([]byte) string, log func(string)) {
	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, cap(buf))

		for scanner.Scan() {
			var s string
			if encodingFunc != nil {
				s = encodingFunc(scanner.Bytes())
			} else {
				s = scanner.Text()
			}
			log(s)
		}

		// 处理扫描错误
		if err := scanner.Err(); err != nil {
			logger.Error("标准输出扫描错误: %v", err)
			log(fmt.Sprintf("\n[SYSTEM ERROR] stdout扫描失败: %v\n", err))
		}
	}()
}
