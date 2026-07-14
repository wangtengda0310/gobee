// rain-qa-func — QA 测试工具集
//
// 应用入口。通过 cobra 组织命令行接口：
//   - 无子命令时启动 Wails GUI 应用
//   - 指定子命令时执行对应的 CLI 功能（如 proto-test、mcp）
//
// Go module: git.devcloud.ztgame.com/v-tangfangda/rain-qa-func
package main

import (
	cmd "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/cmd/rain-qa-func"
)

// main 是应用程序入口。
// 委托给 cobra RootCmd 处理：无子命令启动 GUI，有子命令执行 CLI 逻辑。
func main() {
	// attach 父进程控制台（仅 Windows 生效，详见 attachconsole_windows.go）。
	// rain-qa-func.exe 用 -H windowsgui 构建，命令行带参数时 stdout/stderr 会丢失，
	// 在 cobra Execute 前 attach 父控制台并重建 os.Stdout/Stderr，使 --help、
	// proto-test 等 CLI 输出可见。双击启动（父=explorer.exe 无控制台）时自动跳过。
	attachParentConsole()
	defer releaseParentConsole() // LIFO: 后执行 —— FreeConsole 释放控制台 attachment
	defer nudgeConsoleRefresh()  // LIFO: 先执行 —— 写假回车刷新 PSReadLine 提示符（须在 FreeConsole 前）
	cmd.SetAssets(assets)
	cmd.Execute()
}
