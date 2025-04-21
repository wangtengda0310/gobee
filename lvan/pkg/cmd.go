package pkg

import (
	"bufio"
	"fmt"
	"github.com/wangtengda/gobee/lvan/pkg/logger"
	"io"
	"os/exec"
	"sync/atomic"
)

var a atomic.Int32

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
