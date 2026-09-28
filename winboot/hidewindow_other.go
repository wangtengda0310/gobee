//go:build !windows

package main

import "syscall"

// 非 Windows: 无控制台窗口概念, 直接返回 nil (boot 工具实际仅面向 Windows).
func hideWindow() *syscall.SysProcAttr {
	return nil
}
