package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func main() {
	// Open the shared memory object
	hMap, err := syscall.CreateFileMapping(
		^syscall.Handle(0),
		nil,
		syscall.PAGE_READWRITE,
		0,
		1024,
		syscall.StringToUTF16Ptr("Local\\MySharedMemory"))
	if err != nil {
		fmt.Println("Failed to open file mapping:", err)
		return
	}
	defer syscall.CloseHandle(hMap)

	// Map the shared memory into the process address space
	p, err := syscall.MapViewOfFile(hMap, syscall.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		fmt.Println("Failed to map view of file:", err)
		return
	}
	defer syscall.UnmapViewOfFile(p)

	// Read data from the shared memory
	data := (*[1024]byte)(unsafe.Pointer(p))
	fmt.Println("Data read from shared memory:", string(data[:]))
}
