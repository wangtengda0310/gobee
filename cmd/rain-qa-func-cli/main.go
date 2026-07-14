// rain-qa-func-cli — 纯 CLI 入口（无 GUI/Wails 依赖）。
//
// 用途：
//   - Windows PowerShell 下推荐使用本 CLI：主 exe（rain-qa-func.exe）是 GUI 子系统
//     （-H windowsgui），其命令行输出经 AttachConsole 直接写到 CONOUT$ 控制台屏幕缓冲区，
//     会绕过 PowerShell 的输出管道，导致 PSReadLine 重绘提示符时覆盖输出；本 CLI 是
//     console 子系统，PowerShell 用管道读其 stdout，PSReadLine 能正确跟踪，提示符渲染正常。
//   - 交叉编译到非 Windows 平台（macOS/Linux），提供与主 exe 一致的 cobra 子命令。
//
// 通过 CGO_ENABLED=0 编译，排除所有 wails 依赖。
// 子命令集与主入口（main.go）的 CLI 模式一致：proto-test、mcp。
//
// Go module: git.devcloud.ztgame.com/v-tangfangda/rain-qa-func
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	prototest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test"
	mcp "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/mcp"
)

func main() {
	root := &cobra.Command{
		Use:   "rain-qa-func",
		Short: "Rain QA 测试工具 (CLI)",
		Long:  "Rain QA 测试工具集 —— CLI 模式（无 GUI 支持）。",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	root.AddCommand(
		prototest.NewProtoTestCmd(),
		mcp.NewMCPCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
