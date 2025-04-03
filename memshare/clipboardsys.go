package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
	MAX_RETRIES    = 3
)

func readClipboard() (string, error) {
	user32 := syscall.NewLazyDLL("user32.dll")
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	openClipboard := user32.NewProc("OpenClipboard")
	closeClipboard := user32.NewProc("CloseClipboard")
	getClipboardData := user32.NewProc("GetClipboardData")
	globalLock := kernel32.NewProc("GlobalLock")
	globalUnlock := kernel32.NewProc("GlobalUnlock")
	isFormatAvailable := user32.NewProc("IsClipboardFormatAvailable")
	getLastError := kernel32.NewProc("GetLastError")

	// 沙箱环境重试逻辑
	for i := 0; i < MAX_RETRIES; i++ {
		// 检查格式可用性
		available, _, _ := isFormatAvailable.Call(CF_UNICODETEXT)
		if available != 0 {
			break
		}
		if i == MAX_RETRIES-1 {
			return "", fmt.Errorf("沙箱剪贴板无文本数据（%d次重试）", MAX_RETRIES)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 打开剪贴板
	r, _, _ := openClipboard.Call(0)
	if r == 0 {
		errCode, _, _ := getLastError.Call()
		return "", fmt.Errorf("OpenClipboard失败 (错误码 0x%x)", errCode)
	}
	defer closeClipboard.Call()

	// 获取数据句柄
	h, _, _ := getClipboardData.Call(CF_UNICODETEXT)
	if h == 0 {
		errCode, _, _ := getLastError.Call()
		switch errCode {
		case 0:
			return "", fmt.Errorf("剪贴板无CF_UNICODETEXT数据")
		case 5:
			return "", fmt.Errorf("访问被拒绝，请提升权限")
		default:
			return "", fmt.Errorf("GetClipboardData错误 (0x%x)", errCode)
		}
	}

	// 锁定内存
	p, _, _ := globalLock.Call(h)
	if p == 0 {
		return "", fmt.Errorf("GlobalLock失败")
	}
	defer globalUnlock.Call(h)

	// 安全转换
	text := utf16PtrToString((*uint16)(unsafe.Pointer(p)))
	return text, nil
}

func utf16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}

	const maxLen = 1024 * 1024 // 安全限制
	ptr := uintptr(unsafe.Pointer(p))

	var length int
	for length = 0; length < maxLen; length++ {
		if *(*uint16)(unsafe.Pointer(ptr + uintptr(length)*2)) == 0 {
			break
		}
	}

	return syscall.UTF16ToString(unsafe.Slice(p, length))
}

func main() {
	text, err := readClipboard()
	if err != nil {
		fmt.Printf("读取失败: %v\n", err)

		// 沙箱特殊处理建议
		fmt.Println(`
		若在沙箱环境中，请检查：
		1. 剪贴板共享是否启用
		2. 尝试在沙箱内执行：
		   Set-Clipboard -Value "测试数据"
		3. 等待5秒后再重试`)
		return
	}

	fmt.Println("剪贴板内容:", text)
}
func monitorClipboard() {
	user32 := syscall.NewLazyDLL("user32.dll")
	addClipboardListener := user32.NewProc("AddClipboardFormatListener")
	hwnd := uintptr(0) // 控制台窗口句柄
	addClipboardListener.Call(hwnd)

	// 处理WM_CLIPBOARDUPDATE消息...
}
