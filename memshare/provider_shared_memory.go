package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

func main() {
	// Create a shared memory object
	hMap, err := syscall.CreateFileMapping(
		^syscall.Handle(0),
		nil,
		syscall.PAGE_READWRITE,
		0,
		1024,
		syscall.StringToUTF16Ptr("Local\\MySharedMemory"))
	if err != nil {
		fmt.Println("Failed to create file mapping:", err)
		return
	}
	defer syscall.CloseHandle(hMap)

	// Map the shared memory into the process address space
	p, err := syscall.MapViewOfFile(hMap, syscall.FILE_MAP_WRITE, 0, 0, 0)
	if err != nil {
		fmt.Println("Failed to map view of file:", err)
		return
	}
	defer syscall.UnmapViewOfFile(p)

	// Write data to the shared memory
	data := []byte("Hello from shared memory!")
	copy((*[1 << 30]byte)(unsafe.Pointer(p))[:len(data)], data)

	fmt.Println("Data written to shared memory.")

	// Keep the provider running
	for {
		time.Sleep(time.Second)
	}
}
