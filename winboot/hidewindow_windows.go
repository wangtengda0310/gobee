//go:build windows

package main

import "syscall"

// Windows: 隐藏子进程控制台窗口, 避免 GUI 每次执行命令闪黑窗.
func hideWindow() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}
