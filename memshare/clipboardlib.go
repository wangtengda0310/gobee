package main

import (
	"fmt"

	"github.com/atotto/clipboard"
)

func main() {
	// 写入剪贴板
	err := clipboard.WriteAll("Hello, Clipboard!")
	if err != nil {
		fmt.Println("Error writing to clipboard:", err)
		return
	}

	// 读取剪贴板
	/* content, err := clipboard.ReadAll()
	   if err != nil {
	       fmt.Println("Error reading from clipboard:", err)
	       return
	   }
	   fmt.Println("Clipboard content:", content) */
}
