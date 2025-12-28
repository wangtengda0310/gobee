package customizeCmd

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/cmd/exporter/customCmd"
	"github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/internal/execute"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

func init() {

	customCmd.RegisterCommand("cmd", cmdCommand)
	customCmd.RegisterCommand("command", cmdCommand)

}

var cmdCommand = func(args []string) {
	var cmd = args[0]
	cmdArgs := args[1:]
	logger.Info("执行命令: %s, %s", cmd, cmdArgs)

	flags := pflag.NewFlagSet("cmd", pflag.ExitOnError)
	version := flags.StringP("version", "v", "", "被调用的版本号")
	err := flags.Parse(cmdArgs)
	if err != nil {
		logger.Warn("%v", err)
		return
	}
	//
	//// 必要参数检查
	//if *message == "" {
	//	fmt.Println("必须提供提交信息（-m）")
	//	cmd.PrintDefaults()
	//	os.Exit(1)
	//}

	var req = internal.CommandRequest{
		Cmd:     cmd,
		Version: *version,
		Args:    cmdArgs,
	}
	var task = execute.CreateTask(req, os.Stdout)
	execute.ExecuteTask(task)
}
