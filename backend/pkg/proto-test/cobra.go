package prototest

// Proto 测试 CLI 子命令定义。
// 子命令实现拆分到 cobra_case.go / cobra_replay.go / cobra_send.go，
// 当前均为占位实现（打印 debug 信息），待逐步实现。

import (
	_ "embed"

	"github.com/spf13/cobra"
)

//go:embed cobra-help.md
var helpText string

// NewProtoTestCmd 创建 proto-test 子命令。
// 该命令用于 proto 协议相关的测试功能。
func NewProtoTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proto-test",
		Short: "Proto 协议测试工具",
		Long:  helpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	cmd.AddCommand(
		newCaseCmd(),
		newReplayCmd(),
		newUnityLogCmd(),
		newSendMsgCmd(),
	)

	return cmd
}
