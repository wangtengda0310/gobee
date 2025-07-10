package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wangtengda0310/gobee/lvan/glog/generator"
)

//go:generate go run main.go -xml testdata/demo.xml -mapping testdata/mapping.xml -out output

func main() {
	xmlPath := flag.String("xml", "testdata/demo.xml", "xml文件路径")
	mappingPath := flag.String("mapping", "testdata/mapping.xml", "类型映射xml路径")
	outDir := flag.String("out", "output", "代码输出目录")
	flag.Parse()

	if err := generator.Generate(*xmlPath, *mappingPath, *outDir); err != nil {
		fmt.Println("生成失败:", err)
		os.Exit(1)
	}
	fmt.Println("代码生成成功，开始自动测试...")

	absOut, _ := filepath.Abs(*outDir)
	cmd := exec.Command("go", "test", "-v", absOut)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("自动测试失败:", err)
		os.Exit(2)
	}
	fmt.Println("自动测试通过！")
}
