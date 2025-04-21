package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/wangtengda/gobee/lvan/internal"
	"github.com/wangtengda/gobee/lvan/pkg"
	"io"
	"os"
	"os/exec"
)

func main() {
	from := pflag.String("from", "", "from charset")
	to := pflag.String("to", "", "to charset")
	pflag.Parse()

	args := pflag.Args()
	command := exec.Command(args[0], args[1:]...)

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
		} else {
			os.Exit(command.ProcessState.ExitCode())
		}
	} else {
		os.Exit(command.ProcessState.ExitCode())
	}
}
