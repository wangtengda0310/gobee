package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
	prototest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test"
	mcp "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings/mcp"
)

// AssetsFS 存储嵌入的前端静态资源。
// 由 main.go 通过 SetAssets() 设置（因为 go:embed 不支持 ../ 路径，必须在根目录嵌入）。
var AssetsFS embed.FS

// SetAssets 设置嵌入的前端静态资源。
// 必须在 Execute() 之前调用。
func SetAssets(fs embed.FS) {
	AssetsFS = fs
}

// RootCmd 是应用程序的根命令。
// 无子命令时执行 Wails GUI 启动逻辑，指定子命令时执行对应的 CLI 功能。
var RootCmd = &cobra.Command{
	Use:   "rain-qa-func",
	Short: "Rain QA 测试工具",
	Long:  "Rain QA 测试工具集 —— 支持 GUI 模式和 CLI 子命令模式。\n无子命令时启动 Wails GUI，指定子命令时执行对应功能。",
	Run: func(cmd *cobra.Command, args []string) {
		// 无子命令时启动 Wails GUI
		runWails()
	},
}

func init() {
	// 禁用 cobra Mousetrap 提示。本应用是 GUI(Wails)+CLI 双模式，双击应启动 GUI（runWails），
	// 而非被 cobra 的 Mousetrap（检测到从 explorer.exe 双击启动）阻塞并显示
	// "This is a command line tool. You need to open cmd.exe and run it from there."。
	cobra.MousetrapHelpText = ""

	// 注册一级子命令
	RootCmd.AddCommand(
		prototest.NewProtoTestCmd(),
		functiontest.NewFightTestCmd(),
		mcp.NewMCPCmd(),
	)
}

// Execute 执行根命令。
// 由 main.go 调用，是应用程序的唯一入口。
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
