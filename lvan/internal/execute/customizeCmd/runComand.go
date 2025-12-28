package customizeCmd

import (
	"os"
	"os/exec"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/cmd/exporter/customCmd"
	"github.com/wangtengda0310/gobee/lvan/internal/execute"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
	"github.com/wangtengda0310/gobee/lvan/pkg/utf8"
)

func init() {

	customCmd.RegisterCommand("exec", runCommand)
	customCmd.RegisterCommand("run", runCommand)

}

var runCommand = func(args []string) {
	var cmd = args[0]
	cmdArgs := args[1:]

	logger.Info("执行命令: %s, %s", cmd, cmdArgs)

	flags := pflag.NewFlagSet("cmd", pflag.ExitOnError)
	encoding := flags.String("encoding", "", "被调用的版本号")
	err := flags.Parse(cmdArgs)
	if err != nil {
		logger.Warn("parse flags error %v %v", cmdArgs, err)
		return
	}

	var encodingFunc func([]byte) string
	if *encoding != "" {
		encodingFunc = func(s []byte) string {
			return utf8.From(s, utf8.Charset(*encoding))
		}
	}

	c := exec.Command(cmd, cmdArgs...)
	dir, err := os.Getwd()
	if err != nil {
		logger.Warn("获取当前工作目录失败: %v", err)
	}

	log := func(s string) {
		logger.Info("%s", s)
	}
	err, stdout, stderr := execute.Cmd(c, dir, []string{})
	if err != nil {
		logger.Warn("命令执行失败: %v", err)
	}

	execute.CatchStdout(stdout, encodingFunc, log)

	execute.CatchStderr(stderr, encodingFunc, log)

	if err = c.Wait(); err != nil {
		logger.Warn("等待命令完成失败: %v", err)
	} else {
		logger.Info("命令执行完成")
	}

}
