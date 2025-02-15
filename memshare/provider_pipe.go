package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	PIPE_ACCESS_DUPLEX    = 0x00000003
	PIPE_TYPE_MESSAGE     = 0x00000004
	PIPE_READMODE_MESSAGE = 0x00000002
	PIPE_WAIT            = 0x00000000
)

func main() {
	pipeName := `\\.\pipe\SandboxPipe`

	// 创建命名管道
	hPipe, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateNamedPipeW").Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(pipeName))),
		uintptr(PIPE_ACCESS_DUPLEX),
		uintptr(PIPE_TYPE_MESSAGE|PIPE_READMODE_MESSAGE|PIPE_WAIT),
		uintptr(1), // 最大实例数
		4096, 4096,
		0,
		0,
	)
	if hPipe == 0 {
		fmt.Println("Failed to create named pipe")
		return
	}
	fmt.Println("Named pipe created successfully")

	// 等待客户端连接
	fmt.Println("Waiting for client connection...")
	syscall.NewLazyDLL("kernel32.dll").NewProc("ConnectNamedPipe").Call(hPipe, 0)
	fmt.Println("Client connected")

	// 写入数据
	message := "Data from Sandbox"
	messageBytes := []byte(message)
	var bytesWritten uint32
	syscall.NewLazyDLL("kernel32.dll").NewProc("WriteFile").Call(
		hPipe,
		uintptr(unsafe.Pointer(&messageBytes[0])),
		uintptr(len(messageBytes)),
		uintptr(unsafe.Pointer(&bytesWritten)),
		0,
	)
	if bytesWritten == 0 {
		fmt.Println("Failed to write data to pipe")
	} else {
		fmt.Printf("Data written to pipe: %s\n", message)
	}

	fmt.Scanln() // 保持进程运行
}
