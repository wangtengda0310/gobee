//go:build !windows

// rain-qa-func — QA 测试工具集
// attachconsole_other.go: 非 Windows 平台的空实现。
//
// Go module: git.devcloud.ztgame.com/v-tangfangda/rain-qa-func
package main

// attachParentConsole 在非 Windows 平台是空操作（详见 Windows 版本的注释）。
//
// Unix 系（Linux/macOS）使用 ELF/Mach-O，不存在 Windows PE 子系统概念，
// 进程默认继承父进程文件描述符，stdout 永远是有效句柄，不需要 attach 控制台。
func attachParentConsole() {}

// releaseParentConsole 在非 Windows 平台是空操作（与 attachParentConsole 配对）。
func releaseParentConsole() {}

// nudgeConsoleRefresh 在非 Windows 平台是空操作（PowerShell/PSReadLine 是 Windows 现象）。
func nudgeConsoleRefresh() {}
