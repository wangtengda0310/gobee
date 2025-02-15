package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"unsafe"
)

func main() {
	pipeName := `\\.\pipe\SandboxPipe`

	// 连接管道
	hPipe, _, _ := windows.NewLazyDLL("kernel32.dll").NewProc("CreateFileW").Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(pipeName))),
		uintptr(windows.GENERIC_READ),
		0,
		0,
		uintptr(windows.OPEN_EXISTING),
		0,
		0,
	)

	// 读取数据
	buffer := make([]byte, 4096)
	var bytesRead uint32
	windows.NewLazyDLL("kernel32.dll").NewProc("ReadFile").Call(
		hPipe,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&bytesRead)),
		0,
	)

	fmt.Printf("Received: %s\n", string(buffer[:bytesRead]))
}
