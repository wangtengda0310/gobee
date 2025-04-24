package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg"
)

func main() {
	from := pflag.String("from", "", "from charset")
	to := pflag.String("to", "", "to charset")
	pflag.Parse()

	args := pflag.Args()
	command := exec.Command(args[0], args[1:]...)

	var err error
	command.Dir, err = os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前工作目录失败: %v\n", err)
		os.Exit(1)
	}

	// 创建管道获取命令输出
	stdout, err := command.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		panic(err)
	}

	pipe := func(stdout io.ReadCloser) {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var toString string
			if *from != "" {
				toString = pkg.UtfFrom(scanner.Bytes(), internal.Charset(*from))
			} else {
				toString = scanner.Text()
			}
			if *to != "" {
				toString = pkg.UtfTo([]byte(toString), internal.Charset(*to))
			}
			fmt.Fprintln(os.Stdout, toString)
		}
	}

	go pipe(stdout)
	go pipe(stderr)

	err = command.Run()
	// 检查命令执行是否出错
	if err != nil {
		// 获取退出状态码
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else if command.ProcessState != nil {
			os.Exit(command.ProcessState.ExitCode())
		} else {
			os.Exit(1) // 如果无法获取状态码，则返回通用错误码
		}
	} else {
		os.Exit(command.ProcessState.ExitCode())
	}
}
