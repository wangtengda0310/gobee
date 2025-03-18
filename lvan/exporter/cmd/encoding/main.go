package main

import (
	"bufio"
	"fmt"
	"github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg"
	"os"
	"os/exec"
)

func main() {
	args := os.Args
	command := exec.Command(args[2], args[3:]...)

	command.Dir, _ = os.Getwd()

	// 创建管道获取命令输出
	stdout, err := command.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		panic(err)
	}

	encoding := args[1]
	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			toString := pkg.ByteToString(scanner.Bytes(), internal.Charset(encoding))
			fmt.Fprintln(os.Stdout, toString)
		}
	}()

	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			toString := pkg.ByteToString(scanner.Bytes(), internal.Charset(encoding))
			fmt.Fprintln(os.Stderr, toString)
		}
	}()

	err = command.Run()
	// 检查命令执行是否出错
	if err != nil {
		// 获取退出状态码
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else {
			os.Exit(command.ProcessState.ExitCode())
		}
	} else {
		os.Exit(command.ProcessState.ExitCode())
	}
}
